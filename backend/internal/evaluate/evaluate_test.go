package evaluate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristiangirlea/classify-sessions/backend/internal/classify"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
)

var synthDir = filepath.Join("..", "..", "..", "fixtures", "synthetic")

const sid = "5ea00000-0000-4000-8000-00000000000"

func writeLabels(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "labels.json")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestEvaluateLoadsLabelsAndReportsDisagreements uses labels files written by the test over
// fixtures/synthetic, never the held-out file.
func TestEvaluateLoadsLabelsAndReportsDisagreements(t *testing.T) {
	res, err := ingest.Dir(synthDir)
	if err != nil {
		t.Fatal(err)
	}
	known := SessionIDs(res.Sessions)
	pred := Predict(res.Sessions)

	wantPred := map[string][2]string{
		sid + "1": {"debug", "debug.fix"},
		sid + "2": {"review", "review.review"},
		sid + "3": {"implement", "implement.add"},
		sid + "4": {"refactor", "refactor.refactor"},
		sid + "5": {"test", "test.write-tests"},
		sid + "6": {"research", "research.explain"},
	}
	for id, w := range wantPred {
		if p := pred[id]; p.Category != w[0] || p.Rule != w[1] {
			t.Errorf("%s: predicted %s/%s, want %s/%s", id, p.Category, p.Rule, w[0], w[1])
		}
	}

	t.Run("agreements and disagreements", func(t *testing.T) {
		path := writeLabels(t, `{"description":"test","taxonomy":["debug"],"labels":[`+
			`{"session_id":"`+sid+`6","category":"research"},`+
			`{"session_id":"`+sid+`4","category":"plan"},`+
			`{"session_id":"`+sid+`1","category":"debug"},`+
			`{"session_id":"`+sid+`2","category":"review"}]}`)
		labels, err := Load(path, known)
		if err != nil {
			t.Fatal(err)
		}
		h := Compare(labels, pred)
		if h.Labelled != 4 || h.Agreements != 3 || h.AgreementFraction == nil || *h.AgreementFraction != 0.75 {
			t.Fatalf("heldout %+v, want 4 labelled, 3 agreements, 0.75", h)
		}
		if len(h.Disagreements) != 1 {
			t.Fatalf("disagreements %+v", h.Disagreements)
		}
		d := h.Disagreements[0]
		if d.SessionID != sid+"4" || d.Labelled != "plan" || d.Predicted != "refactor" || d.Rule == nil || *d.Rule != "refactor.refactor" {
			t.Errorf("disagreement %+v", d)
		}
	})

	t.Run("no rule matched marshals rule null", func(t *testing.T) {
		labels := []Label{{SessionID: "b", Category: "ops"}, {SessionID: "a", Category: "plan"}}
		h := Compare(labels, map[string]classify.Result{"a": {Category: classify.Unknown}})
		raw, err := json.Marshal(h)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"labelled":2,"agreements":0,"agreement_fraction":0,"disagreements":[` +
			`{"session_id":"a","labelled":"plan","predicted":"unknown","rule":null},` +
			`{"session_id":"b","labelled":"ops","predicted":"unknown","rule":null}]}`
		if string(raw) != want {
			t.Errorf("got %s\nwant %s", raw, want)
		}
	})

	t.Run("empty labels array", func(t *testing.T) {
		labels, err := Load(writeLabels(t, `{"labels":[]}`), known)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(Compare(labels, pred))
		if string(raw) != `{"labelled":0,"agreements":0,"agreement_fraction":null,"disagreements":[]}` {
			t.Errorf("got %s", raw)
		}
		raw, _ = json.Marshal(Empty())
		if string(raw) != `{"labelled":0,"agreements":0,"agreement_fraction":null,"disagreements":[]}` {
			t.Errorf("Empty() = %s", raw)
		}
	})

	invalid := map[string]struct{ body, msg string }{
		"category not in taxonomy": {`{"labels":[{"session_id":"` + sid + `1","category":"chores"}]}`, "not in the taxonomy"},
		"duplicate session id":     {`{"labels":[{"session_id":"` + sid + `1","category":"debug"},{"session_id":"` + sid + `1","category":"test"}]}`, "repeats session"},
		"session not in input":     {`{"labels":[{"session_id":"00000000-0000-4000-8000-0000000000ff","category":"debug"}]}`, "not in the input"},
		"no labels array":          {`{"label":[]}`, "no labels array"},
		"labels not an array":      {`{"labels":{}}`, "expected shape"},
		"invalid JSON":             {`{"labels":[`, "expected shape"},
		"empty session id":         {`{"labels":[{"session_id":"","category":"debug"}]}`, "no session_id"},
		"missing category":         {`{"labels":[{"session_id":"` + sid + `1"}]}`, "no category"},
	}
	for name, c := range invalid {
		t.Run("invalid/"+name, func(t *testing.T) {
			_, err := Load(writeLabels(t, c.body), known)
			if err == nil || !strings.Contains(err.Error(), c.msg) {
				t.Errorf("error %v, want one containing %q", err, c.msg)
			}
		})
	}

	t.Run("invalid/missing file names no path", func(t *testing.T) {
		dir := t.TempDir()
		_, err := Load(filepath.Join(dir, "absent.json"), known)
		if err == nil || strings.Contains(err.Error(), dir) || strings.Contains(err.Error(), filepath.ToSlash(dir)) {
			t.Errorf("error %v: want an error without the path", err)
		}
	})
}
