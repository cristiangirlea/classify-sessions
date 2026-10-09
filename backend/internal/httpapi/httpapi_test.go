package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		for _, p := range []string{"/", "/api", "/api/summary/", "/api/models", "/api/unknown", "/index.html"} {
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
			for _, p := range []string{SummaryPath, "/nope"} {
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
