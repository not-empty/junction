package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/not-empty/bridge/request"
)

type responseTracker struct {
	http.ResponseWriter
	wroteHeader bool
}

func (t *responseTracker) WriteHeader(statusCode int) {
	t.wroteHeader = true
	t.ResponseWriter.WriteHeader(statusCode)
}

func (t *responseTracker) Write(content []byte) (int, error) {
	t.wroteHeader = true
	return t.ResponseWriter.Write(content)
}

func (t *responseTracker) Unwrap() http.ResponseWriter {
	return t.ResponseWriter
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracker := &responseTracker{ResponseWriter: w}

		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			slog.ErrorContext(r.Context(), "panic recovered",
				"panic", rec,
				"method", r.Method,
				"path", r.URL.Path,
				"stack", string(debug.Stack()),
			)

			if tracker.wroteHeader {
				panic(http.ErrAbortHandler)
			}

			request.ResponseInternalError(tracker)
		}()

		next.ServeHTTP(tracker, r)
	})
}
