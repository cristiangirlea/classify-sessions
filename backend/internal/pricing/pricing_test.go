package pricing

import (
	"math"
	"reflect"
	"testing"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-12 }

func TestPriceTableComponentsAndAliases(t *testing.T) {
	aliases := []struct {
		logged, family string
		long, known    bool
	}{
		{"claude-sonnet-4-5-20250929", "claude-sonnet-4-5", false, true},
		{"claude-sonnet-4-5-20250929[1m]", "claude-sonnet-4-5", true, true},
		{"claude-opus-4-1-20250805", "claude-opus-4-1", false, true},
		{"claude-haiku-4-5-20251001", "claude-haiku-4-5", false, true},
		{"claude-sonnet-5", "claude-sonnet-5", false, true},
		{"claude-imaginary-9", "claude-imaginary-9", false, false},
		{"", "", false, false},
	}
	for _, a := range aliases {
		fam, long := CanonicalModel(a.logged)
		if fam != a.family || long != a.long {
			t.Errorf("CanonicalModel(%q) = %q, %v; want %q, %v", a.logged, fam, long, a.family, a.long)
		}
		key, _, ok := Lookup(a.logged)
		if ok != a.known || (ok && key != a.family) {
			t.Errorf("Lookup(%q) = %q, %v", a.logged, key, ok)
		}
	}

	wantRates := map[string]Rate{
		"claude-sonnet-4-5": {3, 15}, "claude-opus-4-1": {15, 75}, "claude-haiku-4-5": {1, 5}, "claude-sonnet-5": {2, 10},
	}
	if !reflect.DeepEqual(Rates, wantRates) {
		t.Errorf("rate table = %v", Rates)
	}

	// Synthetic session 6 line 3: in 1000, out 2000, cache read 10000, cache write 4000.
	e := model.Event{Kind: model.KindAssistantText, Model: "claude-sonnet-4-5-20250929[1m]",
		Usage: &model.Usage{Input: 1000, Output: 2000, CacheRead: 10000, CacheWrite: 4000, HasCacheFields: true}}
	r := PriceEvent(&e)
	want := Components{Input: 0.003, Output: 0.03, CacheRead: 0.003, CacheWrite: 0.015}
	if !near(r.Known.Input, want.Input) || !near(r.Known.Output, want.Output) ||
		!near(r.Known.CacheRead, want.CacheRead) || !near(r.Known.CacheWrite, want.CacheWrite) {
		t.Errorf("components = %+v, want %+v", r.Known, want)
	}
	if !near(r.Known.Total(), 0.051) {
		t.Errorf("total = %v, want 0.051", r.Known.Total())
	}
	if !reflect.DeepEqual(r.Flags, []Flag{FlagLongContext}) || r.Complete() || r.PriceKey != "claude-sonnet-4-5" {
		t.Errorf("flags %v complete %v key %q", r.Flags, r.Complete(), r.PriceKey)
	}

	flagCases := []struct {
		name     string
		e        model.Event
		usd      float64
		flags    []Flag
		complete bool
	}{
		{"missing usage", model.Event{Kind: model.KindAssistantText, Model: "claude-sonnet-4-5-20250929"}, 0, []Flag{FlagMissingUsage}, false},
		{"no cache fields", model.Event{Kind: model.KindAssistantText, Model: "claude-opus-4-1-20250805",
			Usage: &model.Usage{Input: 100, Output: 200}}, 0.0165, []Flag{FlagNoCacheFields}, false},
		{"sidechain", model.Event{Kind: model.KindAssistantText, Model: "claude-haiku-4-5-20251001", Sidechain: true,
			Usage: &model.Usage{Input: 1000, Output: 1000, HasCacheFields: true}}, 0.006, []Flag{FlagSidechain}, true},
		{"unknown model", model.Event{Kind: model.KindAssistantText, Model: "claude-imaginary-9",
			Usage: &model.Usage{Input: 1000, Output: 1000, HasCacheFields: true}}, 0, []Flag{FlagUnknownModel}, false},
		{"sonnet 5", model.Event{Kind: model.KindToolCall, Model: "claude-sonnet-5",
			Usage: &model.Usage{Input: 1e6, Output: 1e6, CacheRead: 1e6, CacheWrite: 1e6, HasCacheFields: true}}, 2 + 10 + 0.2 + 2.5, nil, true},
	}
	for _, c := range flagCases {
		r := PriceEvent(&c.e)
		if !near(r.Known.Total(), c.usd) || !reflect.DeepEqual(r.Flags, c.flags) || r.Complete() != c.complete {
			t.Errorf("%s: usd %v flags %v complete %v; want %v %v %v", c.name, r.Known.Total(), r.Flags, r.Complete(), c.usd, c.flags, c.complete)
		}
		if r.PriceKey == "" && r.KnownModel {
			t.Errorf("%s: known model without a key", c.name)
		}
	}

	// A session with a malformed or unknown record is incomplete even if every request priced.
	s := &model.Session{Counts: model.Counts{Records: 2, Parsed: 1, Unknown: 1}, Events: []model.Event{flagCases[2].e}}
	if sc := Session(s); sc.Complete || !near(sc.Known.Total(), 0.006) {
		t.Errorf("session with an unknown record: complete %v usd %v", sc.Complete, sc.Known.Total())
	}
}
