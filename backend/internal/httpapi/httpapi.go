// Package httpapi serves the summary over loopback HTTP.
//
// Boundary, in check order: the Host header must name a loopback host (127.0.0.1, localhost
// or [::1], any port) or the request is 403; an Origin header, when present, must equal
// scheme://Host or 403; Sec-Fetch-Site: cross-site is 403; methods other than GET and HEAD
// are 405 with Allow: GET, HEAD; unknown paths are 404. Every error body is the contract's
// ErrorResponse. Every response carries the five security headers below.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cristiangirlea/classify-sessions/backend/internal/contract"
)

// Routes. /api/models serves the summary's model_choice object, byte for byte.
const (
	SummaryPath = "/api/summary"
	ModelsPath  = "/api/models"
)

// Allow is the Allow header of a 405 response.
const Allow = "GET, HEAD"

// SecurityHeaders are set on every response, errors included.
var SecurityHeaders = map[string]string{
	"Cache-Control":                "no-store",
	"X-Content-Type-Options":       "nosniff",
	"Referrer-Policy":              "no-referrer",
	"Content-Security-Policy":      "default-src 'none'",
	"Cross-Origin-Resource-Policy": "same-origin",
}

type handler struct {
	routes map[string][]byte
}

// New returns the API handler. Both bodies are marshalled once, here.
func New(s contract.SummaryResponse) (http.Handler, error) {
	sb, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	mb, err := json.Marshal(s.ModelChoice)
	if err != nil {
		return nil, err
	}
	return &handler{routes: map[string][]byte{SummaryPath: sb, ModelsPath: mb}}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range SecurityHeaders {
		w.Header().Set(k, v)
	}
	if !IsLoopbackHost(r.Host) {
		writeError(w, r, http.StatusForbidden, "forbidden", "host must be a loopback address")
		return
	}
	if origin, ok := r.Header["Origin"]; ok {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		if len(origin) != 1 || origin[0] != scheme+"://"+r.Host {
			writeError(w, r, http.StatusForbidden, "forbidden", "cross-origin requests are not allowed")
			return
		}
	}
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
		writeError(w, r, http.StatusForbidden, "forbidden", "cross-site requests are not allowed")
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", Allow)
		writeError(w, r, http.StatusMethodNotAllowed, "method_not_allowed", "only GET and HEAD are allowed")
		return
	}
	body, ok := h.routes[r.URL.Path]
	if !ok {
		writeError(w, r, http.StatusNotFound, "not_found", "no such path")
		return
	}
	writeJSON(w, r, http.StatusOK, body)
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		w.Write(body)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	body, _ := json.Marshal(contract.ErrorResponse{Error: contract.ErrorBody{Code: code, Message: message}})
	writeJSON(w, r, status, body)
}

// IsLoopbackHost reports whether a Host header value names 127.0.0.1, localhost or [::1],
// with or without a port.
func IsLoopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}
	switch strings.ToLower(host) {
	case "127.0.0.1", "localhost", "::1":
		return true
	}
	return false
}

// CheckListen accepts a listen address on a loopback host; any host is accepted when bindAny
// is set. An empty host (":8080") means every interface and needs bindAny.
func CheckListen(addr string, bindAny bool) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return errors.New("listen address must be host:port")
	}
	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return errors.New("listen port must be a number from 0 to 65535")
	}
	if bindAny {
		return nil
	}
	if host == "" || !IsLoopbackHost(host) {
		return errors.New("listen address must be loopback (127.0.0.1, localhost or [::1]) unless --bind-any is given")
	}
	return nil
}

// Serve serves h on ln until ctx is done, then shuts down gracefully.
func Serve(ctx context.Context, ln net.Listener, h http.Handler) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			return err
		}
		if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}
