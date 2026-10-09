package ingest

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
)

var (
	fixtures  = filepath.Join("..", "..", "..", "fixtures")
	corpusDir = filepath.Join(fixtures, "claude-code")
	synthDir  = filepath.Join(fixtures, "synthetic")
)

const (
	synth1 = "5ea00000-0000-4000-8000-000000000001"
	synth2 = "5ea00000-0000-4000-8000-000000000002"
	synth3 = "5ea00000-0000-4000-8000-000000000003"
	synth4 = "5ea00000-0000-4000-8000-000000000004"
	synth5 = "5ea00000-0000-4000-8000-000000000005"
	synth6 = "5ea00000-0000-4000-8000-000000000006"
)

func ingestDir(t *testing.T, dir string) *Result {
	t.Helper()
	res, err := Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	checkInvariant(t, res)
	return res
}

// checkInvariant: every nonblank line has exactly one outcome and the totals agree at
// directory and session level.
func checkInvariant(t *testing.T, res *Result) {
	t.Helper()
	c := res.Counts
	if c.Records != c.Parsed+c.Malformed+c.Unknown+c.Duplicate {
		t.Errorf("total %d != parsed %d + malformed %d + unknown %d + duplicate %d", c.Records, c.Parsed, c.Malformed, c.Unknown, c.Duplicate)
	}
	if len(res.Outcomes) != c.Records {
		t.Errorf("outcomes %d != records %d", len(res.Outcomes), c.Records)
	}
	seen := map[string]bool{}
	byStatus := map[Status]int{}
	for _, o := range res.Outcomes {
		k := o.File + ":" + itoa(o.Line)
		if seen[k] {
			t.Errorf("line %s counted twice", k)
		}
		seen[k] = true
		byStatus[o.Status]++
	}
	if byStatus[StatusParsed] != c.Parsed || byStatus[StatusMalformed] != c.Malformed ||
		byStatus[StatusUnknown] != c.Unknown || byStatus[StatusDuplicate] != c.Duplicate {
		t.Errorf("outcome statuses %v disagree with counts %+v", byStatus, c)
	}
	events := 0
	for _, s := range res.Sessions {
		sc := s.Counts
		if sc.Records != sc.Parsed+sc.Malformed+sc.Unknown+sc.Duplicate {
			t.Errorf("session %s: counts do not add up: %+v", s.ID, sc)
		}
		if len(s.Events) != sc.Parsed {
			t.Errorf("session %s: %d events for %d parsed", s.ID, len(s.Events), sc.Parsed)
		}
		events += len(s.Events)
	}
	if events != c.Parsed {
		t.Errorf("events %d != parsed %d", events, c.Parsed)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func outcomesOf(res *Result, session string) []Outcome {
	var out []Outcome
	for _, o := range res.Outcomes {
		if strings.HasSuffix(o.File, session+".jsonl") {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Line < out[j].Line })
	return out
}

func statuses(outs []Outcome) map[int]Status {
	m := map[int]Status{}
	for _, o := range outs {
		m[o.Line] = o.Status
	}
	return m
}

func mustSession(t *testing.T, res *Result, id string) *model.Session {
	t.Helper()
	s := res.Session(id)
	if s == nil {
		t.Fatalf("session %s not found", id)
	}
	return s
}

func eventByUUID(s *model.Session, uuid string) *model.Event {
	for i := range s.Events {
		if s.Events[i].UUID == uuid {
			return &s.Events[i]
		}
	}
	return nil
}

func TestCorpusAccounting(t *testing.T) {
	res := ingestDir(t, corpusDir)
	if len(res.Files) != 34 {
		t.Errorf("files = %d, want 34", len(res.Files))
	}
	want := model.Counts{Records: 1107, Parsed: 1084, Malformed: 1, Unknown: 0, Duplicate: 22}
	if res.Counts != want {
		t.Errorf("counts = %+v, want %+v", res.Counts, want)
	}
	wantD := Diagnostics{MissingUsage: 10, MissingTimestamp: 12, MissingCacheFields: 59, Sidechain: 2}
	if res.Diagnostics != wantD {
		t.Errorf("diagnostics = %+v, want %+v", res.Diagnostics, wantD)
	}
	if len(res.Sessions) != 34 {
		t.Errorf("sessions = %d, want 34", len(res.Sessions))
	}
	for _, o := range res.Outcomes {
		if o.Status == StatusMalformed && !strings.HasSuffix(o.File, "6061f190-d64b-41d4-8d7b-de81a3d6b6ee.jsonl") {
			t.Errorf("unexpected malformed line in session %s line %d (%s)", o.SessionID, o.Line, o.Reason)
		}
		if o.Status == StatusDuplicate && !strings.HasSuffix(o.File, "b7dab20e-479d-4d1b-b6e4-8ac4e6451e08.jsonl") {
			t.Errorf("unexpected duplicate in session %s line %d", o.SessionID, o.Line)
		}
	}
	// All 22 duplicates belong to one session, b7dab20e…, by the session they are attributed to.
	dupSessions := map[string]int{}
	for _, o := range res.Outcomes {
		if o.Status == StatusDuplicate {
			dupSessions[o.SessionID]++
		}
	}
	if len(dupSessions) != 1 || dupSessions["b7dab20e-479d-4d1b-b6e4-8ac4e6451e08"] != 22 {
		t.Errorf("duplicates by session = %v, want all 22 in b7dab20e…", dupSessions)
	}
	if s := mustSession(t, res, "b7dab20e-479d-4d1b-b6e4-8ac4e6451e08"); s.Counts.Duplicate != 22 {
		t.Errorf("b7dab20e… session duplicate count = %d, want 22", s.Counts.Duplicate)
	}
}

func TestSyntheticDirectoryAccounting(t *testing.T) {
	res := ingestDir(t, synthDir)
	want := model.Counts{Records: 34, Parsed: 27, Malformed: 3, Unknown: 2, Duplicate: 2}
	if res.Counts != want {
		t.Errorf("counts = %+v, want %+v", res.Counts, want)
	}
	wantD := Diagnostics{MissingUsage: 1, MissingTimestamp: 2, MissingCacheFields: 1, Sidechain: 1}
	if res.Diagnostics != wantD {
		t.Errorf("diagnostics = %+v, want %+v", res.Diagnostics, wantD)
	}
}

func TestMalformedLineCountedNotDropped(t *testing.T) {
	res := ingestDir(t, synthDir)
	s := mustSession(t, res, synth1)
	want := model.Counts{Records: 5, Parsed: 3, Malformed: 2}
	if s.Counts != want {
		t.Errorf("session 1 counts = %+v, want %+v", s.Counts, want)
	}
	got := statuses(outcomesOf(res, synth1))
	wantLines := map[int]Status{1: StatusParsed, 2: StatusParsed, 3: StatusMalformed, 4: StatusMalformed, 6: StatusParsed}
	if len(got) != len(wantLines) {
		t.Errorf("line outcomes = %v, want %v (blank line 5 not counted)", got, wantLines)
	}
	for l, st := range wantLines {
		if got[l] != st {
			t.Errorf("line %d = %s, want %s", l, got[l], st)
		}
	}
	for _, o := range outcomesOf(res, synth1) {
		if o.SessionID != synth1 {
			t.Errorf("line %d attributed to %s", o.Line, o.SessionID)
		}
	}

	t.Run("oversize line", func(t *testing.T) {
		dir := t.TempDir()
		big := `{"type":"user","pad":"` + strings.Repeat("x", MaxLineBytes) + `"}`
		ok := `{"type":"summary","summary":"s","leafUuid":"u"}`
		if err := os.WriteFile(filepath.Join(dir, "a.jsonl"), []byte(big+"\n"+ok+"\n\n"+ok), 0o600); err != nil {
			t.Fatal(err)
		}
		res := ingestDir(t, dir)
		want := model.Counts{Records: 3, Parsed: 2, Malformed: 1}
		if res.Counts != want {
			t.Errorf("counts = %+v, want %+v", res.Counts, want)
		}
		if st := statuses(res.Outcomes); st[1] != StatusMalformed || st[2] != StatusParsed || st[4] != StatusParsed {
			t.Errorf("outcomes = %v", st)
		}
	})

	t.Run("bad shapes", func(t *testing.T) {
		cases := map[string]string{
			"non-integer tokens": `{"type":"assistant","uuid":"a","sessionId":"s","message":{"model":"m","content":[],"usage":{"input_tokens":1.5,"output_tokens":1}}}`,
			"string tokens":      `{"type":"assistant","uuid":"a","sessionId":"s","message":{"model":"m","content":[],"usage":{"input_tokens":"1","output_tokens":1}}}`,
			"bad timestamp":      `{"type":"user","uuid":"a","sessionId":"s","timestamp":"yesterday","message":{"content":"x"}}`,
			"no uuid":            `{"type":"user","sessionId":"s","message":{"content":"x"}}`,
			"content number":     `{"type":"user","uuid":"a","sessionId":"s","message":{"content":3}}`,
			"not an object":      `[1,2]`,
			"trailing data":      `{"type":"summary"} {}`,
		}
		for name, line := range cases {
			lr := parseLine([]byte(line))
			if lr.status != StatusMalformed {
				t.Errorf("%s: status %s, want malformed", name, lr.status)
			}
		}
	})
}

func TestUnknownTypeAndUnknownBlockAreUnknown(t *testing.T) {
	res := ingestDir(t, synthDir)
	s := mustSession(t, res, synth2)
	want := model.Counts{Records: 4, Parsed: 2, Unknown: 2}
	if s.Counts != want {
		t.Errorf("session 2 counts = %+v, want %+v", s.Counts, want)
	}
	got := statuses(outcomesOf(res, synth2))
	for l, st := range map[int]Status{1: StatusParsed, 2: StatusUnknown, 3: StatusUnknown, 4: StatusParsed} {
		if got[l] != st {
			t.Errorf("line %d = %s, want %s", l, got[l], st)
		}
	}
	if eventByUUID(s, "5ea00002-0000-4000-8000-000000000003") != nil {
		t.Error("the record with an unknown block must not become an event")
	}
}

func TestDuplicateVsConflictingUUID(t *testing.T) {
	res := ingestDir(t, synthDir)
	s := mustSession(t, res, synth3)
	want := model.Counts{Records: 5, Parsed: 2, Duplicate: 2, Malformed: 1}
	if s.Counts != want {
		t.Errorf("session 3 counts = %+v, want %+v", s.Counts, want)
	}
	got := statuses(outcomesOf(res, synth3))
	for l, st := range map[int]Status{1: StatusParsed, 2: StatusParsed, 3: StatusDuplicate, 4: StatusDuplicate, 5: StatusMalformed} {
		if got[l] != st {
			t.Errorf("line %d = %s, want %s", l, got[l], st)
		}
	}
	assistants := 0
	for _, e := range s.Events {
		if e.IsAssistant() {
			assistants++
		}
	}
	if assistants != 1 {
		t.Errorf("assistant events = %d, want 1 (billed once)", assistants)
	}
}

// genuinePrompts returns the dated user_prompt events with non-empty text.
func genuinePrompts(s *model.Session) []*model.Event {
	var out []*model.Event
	for i := range s.Events {
		e := &s.Events[i]
		if e.Kind == model.KindUserPrompt && e.HasTime && e.Text != "" {
			out = append(out, e)
		}
	}
	return out
}

func TestFirstPromptUsesTimestampOrder(t *testing.T) {
	t.Run("synthetic session 4", func(t *testing.T) {
		s := mustSession(t, ingestDir(t, synthDir), synth4)
		if !s.HasFirstPrompt || s.FirstPromptUUID != "5ea00004-0000-4000-8000-000000000003" {
			t.Fatalf("first prompt uuid = %q, want line 3's record", s.FirstPromptUUID)
		}
		if e := eventByUUID(s, "5ea00004-0000-4000-8000-000000000004"); e == nil || e.HasTime {
			t.Error("line 4 must be parsed and undated")
		}
		if last := s.Events[len(s.Events)-1]; last.HasTime {
			t.Error("undated events must sort last")
		}
	})
	t.Run("corpus 3888cf6f", func(t *testing.T) {
		res := ingestDir(t, corpusDir)
		s := mustSession(t, res, "3888cf6f-f20f-4113-9752-143c54690953")
		ps := genuinePrompts(s)
		if len(ps) == 0 || !s.HasFirstPrompt {
			t.Fatal("no genuine prompt")
		}
		earliest, fileFirst := ps[0], ps[0]
		for _, p := range ps {
			if p.At.Before(earliest.At) {
				earliest = p
			}
			if p.Line < fileFirst.Line {
				fileFirst = p
			}
		}
		if s.FirstPromptUUID != earliest.UUID || !s.FirstPromptAt.Equal(earliest.At) {
			t.Errorf("first prompt is line %d, want the earliest dated prompt at line %d", eventByUUID(s, s.FirstPromptUUID).Line, earliest.Line)
		}
		if fileFirst.UUID == earliest.UUID {
			t.Error("fixture no longer exercises timestamp order: file-order first prompt is also the earliest")
		}
		for i := 1; i < len(s.Events); i++ {
			a, b := s.Events[i-1], s.Events[i]
			if a.HasTime && b.HasTime && b.At.Before(a.At) {
				t.Fatalf("events not sorted at index %d", i)
			}
		}
	})
	// b7dab20e… is out of timestamp order only because lines 45–66 are byte copies of lines
	// 1–22: the raw timestamps jump back at line 45. The copies are duplicates, the parsed
	// records are already in file order, so the first prompt is line 1 under both orders.
	t.Run("corpus b7dab20e", func(t *testing.T) {
		const id = "b7dab20e-479d-4d1b-b6e4-8ac4e6451e08"
		res := ingestDir(t, corpusDir)
		s := mustSession(t, res, id)

		raw, err := os.ReadFile(filepath.Join(corpusDir, filepath.FromSlash(s.Files[0])))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
		if len(lines) != 66 {
			t.Fatalf("file has %d lines, want 66", len(lines))
		}
		stamp := func(n int) string {
			var r struct {
				Timestamp string `json:"timestamp"`
			}
			if err := json.Unmarshal([]byte(lines[n-1]), &r); err != nil {
				t.Fatalf("line %d: %v", n, err)
			}
			return r.Timestamp
		}
		if !(stamp(45) < stamp(44)) {
			t.Error("fixture no longer jumps back in time at line 45")
		}
		for n := 45; n <= 66; n++ {
			if lines[n-1] != lines[n-45] {
				t.Errorf("line %d is not a byte copy of line %d", n, n-44)
			}
		}
		for _, o := range outcomesOf(res, id) {
			if (o.Status == StatusDuplicate) != (o.Line >= 45) {
				t.Errorf("line %d status %s", o.Line, o.Status)
			}
		}

		byLine := map[int]*model.Event{}
		for i := range s.Events {
			byLine[s.Events[i].Line] = &s.Events[i]
		}
		var prev *model.Event
		for n := 1; n <= 66; n++ {
			e := byLine[n]
			if e == nil || !e.HasTime {
				continue
			}
			if prev != nil && e.At.Before(prev.At) {
				t.Errorf("parsed line %d is earlier than parsed line %d", e.Line, prev.Line)
			}
			prev = e
		}

		ps := genuinePrompts(s)
		if len(ps) == 0 || !s.HasFirstPrompt {
			t.Fatal("no genuine prompt")
		}
		first := eventByUUID(s, s.FirstPromptUUID)
		if first == nil || first.Line != 1 || s.Events[0].UUID != s.FirstPromptUUID {
			t.Error("first prompt must be line 1, the earliest event of the session")
		}
		copies := 0
		for _, e := range s.Events {
			if e.UUID == s.FirstPromptUUID {
				copies++
			}
		}
		if copies != 1 {
			t.Errorf("first prompt record appears %d times among events, want 1", copies)
		}
	})
}

func TestStringAndBlockContentArePrompts(t *testing.T) {
	res := ingestDir(t, synthDir)
	s1 := mustSession(t, res, synth1)
	if s1.FirstPromptUUID != "5ea00001-0000-4000-8000-000000000001" {
		t.Error("session 1: string content on line 1 must be the first prompt")
	}
	if s1.FirstPrompt != "Fix the failing login test" {
		t.Error("session 1: first prompt text differs from line 1's string content")
	}
	s4 := mustSession(t, res, synth4)
	if s4.FirstPrompt != "Refactor the config loader into its own package" {
		t.Error("session 4: first prompt text differs from line 3's text block")
	}
	if e := eventByUUID(s4, "5ea00004-0000-4000-8000-000000000003"); e == nil || e.Kind != model.KindUserPrompt {
		t.Error("session 4: text-block record must be a user_prompt event")
	}
	b, err := json.Marshal(s4)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), s4.FirstPrompt) || strings.Contains(string(b), "changelog") {
		t.Error("marshalled session carries prompt text")
	}
	s6 := mustSession(t, res, synth6)
	if s6.FirstPromptUUID != "5ea00006-0000-4000-8000-000000000002" {
		t.Error("session 6: first prompt must be line 2, after the summary record")
	}
}

func TestToolResultsAreNotPrompts(t *testing.T) {
	s := mustSession(t, ingestDir(t, synthDir), synth5)
	if s.FirstPromptUUID != "5ea00005-0000-4000-8000-000000000003" {
		t.Error("session 5: first prompt must be line 3, not the tool result on line 1")
	}
	calls, results, errs := 0, 0, 0
	outcomes := map[string]model.Outcome{}
	for _, e := range s.Events {
		calls += len(e.ToolCalls)
		results += len(e.ToolResults)
		if len(e.ToolResults) > 0 && e.Kind != model.KindToolResult {
			t.Errorf("record %s carries a tool result but is %s", e.UUID, e.Kind)
		}
		for _, r := range e.ToolResults {
			outcomes[r.ToolUseID] = r.Outcome
			if r.Outcome == model.OutcomeError {
				errs++
			}
		}
	}
	if calls != 2 || results != 3 || errs != 1 {
		t.Errorf("tool calls %d results %d errors %d, want 2, 3, 1", calls, results, errs)
	}
	want := map[string]model.Outcome{"toolu_s5_0": model.OutcomeSuccess, "toolu_s5_1": model.OutcomeError, "toolu_s5_2": model.OutcomeSuccess}
	for id, o := range want {
		if outcomes[id] != o {
			t.Errorf("%s outcome %s, want %s", id, outcomes[id], o)
		}
	}
}

func TestSystemReminderStrippedFromSyntheticPrompt(t *testing.T) {
	s := mustSession(t, ingestDir(t, synthDir), synth5)
	if s.FirstPrompt != "Write unit tests for the date parser" {
		t.Error("session 5: first prompt is not line 3's text with the reminder removed")
	}
	e := eventByUUID(s, "5ea00005-0000-4000-8000-000000000002")
	if e == nil || e.Kind != model.KindUserPrompt || e.Text != "" {
		t.Error("session 5 line 2: must be a user record whose text is empty after stripping")
	}
	cases := map[string]string{
		"<system-reminder>a</system-reminder>":                                        "",
		"x<system-reminder>a\nb</system-reminder>y":                                   "xy",
		"<system-reminder>a</system-reminder> k <system-reminder>b</system-reminder>": "k",
		"no tags": "no tags",
	}
	for in, want := range cases {
		if got := StripSystemReminders(in); got != want {
			t.Errorf("strip case %d chars: got %d chars, want %d", len(in), len(got), len(want))
		}
	}
}

func TestSummaryRecordsAttachToFileSession(t *testing.T) {
	t.Run("synthetic session 6", func(t *testing.T) {
		res := ingestDir(t, synthDir)
		s := mustSession(t, res, synth6)
		if s.Counts.Records != 7 || s.Counts.Parsed != 7 {
			t.Errorf("session 6 counts = %+v, want 7 parsed", s.Counts)
		}
		last := s.Events[len(s.Events)-1]
		if last.Kind != model.KindSummary || last.HasTime || last.Line != 1 {
			t.Errorf("last event = %s line %d dated %v, want the timeless summary from line 1", last.Kind, last.Line, last.HasTime)
		}
	})
	t.Run("corpus", func(t *testing.T) {
		res := ingestDir(t, corpusDir)
		n := 0
		for _, s := range res.Sessions {
			for _, e := range s.Events {
				if e.Kind != model.KindSummary {
					continue
				}
				n++
				if !strings.HasSuffix(e.File, s.ID+".jsonl") {
					t.Errorf("summary from file %s attached to session %s", path.Base(e.File), s.ID)
				}
				if e.HasTime {
					t.Error("summary event must be timeless")
				}
			}
		}
		if n != 12 {
			t.Errorf("summary events = %d, want 12", n)
		}
	})
}

func TestFixturesMatchManifest(t *testing.T) {
	f, err := os.Open(filepath.Join(fixtures, "claude-code.sha256"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	listed := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		sum, rel, ok := strings.Cut(line, " ")
		if !ok {
			t.Fatalf("manifest line without a path")
		}
		rel = strings.TrimPrefix(strings.TrimSpace(rel), "*")
		raw, err := os.ReadFile(filepath.Join(fixtures, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("%s: %v", path.Base(rel), err)
			continue
		}
		h := sha256.Sum256(raw)
		if hex.EncodeToString(h[:]) != strings.ToLower(sum) {
			t.Errorf("%s: SHA-256 differs from the manifest", path.Base(rel))
		}
		listed[rel] = true
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 34 {
		t.Errorf("manifest lists %d files, want 34", len(listed))
	}
	res := ingestDir(t, corpusDir)
	for _, rel := range res.Files {
		if !listed["claude-code/"+rel] {
			t.Errorf("%s is ingested but not in the manifest", path.Base(rel))
		}
	}
}
