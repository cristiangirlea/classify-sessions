// Command sessions classifies Claude Code session logs and reports spend.
//
//	sessions summary    --input DIR [--labels FILE]       SummaryResponse JSON on stdout
//	sessions evaluate   --input DIR --labels FILE         classifier.heldout JSON on stdout
//	sessions serve      --input DIR [--labels FILE] [--listen ADDR] [--bind-any]
//	                                                      GET /api/summary on loopback
//	sessions prompts    --input DIR --exclude-labels FILE first prompts of unlabelled sessions, stdout only
//	sessions crosscheck --input DIR --otel FILE           text reconciliation with an OTel export
//	sessions version
//
// --input is mandatory; there is no home-directory default. serve listens on 127.0.0.1:8080
// by default and refuses a non-loopback --listen unless --bind-any is given; with --bind-any
// and no --listen it listens on 0.0.0.0:8080. The Host check stays loopback-only either way.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/crosscheck"
	"github.com/cristiangirlea/classify-sessions/backend/internal/evaluate"
	"github.com/cristiangirlea/classify-sessions/backend/internal/httpapi"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/summary"
)

const version = "0.2.0"

const (
	defaultListen    = "127.0.0.1:8080"
	defaultListenAny = "0.0.0.0:8080"
)

const usage = `usage:
  sessions summary    --input DIR [--labels FILE]
  sessions evaluate   --input DIR --labels FILE
  sessions serve      --input DIR [--labels FILE] [--listen 127.0.0.1:8080] [--bind-any]
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
	case "evaluate":
		err = cmdEvaluate(args[1:], stdout)
	case "serve":
		err = cmdServe(args[1:], stderr)
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

// flags is the parsed command line of one subcommand.
type flags struct {
	str   map[string]string
	bools map[string]bool
	set   map[string]bool // flags given explicitly
}

// parseFlags parses required string flags, optional string flags and bool flags. A required
// flag must be given and non-empty; an optional one, when given, must be non-empty.
func parseFlags(name string, args []string, required, optional, bools []string) (flags, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	strs := map[string]*string{}
	for _, r := range append(append([]string{}, required...), optional...) {
		strs[r] = fs.String(r, "", "")
	}
	bs := map[string]*bool{}
	for _, b := range bools {
		bs[b] = fs.Bool(b, false, "")
	}
	if err := fs.Parse(args); err != nil {
		return flags{}, usageError(err.Error())
	}
	if fs.NArg() > 0 {
		return flags{}, usageError("unexpected argument")
	}
	out := flags{str: map[string]string{}, bools: map[string]bool{}, set: map[string]bool{}}
	fs.Visit(func(f *flag.Flag) { out.set[f.Name] = true })
	for _, r := range required {
		if *strs[r] == "" {
			return flags{}, usageError("--" + r + " is required")
		}
	}
	for _, o := range optional {
		if out.set[o] && *strs[o] == "" {
			return flags{}, usageError("--" + o + " must not be empty")
		}
	}
	for k, v := range strs {
		out.str[k] = *v
	}
	for k, v := range bs {
		out.bools[k] = *v
	}
	return out, nil
}

// load ingests --input and, when a labels path is given, loads and validates the labels
// against the input's session ids. labels is nil when path is empty.
func load(input, labelsPath string) (*ingest.Result, []evaluate.Label, error) {
	res, err := ingest.Dir(input)
	if err != nil {
		return nil, nil, err
	}
	if labelsPath == "" {
		return res, nil, nil
	}
	labels, err := evaluate.Load(labelsPath, evaluate.SessionIDs(res.Sessions))
	if err != nil {
		return nil, nil, err
	}
	return res, labels, nil
}

func buildSummary(f flags) (contract.SummaryResponse, error) {
	res, labels, err := load(f.str["input"], f.str["labels"])
	if err != nil {
		return contract.SummaryResponse{}, err
	}
	return summary.Build(res, summary.Options{Input: f.str["input"], Labels: labels}), nil
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func cmdSummary(args []string, stdout io.Writer) error {
	f, err := parseFlags("summary", args, []string{"input"}, []string{"labels"}, nil)
	if err != nil {
		return err
	}
	s, err := buildSummary(f)
	if err != nil {
		return err
	}
	return writeJSON(stdout, s)
}

func cmdEvaluate(args []string, stdout io.Writer) error {
	f, err := parseFlags("evaluate", args, []string{"input", "labels"}, nil, nil)
	if err != nil {
		return err
	}
	res, labels, err := load(f.str["input"], f.str["labels"])
	if err != nil {
		return err
	}
	return writeJSON(stdout, evaluate.Compare(labels, evaluate.Predict(res.Sessions)))
}

// listenAddr resolves --listen and --bind-any to an address and validates it.
func listenAddr(f flags) (string, error) {
	addr := f.str["listen"]
	if !f.set["listen"] {
		addr = defaultListen
		if f.bools["bind-any"] {
			addr = defaultListenAny
		}
	}
	if err := httpapi.CheckListen(addr, f.bools["bind-any"]); err != nil {
		return "", usageError(err.Error())
	}
	return addr, nil
}

func cmdServe(args []string, stderr io.Writer) error {
	f, err := parseFlags("serve", args, []string{"input"}, []string{"labels", "listen"}, []string{"bind-any"})
	if err != nil {
		return err
	}
	addr, err := listenAddr(f)
	if err != nil {
		return err
	}
	s, err := buildSummary(f)
	if err != nil {
		return err
	}
	h, err := httpapi.New(s)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	fmt.Fprintf(stderr, "sessions serve: %s on %s (Ctrl+C to stop)\n", httpapi.SummaryPath, ln.Addr())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return httpapi.Serve(ctx, ln, h)
}

func cmdCrosscheck(args []string, stdout io.Writer) error {
	f, err := parseFlags("crosscheck", args, []string{"input", "otel"}, nil, nil)
	if err != nil {
		return err
	}
	res, err := ingest.Dir(f.str["input"])
	if err != nil {
		return err
	}
	oracle, err := crosscheck.ReadOTelFile(f.str["otel"])
	if err != nil {
		return err
	}
	crosscheck.Reconcile(res, oracle).Write(stdout)
	return nil
}

func cmdPrompts(args []string, stdout io.Writer) error {
	f, err := parseFlags("prompts", args, []string{"input", "exclude-labels"}, nil, nil)
	if err != nil {
		return err
	}
	exclude, err := readLabelIDs(f.str["exclude-labels"])
	if err != nil {
		return err
	}
	res, err := ingest.Dir(f.str["input"])
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
