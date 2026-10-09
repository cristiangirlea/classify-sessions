// Package pricing prices assistant records at dated public list rates.
//
// Cost of one parsed (deduplicated) assistant record, in USD:
//
//	(in·I + out·O + cache_read·I·0.1 + cache_write·I·1.25) / 1e6
//
// where I and O are the per-million input and output rates of the model's price key.
// Cache writes are assumed to be 5-minute writes. A request that cannot be priced in full
// contributes only its known part and marks its session incomplete: unknown cost is never 0.
package pricing

import (
	"regexp"
	"strings"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

const (
	SourceURL            = "https://docs.claude.com/en/docs/about-claude/pricing"
	AsOf                 = "2026-10-09"
	CacheReadMultiplier  = 0.1
	CacheWriteMultiplier = 1.25
	// LongContextMarker is the suffix Claude Code appends to a model id for the 1M-token
	// context window. Such requests are priced at base rates and flagged.
	LongContextMarker = "[1m]"
)

// Rate is a list price in USD per million tokens.
type Rate struct {
	Input  float64
	Output float64
}

// Rates is the dated price table, keyed by canonical model family.
var Rates = map[string]Rate{
	"claude-sonnet-4-5": {Input: 3, Output: 15},
	"claude-opus-4-1":   {Input: 15, Output: 75},
	"claude-haiku-4-5":  {Input: 1, Output: 5},
	"claude-sonnet-5":   {Input: 2, Output: 10},
}

// Assumptions are stated in the summary next to the prices.
var Assumptions = []string{
	"Cache writes are priced as 5-minute writes (1.25x input).",
	"Requests marked [1m] are priced at base rates because no request exceeds the long-context threshold; they are flagged.",
	"A request with no usage contributes no cost and marks its session incomplete.",
	"Sidechain (subagent) requests are billed like any other request and flagged.",
}

var dateSuffix = regexp.MustCompile(`-\d{8}$`)

// CanonicalModel strips the long-context marker and a -YYYYMMDD date suffix from a logged
// model id. longContext reports whether the marker was present.
func CanonicalModel(logged string) (family string, longContext bool) {
	family = strings.TrimSpace(logged)
	if strings.HasSuffix(family, LongContextMarker) {
		family, longContext = strings.TrimSuffix(family, LongContextMarker), true
	}
	return dateSuffix.ReplaceAllString(family, ""), longContext
}

// Lookup returns the price key and rate for a logged model id; ok is false when the
// canonical family is not in the table.
func Lookup(logged string) (key string, rate Rate, ok bool) {
	family, _ := CanonicalModel(logged)
	rate, ok = Rates[family]
	if !ok {
		return "", Rate{}, false
	}
	return family, rate, true
}

// Components split a cost into its four token classes, in USD.
type Components struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
}

// Total is the sum of the four components.
func (c Components) Total() float64 { return c.Input + c.Output + c.CacheRead + c.CacheWrite }

// Add accumulates o into c.
func (c *Components) Add(o Components) {
	c.Input += o.Input
	c.Output += o.Output
	c.CacheRead += o.CacheRead
	c.CacheWrite += o.CacheWrite
}

// Cost prices usage at rate.
func Cost(rate Rate, u model.Usage) Components {
	return Components{
		Input:      float64(u.Input) * rate.Input / 1e6,
		Output:     float64(u.Output) * rate.Output / 1e6,
		CacheRead:  float64(u.CacheRead) * rate.Input * CacheReadMultiplier / 1e6,
		CacheWrite: float64(u.CacheWrite) * rate.Input * CacheWriteMultiplier / 1e6,
	}
}

// Flag values are exactly the contract's ModelSpend flag enum.
type Flag string

const (
	FlagMissingUsage  Flag = "missing_usage"
	FlagUnknownModel  Flag = "unknown_model"
	FlagNoCacheFields Flag = "no_cache_fields"
	FlagLongContext   Flag = "long_context_marker"
	FlagSidechain     Flag = "sidechain"
)

// FlagOrder is the contract enum order, used to sort flags.
var FlagOrder = []Flag{FlagMissingUsage, FlagUnknownModel, FlagNoCacheFields, FlagLongContext, FlagSidechain}

// Incomplete reports whether the flag means the request's cost is not fully known.
// Sidechain requests are billed in full, so that flag alone keeps a request complete.
func (f Flag) Incomplete() bool { return f != FlagSidechain }

// Request is the priced view of one assistant event.
type Request struct {
	Model      string // as logged
	PriceKey   string // "" when unknown
	Known      Components
	Flags      []Flag // in FlagOrder
	Event      *model.Event
	LongCtx    bool
	HasUsage   bool
	KnownModel bool
}

// Complete reports whether every token of the request was priced.
func (r Request) Complete() bool {
	for _, f := range r.Flags {
		if f.Incomplete() {
			return false
		}
	}
	return true
}

// PriceEvent prices one assistant event. A request it cannot price contributes zero known
// components and carries a flag saying why.
func PriceEvent(e *model.Event) Request {
	r := Request{Model: e.Model, Event: e, HasUsage: e.Usage != nil}
	key, rate, ok := Lookup(e.Model)
	r.PriceKey, r.KnownModel = key, ok
	_, r.LongCtx = CanonicalModel(e.Model)
	set := map[Flag]bool{}
	if e.Usage == nil {
		set[FlagMissingUsage] = true
	} else if !e.Usage.HasCacheFields {
		set[FlagNoCacheFields] = true
	}
	if !ok {
		set[FlagUnknownModel] = true
	}
	if r.LongCtx {
		set[FlagLongContext] = true
	}
	if e.Sidechain {
		set[FlagSidechain] = true
	}
	if ok && e.Usage != nil {
		r.Known = Cost(rate, *e.Usage)
	}
	for _, f := range FlagOrder {
		if set[f] {
			r.Flags = append(r.Flags, f)
		}
	}
	return r
}

// SessionCost is the priced view of one session.
type SessionCost struct {
	Known    Components
	Requests []Request
	// Complete is false when any request is incomplete or any malformed or unknown record is
	// attributed to the session (its tokens, if any, are not known).
	Complete bool
}

// Session prices every assistant event of s once.
func Session(s *model.Session) SessionCost {
	sc := SessionCost{Complete: s.Counts.Malformed == 0 && s.Counts.Unknown == 0}
	for i := range s.Events {
		e := &s.Events[i]
		if !e.IsAssistant() {
			continue
		}
		r := PriceEvent(e)
		sc.Known.Add(r.Known)
		sc.Requests = append(sc.Requests, r)
		if !r.Complete() {
			sc.Complete = false
		}
	}
	return sc
}
