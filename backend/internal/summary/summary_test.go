package summary

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/pricing"
)

var (
	repo      = filepath.Join("..", "..", "..")
	corpusDir = filepath.Join(repo, "fixtures", "claude-code")
	synthDir  = filepath.Join(repo, "fixtures", "synthetic")
)

func build(t *testing.T, dir, input string) contract.SummaryResponse {
	t.Helper()
	res, err := ingest.Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	return Build(res, Options{Input: input, GeneratedAt: time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)})
}

func within(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func sumComponents(c contract.Components) float64 {
	return c.InputUSD + c.OutputUSD + c.CacheReadUSD + c.CacheWriteUSD
}

func TestUsageBilledOncePerRecord(t *testing.T) {
	t.Run("corpus", func(t *testing.T) {
		s := build(t, corpusDir, "../fixtures/claude-code")
		if !within(s.Spend.TotalUSD, 38.0136, 0.0001) {
			t.Errorf("known USD = %v, want 38.0136 ± 0.0001", s.Spend.TotalUSD)
		}
		if !within(sumComponents(s.Spend.Components), s.Spend.TotalUSD, 1e-6) {
			t.Errorf("components %+v do not add up to %v", s.Spend.Components, s.Spend.TotalUSD)
		}
		// Per logged model: deduplicated assistant records and known USD. The two Sonnet 4.5
		// ids and Opus agree with the OTel oracle to rounding (see TestOTelCrossCheck).
		want := map[string]struct {
			requests int
			usd      float64
		}{
			"claude-haiku-4-5-20251001":      {2, 0.0189946},
			"claude-opus-4-1-20250805":       {175, 28.637403},
			"claude-sonnet-4-5-20250929":     {195, 6.50206785},
			"claude-sonnet-4-5-20250929[1m]": {58, 1.9626297},
			"claude-sonnet-5":                {37, 0.8924951},
		}
		if len(s.Spend.ByModel) != len(want) {
			t.Errorf("by_model has %d rows, want %d", len(s.Spend.ByModel), len(want))
		}
		sum, requests := 0.0, 0
		for _, m := range s.Spend.ByModel {
			w, ok := want[m.Model]
			if !ok {
				t.Errorf("unexpected model %s", m.Model)
				continue
			}
			if m.Requests != w.requests || !within(m.CostUSD, w.usd, 1e-6) {
				t.Errorf("%s: requests %d usd %v, want %d %v", m.Model, m.Requests, m.CostUSD, w.requests, w.usd)
			}
			sum += m.CostUSD
			requests += m.Requests
		}
		if !within(sum, s.Spend.TotalUSD, 1e-6) {
			t.Errorf("per-model sum %v != total %v", sum, s.Spend.TotalUSD)
		}
		if requests != 467 {
			t.Errorf("requests = %d, want 467 (476 assistant lines minus 9 duplicates)", requests)
		}
		// Per-model components. The contract's ModelSpend has no components, so they are
		// aggregated here from the priced requests. Values measured by this code; the OTel
		// cross-check confirms per-model totals, not the split into components.
		res, err := ingest.Dir(corpusDir)
		if err != nil {
			t.Fatal(err)
		}
		comps := map[string]*pricing.Components{}
		for _, sess := range res.Sessions {
			for _, r := range pricing.Session(sess).Requests {
				if comps[r.Model] == nil {
					comps[r.Model] = &pricing.Components{}
				}
				comps[r.Model].Add(r.Known)
			}
		}
		wantComps := map[string]pricing.Components{
			"claude-haiku-4-5-20251001":      {Input: 0.000427, Output: 0.001215, CacheRead: 0.0124376, CacheWrite: 0.004915},
			"claude-opus-4-1-20250805":       {Input: 1.606305, Output: 10.998375, CacheRead: 10.621998, CacheWrite: 5.410725},
			"claude-sonnet-4-5-20250929":     {Input: 0.386703, Output: 2.666835, CacheRead: 2.2970061, CacheWrite: 1.15152375},
			"claude-sonnet-4-5-20250929[1m]": {Input: 0.148008, Output: 0.8538, CacheRead: 0.6550467, CacheWrite: 0.305775},
			"claude-sonnet-5":                {Input: 0.014132, Output: 0.33202, CacheRead: 0.3745206, CacheWrite: 0.1718225},
		}
		if len(comps) != len(wantComps) {
			t.Errorf("models with priced requests = %d, want %d", len(comps), len(wantComps))
		}
		for _, m := range s.Spend.ByModel {
			got, w := comps[m.Model], wantComps[m.Model]
			if got == nil {
				t.Errorf("%s: no priced requests", m.Model)
				continue
			}
			if !within(got.Input, w.Input, 1e-9) || !within(got.Output, w.Output, 1e-9) ||
				!within(got.CacheRead, w.CacheRead, 1e-9) || !within(got.CacheWrite, w.CacheWrite, 1e-9) {
				t.Errorf("%s: components %+v, want %+v", m.Model, *got, w)
			}
			if !within(got.Total(), m.CostUSD, 1e-9) || !within(w.Total(), m.CostUSD, 1e-9) {
				t.Errorf("%s: components sum %v, model USD %v", m.Model, got.Total(), m.CostUSD)
			}
		}

		rowSum := 0.0
		for _, r := range s.SessionList {
			rowSum += r.CostUSD
		}
		if !within(rowSum, s.Spend.TotalUSD, 1e-6) {
			t.Errorf("session rows sum %v != total %v", rowSum, s.Spend.TotalUSD)
		}
	})

	t.Run("multi-block record and its copies", func(t *testing.T) {
		dir := t.TempDir()
		rec := `{"type":"assistant","uuid":"u1","sessionId":"s1","timestamp":"2026-08-10T09:00:00Z","requestId":"r1",` +
			`"message":{"model":"claude-sonnet-4-5-20250929","content":[{"type":"text","text":"a"},{"type":"tool_use","id":"t1","name":"Read","input":{}},{"type":"text","text":"b"}],` +
			`"usage":{"input_tokens":1000,"output_tokens":1000,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`
		if err := os.WriteFile(filepath.Join(dir, "s1.jsonl"), []byte(rec+"\n"+rec+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		s := build(t, dir, "x")
		if s.Records.Parsed != 1 || s.Records.Duplicate != 1 {
			t.Errorf("records %+v, want 1 parsed 1 duplicate", s.Records)
		}
		if !within(s.Spend.TotalUSD, 0.018, 1e-12) || s.SessionList[0].Requests != 1 || s.SessionList[0].ToolCalls != 1 {
			t.Errorf("usd %v requests %d tool calls %d, want 0.018 1 1", s.Spend.TotalUSD, s.SessionList[0].Requests, s.SessionList[0].ToolCalls)
		}
	})
}

func TestSyntheticSessionCosts(t *testing.T) {
	s := build(t, synthDir, "../fixtures/synthetic")
	want := map[string]struct {
		usd      float64
		complete bool
	}{
		"5ea00000-0000-4000-8000-000000000001": {0.00345, false},
		"5ea00000-0000-4000-8000-000000000002": {0.0018, false},
		"5ea00000-0000-4000-8000-000000000003": {0.0018, false},
		"5ea00000-0000-4000-8000-000000000004": {0.0036, true},
		"5ea00000-0000-4000-8000-000000000005": {0.0054, true},
		"5ea00000-0000-4000-8000-000000000006": {0.0735, false},
	}
	if len(s.SessionList) != len(want) {
		t.Fatalf("sessions = %d", len(s.SessionList))
	}
	for _, r := range s.SessionList {
		w := want[r.ID]
		if !within(r.CostUSD, w.usd, 1e-12) || r.CostComplete != w.complete {
			t.Errorf("%s: usd %v complete %v, want %v %v", r.ID, r.CostUSD, r.CostComplete, w.usd, w.complete)
		}
		if r.Category != "unknown" || r.Rule != nil {
			t.Errorf("%s: category %s rule %v, want unknown and null", r.ID, r.Category, r.Rule)
		}
	}
	if !within(s.Spend.TotalUSD, 0.08955, 1e-12) || s.Spend.IncompleteSessions != 4 {
		t.Errorf("total %v incomplete %d, want 0.08955 and 4", s.Spend.TotalUSD, s.Spend.IncompleteSessions)
	}
	flags := map[string]string{}
	for _, m := range s.Spend.ByModel {
		flags[m.Model] = strings.Join(m.Flags, ",")
		if (m.PriceKey == nil) != (m.Model == "claude-imaginary-9") {
			t.Errorf("%s: price_key %v", m.Model, m.PriceKey)
		}
	}
	wantFlags := map[string]string{
		"claude-sonnet-4-5-20250929[1m]": "long_context_marker",
		"claude-sonnet-4-5-20250929":     "missing_usage",
		"claude-opus-4-1-20250805":       "no_cache_fields",
		"claude-haiku-4-5-20251001":      "sidechain",
		"claude-imaginary-9":             "unknown_model",
	}
	for m, f := range wantFlags {
		if flags[m] != f {
			t.Errorf("%s flags %q, want %q", m, flags[m], f)
		}
	}
	if s.Records.Total != 34 || s.Sessions.Total != 6 || len(s.Categories) != 1 || s.Categories[0].Records != 34 {
		t.Errorf("records %d sessions %d categories %+v", s.Records.Total, s.Sessions.Total, s.Categories)
	}
}

func TestInputHint(t *testing.T) {
	cases := map[string]string{
		"../fixtures/claude-code":    "../fixtures/claude-code",
		`..\fixtures\claude-code`:    "../fixtures/claude-code",
		"fixtures/claude-code":       "fixtures/claude-code",
		"/srv/data/claude-code":      "claude-code",
		"/srv/data/claude-code/":     "claude-code",
		`C:\data\claude-code`:        "claude-code",
		"C:/data/claude-code":        "claude-code",
		`\\server\share\claude-code`: "claude-code",
		"/":                          ".",
		"":                           ".",
	}
	for in, want := range cases {
		if got := InputHint(in); got != want {
			t.Errorf("InputHint case %d chars = %q, want %q", len(in), got, want)
		}
	}
}

// emptyArrayPaths are key-set prefixes skipped by TestLiveSummaryMatchesContract because the
// array is empty on one side, so its item keys cannot be compared. PR2 (classifier) and PR3
// (model choice) remove entries as the arrays fill; the test fails while an entry is stale.
var emptyArrayPaths = []string{
	".classifier.heldout.disagreements[]", // PR2: empty until the held-out evaluation runs
	".model_choice.categories[]",          // PR3: placeholder
	".model_choice.model_totals[]",        // PR3: placeholder
}

func TestLiveSummaryMatchesContract(t *testing.T) {
	live := build(t, corpusDir, "../fixtures/claude-code")
	raw, err := json.Marshal(live)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var strict contract.SummaryResponse
	if err := dec.Decode(&strict); err != nil {
		t.Fatalf("strict decode of the live summary: %v", err)
	}
	example, err := os.ReadFile(filepath.Join(repo, "contract", "examples", "summary.json"))
	if err != nil {
		t.Fatal(err)
	}
	var liveAny, exAny any
	if err := json.Unmarshal(raw, &liveAny); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(example, &exAny); err != nil {
		t.Fatal(err)
	}
	liveKeys, exKeys := contract.KeySet(liveAny), contract.KeySet(exAny)
	skipped := func(k string) bool {
		for _, p := range emptyArrayPaths {
			if strings.HasPrefix(k, p+".") {
				return true
			}
		}
		return false
	}
	has := func(keys []string, p string) bool {
		for _, k := range keys {
			if strings.HasPrefix(k, p+".") {
				return true
			}
		}
		return false
	}
	for _, p := range emptyArrayPaths {
		if has(liveKeys, p) && has(exKeys, p) {
			t.Errorf("%s is populated on both sides: remove it from emptyArrayPaths", p)
		}
	}
	filter := func(keys []string) map[string]bool {
		m := map[string]bool{}
		for _, k := range keys {
			if !skipped(k) {
				m[k] = true
			}
		}
		return m
	}
	lm, em := filter(liveKeys), filter(exKeys)
	for k := range em {
		if !lm[k] {
			t.Errorf("key %s is in the example but not in the live summary", k)
		}
	}
	for k := range lm {
		if !em[k] {
			t.Errorf("key %s is in the live summary but not in the example", k)
		}
	}

	// Values the contract constrains beyond keys.
	for _, m := range live.Spend.ByModel {
		for _, f := range m.Flags {
			switch f {
			case "missing_usage", "unknown_model", "no_cache_fields", "long_context_marker", "sidechain":
			default:
				t.Errorf("flag %q is not in the contract enum", f)
			}
		}
	}
	if live.Records.Total != live.Records.Parsed+live.Records.Malformed+live.Records.Unknown+live.Records.Duplicate {
		t.Error("records do not add up")
	}
	if live.Source.InputHint != "../fixtures/claude-code" {
		t.Errorf("input_hint = %q", live.Source.InputHint)
	}
	if live.ModelChoice.PricingSource != live.Pricing.SourceURL || live.ModelChoice.PricingDate != live.Pricing.AsOf {
		t.Error("model_choice pricing source or date differs from pricing")
	}
	var exStrict contract.SummaryResponse
	if err := json.Unmarshal(example, &exStrict); err != nil {
		t.Fatal(err)
	}
	if live.ModelChoice.Definitions != exStrict.ModelChoice.Definitions {
		t.Error("model_choice definitions differ from the contract example")
	}
	if strings.Join(live.Pricing.Assumptions[:3], "|") != strings.Join(exStrict.Pricing.Assumptions, "|") {
		t.Error("the first three pricing assumptions differ from the contract example")
	}
}
