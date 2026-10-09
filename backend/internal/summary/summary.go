// Package summary builds the contract SummaryResponse from ingested sessions. PR1 has no
// classifier: every session is category "unknown" with rule null, and model_choice is a
// schema-valid placeholder that PR3 replaces.
package summary

import (
	"math"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/pricing"
)

// Taxonomy is the contract's Category enum, in schema order.
var Taxonomy = []string{"debug", "implement", "refactor", "test", "review", "research", "plan", "ops", "unknown"}

// ExampleRequest is shown on the page header.
const ExampleRequest = "curl -s http://127.0.0.1:8080/api/summary"

// Options are the inputs that do not come from the logs.
type Options struct {
	Input       string    // --input as given; reduced by InputHint
	GeneratedAt time.Time // zero means now
}

// InputHint returns the --input value with forward slashes, or only its last path element
// when it is absolute or rooted, so no absolute path reaches the output.
func InputHint(input string) string {
	s := strings.ReplaceAll(input, `\`, "/")
	drive := len(s) >= 2 && s[1] == ':'
	if !filepath.IsAbs(input) && !strings.HasPrefix(s, "/") && !drive {
		if s == "" {
			return "."
		}
		return s
	}
	if drive {
		s = s[2:]
	}
	s = strings.Trim(s, "/")
	if s == "" {
		return "."
	}
	return path.Base(s)
}

// usd rounds to 1e-9 USD: every list price here is a multiple of 1e-8 USD per token, so this
// removes float noise without changing any value.
func usd(v float64) float64 { return math.Round(v*1e9) / 1e9 }

func components(c pricing.Components) contract.Components {
	return contract.Components{
		InputUSD:      usd(c.Input),
		OutputUSD:     usd(c.Output),
		CacheReadUSD:  usd(c.CacheRead),
		CacheWriteUSD: usd(c.CacheWrite),
	}
}

func fraction(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func timePtr(t time.Time, ok bool) *string {
	if !ok {
		return nil
	}
	s := t.UTC().Format(time.RFC3339Nano)
	return &s
}

type modelAgg struct {
	key      string
	known    bool
	requests int
	cost     pricing.Components
	flags    map[pricing.Flag]bool
}

// Build maps the ingest result onto the contract.
func Build(res *ingest.Result, opt Options) contract.SummaryResponse {
	gen := opt.GeneratedAt
	if gen.IsZero() {
		gen = time.Now()
	}
	c := res.Counts
	out := contract.SummaryResponse{
		GeneratedAt: gen.UTC().Format(time.RFC3339),
		Source:      contract.Source{Files: len(res.Files), InputHint: InputHint(opt.Input)},
		Records: contract.Records{
			Total:            c.Records,
			Parsed:           c.Parsed,
			Malformed:        c.Malformed,
			Unknown:          c.Unknown,
			Duplicate:        c.Duplicate,
			UnparsedFraction: fraction(c.Malformed+c.Unknown, c.Records),
			Diagnostics: contract.Diagnostics{
				MissingUsage:       res.Diagnostics.MissingUsage,
				MissingTimestamp:   res.Diagnostics.MissingTimestamp,
				MissingCacheFields: res.Diagnostics.MissingCacheFields,
				Sidechain:          res.Diagnostics.Sidechain,
			},
		},
		Categories: []contract.CategorySummary{},
		Pricing: contract.Pricing{
			SourceURL:            pricing.SourceURL,
			AsOf:                 pricing.AsOf,
			CacheReadMultiplier:  pricing.CacheReadMultiplier,
			CacheWriteMultiplier: pricing.CacheWriteMultiplier,
			Assumptions:          append([]string{}, pricing.Assumptions...),
		},
		Classifier: contract.Classifier{
			Method:   "none",
			Taxonomy: append([]string{}, Taxonomy...),
			Heldout:  contract.Heldout{Disagreements: []contract.Disagreement{}},
		},
		SessionList:    []contract.SessionRow{},
		ExampleRequest: ExampleRequest,
		ModelChoice:    PlaceholderModelChoice(),
	}

	var total pricing.Components
	incomplete := 0
	byModel := map[string]*modelAgg{}
	unknownCat := contract.CategorySummary{Category: "unknown"}
	var unknownCost pricing.Components
	unclassifiedRecords := 0

	for _, s := range res.Sessions {
		sc := pricing.Session(s)
		total.Add(sc.Known)
		if !sc.Complete {
			incomplete++
		}
		models := map[string]bool{}
		toolCalls, toolErrors := 0, 0
		for i := range s.Events {
			e := &s.Events[i]
			toolCalls += len(e.ToolCalls)
			for _, r := range e.ToolResults {
				if r.Outcome == model.OutcomeError {
					toolErrors++
				}
			}
		}
		for _, r := range sc.Requests {
			models[r.Model] = true
			a := byModel[r.Model]
			if a == nil {
				a = &modelAgg{key: r.PriceKey, known: r.KnownModel, flags: map[pricing.Flag]bool{}}
				byModel[r.Model] = a
			}
			a.requests++
			a.cost.Add(r.Known)
			for _, f := range r.Flags {
				a.flags[f] = true
			}
		}
		modelList := make([]string, 0, len(models))
		for m := range models {
			modelList = append(modelList, m)
		}
		sort.Strings(modelList)

		// PR1: no classifier, every session is unknown.
		unknownCat.Sessions++
		unknownCat.Records += s.Counts.Records
		unknownCost.Add(sc.Known)
		if !sc.Complete {
			unknownCat.CostIncompleteSessions++
		}
		unclassifiedRecords += s.Counts.Records

		out.SessionList = append(out.SessionList, contract.SessionRow{
			ID:           s.ID,
			StartedAt:    timePtr(s.Start, s.HasStart),
			EndedAt:      timePtr(s.End, s.HasStart),
			Records:      s.Counts.Records,
			Category:     "unknown",
			Rule:         nil,
			CostUSD:      usd(sc.Known.Total()),
			CostComplete: sc.Complete,
			Requests:     len(sc.Requests),
			Models:       modelList,
			ToolCalls:    toolCalls,
			ToolErrors:   toolErrors,
		})
	}
	if unknownCat.Sessions > 0 {
		unknownCat.CostUSD = usd(unknownCost.Total())
		unknownCat.Components = components(unknownCost)
		out.Categories = append(out.Categories, unknownCat)
	}

	n := len(res.Sessions)
	out.Sessions = contract.Sessions{
		Total:                        n,
		Classified:                   0,
		Unclassified:                 n,
		UnclassifiedFractionSessions: fraction(n, n),
		UnclassifiedFractionRecords:  fraction(unclassifiedRecords, c.Records),
	}

	out.Spend = contract.Spend{
		TotalUSD:           usd(total.Total()),
		IncompleteSessions: incomplete,
		Components:         components(total),
		ByModel:            []contract.ModelSpend{},
	}
	names := make([]string, 0, len(byModel))
	for m := range byModel {
		names = append(names, m)
	}
	sort.Strings(names)
	for _, m := range names {
		a := byModel[m]
		ms := contract.ModelSpend{Model: m, Requests: a.requests, CostUSD: usd(a.cost.Total()), Flags: []string{}}
		if a.known {
			k := a.key
			ms.PriceKey = &k
		}
		for _, f := range pricing.FlagOrder {
			if a.flags[f] {
				ms.Flags = append(ms.Flags, string(f))
			}
		}
		out.Spend.ByModel = append(out.Spend.ByModel, ms)
	}
	return out
}

// PlaceholderModelChoice is the schema-valid model_choice emitted until PR3 computes it.
func PlaceholderModelChoice() contract.ModelsResponse {
	return contract.ModelsResponse{
		Models:           []string{},
		Categories:       []contract.ModelCategory{},
		ModelTotals:      []contract.ModelCell{},
		ExcludedSessions: []string{},
		Advice: contract.ModelAdvice{
			SharedCategories: []string{},
			Verdict:          "no_comparison_possible",
			Statement:        "Model-choice analysis is not computed in this build.",
			Caveats:          []string{},
		},
		Definitions: contract.ModelDefinitions{
			DominantModel:      "Canonical model family (date suffix and [1m] stripped) with the most assistant records after dedupe; ties broken by known USD, then name. The whole session is attributed to it.",
			ComparableCategory: "A non-unknown category where at least two models have a session.",
			SupportedCategory:  "A comparable category with at least three sessions per model; only supported categories inform the advice.",
			CostPerSession:     "Known USD at dated public rates; a lower bound when the session's cost is incomplete.",
			CorrectionMessage:  "A genuine user message, after the first, that matches the anchored opener or phrase rules, case-insensitive.",
			CorrectionRate:     "Correction messages divided by genuine user messages, pooled per cell.",
			ToolFailureRate:    "Failed results divided by known results, first result per tool-call id in chronological order.",
		},
		PricingSource: pricing.SourceURL,
		PricingDate:   pricing.AsOf,
	}
}
