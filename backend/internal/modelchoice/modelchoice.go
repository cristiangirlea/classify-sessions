// Package modelchoice compares the models found in the logs per purpose category: cost per
// session, correction messages and tool failures. It is descriptive and makes no causal
// claim. Output carries canonical model families, categories, numbers and session ids only.
package modelchoice

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/cristiangirlea/classify-sessions/backend/internal/classify"
	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/pricing"
)

// Definitions are stated verbatim in every ModelsResponse (text as in contract/examples).
var Definitions = contract.ModelDefinitions{
	DominantModel:      "Canonical model family (date suffix and [1m] stripped) with the most assistant records after dedupe; ties broken by known USD, then name. The whole session is attributed to it.",
	ComparableCategory: "A non-unknown category where at least two models have a session.",
	SupportedCategory:  "A comparable category with at least three sessions per model; only supported categories inform the advice.",
	CostPerSession:     "Known USD at dated public rates; a lower bound when the session's cost is incomplete.",
	CorrectionMessage:  "A genuine user message, after the first, that matches the anchored opener or phrase rules, case-insensitive.",
	CorrectionRate:     "Correction messages divided by genuine user messages, pooled per cell.",
	ToolFailureRate:    "Failed results divided by known results, first result per tool-call id in chronological order.",
}

// Thresholds of the advice.
const (
	SupportedMinSessions = 3  // per model, for a supported category; also for the expensive model
	VerdictMinSessions   = 10 // pooled per model over the supported shared categories
)

// Verdicts, as in the contract enum.
const (
	VerdictInsufficient = "insufficient_evidence"
	VerdictNoBenefit    = "no_premium_benefit_observed"
	VerdictFewer        = "fewer_corrections_at_premium"
	VerdictNone         = "no_comparison_possible"
)

// Reasons shown per category.
const (
	ReasonUnknown     = "unknown is reported, never compared"
	ReasonSingleModel = "only one model has a session in this category"
	ReasonThin        = "fewer than three sessions for at least one model"
	ReasonSupported   = "every model has at least three sessions"
)

// AllCategory is the category of a model_totals cell.
const AllCategory = "all"

// Correction rules, applied to the lowercased, trimmed message.
var (
	correctionOpener = regexp.MustCompile(`^(no|nope|wrong|not quite|that's not|that is not|still (failing|fails|broken|wrong)|too much|undo|revert|you hallucinated|try again)\b`)
	correctionPhrase = regexp.MustCompile(`(revert|undo|hallucinated|does not exist|doesn't exist|not what i asked|that broke|still fails|still failing|try again|wrong file|doesn't work|does not work|not working|instead of|actually,)`)
)

// IsCorrection reports whether a message matches the opener or phrase rules. It does not
// know the message's position; CountCorrections skips the first message.
func IsCorrection(message string) bool {
	m := strings.ToLower(strings.TrimSpace(message))
	return correctionOpener.MatchString(m) || correctionPhrase.MatchString(m)
}

// CountCorrections counts correction messages among genuine user messages in chronological
// order. The first message opens the session and is never a correction.
func CountCorrections(prompts []string) (corrections, userMessages int) {
	for i, p := range prompts {
		if i > 0 && IsCorrection(p) {
			corrections++
		}
	}
	return corrections, len(prompts)
}

// Dominance is the result of DominantModel.
type Dominance struct {
	Family      string  // canonical family; "" when the session has no assistant record
	Records     int     // assistant records of Family
	Total       int     // assistant records with a model
	Share       float64 // Records / Total; 0 when Total is 0
	Mixed       bool    // more than one family
	LongContext bool    // any assistant record carries the [1m] marker
}

// DominantModel attributes a session to the canonical family with the most assistant records
// ([1m] folded into its family and counted), ties by known USD, then name.
func DominantModel(s *model.Session) Dominance {
	records := map[string]int{}
	cost := map[string]float64{}
	d := Dominance{}
	for i := range s.Events {
		e := &s.Events[i]
		if !e.IsAssistant() || strings.TrimSpace(e.Model) == "" {
			continue
		}
		fam, long := pricing.CanonicalModel(e.Model)
		if long {
			d.LongContext = true
		}
		records[fam]++
		cost[fam] += pricing.PriceEvent(e).Known.Total()
		d.Total++
	}
	for fam, n := range records {
		switch {
		case d.Family == "",
			n > d.Records,
			n == d.Records && cost[fam] > cost[d.Family],
			n == d.Records && cost[fam] == cost[d.Family] && fam < d.Family:
			d.Family, d.Records = fam, n
		}
	}
	d.Mixed = len(records) > 1
	if d.Total > 0 {
		d.Share = float64(d.Records) / float64(d.Total)
	}
	return d
}

// Signals are the per-session numbers the comparison pools. No text.
type Signals struct {
	SessionID        string  `json:"session_id"`
	Category         string  `json:"category"`
	Model            string  `json:"model"`    // dominant family; "" when excluded
	Excluded         bool    `json:"excluded"` // no assistant record
	Mixed            bool    `json:"mixed"`
	LongContext      bool    `json:"long_context"`
	KnownUSD         float64 `json:"known_usd"`
	CompleteCost     bool    `json:"complete_cost"`
	UserMessages     int     `json:"user_messages"`
	Corrections      int     `json:"corrections"`
	ToolResultsKnown int     `json:"tool_results_known"`
	ToolFailures     int     `json:"tool_failures"`
	UnknownOutcomes  int     `json:"unknown_outcome_results"`
}

// ComputeSignals measures one session. Its category is the frozen rules-v1 classification.
// Genuine user messages are non-sidechain user prompts with text, in timestamp order; tool
// outcomes use the first result per tool-call id in that order.
func ComputeSignals(s *model.Session) Signals {
	d := DominantModel(s)
	sc := pricing.Session(s)
	sig := Signals{
		SessionID:    s.ID,
		Category:     classify.Session(s).Category,
		Model:        d.Family,
		Excluded:     d.Family == "",
		Mixed:        d.Mixed,
		LongContext:  d.LongContext,
		KnownUSD:     sc.Known.Total(),
		CompleteCost: sc.Complete,
	}
	var prompts []string
	seen := map[string]bool{}
	for i := range s.Events {
		e := &s.Events[i]
		if e.Kind == model.KindUserPrompt && !e.Sidechain && e.Text != "" {
			prompts = append(prompts, e.Text)
		}
		for _, r := range e.ToolResults {
			if seen[r.ToolUseID] {
				continue
			}
			seen[r.ToolUseID] = true
			switch r.Outcome {
			case model.OutcomeError:
				sig.ToolResultsKnown++
				sig.ToolFailures++
			case model.OutcomeSuccess:
				sig.ToolResultsKnown++
			default:
				sig.UnknownOutcomes++
			}
		}
	}
	sig.Corrections, sig.UserMessages = CountCorrections(prompts)
	return sig
}

// CompareModels measures every session and compares the models.
func CompareModels(sessions []*model.Session) contract.ModelsResponse {
	sigs := make([]Signals, 0, len(sessions))
	for _, s := range sessions {
		sigs = append(sigs, ComputeSignals(s))
	}
	return CompareSignals(sigs)
}

func usd(v float64) float64 { return math.Round(v*1e9) / 1e9 }

func ratio(n, d int) *float64 {
	if d == 0 {
		return nil
	}
	r := float64(n) / float64(d)
	return &r
}

// cell pools signals of one model in one category (or "all").
func cell(modelName, category string, sigs []Signals) contract.ModelCell {
	c := contract.ModelCell{Model: modelName, Category: category, SessionIDs: []string{}}
	costs := make([]float64, 0, len(sigs))
	total := 0.0
	for _, s := range sigs {
		c.Sessions++
		if s.Mixed {
			c.MixedModelSessions++
		}
		if s.LongContext {
			c.LongContextSessions++
		}
		if s.CompleteCost {
			c.CompleteCostSessions++
		}
		costs = append(costs, s.KnownUSD)
		total += s.KnownUSD
		c.UserMessages += s.UserMessages
		c.CorrectionMessages += s.Corrections
		if s.Corrections > 0 {
			c.CorrectedSessions++
		}
		c.ToolResultsKnown += s.ToolResultsKnown
		c.ToolFailures += s.ToolFailures
		c.UnknownOutcomeResults += s.UnknownOutcomes
		c.SessionIDs = append(c.SessionIDs, s.SessionID)
	}
	sort.Strings(c.SessionIDs)
	if n := len(costs); n > 0 {
		sort.Float64s(costs)
		c.TotalKnownUSD = usd(total)
		c.MeanKnownUSD = usd(total / float64(n))
		if n%2 == 1 {
			c.MedianKnownUSD = usd(costs[n/2])
		} else {
			c.MedianKnownUSD = usd((costs[n/2-1] + costs[n/2]) / 2)
		}
	}
	c.CorrectionRate = ratio(c.CorrectionMessages, c.UserMessages)
	c.ToolFailureRate = ratio(c.ToolFailures, c.ToolResultsKnown)
	return c
}

// CompareSignals builds the ModelsResponse from per-session signals.
func CompareSignals(sigs []Signals) contract.ModelsResponse {
	out := contract.ModelsResponse{
		Models:           []string{},
		Categories:       []contract.ModelCategory{},
		ModelTotals:      []contract.ModelCell{},
		ExcludedSessions: []string{},
		Definitions:      Definitions,
		PricingSource:    pricing.SourceURL,
		PricingDate:      pricing.AsOf,
	}
	byModel := map[string][]Signals{}
	byCat := map[string]map[string][]Signals{}
	for _, s := range sigs {
		if s.Excluded {
			out.ExcludedSessions = append(out.ExcludedSessions, s.SessionID)
			continue
		}
		byModel[s.Model] = append(byModel[s.Model], s)
		if byCat[s.Category] == nil {
			byCat[s.Category] = map[string][]Signals{}
		}
		byCat[s.Category][s.Model] = append(byCat[s.Category][s.Model], s)
	}
	sort.Strings(out.ExcludedSessions)
	for m := range byModel {
		out.Models = append(out.Models, m)
	}
	sort.Strings(out.Models)
	for _, m := range out.Models {
		out.ModelTotals = append(out.ModelTotals, cell(m, AllCategory, byModel[m]))
	}
	for _, cat := range classify.Taxonomy {
		cm := byCat[cat]
		if cm == nil {
			continue
		}
		mc := contract.ModelCategory{Category: cat, Cells: []contract.ModelCell{}}
		models := make([]string, 0, len(cm))
		for m := range cm {
			models = append(models, m)
		}
		sort.Strings(models)
		thin := false
		for _, m := range models {
			c := cell(m, cat, cm[m])
			if c.Sessions < SupportedMinSessions {
				thin = true
			}
			mc.Cells = append(mc.Cells, c)
		}
		switch {
		case cat == classify.Unknown:
			mc.Reason = ReasonUnknown
		case len(models) < 2:
			mc.Reason = ReasonSingleModel
		case thin:
			mc.Comparable, mc.Reason = true, ReasonThin
		default:
			mc.Comparable, mc.Supported, mc.Reason = true, true, ReasonSupported
		}
		out.Categories = append(out.Categories, mc)
	}
	out.Advice = advise(out, sigs)
	return out
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func pct(r *float64) string {
	if r == nil {
		return "no rate"
	}
	return fmt.Sprintf("%.1f%%", *r*100)
}

func findCell(cells []contract.ModelCell, m string) *contract.ModelCell {
	for i := range cells {
		if cells[i].Model == m {
			return &cells[i]
		}
	}
	return nil
}

// advise names the expensive model (highest mean with three or more sessions) and the
// comparator (the other model with most sessions), lists their shared comparable categories,
// pools the supported ones and states a verdict. Deterministic: inputs are sorted and the
// statement is a fixed template.
func advise(r contract.ModelsResponse, sigs []Signals) contract.ModelAdvice {
	a := contract.ModelAdvice{SharedCategories: []string{}, Caveats: caveats(sigs)}

	anyComparable := false
	for _, c := range r.Categories {
		anyComparable = anyComparable || c.Comparable
	}
	if !anyComparable {
		a.Verdict = VerdictNone
		a.Statement = "No category has sessions from two models, so the logs allow no model comparison."
		return a
	}

	var exp *contract.ModelCell
	for i := range r.ModelTotals {
		c := &r.ModelTotals[i]
		if c.Sessions >= SupportedMinSessions && (exp == nil || c.MeanKnownUSD > exp.MeanKnownUSD) {
			exp = c
		}
	}
	if exp == nil {
		a.Verdict = VerdictNone
		a.Statement = "No model has three or more sessions, so no expensive model is named and the logs allow no model comparison."
		return a
	}
	var cmp *contract.ModelCell
	for i := range r.ModelTotals {
		c := &r.ModelTotals[i]
		if c.Model != exp.Model && (cmp == nil || c.Sessions > cmp.Sessions) {
			cmp = c
		}
	}
	e := exp.Model
	a.ExpensiveModel = &e
	if cmp == nil {
		a.Verdict = VerdictNone
		a.Statement = fmt.Sprintf("Only %s has sessions, so the logs allow no model comparison.", e)
		return a
	}
	cm := cmp.Model
	a.ComparatorModel = &cm
	var costSentence string
	if cmp.MeanKnownUSD > 0 {
		cr := exp.MeanKnownUSD / cmp.MeanKnownUSD
		a.CostRatio = &cr
		costSentence = fmt.Sprintf("%s has the highest mean known cost per session, USD %.3f over %d %s, against USD %.3f over %d %s for %s, a ratio of %.1fx.",
			e, exp.MeanKnownUSD, exp.Sessions, plural(exp.Sessions, "session", "sessions"),
			cmp.MeanKnownUSD, cmp.Sessions, plural(cmp.Sessions, "session", "sessions"), cm, cr)
	} else {
		costSentence = fmt.Sprintf("%s has the highest mean known cost per session, USD %.3f over %d %s; %s has no known cost, so no ratio is given.",
			e, exp.MeanKnownUSD, exp.Sessions, plural(exp.Sessions, "session", "sessions"), cm)
	}

	var pe, pc contract.ModelCell
	supported := []string{}
	for _, c := range r.Categories {
		if !c.Comparable {
			continue
		}
		ce, cc := findCell(c.Cells, e), findCell(c.Cells, cm)
		if ce == nil || cc == nil {
			continue
		}
		a.SharedCategories = append(a.SharedCategories, c.Category)
		if !c.Supported {
			continue
		}
		supported = append(supported, c.Category)
		pe.Sessions += ce.Sessions
		pe.UserMessages += ce.UserMessages
		pe.CorrectionMessages += ce.CorrectionMessages
		pc.Sessions += cc.Sessions
		pc.UserMessages += cc.UserMessages
		pc.CorrectionMessages += cc.CorrectionMessages
	}
	if len(a.SharedCategories) == 0 {
		a.Verdict = VerdictNone
		a.Statement = costSentence + " " + fmt.Sprintf("No category has sessions from both %s and %s, so the logs allow no model comparison.", e, cm)
		return a
	}
	sharedSentence := fmt.Sprintf("They share %d comparable %s (%s), of which %d %s at least three sessions per model.",
		len(a.SharedCategories), plural(len(a.SharedCategories), "category", "categories"), strings.Join(a.SharedCategories, ", "),
		len(supported), plural(len(supported), "has", "have"))
	re, rc := ratio(pe.CorrectionMessages, pe.UserMessages), ratio(pc.CorrectionMessages, pc.UserMessages)
	var poolSentence string
	if len(supported) == 0 {
		poolSentence = "No shared category is supported, so no correction rates are pooled."
	} else {
		poolSentence = fmt.Sprintf("Pooled over the supported shared categories, %s has %d corrections in %d user messages (%s) across %d sessions, and %s has %d in %d (%s) across %d sessions.",
			e, pe.CorrectionMessages, pe.UserMessages, pct(re), pe.Sessions,
			cm, pc.CorrectionMessages, pc.UserMessages, pct(rc), pc.Sessions)
	}
	var verdictSentence string
	switch {
	case pe.Sessions < VerdictMinSessions || pc.Sessions < VerdictMinSessions || re == nil || rc == nil:
		a.Verdict = VerdictInsufficient
		verdictSentence = fmt.Sprintf("With fewer than %d pooled sessions for at least one model, the logs cannot show whether the premium buys fewer corrections.", VerdictMinSessions)
	case *re < *rc:
		a.Verdict = VerdictFewer
		verdictSentence = fmt.Sprintf("%s shows a lower pooled correction rate than %s; this is descriptive, not causal.", e, cm)
	default:
		a.Verdict = VerdictNoBenefit
		verdictSentence = fmt.Sprintf("%s shows no lower pooled correction rate than %s, so no benefit from the premium is observed.", e, cm)
	}
	a.Statement = strings.Join([]string{costSentence, sharedSentence, poolSentence, verdictSentence}, " ")
	return a
}

// caveats are generated from the analysed (not excluded) sessions.
func caveats(sigs []Signals) []string {
	incomplete, long, mixed, excluded := 0, 0, 0, 0
	for _, s := range sigs {
		if s.Excluded {
			excluded++
			continue
		}
		if !s.CompleteCost {
			incomplete++
		}
		if s.LongContext {
			long++
		}
		if s.Mixed {
			mixed++
		}
	}
	out := []string{}
	if incomplete > 0 {
		out = append(out, fmt.Sprintf("%d %s incomplete cost; %s known cost is a lower bound.",
			incomplete, plural(incomplete, "session has", "sessions have"), plural(incomplete, "its", "their")))
	}
	if long > 0 {
		out = append(out, fmt.Sprintf("%d %s [1m] requests, priced at base rates and folded into their model family.",
			long, plural(long, "session includes", "sessions include")))
	}
	if mixed > 0 {
		out = append(out, fmt.Sprintf("%d %s more than one model; each is attributed whole to its dominant model.",
			mixed, plural(mixed, "session uses", "sessions use")))
	}
	if excluded > 0 {
		out = append(out, fmt.Sprintf("%d %s no assistant record and %s excluded.",
			excluded, plural(excluded, "session has", "sessions have"), plural(excluded, "is", "are")))
	}
	return append(out,
		"Correction messages are found by keyword rules and can miss or misread corrections.",
		"The comparison is descriptive and makes no causal claim.")
}
