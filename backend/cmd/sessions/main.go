// Command sessions classifies Claude Code session logs and reports spend.
//
//	sessions summary    --input DIR                       SummaryResponse JSON on stdout
//	sessions prompts    --input DIR --exclude-labels FILE first prompts of unlabelled sessions, stdout only
//	sessions crosscheck --input DIR --otel FILE           text reconciliation with an OTel export
//	sessions version
//
// --input is mandatory; there is no home-directory default.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"

	"github.com/cristiangirlea/classify-sessions/backend/internal/crosscheck"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/summary"
)

const version = "0.1.0"

const usage = `usage:
  sessions summary    --input DIR
  sessions prompts    --input DIR --exclude-labels FILE
  sessions crosscheck --input DIR --otel FILE
  sessions version`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return 2
	}
	var err error
	switch args[0] {
	case "version":
		fmt.Fprintf(stdout, "sessions %s (%s)\n", version, runtime.Version())
		return 0
	case "summary":
		err = cmdSummary(args[1:], stdout)
	case "prompts":
		err = cmdPrompts(args[1:], stdout)
	case "crosscheck":
		err = cmdCrosscheck(args[1:], stdout)
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
	var ue usageError
	if errors.As(err, &ue) {
		fmt.Fprintf(stderr, "sessions %s: %s\n%s\n", args[0], ue, usage)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "sessions %s: %v\n", args[0], err)
		return 1
	}
	return 0
}

type usageError string

func (e usageError) Error() string { return string(e) }

func parseFlags(name string, args []string, required ...string) (map[string]string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	vals := map[string]*string{}
	for _, r := range required {
		vals[r] = fs.String(r, "", "")
	}
	if err := fs.Parse(args); err != nil {
		return nil, usageError(err.Error())
	}
	if fs.NArg() > 0 {
		return nil, usageError("unexpected argument")
	}
	out := map[string]string{}
	for _, r := range required {
		if *vals[r] == "" {
			return nil, usageError("--" + r + " is required")
		}
		out[r] = *vals[r]
	}
	return out, nil
}

func cmdSummary(args []string, stdout io.Writer) error {
	f, err := parseFlags("summary", args, "input")
	if err != nil {
		return err
	}
	res, err := ingest.Dir(f["input"])
	if err != nil {
		return err
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(summary.Build(res, summary.Options{Input: f["input"]}))
}

func cmdCrosscheck(args []string, stdout io.Writer) error {
	f, err := parseFlags("crosscheck", args, "input", "otel")
	if err != nil {
		return err
	}
	res, err := ingest.Dir(f["input"])
	if err != nil {
		return err
	}
	oracle, err := crosscheck.ReadOTelFile(f["otel"])
	if err != nil {
		return err
	}
	crosscheck.Reconcile(res, oracle).Write(stdout)
	return nil
}

func cmdPrompts(args []string, stdout io.Writer) error {
	f, err := parseFlags("prompts", args, "input", "exclude-labels")
	if err != nil {
		return err
	}
	exclude, err := readLabelIDs(f["exclude-labels"])
	if err != nil {
		return err
	}
	res, err := ingest.Dir(f["input"])
	if err != nil {
		return err
	}
	return writePrompts(stdout, res.Sessions, exclude)
}

// readLabelIDs reads only labels[].session_id from a labels file.
func readLabelIDs(path string) (map[string]bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		var pe *os.PathError
		if errors.As(err, &pe) {
			err = pe.Err
		}
		return nil, fmt.Errorf("read labels file: %w", err)
	}
	var doc struct {
		Labels []struct {
			SessionID string `json:"session_id"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, errors.New("labels file is not valid JSON of the expected shape")
	}
	if doc.Labels == nil {
		return nil, errors.New("labels file has no labels array")
	}
	ids := map[string]bool{}
	for i, l := range doc.Labels {
		if l.SessionID == "" {
			return nil, fmt.Errorf("labels[%d] has no session_id", i)
		}
		ids[l.SessionID] = true
	}
	return ids, nil
}

// promptLine is one line of the prompts output. Printed to stdout only, for rule tuning.
type promptLine struct {
	SessionID   string  `json:"session_id"`
	FirstPrompt *string `json:"first_prompt"`
}

func writePrompts(w io.Writer, sessions []*model.Session, exclude map[string]bool) error {
	ids := make([]*model.Session, 0, len(sessions))
	for _, s := range sessions {
		if !exclude[s.ID] {
			ids = append(ids, s)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].ID < ids[j].ID })
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, s := range ids {
		pl := promptLine{SessionID: s.ID}
		if s.HasFirstPrompt {
			p := s.FirstPrompt
			pl.FirstPrompt = &p
		}
		if err := enc.Encode(pl); err != nil {
			return err
		}
	}
	return nil
}
