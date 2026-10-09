package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var synthDir = filepath.Join("..", "..", "..", "fixtures", "synthetic")

// TestPromptsExcludesLabelledSessions uses a labels file written by the test, never the
// held-out file, and compares prompt text in code without printing it.
func TestPromptsExcludesLabelledSessions(t *testing.T) {
	labels := filepath.Join(t.TempDir(), "labels.json")
	doc := `{"description":"test","provenance":"test","selection_rule":"test","taxonomy":["unknown"],` +
		`"labels":[{"session_id":"5ea00000-0000-4000-8000-000000000001","category":"debug"},` +
		`{"session_id":"5ea00000-0000-4000-8000-000000000002","category":"review"},` +
		`{"session_id":"5ea00000-0000-4000-8000-000000000003","category":"implement"}]}`
	if err := os.WriteFile(labels, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"prompts", "--input", synthDir, "--exclude-labels", labels}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	want := map[string]string{
		"5ea00000-0000-4000-8000-000000000004": "Refactor the config loader into its own package",
		"5ea00000-0000-4000-8000-000000000005": "Write unit tests for the date parser",
		"5ea00000-0000-4000-8000-000000000006": "Explain how the cache layer works",
	}
	sc := bufio.NewScanner(&stdout)
	got := map[string]string{}
	for sc.Scan() {
		var pl promptLine
		if err := json.Unmarshal(sc.Bytes(), &pl); err != nil {
			t.Fatalf("output line is not JSON")
		}
		if pl.FirstPrompt == nil {
			t.Errorf("%s: no first prompt", pl.SessionID)
			continue
		}
		got[pl.SessionID] = *pl.FirstPrompt
	}
	if len(got) != len(want) {
		t.Errorf("printed %d sessions, want %d", len(got), len(want))
	}
	for id, p := range want {
		if got[id] != p {
			t.Errorf("%s: first prompt differs from the synthetic fixture", id)
		}
	}
}

func TestInputIsMandatory(t *testing.T) {
	for _, args := range [][]string{
		{"summary"},
		{"prompts", "--input", synthDir},
		{"crosscheck", "--input", synthDir},
		{"summary", "--input", synthDir, "extra"},
		{"summary", "--input", synthDir, "--labels", ""},
		{"evaluate", "--input", synthDir},
		{"evaluate", "--labels", "x.json"},
		{"serve"},
		{"serve", "--input", synthDir, "--listen", "0.0.0.0:8080"},
		{"serve", "--input", synthDir, "--listen", ":8080"},
		{"serve", "--input", synthDir, "--listen", "8080"},
		{"nonsense"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 || stdout.Len() != 0 {
			t.Errorf("%v: exit %d, stdout %d bytes; want 2 and nothing", args, code, stdout.Len())
		}
	}
}

func TestSummaryCommandWritesContractJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"summary", "--input", synthDir}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	var v map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	out := stdout.String()
	for _, banned := range []string{"Refactor the config loader", "Write unit tests", "Explain how the cache", "Fix the failing"} {
		if strings.Contains(out, banned) {
			t.Error("summary output carries prompt text")
		}
	}
	abs, err := filepath.Abs(synthDir)
	if err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if code := run([]string{"summary", "--input", abs}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.Contains(stdout.String(), filepath.ToSlash(filepath.Dir(abs))) || strings.Contains(stdout.String(), filepath.Dir(abs)) {
		t.Error("summary output carries the absolute input path")
	}
	if !strings.Contains(stdout.String(), `"input_hint": "synthetic"`) {
		t.Error("absolute input must reduce to its last path element")
	}
}

// testLabels writes a labels file over fixtures/synthetic: two agreements and one
// disagreement (session 4 is predicted refactor by refactor.refactor).
func testLabels(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "labels.json")
	doc := `{"labels":[{"session_id":"5ea00000-0000-4000-8000-000000000002","category":"review"},` +
		`{"session_id":"5ea00000-0000-4000-8000-000000000004","category":"plan"},` +
		`{"session_id":"5ea00000-0000-4000-8000-000000000005","category":"test"}]}`
	if err := os.WriteFile(p, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

const wantHeldout = `{"labelled":3,"agreements":2,"agreement_fraction":0.6666666666666666,"disagreements":[` +
	`{"session_id":"5ea00000-0000-4000-8000-000000000004","labelled":"plan","predicted":"refactor","rule":"refactor.refactor"}]}`

func compact(t *testing.T, b []byte) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestEvaluateCommandPrintsHeldout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"evaluate", "--input", synthDir, "--labels", testLabels(t)}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	if got := compact(t, stdout.Bytes()); got != wantHeldout {
		t.Errorf("got %s\nwant %s", got, wantHeldout)
	}

	bad := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(bad, []byte(`{"labels":[{"session_id":"5ea00000-0000-4000-8000-000000000002","category":"chores"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"evaluate", "--input", synthDir, "--labels", bad}, &stdout, &stderr); code != 1 || stdout.Len() != 0 {
		t.Errorf("invalid labels: exit %d, stdout %d bytes; want 1 and nothing", code, stdout.Len())
	}
	if strings.Contains(stderr.String(), filepath.Dir(bad)) {
		t.Error("error message carries the labels path")
	}
}

func TestSummaryLabelsFillHeldout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"summary", "--input", synthDir, "--labels", testLabels(t)}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr.String())
	}
	var s struct {
		Classifier struct {
			Method  string          `json:"method"`
			Heldout json.RawMessage `json:"heldout"`
		} `json:"classifier"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	if s.Classifier.Method != "rules-v1" || compact(t, s.Classifier.Heldout) != wantHeldout {
		t.Errorf("method %s heldout %s", s.Classifier.Method, s.Classifier.Heldout)
	}

	stdout.Reset()
	if code := run([]string{"summary", "--input", synthDir}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if err := json.Unmarshal(stdout.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	if got := compact(t, s.Classifier.Heldout); got != `{"labelled":0,"agreements":0,"agreement_fraction":null,"disagreements":[]}` {
		t.Errorf("without labels heldout = %s", got)
	}
}

func TestServeListenAddress(t *testing.T) {
	cases := []struct {
		args []string
		want string
		ok   bool
	}{
		{[]string{"--input", "x"}, "127.0.0.1:8080", true},
		{[]string{"--input", "x", "--listen", "localhost:9000"}, "localhost:9000", true},
		{[]string{"--input", "x", "--bind-any"}, "0.0.0.0:8080", true},
		{[]string{"--input", "x", "--bind-any", "--listen", ":9000"}, ":9000", true},
		{[]string{"--input", "x", "--listen", "0.0.0.0:9000"}, "", false},
	}
	for i, c := range cases {
		f, err := parseFlags("serve", c.args, []string{"input"}, []string{"labels", "listen"}, []string{"bind-any"})
		if err != nil {
			t.Fatal(err)
		}
		got, err := listenAddr(f)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("case %d: %q, %v; want %q ok=%v", i, got, err, c.want, c.ok)
		}
	}
}
