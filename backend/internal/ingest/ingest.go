// Package ingest reads Claude Code session logs (*.jsonl) under one directory and accounts
// for every nonblank line exactly once: parsed, malformed, unknown or duplicate.
//
//   - malformed: invalid JSON, a line over the 8 MiB cap, a bad field shape, negative or
//     non-integer token counts, an unparsable timestamp, or a uuid that conflicts with an
//     earlier record of the same session.
//   - unknown: a record whose type is not user, assistant or summary, or that carries a
//     content block the parser does not know; the whole record is set aside, never billed.
//   - duplicate: same (sessionId, uuid) as an earlier parsed record with identical
//     canonical JSON (keys sorted), so a byte copy and a key-reordered copy both count.
//
// Records without a usable sessionId (summary records, invalid JSON) are attributed to the
// file's session: the first sessionId seen in the file, else the file name without .jsonl.
package ingest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

// MaxLineBytes is the longest line accepted; a longer line is counted as malformed.
const MaxLineBytes = 8 << 20

// Status of one nonblank line.
type Status string

const (
	StatusParsed    Status = "parsed"
	StatusMalformed Status = "malformed"
	StatusUnknown   Status = "unknown"
	StatusDuplicate Status = "duplicate"
)

// Outcome records what happened to one nonblank line. Reason never contains record text.
type Outcome struct {
	File      string // relative to the input directory, forward slashes
	Line      int
	Status    Status
	Reason    string
	SessionID string // the session the line is attributed to
}

// Diagnostics count conditions on parsed records (duplicates excluded).
type Diagnostics struct {
	MissingUsage       int // assistant records with no usage object
	MissingTimestamp   int // parsed records with no timestamp (summary records included)
	MissingCacheFields int // assistant usage without both cache token fields
	Sidechain          int // parsed records with isSidechain true
}

// Result is the whole directory.
type Result struct {
	Files       []string // relative, forward slashes, sorted
	Counts      model.Counts
	Diagnostics Diagnostics
	Sessions    []*model.Session // sorted by id
	Outcomes    []Outcome        // one per nonblank line, in walk order
}

// Session returns the session with the given id, or nil.
func (r *Result) Session(id string) *model.Session {
	for _, s := range r.Sessions {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// Dir ingests every *.jsonl file under root, in sorted path order.
func Dir(root string) (*Result, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("input directory: %w", pathless(err))
	}
	if !info.IsDir() {
		return nil, errors.New("input is not a directory")
	}
	var files []string
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() && strings.HasSuffix(d.Name(), ".jsonl") {
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk input: %w", pathless(err))
	}
	sort.Strings(files)

	in := newIngester()
	for _, rel := range files {
		f, err := os.Open(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path.Base(rel), pathless(err))
		}
		err = in.file(rel, f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path.Base(rel), err)
		}
	}
	return in.finish(files), nil
}

type dedupeKey struct{ session, uuid string }

type ingester struct {
	sessions map[string]*model.Session
	seen     map[dedupeKey]string // canonical JSON of the first parsed record with this key
	res      Result
}

func newIngester() *ingester {
	return &ingester{sessions: map[string]*model.Session{}, seen: map[dedupeKey]string{}}
}

// lineResult is the file-local verdict on one line before session attribution and dedupe.
type lineResult struct {
	line      int
	status    Status // parsed, malformed or unknown; parsed may still become duplicate or malformed
	reason    string
	sessionID string // explicit sessionId on the record, "" if absent
	canonical string
	event     *model.Event
}

func (in *ingester) file(rel string, r io.Reader) error {
	var lines []lineResult
	err := readLines(r, func(n int, raw []byte, tooLong bool) {
		if tooLong {
			lines = append(lines, lineResult{line: n, status: StatusMalformed, reason: "line exceeds 8 MiB"})
			return
		}
		if len(bytes.TrimSpace(raw)) == 0 {
			return
		}
		lr := parseLine(raw)
		lr.line = n
		if lr.event != nil {
			lr.event.File = rel
			lr.event.Line = n
		}
		lines = append(lines, lr)
	})
	if err != nil {
		return err
	}

	fileSession := strings.TrimSuffix(path.Base(rel), ".jsonl")
	for _, lr := range lines {
		if lr.sessionID != "" {
			fileSession = lr.sessionID
			break
		}
	}

	for _, lr := range lines {
		sid := lr.sessionID
		if sid == "" || (lr.event != nil && lr.event.Kind == model.KindSummary) {
			sid = fileSession
		}
		status, reason := lr.status, lr.reason
		if status == StatusParsed && lr.event.Kind != model.KindSummary {
			key := dedupeKey{sid, lr.event.UUID}
			if prev, ok := in.seen[key]; ok {
				if prev == lr.canonical {
					status, reason = StatusDuplicate, "same sessionId and uuid, identical canonical JSON"
				} else {
					status, reason = StatusMalformed, "uuid conflicts with an earlier record of the session"
				}
			} else {
				in.seen[key] = lr.canonical
			}
		}
		s := in.session(sid, rel)
		s.Counts.Records++
		switch status {
		case StatusParsed:
			s.Counts.Parsed++
			s.Events = append(s.Events, *lr.event)
			in.diagnose(lr.event)
		case StatusMalformed:
			s.Counts.Malformed++
		case StatusUnknown:
			s.Counts.Unknown++
		case StatusDuplicate:
			s.Counts.Duplicate++
		}
		in.res.Outcomes = append(in.res.Outcomes, Outcome{File: rel, Line: lr.line, Status: status, Reason: reason, SessionID: sid})
	}
	return nil
}

func (in *ingester) diagnose(e *model.Event) {
	d := &in.res.Diagnostics
	if !e.HasTime {
		d.MissingTimestamp++
	}
	if e.Sidechain {
		d.Sidechain++
	}
	if e.IsAssistant() {
		if e.Usage == nil {
			d.MissingUsage++
		} else if !e.Usage.HasCacheFields {
			d.MissingCacheFields++
		}
	}
}

func (in *ingester) session(id, file string) *model.Session {
	s, ok := in.sessions[id]
	if !ok {
		s = &model.Session{ID: id}
		in.sessions[id] = s
	}
	if n := len(s.Files); n == 0 || s.Files[n-1] != file {
		s.Files = append(s.Files, file)
	}
	return s
}

func (in *ingester) finish(files []string) *Result {
	res := &in.res
	res.Files = files
	for _, s := range in.sessions {
		finishSession(s)
		res.Counts.Add(s.Counts)
		res.Sessions = append(res.Sessions, s)
	}
	sort.Slice(res.Sessions, func(i, j int) bool { return res.Sessions[i].ID < res.Sessions[j].ID })
	return res
}

// finishSession orders events by timestamp (timeless last, file order kept on ties), sets
// the session span and picks the first genuine prompt in timestamp order.
func finishSession(s *model.Session) {
	sort.SliceStable(s.Events, func(i, j int) bool {
		a, b := &s.Events[i], &s.Events[j]
		if a.HasTime != b.HasTime {
			return a.HasTime
		}
		return a.HasTime && a.At.Before(b.At)
	})
	for i := range s.Events {
		e := &s.Events[i]
		if !e.HasTime {
			continue
		}
		if !s.HasStart || e.At.Before(s.Start) {
			s.Start = e.At
		}
		if !s.HasStart || e.At.After(s.End) {
			s.End = e.At
		}
		s.HasStart = true
		if !s.HasFirstPrompt && e.Kind == model.KindUserPrompt && e.Text != "" {
			s.FirstPrompt, s.FirstPromptAt, s.FirstPromptUUID, s.HasFirstPrompt = e.Text, e.At, e.UUID, true
		}
	}
}

// readLines calls fn for every line (1-based), with the trailing newline and any carriage
// return removed. A line longer than MaxLineBytes is reported once with tooLong set and its
// bytes discarded.
func readLines(r io.Reader, fn func(n int, line []byte, tooLong bool)) error {
	br := bufio.NewReaderSize(r, 64<<10)
	var buf []byte
	n := 0
	tooLong := false
	for {
		chunk, err := br.ReadSlice('\n')
		if !tooLong {
			if len(buf)+len(chunk) > MaxLineBytes+2 { // allow for "\r\n"
				tooLong = true
				buf = buf[:0]
			} else {
				buf = append(buf, chunk...)
			}
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		atEOF := err != nil
		if atEOF && len(buf) == 0 && !tooLong && len(chunk) == 0 {
			return nil
		}
		n++
		line := bytes.TrimRight(buf, "\r\n")
		if n == 1 {
			line = bytes.TrimPrefix(line, []byte("\xef\xbb\xbf"))
		}
		if !tooLong && len(line) > MaxLineBytes {
			tooLong = true
		}
		fn(n, line, tooLong)
		buf, tooLong = buf[:0], false
		if atEOF {
			return nil
		}
	}
}

var systemReminder = regexp.MustCompile(`(?s)<system-reminder>.*?</system-reminder>`)

// StripSystemReminders removes <system-reminder>…</system-reminder> spans and trims space.
func StripSystemReminders(s string) string {
	return strings.TrimSpace(systemReminder.ReplaceAllString(s, ""))
}

// shapeError is a malformed-record reason.
type shapeError string

func (e shapeError) Error() string { return string(e) }

func malformed(reason string) lineResult { return lineResult{status: StatusMalformed, reason: reason} }

// parseLine decodes and validates one nonblank line.
func parseLine(raw []byte) lineResult {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return malformed("invalid JSON")
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return malformed("invalid JSON: trailing data")
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return malformed("record is not a JSON object")
	}
	lr := lineResult{}
	if sid, present, ok := optString(obj, "sessionId"); present {
		if !ok || sid == "" {
			return malformed("sessionId is not a non-empty string")
		}
		lr.sessionID = sid
	}
	typ, ok := obj["type"].(string)
	if !ok {
		lr.status, lr.reason = StatusMalformed, "type missing or not a string"
		return lr
	}
	var ev *model.Event
	var unknownBlock bool
	var err error
	switch typ {
	case "summary":
		ev, err = parseSummary(obj)
	case "user":
		ev, unknownBlock, err = parseUser(obj)
	case "assistant":
		ev, unknownBlock, err = parseAssistant(obj)
	default:
		lr.status, lr.reason = StatusUnknown, "unknown record type"
		return lr
	}
	if err != nil {
		lr.status, lr.reason = StatusMalformed, err.Error()
		return lr
	}
	if unknownBlock {
		lr.status, lr.reason = StatusUnknown, "unknown content block type"
		return lr
	}
	canon, err := json.Marshal(v)
	if err != nil {
		lr.status, lr.reason = StatusMalformed, "cannot canonicalise"
		return lr
	}
	lr.status, lr.canonical, lr.event = StatusParsed, string(canon), ev
	return lr
}

func parseSummary(obj map[string]any) (*model.Event, error) {
	if _, present, ok := optString(obj, "summary"); present && !ok {
		return nil, shapeError("summary is not a string")
	}
	if _, present, ok := optString(obj, "leafUuid"); present && !ok {
		return nil, shapeError("leafUuid is not a string")
	}
	ev := &model.Event{Kind: model.KindSummary}
	if err := parseCommonTime(obj, ev); err != nil {
		return nil, err
	}
	// Summary records are timeless events of their file's session whatever they carry.
	ev.HasTime, ev.At = false, time.Time{}
	return ev, nil
}

// parseCommon reads the identity fields shared by user and assistant records.
func parseCommon(obj map[string]any) (*model.Event, error) {
	ev := &model.Event{}
	uuid, present, ok := optString(obj, "uuid")
	if !present || !ok || uuid == "" {
		return nil, shapeError("uuid missing or not a non-empty string")
	}
	ev.UUID = uuid
	if err := parseCommonTime(obj, ev); err != nil {
		return nil, err
	}
	if v, present := obj["isSidechain"]; present {
		b, ok := v.(bool)
		if !ok {
			return nil, shapeError("isSidechain is not a boolean")
		}
		ev.Sidechain = b
	}
	for _, f := range []struct {
		key string
		dst *string
	}{{"version", &ev.Version}, {"requestId", &ev.RequestID}} {
		s, present, ok := optString(obj, f.key)
		if present && !ok {
			return nil, shapeError(f.key + " is not a string")
		}
		*f.dst = s
	}
	if v, present := obj["parentUuid"]; present && v != nil {
		if _, ok := v.(string); !ok {
			return nil, shapeError("parentUuid is not a string or null")
		}
	}
	return ev, nil
}

// parseCommonTime sets ev.At when the record carries a timestamp; absence is not an error.
func parseCommonTime(obj map[string]any, ev *model.Event) error {
	ts, present, ok := optString(obj, "timestamp")
	if !present {
		return nil
	}
	if !ok {
		return shapeError("timestamp is not a string")
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return shapeError("timestamp is not RFC 3339")
	}
	ev.At, ev.HasTime = t.UTC(), true
	return nil
}

// pathless drops the path from a filesystem error so no absolute path reaches any output.
func pathless(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

func message(obj map[string]any) (map[string]any, error) {
	m, ok := obj["message"].(map[string]any)
	if !ok {
		return nil, shapeError("message missing or not an object")
	}
	return m, nil
}

func parseUser(obj map[string]any) (*model.Event, bool, error) {
	ev, err := parseCommon(obj)
	if err != nil {
		return nil, false, err
	}
	msg, err := message(obj)
	if err != nil {
		return nil, false, err
	}
	var texts []string
	unknown := false
	switch c := msg["content"].(type) {
	case string:
		texts = append(texts, c)
	case []any:
		for _, b := range c {
			blk, ok := b.(map[string]any)
			if !ok {
				return nil, false, shapeError("content block is not an object")
			}
			bt, ok := blk["type"].(string)
			if !ok {
				return nil, false, shapeError("content block type missing or not a string")
			}
			switch bt {
			case "text":
				t, ok := blk["text"].(string)
				if !ok {
					return nil, false, shapeError("text block without a string text")
				}
				texts = append(texts, t)
			case "tool_result":
				id, ok := blk["tool_use_id"].(string)
				if !ok {
					return nil, false, shapeError("tool_result without a string tool_use_id")
				}
				out := model.OutcomeUnknown
				if v, present := blk["is_error"]; present {
					isErr, ok := v.(bool)
					if !ok {
						return nil, false, shapeError("is_error is not a boolean")
					}
					out = model.OutcomeSuccess
					if isErr {
						out = model.OutcomeError
					}
				}
				ev.ToolResults = append(ev.ToolResults, model.ToolResult{ToolUseID: id, Outcome: out})
			default:
				unknown = true
			}
		}
	default:
		return nil, false, shapeError("user content is neither a string nor an array")
	}
	if len(ev.ToolResults) > 0 {
		ev.Kind = model.KindToolResult
	} else {
		ev.Kind = model.KindUserPrompt
		ev.Text = StripSystemReminders(strings.Join(texts, "\n"))
	}
	return ev, unknown, nil
}

func parseAssistant(obj map[string]any) (*model.Event, bool, error) {
	ev, err := parseCommon(obj)
	if err != nil {
		return nil, false, err
	}
	msg, err := message(obj)
	if err != nil {
		return nil, false, err
	}
	if v, present := msg["model"]; present {
		s, ok := v.(string)
		if !ok {
			return nil, false, shapeError("model is not a string")
		}
		ev.Model = s
	}
	unknown := false
	switch c := msg["content"].(type) {
	case string:
	case []any:
		for _, b := range c {
			blk, ok := b.(map[string]any)
			if !ok {
				return nil, false, shapeError("content block is not an object")
			}
			bt, ok := blk["type"].(string)
			if !ok {
				return nil, false, shapeError("content block type missing or not a string")
			}
			switch bt {
			case "text":
				if _, ok := blk["text"].(string); !ok {
					return nil, false, shapeError("text block without a string text")
				}
			case "tool_use":
				id, ok1 := blk["id"].(string)
				name, ok2 := blk["name"].(string)
				if !ok1 || !ok2 {
					return nil, false, shapeError("tool_use without a string id and name")
				}
				ev.ToolCalls = append(ev.ToolCalls, model.ToolCall{ID: id, Name: name})
			default:
				unknown = true
			}
		}
	default:
		return nil, false, shapeError("assistant content is neither a string nor an array")
	}
	if len(ev.ToolCalls) > 0 {
		ev.Kind = model.KindToolCall
	} else {
		ev.Kind = model.KindAssistantText
	}
	if v, present := msg["usage"]; present && v != nil {
		u, ok := v.(map[string]any)
		if !ok {
			return nil, false, shapeError("usage is not an object")
		}
		usage, err := parseUsage(u)
		if err != nil {
			return nil, false, err
		}
		ev.Usage = usage
	}
	return ev, unknown, nil
}

func parseUsage(u map[string]any) (*model.Usage, error) {
	out := &model.Usage{}
	for _, f := range []struct {
		key      string
		dst      *int64
		required bool
	}{
		{"input_tokens", &out.Input, true},
		{"output_tokens", &out.Output, true},
		{"cache_read_input_tokens", &out.CacheRead, false},
		{"cache_creation_input_tokens", &out.CacheWrite, false},
	} {
		v, present := u[f.key]
		if !present {
			if f.required {
				return nil, shapeError(f.key + " missing from usage")
			}
			continue
		}
		n, err := tokenCount(v)
		if err != nil {
			return nil, shapeError(f.key + ": " + err.Error())
		}
		*f.dst = n
	}
	_, hasRead := u["cache_read_input_tokens"]
	_, hasWrite := u["cache_creation_input_tokens"]
	out.HasCacheFields = hasRead && hasWrite
	if s, present, ok := optString(u, "service_tier"); present {
		if !ok {
			return nil, shapeError("service_tier is not a string")
		}
		out.ServiceTier = s
	}
	return out, nil
}

// tokenCount accepts only a non-negative integer literal.
func tokenCount(v any) (int64, error) {
	num, ok := v.(json.Number)
	if !ok {
		return 0, errors.New("not a number")
	}
	n, err := strconv.ParseInt(string(num), 10, 64)
	if err != nil {
		return 0, errors.New("not an integer")
	}
	if n < 0 {
		return 0, errors.New("negative")
	}
	return n, nil
}

// optString reports whether key is present and, if so, whether it is a string.
func optString(obj map[string]any, key string) (s string, present, ok bool) {
	v, present := obj[key]
	if !present {
		return "", false, false
	}
	s, ok = v.(string)
	return s, true, ok
}
