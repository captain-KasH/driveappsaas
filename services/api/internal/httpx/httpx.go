// Package httpx holds the API's JSON, error and middleware conventions.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Error is a client-facing error. Anything else becomes a 500 with no detail.
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func Errorf(status int, code, format string, args ...any) *Error {
	return &Error{Status: status, Code: code, Message: fmt.Sprintf(format, args...)}
}

var (
	ErrUnauthenticated = &Error{http.StatusUnauthorized, "unauthenticated", "sign in required"}
	ErrForbidden       = &Error{http.StatusForbidden, "forbidden", "you do not have permission for this action"}
	ErrNotFound        = &Error{http.StatusNotFound, "not_found", "resource not found"}
)

type ctxKey int

const requestIDKey ctxKey = 0

func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// Handler lets endpoints return errors instead of writing them by hand.
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(w, r)
	if err == nil {
		return
	}
	var e *Error
	if !errors.As(err, &e) {
		slog.ErrorContext(r.Context(), "unhandled error", "err", err, "request_id", RequestID(r.Context()))
		e = &Error{http.StatusInternalServerError, "internal", "something went wrong"}
	}
	WriteJSON(w, e.Status, map[string]any{"error": map[string]string{
		"code": e.Code, "message": e.Message, "requestId": RequestID(r.Context()),
	}})
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// DecodeJSON rejects unknown fields and bodies over 1 MiB.
func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return Errorf(http.StatusBadRequest, "invalid_json", "request body is not valid JSON: %v", err)
	}
	return nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) { s.status = code; s.ResponseWriter.WriteHeader(code) }

// Middleware assigns a request ID, recovers panics and logs one line per request.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey, id))
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		start := time.Now()
		defer func() {
			if p := recover(); p != nil {
				slog.Error("panic", "panic", p, "request_id", id)
				Handler(func(http.ResponseWriter, *http.Request) error { return fmt.Errorf("panic: %v", p) }).ServeHTTP(rec, r)
			}
			slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(), "request_id", id)
		}()
		next.ServeHTTP(rec, r)
	})
}
