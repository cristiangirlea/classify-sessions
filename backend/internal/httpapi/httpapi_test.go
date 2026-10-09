package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
	"github.com/cristiangirlea/classify-sessions/backend/internal/ingest"
	"github.com/cristiangirlea/classify-sessions/backend/internal/model"
	"github.com/cristiangirlea/classify-sessions/backend/internal/summary"
)

var corpusDir = filepath.Join("..", "..", "..", "fixtures", "claude-code")

type req struct {
	method, path, host string
	header             map[string]string
}

func do(h http.Handler, q req) *httptest.ResponseRecorder {
	r := httptest.NewRequest(q.method, q.path, nil)
	r.Host = q.host
	for k, v := range q.header {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

var errorCode = regexp.MustCompile(`^[a-z_]+$`)

// checkError decodes an error body strictly into the contract's ErrorResponse.
func checkError(t *testing.T, w *httptest.ResponseRecorder, code string) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	dec.DisallowUnknownFields()
	var e contract.ErrorResponse
	if err := dec.Decode(&e); err != nil {
		t.Fatalf("error body does not match the contract: %v", err)
	}
	if e.Error.Code != code || !errorCode.MatchString(e.Error.Code) || e.Error.Message == "" {
		t.Errorf("error %+v, want code %s", e.Error, code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("error content type %q", ct)
	}
}

func checkSecurityHeaders(t *testing.T, h http.Header) {
	t.Helper()
	want := map[string]string{
		"Cache-Control":                "no-store",
		"X-Content-Type-Options":       "nosniff",
		"Referrer-Policy":              "no-referrer",
		"Content-Security-Policy":      "default-src 'none'",
		"Cross-Origin-Resource-Policy": "same-origin",
	}
	for k, v := range want {
		if got := h.Get(k); got != v {
			t.Errorf("header %s = %q, want %q", k, got, v)
		}
	}
}

func TestHTTPBoundary(t *testing.T) {
	res, err := ingest.Dir(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	sum := summary.Build(res, summary.Options{Input: "../fixtures/claude-code", GeneratedAt: time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)})
	h, err := New(sum)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(sum)
	if err != nil {
		t.Fatal(err)
	}
	const host = "127.0.0.1:8080"

	t.Run("GET summary", func(t *testing.T) {
		w := do(h, req{method: "GET", path: SummaryPath, host: host})
		if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), want) {
			t.Fatalf("status %d, body equal %v", w.Code, bytes.Equal(w.Body.Bytes(), want))
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("content type %q", w.Header().Get("Content-Type"))
		}
		dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
		dec.DisallowUnknownFields()
		var s contract.SummaryResponse
		if err := dec.Decode(&s); err != nil {
			t.Errorf("body does not decode strictly: %v", err)
		}
		checkSecurityHeaders(t, w.Header())
		checkNoPromptText(t, res, w.Body.String())
	})

	t.Run("HEAD summary", func(t *testing.T) {
		w := do(h, req{method: "HEAD", path: SummaryPath, host: host})
		if w.Code != 200 || w.Body.Len() != 0 {
			t.Errorf("status %d, body %d bytes; want 200 and no body", w.Code, w.Body.Len())
		}
		if w.Header().Get("Content-Length") != itoa(len(want)) {
			t.Errorf("content length %q, want %d", w.Header().Get("Content-Length"), len(want))
		}
		checkSecurityHeaders(t, w.Header())
	})

	t.Run("GET and HEAD models", func(t *testing.T) {
		w := do(h, req{method: "GET", path: ModelsPath, host: host})
		if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("status %d content type %q", w.Code, w.Header().Get("Content-Type"))
		}
		var full struct {
			ModelChoice json.RawMessage `json:"model_choice"`
		}
		if err := json.Unmarshal(want, &full); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(w.Body.Bytes(), full.ModelChoice) {
			t.Error("/api/models body differs from the summary's model_choice")
		}
		dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
		dec.DisallowUnknownFields()
		var m contract.ModelsResponse
		if err := dec.Decode(&m); err != nil {
			t.Errorf("models body does not decode strictly: %v", err)
		}
		checkSecurityHeaders(t, w.Header())
		checkNoPromptText(t, res, w.Body.String())

		hw := do(h, req{method: "HEAD", path: ModelsPath, host: host})
		if hw.Code != 200 || hw.Body.Len() != 0 || hw.Header().Get("Content-Length") != itoa(w.Body.Len()) {
			t.Errorf("HEAD: status %d, body %d bytes, length %q", hw.Code, hw.Body.Len(), hw.Header().Get("Content-Length"))
		}
		checkSecurityHeaders(t, hw.Header())

		if fw := do(h, req{method: "GET", path: ModelsPath, host: "example.test"}); fw.Code != 403 {
			t.Errorf("foreign host: status %d", fw.Code)
		}
	})

	t.Run("loopback hosts and same-origin requests are allowed", func(t *testing.T) {
		for i, q := range []req{
			{method: "GET", path: SummaryPath, host: "localhost:5173"},
			{method: "GET", path: SummaryPath, host: "[::1]:8080"},
			{method: "GET", path: SummaryPath, host: "localhost"},
			{method: "GET", path: SummaryPath + "?x=1", host: "LOCALHOST:8080"},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": "http://" + host}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Sec-Fetch-Site": "same-origin"}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Sec-Fetch-Site": "none"}},
		} {
			if w := do(h, q); w.Code != 200 {
				t.Errorf("case %d: status %d, want 200", i, w.Code)
			}
		}
	})

	t.Run("404", func(t *testing.T) {
		for _, p := range []string{"/", "/api", "/api/summary/", "/api/models/", "/api/model", "/api/unknown", "/index.html"} {
			w := do(h, req{method: "GET", path: p, host: host})
			if w.Code != 404 {
				t.Errorf("%s: status %d, want 404", p, w.Code)
				continue
			}
			checkError(t, w, "not_found")
			checkSecurityHeaders(t, w.Header())
		}
		w := do(h, req{method: "HEAD", path: "/nope", host: host})
		if w.Code != 404 || w.Body.Len() != 0 {
			t.Errorf("HEAD unknown path: status %d, body %d bytes", w.Code, w.Body.Len())
		}
	})

	t.Run("405", func(t *testing.T) {
		for _, m := range []string{"POST", "PUT", "PATCH", "DELETE", "OPTIONS", "TRACE"} {
			for _, p := range []string{SummaryPath, ModelsPath, "/nope"} {
				w := do(h, req{method: m, path: p, host: host})
				if w.Code != 405 || w.Header().Get("Allow") != "GET, HEAD" {
					t.Errorf("%s %s: status %d Allow %q, want 405 and GET, HEAD", m, p, w.Code, w.Header().Get("Allow"))
					continue
				}
				checkError(t, w, "method_not_allowed")
				checkSecurityHeaders(t, w.Header())
			}
		}
	})

	t.Run("403", func(t *testing.T) {
		for i, q := range []req{
			{method: "GET", path: SummaryPath, host: "example.test"},
			{method: "GET", path: SummaryPath, host: "example.test:8080"},
			{method: "GET", path: SummaryPath, host: "127.0.0.1.example.test:8080"},
			{method: "GET", path: SummaryPath, host: "192.168.1.10:8080"},
			{method: "GET", path: SummaryPath, host: "0.0.0.0:8080"},
			{method: "GET", path: SummaryPath, host: ""},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": "http://example.test"}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": "null"}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": "https://" + host}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": "http://localhost:8080"}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Origin": ""}},
			{method: "GET", path: SummaryPath, host: host, header: map[string]string{"Sec-Fetch-Site": "cross-site"}},
			{method: "HEAD", path: SummaryPath, host: host, header: map[string]string{"Sec-Fetch-Site": "cross-site"}},
			{method: "POST", path: "/nope", host: "example.test"},
		} {
			w := do(h, q)
			if w.Code != 403 {
				t.Errorf("case %d: status %d, want 403", i, w.Code)
				continue
			}
			checkSecurityHeaders(t, w.Header())
			if q.method == "HEAD" {
				if w.Body.Len() != 0 {
					t.Errorf("case %d: HEAD error carries a body", i)
				}
				continue
			}
			checkError(t, w, "forbidden")
		}
	})

	t.Run("live loopback round trip", func(t *testing.T) {
		srv := httptest.NewServer(h)
		defer srv.Close()
		resp, err := http.Get(srv.URL + SummaryPath)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 || !bytes.Equal(body, want) {
			t.Errorf("status %d, body equal %v", resp.StatusCode, bytes.Equal(body, want))
		}
		checkSecurityHeaders(t, resp.Header)
	})
}

// TestLiveModelsMatchesContract decodes the live /api/models body strictly and compares its
// recursive key set with contract/examples/models.json.
func TestLiveModelsMatchesContract(t *testing.T) {
	res, err := ingest.Dir(corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	h, err := New(summary.Build(res, summary.Options{Input: "../fixtures/claude-code"}))
	if err != nil {
		t.Fatal(err)
	}
	w := do(h, req{method: "GET", path: ModelsPath, host: "127.0.0.1:8080"})
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	dec.DisallowUnknownFields()
	var strict contract.ModelsResponse
	if err := dec.Decode(&strict); err != nil {
		t.Fatalf("strict decode: %v", err)
	}
	example, err := os.ReadFile(filepath.Join("..", "..", "..", "contract", "examples", "models.json"))
	if err != nil {
		t.Fatal(err)
	}
	var liveAny, exAny any
	if err := json.Unmarshal(w.Body.Bytes(), &liveAny); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(example, &exAny); err != nil {
		t.Fatal(err)
	}
	live, ex := strings.Join(contract.KeySet(liveAny), "\n"), strings.Join(contract.KeySet(exAny), "\n")
	if live != ex {
		t.Errorf("key sets differ\nexample:\n%s\n\nlive:\n%s", ex, live)
	}
	var exStrict contract.ModelsResponse
	if err := json.Unmarshal(example, &exStrict); err != nil {
		t.Fatal(err)
	}
	if strict.Definitions != exStrict.Definitions {
		t.Error("definitions differ from the contract example")
	}
	verdicts := map[string]bool{"insufficient_evidence": true, "no_premium_benefit_observed": true, "fewer_corrections_at_premium": true, "no_comparison_possible": true}
	if !verdicts[strict.Advice.Verdict] {
		t.Errorf("verdict %q is not in the contract enum", strict.Advice.Verdict)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// checkNoPromptText fails when any session's first prompt, or any genuine user prompt of 12
// characters or more, appears in body, raw or JSON-escaped. Failures name the session only.
func checkNoPromptText(t *testing.T, res *ingest.Result, body string) {
	t.Helper()
	checked := 0
	for _, s := range res.Sessions {
		texts := []string{}
		if s.HasFirstPrompt {
			texts = append(texts, s.FirstPrompt)
		}
		for i := range s.Events {
			if s.Events[i].Kind == model.KindUserPrompt {
				texts = append(texts, s.Events[i].Text)
			}
		}
		for _, p := range texts {
			p = strings.TrimSpace(p)
			if len(p) < 12 {
				continue
			}
			checked++
			esc, _ := json.Marshal(p)
			if strings.Contains(body, p) || strings.Contains(body, strings.Trim(string(esc), `"`)) {
				t.Errorf("session %s: prompt text found in the response body", s.ID)
			}
		}
	}
	if checked == 0 {
		t.Error("no prompt text was checked")
	}
}

func TestListenAddressLoopbackUnlessBindAny(t *testing.T) {
	cases := []struct {
		addr    string
		bindAny bool
		ok      bool
	}{
		{"127.0.0.1:8080", false, true},
		{"localhost:8080", false, true},
		{"[::1]:8080", false, true},
		{"127.0.0.1:0", false, true},
		{"0.0.0.0:8080", false, false},
		{":8080", false, false},
		{"192.168.1.10:8080", false, false},
		{"0.0.0.0:8080", true, true},
		{":8080", true, true},
		{"127.0.0.1", false, false},
		{"127.0.0.1:http", false, false},
		{"127.0.0.1:70000", true, false},
	}
	for i, c := range cases {
		if err := CheckListen(c.addr, c.bindAny); (err == nil) != c.ok {
			t.Errorf("case %d: CheckListen error %v, want ok=%v", i, err, c.ok)
		}
	}
}
