// Package crosscheck reconciles the known USD computed from the session logs with the
// cost_usd of claude_code.api_request events in an OTel export. The OTel file is an oracle
// for pricing only; it is never ingested into the summary.
//
// An OTel request matches a parsed assistant record when session id, model and timestamp
// are equal. The report splits the oracle total into: matched requests (with a per-model
// price delta), oracle requests with no matching record, and records absent from the oracle.
package crosscheck

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/pricing"
)

// OTelRequest is one claude_code.api_request event.
type OTelRequest struct {
	SessionID  string
	Model      string
	At         time.Time
	Input      int64
	Output     int64
	CacheRead  int64
	CacheWrite int64
	CostUSD    float64
}

// ModelRow is the matched-request comparison for one logged model id.
type ModelRow struct {
	Model     string
	Matched   int
	KnownUSD  float64 // our price of the matched requests
	OracleUSD float64 // the oracle's cost_usd of the same requests
}

// Delta is OracleUSD - KnownUSD.
func (m ModelRow) Delta() float64 { return m.OracleUSD - m.KnownUSD }

// Result is the reconciliation.
type Result struct {
	OracleRequests int
	OracleUSD      float64
	LogRequests    int // parsed assistant records
	KnownUSD       float64

	Matched        int
	TokenMismatch  int // matched pairs whose four token counts differ
	ByModel        []ModelRow
	MatchedKnown   float64
	MatchedOracle  float64
	OracleOnly     int     // oracle requests with no matching record
	OracleOnlyUSD  float64 // their oracle cost
	OracleOnlyNoUs int     // of which a record with the same key exists but has no usage
	LogOnly        int     // records with no oracle request
	LogOnlyUSD     float64 // their known cost
	LogOnlyModels  map[string]int
	LogOnlySide    int // of which sidechain
}

type key struct {
	session, model string
	at             int64
}

// ReadOTel reads the api_request events of an OTel JSON-lines export.
func ReadOTel(r io.Reader) ([]OTelRequest, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64<<10), ingest.MaxLineBytes)
	var out []OTelRequest
	n := 0
	for sc.Scan() {
		n++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev struct {
			Timestamp  string `json:"timestamp"`
			Body       string `json:"body"`
			Attributes struct {
				SessionID  string  `json:"session.id"`
				Model      string  `json:"model"`
				Input      int64   `json:"input_tokens"`
				Output     int64   `json:"output_tokens"`
				CacheRead  int64   `json:"cache_read_tokens"`
				CacheWrite int64   `json:"cache_creation_tokens"`
				CostUSD    float64 `json:"cost_usd"`
			} `json:"attributes"`
		}
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, fmt.Errorf("otel line %d: invalid JSON", n)
		}
		if ev.Body != "claude_code.api_request" {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, ev.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("otel line %d: bad timestamp", n)
		}
		a := ev.Attributes
		out = append(out, OTelRequest{
			SessionID: a.SessionID, Model: a.Model, At: at.UTC(),
			Input: a.Input, Output: a.Output, CacheRead: a.CacheRead, CacheWrite: a.CacheWrite,
			CostUSD: a.CostUSD,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, errors.New("otel file: " + err.Error())
	}
	return out, nil
}

// ReadOTelFile opens path and calls ReadOTel. Errors never carry the path.
func ReadOTelFile(path string) ([]OTelRequest, error) {
	f, err := os.Open(path)
	if err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) {
			err = pe.Err
		}
		return nil, fmt.Errorf("open otel file: %w", err)
	}
	defer f.Close()
	return ReadOTel(f)
}

// Reconcile matches oracle requests to parsed assistant records.
func Reconcile(res *ingest.Result, oracle []OTelRequest) Result {
	var r Result
	logs := map[key][]pricing.Request{}
	for _, s := range res.Sessions {
		for i := range s.Events {
			e := &s.Events[i]
			if !e.IsAssistant() {
				continue
			}
			p := pricing.PriceEvent(e)
			r.LogRequests++
			r.KnownUSD += p.Known.Total()
			k := key{s.ID, e.Model, -1}
			if e.HasTime {
				k.at = e.At.UnixNano()
			}
			logs[k] = append(logs[k], p)
		}
	}
	rows := map[string]*ModelRow{}
	for _, o := range oracle {
		r.OracleRequests++
		r.OracleUSD += o.CostUSD
		k := key{o.SessionID, o.Model, o.At.UnixNano()}
		cands := logs[k]
		idx := -1
		for i, p := range cands {
			if p.HasUsage {
				idx = i
				break
			}
		}
		if idx < 0 {
			r.OracleOnly++
			r.OracleOnlyUSD += o.CostUSD
			if len(cands) > 0 {
				r.OracleOnlyNoUs++
				logs[k] = cands[1:]
				if len(logs[k]) == 0 {
					delete(logs, k)
				}
			}
			continue
		}
		p := cands[idx]
		logs[k] = append(cands[:idx:idx], cands[idx+1:]...)
		if len(logs[k]) == 0 {
			delete(logs, k)
		}
		r.Matched++
		u := p.Event.Usage
		if u.Input != o.Input || u.Output != o.Output || u.CacheRead != o.CacheRead || u.CacheWrite != o.CacheWrite {
			r.TokenMismatch++
		}
		row := rows[o.Model]
		if row == nil {
			row = &ModelRow{Model: o.Model}
			rows[o.Model] = row
		}
		row.Matched++
		row.KnownUSD += p.Known.Total()
		row.OracleUSD += o.CostUSD
		r.MatchedKnown += p.Known.Total()
		r.MatchedOracle += o.CostUSD
	}
	r.LogOnlyModels = map[string]int{}
	for _, ps := range logs {
		for _, p := range ps {
			r.LogOnly++
			r.LogOnlyUSD += p.Known.Total()
			r.LogOnlyModels[p.Model]++
			if p.Event.Sidechain {
				r.LogOnlySide++
			}
		}
	}
	for _, row := range rows {
		r.ByModel = append(r.ByModel, *row)
	}
	sort.Slice(r.ByModel, func(i, j int) bool { return r.ByModel[i].Model < r.ByModel[j].Model })
	return r
}

// MatchedEvents returns, for one logged model id, the (record, oracle) pairs that match, so
// a test can reprice the records at an alternative rate.
func MatchedEvents(res *ingest.Result, oracle []OTelRequest, modelID string) (events []*model.Event, oracleUSD []float64) {
	byKey := map[key]*model.Event{}
	for _, s := range res.Sessions {
		for i := range s.Events {
			e := &s.Events[i]
			if e.IsAssistant() && e.HasTime && e.Usage != nil && e.Model == modelID {
				byKey[key{s.ID, e.Model, e.At.UnixNano()}] = e
			}
		}
	}
	for _, o := range oracle {
		if o.Model != modelID {
			continue
		}
		if e, ok := byKey[key{o.SessionID, o.Model, o.At.UnixNano()}]; ok {
			events = append(events, e)
			oracleUSD = append(oracleUSD, o.CostUSD)
		}
	}
	return events, oracleUSD
}

// Write prints the text report. It contains model ids, counts and USD only.
func (r Result) Write(w io.Writer) {
	fmt.Fprintf(w, "oracle total USD %.6f over %d api_request events\n", r.OracleUSD, r.OracleRequests)
	fmt.Fprintf(w, "known total USD %.6f over %d assistant records\n", r.KnownUSD, r.LogRequests)
	fmt.Fprintf(w, "matched %d requests by session id, model and timestamp; token mismatches %d\n", r.Matched, r.TokenMismatch)
	for _, m := range r.ByModel {
		fmt.Fprintf(w, "  %-34s matched %4d  known %10.6f  oracle %10.6f  delta %+.4f\n", m.Model, m.Matched, m.KnownUSD, m.OracleUSD, m.Delta())
	}
	fmt.Fprintf(w, "oracle requests without a priced record: %d, USD %+.4f (%d match a record that has no usage)\n", r.OracleOnly, r.OracleOnlyUSD, r.OracleOnlyNoUs)
	models := make([]string, 0, len(r.LogOnlyModels))
	for m := range r.LogOnlyModels {
		models = append(models, m)
	}
	sort.Strings(models)
	fmt.Fprintf(w, "records absent from the oracle: %d, USD %+.4f (%d sidechain)", r.LogOnly, -r.LogOnlyUSD, r.LogOnlySide)
	for _, m := range models {
		fmt.Fprintf(w, " %s=%d", m, r.LogOnlyModels[m])
	}
	fmt.Fprintln(w)
	recon := r.KnownUSD + (r.MatchedOracle - r.MatchedKnown) + r.OracleOnlyUSD - r.LogOnlyUSD
	residual := r.OracleUSD - recon
	if math.Abs(residual) < 5e-7 {
		residual = 0 // float noise; avoids printing -0.000000
	}
	fmt.Fprintf(w, "reconciliation: known %.6f + matched deltas %+.4f + oracle-only %+.4f - log-only %.4f = %.6f (oracle %.6f, residual %+.6f)\n",
		r.KnownUSD, r.MatchedOracle-r.MatchedKnown, r.OracleOnlyUSD, r.LogOnlyUSD, recon, r.OracleUSD, residual)
}
