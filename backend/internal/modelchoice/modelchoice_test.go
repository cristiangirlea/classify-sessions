package modelchoice

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

// Every prompt in this file is invented for the test.

func at(min int) time.Time {
	return time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC).Add(time.Duration(min) * time.Minute)
}

func user(min int, text string) model.Event {
	return model.Event{UUID: fmt.Sprintf("u%d", min), At: at(min), HasTime: true, Kind: model.KindUserPrompt, Text: text}
}

func asst(min int, m string, in, out int64) model.Event {
	return model.Event{UUID: fmt.Sprintf("a%d", min), At: at(min), HasTime: true, Kind: model.KindAssistantText, Model: m,
		Usage: &model.Usage{Input: in, Output: out, HasCacheFields: true}}
}

func session(id string, evs ...model.Event) *model.Session {
	s := &model.Session{ID: id, Events: evs}
	for _, e := range evs {
		if e.Kind == model.KindUserPrompt && e.Text != "" {
			s.FirstPrompt, s.HasFirstPrompt = e.Text, true
			break
		}
	}
	return s
}

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestCorrectionRateIsCorrectionsOverUserMessages(t *testing.T) {
	// The first message opens with "no" but is never a correction; two of the later three are.
	prompts := []string{
		"No rush: add a dark theme toggle to the settings page",
		"Nope, the toggle belongs in the header",
		"Looks good, also persist the choice",
		"That broke the mobile layout",
	}
	if c, n := CountCorrections(prompts); c != 2 || n != 4 {
		t.Fatalf("CountCorrections = %d of %d, want 2 of 4", c, n)
	}

	tr := model.Event{UUID: "tr", At: at(3), HasTime: true, Kind: model.KindToolResult,
		ToolResults: []model.ToolResult{{ToolUseID: "t1", Outcome: model.OutcomeSuccess}}}
	side := user(5, "Try again with the sidechain helper")
	side.Sidechain = true
	empty := user(7, "")
	s := session("s-rate",
		user(0, prompts[0]), asst(1, "claude-sonnet-4-5-20250929", 100, 100),
		user(2, prompts[1]), tr, side, empty,
		user(4, prompts[2]), user(6, prompts[3]), asst(8, "claude-sonnet-4-5-20250929", 100, 100))
	sig := ComputeSignals(s)
	if sig.Corrections != 2 || sig.UserMessages != 4 {
		t.Errorf("signals %d of %d, want 2 of 4 (sidechain, tool-result and empty records are not user messages)", sig.Corrections, sig.UserMessages)
	}
	r := CompareSignals([]Signals{sig})
	c := r.Categories[0].Cells[0]
	if c.CorrectionRate == nil || *c.CorrectionRate != 0.5 || c.CorrectedSessions != 1 || r.ModelTotals[0].CorrectionRate == nil || *r.ModelTotals[0].CorrectionRate != 0.5 {
		t.Errorf("cell rate %v corrected %d", c.CorrectionRate, c.CorrectedSessions)
	}
	if none := CompareSignals([]Signals{{SessionID: "x", Category: "debug", Model: "m"}}); none.ModelTotals[0].CorrectionRate != nil || none.ModelTotals[0].ToolFailureRate != nil {
		t.Error("rates must be null when there is nothing to divide by")
	}
}

func TestCorrectionRulesIgnoreNewTaskPromptsAndApprovals(t *testing.T) {
	corrections := []string{
		"No, that's the wrong function; use the parser in utils",
		"That's not what I asked for, keep the old signature",
		"Still failing with the same assertion after your change",
		"You hallucinated a flag that does not exist in this CLI",
		"Please revert the last edit to the router",
		"Actually, put the helper in its own file instead of inline",
	}
	others := []string{
		"Why does the invoice show the wrong total for refunds?",
		"Add a regression test for the refund rounding",
		"Great, the tests pass now",
		"Now also update the changelog",
		"Fix the date header next",
		"Looks good, ship it",
	}
	for i, p := range corrections {
		if !IsCorrection(p) {
			t.Errorf("correction case %d does not match", i)
		}
	}
	for i, p := range others {
		if IsCorrection(p) {
			t.Errorf("non-correction case %d matches", i)
		}
	}
	evs := []model.Event{user(0, "Add a CSV export to the orders page"), asst(1, "claude-opus-4-1-20250805", 10, 10)}
	for i, p := range append(append([]string{}, corrections...), others...) {
		evs = append(evs, user(2+i, p))
	}
	sig := ComputeSignals(session("s-rules", evs...))
	if sig.Corrections != len(corrections) || sig.UserMessages != 1+len(corrections)+len(others) {
		t.Errorf("signals %d of %d", sig.Corrections, sig.UserMessages)
	}
	raw, err := json.Marshal(sig)
	if err != nil {
		t.Fatal(err)
	}
	full, err := json.Marshal(CompareSignals([]Signals{sig}))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range evs {
		if e.Text == "" {
			continue
		}
		if strings.Contains(string(raw), e.Text) || strings.Contains(string(full), e.Text) {
			t.Error("marshalled signals carry prompt text")
		}
	}
}

func TestDominantModelAttributesMixedSessionToMajority(t *testing.T) {
	haiku := asst(4, "claude-haiku-4-5-20251001", 1000, 1000)
	haiku.Sidechain = true
	s := session("s-mixed", user(0, "Add paging to the audit log"),
		asst(1, "claude-sonnet-4-5-20250929", 10, 10),
		asst(2, "claude-sonnet-4-5-20250929[1m]", 10, 10),
		asst(3, "claude-sonnet-4-5", 10, 10),
		haiku)
	d := DominantModel(s)
	if d.Family != "claude-sonnet-4-5" || d.Records != 3 || d.Total != 4 || d.Share != 0.75 || !d.Mixed || !d.LongContext {
		t.Errorf("dominance %+v, want claude-sonnet-4-5 3 of 4 (0.75), mixed, long context", d)
	}

	tie := DominantModel(session("s-tie", asst(1, "claude-sonnet-4-5-20250929", 10, 10), asst(2, "claude-opus-4-1-20250805", 10, 10)))
	if tie.Family != "claude-opus-4-1" || !tie.Mixed || tie.LongContext {
		t.Errorf("record tie: %+v, want the family with more known USD", tie)
	}
	nameTie := DominantModel(session("s-name", asst(1, "claude-sonnet-5", 0, 0), asst(2, "claude-haiku-4-5", 0, 0)))
	if nameTie.Family != "claude-haiku-4-5" {
		t.Errorf("cost tie: %+v, want the first name", nameTie)
	}

	none := session("s-none", user(0, "Add paging to the audit log"))
	if d := DominantModel(none); d.Family != "" || d.Total != 0 {
		t.Errorf("no assistant record: %+v", d)
	}
	r := CompareModels([]*model.Session{s, none})
	if len(r.ExcludedSessions) != 1 || r.ExcludedSessions[0] != "s-none" {
		t.Errorf("excluded %v", r.ExcludedSessions)
	}
	if len(r.ModelTotals) != 1 || r.ModelTotals[0].Model != "claude-sonnet-4-5" || r.ModelTotals[0].MixedModelSessions != 1 || r.ModelTotals[0].LongContextSessions != 1 {
		t.Errorf("model totals %+v", r.ModelTotals)
	}
	if strings.Join(r.Models, ",") != "claude-sonnet-4-5" {
		t.Errorf("models %v", r.Models)
	}
}

func sigs(category, m string, n int, usd float64, corrections, messages int) []Signals {
	out := make([]Signals, n)
	for i := range out {
		out[i] = Signals{SessionID: fmt.Sprintf("%s-%s-%02d", category, m, i), Category: category, Model: m,
			KnownUSD: usd, CompleteCost: true, UserMessages: messages, Corrections: corrections}
	}
	return out
}

func join(parts ...[]Signals) []Signals {
	var out []Signals
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func category(r contract.ModelsResponse, c string) *contract.ModelCategory {
	for i := range r.Categories {
		if r.Categories[i].Category == c {
			return &r.Categories[i]
		}
	}
	return nil
}

func TestSingleModelCategoryReportedNotCompared(t *testing.T) {
	r := CompareSignals(join(
		sigs("plan", "claude-sonnet-4-5", 3, 0.3, 0, 2),
		sigs("debug", "claude-opus-4-1", 3, 2, 0, 2),
		sigs("debug", "claude-sonnet-4-5", 1, 0.5, 0, 2),
		sigs("unknown", "claude-opus-4-1", 1, 2, 0, 1),
		sigs("unknown", "claude-sonnet-4-5", 1, 0.5, 0, 1),
	))
	order := []string{}
	for _, c := range r.Categories {
		order = append(order, c.Category)
	}
	if strings.Join(order, ",") != "debug,plan,unknown" {
		t.Errorf("categories %v, want taxonomy order", order)
	}
	plan := category(r, "plan")
	if plan == nil || plan.Comparable || plan.Supported || plan.Reason != ReasonSingleModel || len(plan.Cells) != 1 || plan.Cells[0].Sessions != 3 {
		t.Errorf("plan %+v: want reported with its cell, not comparable", plan)
	}
	if u := category(r, "unknown"); u == nil || u.Comparable || u.Reason != ReasonUnknown || len(u.Cells) != 2 {
		t.Errorf("unknown %+v: want reported, never compared", u)
	}
	debug := category(r, "debug")
	if debug == nil || !debug.Comparable || debug.Supported || debug.Reason != ReasonThin || debug.Cells[0].Model != "claude-opus-4-1" {
		t.Errorf("debug %+v: want comparable, not supported, cells sorted by model", debug)
	}
	if strings.Join(r.Advice.SharedCategories, ",") != "debug" || r.Advice.Verdict != VerdictInsufficient {
		t.Errorf("advice %+v", r.Advice)
	}
}

func TestIncompleteCostCountedSeparately(t *testing.T) {
	missing := asst(2, "claude-sonnet-4-5-20250929", 0, 0)
	missing.Usage = nil
	partial := session("s-partial", user(0, "Add a print view to invoices"), asst(1, "claude-sonnet-4-5-20250929", 1000, 1000), missing)
	whole := session("s-whole", user(0, "Add a print view to receipts"), asst(1, "claude-sonnet-4-5-20250929", 1000, 3000))
	sp, sw := ComputeSignals(partial), ComputeSignals(whole)
	if sp.CompleteCost || !near(sp.KnownUSD, 0.018, 1e-12) {
		t.Errorf("partial: complete %v known %v, want false and the known part 0.018", sp.CompleteCost, sp.KnownUSD)
	}
	if !sw.CompleteCost || !near(sw.KnownUSD, 0.048, 1e-12) {
		t.Errorf("whole: complete %v known %v", sw.CompleteCost, sw.KnownUSD)
	}
	r := CompareSignals([]Signals{sp, sw})
	c := r.ModelTotals[0]
	if c.Sessions != 2 || c.CompleteCostSessions != 1 || !near(c.MeanKnownUSD, 0.033, 1e-9) || !near(c.MedianKnownUSD, 0.033, 1e-9) || !near(c.TotalKnownUSD, 0.066, 1e-9) {
		t.Errorf("cell %+v: want 2 sessions, 1 complete, mean and median 0.033 over all sessions", c)
	}
	if len(r.Advice.Caveats) == 0 || r.Advice.Caveats[0] != "1 session has incomplete cost; its known cost is a lower bound." {
		t.Errorf("caveats %q", r.Advice.Caveats)
	}
}

func TestAdviceDeterministicAndVerdictThresholds(t *testing.T) {
	build := func(opusN int, opusCorr int) []Signals {
		return join(
			sigs("debug", "claude-opus-4-1", opusN, 2, opusCorr, 4),
			sigs("debug", "claude-sonnet-4-5", 10, 0.5, 1, 4),
			sigs("plan", "claude-sonnet-4-5", 2, 0.5, 0, 4),
		)
	}
	marshal := func(s []Signals) string {
		b, err := json.Marshal(CompareSignals(s))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	in := build(10, 0)
	reversed := make([]Signals, len(in))
	for i := range in {
		reversed[len(in)-1-i] = in[i]
	}
	if a, b, c := marshal(in), marshal(in), marshal(reversed); a != b || a != c {
		t.Error("the same signals give different bytes")
	}

	nine := CompareSignals(build(9, 0)).Advice
	ten := CompareSignals(build(10, 0)).Advice
	if nine.Verdict != VerdictInsufficient || ten.Verdict != VerdictFewer {
		t.Errorf("verdicts %s (9 pooled) and %s (10 pooled), want %s then %s", nine.Verdict, ten.Verdict, VerdictInsufficient, VerdictFewer)
	}
	if ten.ExpensiveModel == nil || *ten.ExpensiveModel != "claude-opus-4-1" || ten.ComparatorModel == nil || *ten.ComparatorModel != "claude-sonnet-4-5" ||
		ten.CostRatio == nil || !near(*ten.CostRatio, 4, 1e-9) || strings.Join(ten.SharedCategories, ",") != "debug" {
		t.Errorf("advice %+v", ten)
	}
	if strings.Count(ten.Statement, ". ") != 3 || !strings.Contains(ten.Statement, "a ratio of 4.0x") {
		t.Errorf("statement is not the four-sentence template: %q", ten.Statement)
	}
	if v := CompareSignals(build(10, 2)).Advice.Verdict; v != VerdictNoBenefit {
		t.Errorf("higher premium correction rate: verdict %s, want %s", v, VerdictNoBenefit)
	}

	empty := CompareSignals(nil)
	if empty.Advice.Verdict != VerdictNone || empty.Advice.ExpensiveModel != nil || empty.Advice.CostRatio != nil ||
		empty.Models == nil || empty.Categories == nil || empty.ModelTotals == nil || empty.ExcludedSessions == nil || empty.Advice.SharedCategories == nil {
		t.Errorf("empty input: %+v", empty)
	}
	if raw, _ := json.Marshal(empty); !strings.Contains(string(raw), `"models":[]`) {
		t.Error("empty arrays must marshal as []")
	}
	disjoint := CompareSignals(join(
		sigs("debug", "claude-opus-4-1", 3, 2, 0, 2),
		sigs("debug", "claude-haiku-4-5", 1, 0.1, 0, 2),
		sigs("plan", "claude-sonnet-4-5", 5, 0.5, 0, 2),
	)).Advice
	if disjoint.Verdict != VerdictNone || len(disjoint.SharedCategories) != 0 || disjoint.ComparatorModel == nil || *disjoint.ComparatorModel != "claude-sonnet-4-5" {
		t.Errorf("no shared category: %+v", disjoint)
	}
}

func TestCompareModelsSuppliedCorpus(t *testing.T) {
	res, err := ingest.Dir(filepath.Join("..", "..", "..", "fixtures", "claude-code"))
	if err != nil {
		t.Fatal(err)
	}
	r := CompareModels(res.Sessions)
	want := map[string]struct {
		n    int
		mean float64
	}{
		"claude-opus-4-1":   {12, 2.386},
		"claude-sonnet-4-5": {18, 0.471},
		"claude-sonnet-5":   {4, 0.223},
	}
	if strings.Join(r.Models, ",") != "claude-opus-4-1,claude-sonnet-4-5,claude-sonnet-5" {
		t.Errorf("models %v", r.Models)
	}
	corrections, corrected, failures, known, mixed, long := 0, 0, 0, 0, 0, 0
	for _, c := range r.ModelTotals {
		w := want[c.Model]
		if c.Sessions != w.n || !near(c.MeanKnownUSD, w.mean, 0.0005) {
			t.Errorf("%s: n=%d mean=%v, want n=%d mean %v", c.Model, c.Sessions, c.MeanKnownUSD, w.n, w.mean)
		}
		corrections += c.CorrectionMessages
		corrected += c.CorrectedSessions
		failures += c.ToolFailures
		known += c.ToolResultsKnown
		mixed += c.MixedModelSessions
		long += c.LongContextSessions
	}
	if corrections != 27 || corrected != 8 {
		t.Errorf("corrections %d in %d sessions, want 27 in 8", corrections, corrected)
	}
	if failures != 59 || known != 414 {
		t.Errorf("tool failures %d of %d, want 59 of 414", failures, known)
	}
	if mixed != 2 || long != 4 || len(r.ExcludedSessions) != 0 {
		t.Errorf("mixed %d long-context %d excluded %d, want 2, 4, 0", mixed, long, len(r.ExcludedSessions))
	}
	comparable := []string{}
	for _, c := range r.Categories {
		if c.Comparable {
			comparable = append(comparable, c.Category)
		}
		if c.Supported {
			t.Errorf("%s is supported; no category has three sessions of two models", c.Category)
		}
	}
	if strings.Join(comparable, ",") != "debug,implement,refactor,test,review,research" {
		t.Errorf("comparable categories %v", comparable)
	}
	a := r.Advice
	if a.Verdict != VerdictInsufficient || a.ExpensiveModel == nil || *a.ExpensiveModel != "claude-opus-4-1" ||
		a.ComparatorModel == nil || *a.ComparatorModel != "claude-sonnet-4-5" || a.CostRatio == nil || !near(*a.CostRatio, 5.06, 0.01) {
		t.Errorf("advice %+v", a)
	}
}
