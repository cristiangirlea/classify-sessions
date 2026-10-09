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
