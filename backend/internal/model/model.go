// Package model holds the in-memory shape of ingested sessions. Nothing here is serialised
// to the API directly; the summary package maps it onto the contract structs. Prompt text
// lives only in fields tagged `json:"-"` so an accidental marshal cannot leak it.
package model

import "time"

// Kind is the role of one parsed record.
type Kind string

const (
	KindUserPrompt    Kind = "user_prompt"    // user record without a tool_result block
	KindToolResult    Kind = "tool_result"    // user record carrying at least one tool_result block
	KindAssistantText Kind = "assistant_text" // assistant record without a tool_use block
	KindToolCall      Kind = "tool_call"      // assistant record with at least one tool_use block
	KindSummary       Kind = "summary"        // summary record: no identity, attached to its file's session
)

// Outcome of a tool result, read from is_error.
type Outcome string

const (
	OutcomeUnknown Outcome = "unknown" // is_error absent
	OutcomeSuccess Outcome = "success" // is_error false
	OutcomeError   Outcome = "error"   // is_error true
)

// Usage is the token usage of one assistant record.
type Usage struct {
	Input          int64
	Output         int64
	CacheRead      int64
	CacheWrite     int64
	HasCacheFields bool // both cache_read_input_tokens and cache_creation_input_tokens present
	ServiceTier    string
}

// ToolCall is one tool_use block of an assistant record.
type ToolCall struct {
	ID   string
	Name string
}

// ToolResult is one tool_result block of a user record.
type ToolResult struct {
	ToolUseID string
	Outcome   Outcome
}

// Event is one parsed (not duplicate, not malformed, not unknown) record. One record is one
// event, so usage on an assistant record is billed exactly once however many content blocks
// it carries.
type Event struct {
	UUID        string
	File        string // path relative to the input directory, forward slashes
	Line        int    // 1-based line number in File
	At          time.Time
	HasTime     bool
	Kind        Kind
	Model       string // assistant only, exactly as logged
	Version     string
	RequestID   string
	Sidechain   bool
	ToolCalls   []ToolCall
	ToolResults []ToolResult
	Usage       *Usage // assistant only; nil when the record has no usage
	Text        string `json:"-"` // user_prompt only: text after <system-reminder> stripping, trimmed
}

// IsAssistant reports whether the event is an assistant record.
func (e *Event) IsAssistant() bool {
	return e.Kind == KindAssistantText || e.Kind == KindToolCall
}

// Counts is the accounting of the nonblank lines attributed to a session or a directory.
// Records = Parsed + Malformed + Unknown + Duplicate.
type Counts struct {
	Records   int
	Parsed    int
	Malformed int
	Unknown   int
	Duplicate int
}

// Add accumulates o into c.
func (c *Counts) Add(o Counts) {
	c.Records += o.Records
	c.Parsed += o.Parsed
	c.Malformed += o.Malformed
	c.Unknown += o.Unknown
	c.Duplicate += o.Duplicate
}

// Session is every record attributed to one session id.
type Session struct {
	ID       string
	Files    []string // relative to the input directory, forward slashes
	Start    time.Time
	End      time.Time
	HasStart bool // false when no event carries a timestamp
	Events   []Event
	Counts   Counts

	FirstPrompt     string `json:"-"` // never serialised
	FirstPromptAt   time.Time
	FirstPromptUUID string
	HasFirstPrompt  bool
}
