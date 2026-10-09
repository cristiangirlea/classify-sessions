package crosscheck

import (
	"bytes"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/pricing"
)

func within(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestOTelCrossCheck(t *testing.T) {
	fixtures := filepath.Join("..", "..", "..", "fixtures")
	res, err := ingest.Dir(filepath.Join(fixtures, "claude-code"))
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := ReadOTelFile(filepath.Join(fixtures, "otel", "claude-code-events.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r := Reconcile(res, oracle)

	if r.OracleRequests != 465 || !within(r.OracleUSD, 41.119147, 1e-6) {
		t.Errorf("oracle %d requests USD %v, want 465 and 41.119147", r.OracleRequests, r.OracleUSD)
	}
	if !within(r.KnownUSD, 38.0136, 0.0001) || r.LogRequests != 467 {
		t.Errorf("known %d requests USD %v, want 467 and 38.0136", r.LogRequests, r.KnownUSD)
	}
	if r.Matched != 455 || r.TokenMismatch != 0 {
		t.Errorf("matched %d token mismatches %d, want 455 and 0", r.Matched, r.TokenMismatch)
	}

	deltas := map[string]float64{}
	for _, m := range r.ByModel {
		deltas[m.Model] = m.Delta()
	}
	// Our list prices agree with the oracle to its 6-decimal rounding for Opus and Sonnet 4.5.
	for _, m := range []string{"claude-opus-4-1-20250805", "claude-sonnet-4-5-20250929"} {
		if !within(deltas[m], 0, 0.0001) {
			t.Errorf("%s delta %v, want 0", m, deltas[m])
		}
	}
	if !within(deltas["claude-sonnet-5"], 0.4462, 0.0001) {
		t.Errorf("Sonnet 5 delta %v, want +0.4462", deltas["claude-sonnet-5"])
	}
	if !within(deltas["claude-sonnet-4-5-20250929[1m]"], 1.5357, 0.0001) {
		t.Errorf("[1m] delta %v, want +1.5357", deltas["claude-sonnet-4-5-20250929[1m]"])
	}

	// The deltas are explained by the generator's rates: Sonnet 5 at 3/15 and [1m] at 6/22.50,
	// cache multipliers unchanged. Repricing each matched record reproduces the oracle cost.
	for _, c := range []struct {
		model string
		rate  pricing.Rate
	}{
		{"claude-sonnet-5", pricing.Rate{Input: 3, Output: 15}},
		{"claude-sonnet-4-5-20250929[1m]", pricing.Rate{Input: 6, Output: 22.5}},
	} {
		events, oracleUSD := MatchedEvents(res, oracle, c.model)
		if len(events) == 0 {
			t.Errorf("%s: no matched events", c.model)
		}
		for i, e := range events {
			got := pricing.Cost(c.rate, *e.Usage).Total()
			if !within(got, oracleUSD[i], 6e-7) {
				t.Errorf("%s request %d: repriced %v, oracle %v", c.model, i, got, oracleUSD[i])
				break
			}
		}
	}

	if r.OracleOnly != 10 || r.OracleOnlyNoUs != 10 || !within(r.OracleOnlyUSD, 1.1426, 0.0001) {
		t.Errorf("oracle-only %d (no usage %d) USD %v, want 10, 10, 1.1426", r.OracleOnly, r.OracleOnlyNoUs, r.OracleOnlyUSD)
	}
	if r.LogOnly != 2 || r.LogOnlySide != 2 || r.LogOnlyModels["claude-haiku-4-5-20251001"] != 2 || !within(r.LogOnlyUSD, 0.0190, 0.0001) {
		t.Errorf("log-only %d (sidechain %d, models %v) USD %v, want 2 Haiku sidechains, 0.0190", r.LogOnly, r.LogOnlySide, r.LogOnlyModels, r.LogOnlyUSD)
	}
	recon := r.KnownUSD + (r.MatchedOracle - r.MatchedKnown) + r.OracleOnlyUSD - r.LogOnlyUSD
	if !within(recon, r.OracleUSD, 1e-6) {
		t.Errorf("reconciliation %v != oracle %v", recon, r.OracleUSD)
	}

	var buf bytes.Buffer
	r.Write(&buf)
	out := buf.String()
	for _, want := range []string{"oracle total USD 41.119147 over 465", "matched 455 requests", "residual +0.000000"} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q", want)
		}
	}
}
