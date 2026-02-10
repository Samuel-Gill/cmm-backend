package logging

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	authmw "matchmaking-service/auth/middleware"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

type logRecord map[string]any

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = randomRequestID()
			}
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			requestID, _ := r.Context().Value(RequestIDKey).(string)

			writeLog(logRecord{
				"level":      "info",
				"event":      "request_started",
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
			})

			next.ServeHTTP(rec, r)

			userID := rec.Header().Get("X-User-ID")
			finished := logRecord{
				"level":         "info",
				"event":         "request_finished",
				"request_id":    requestID,
				"method":        r.Method,
				"path":          r.URL.Path,
				"status":        rec.status,
				"duration_ms":   time.Since(start).Milliseconds(),
				"authenticated": userID != "",
			}
			if userID != "" {
				finished["user_id"] = userID
			}
			writeLog(finished)

			if rec.status >= 400 {
				errLog := logRecord{
					"level":      "error",
					"event":      "request_error",
					"request_id": requestID,
					"method":     r.Method,
					"path":       r.URL.Path,
					"status":     rec.status,
				}
				if userID != "" {
					errLog["user_id"] = userID
				}
				if rec.status >= 500 {
					errLog["stack_trace"] = string(debug.Stack())
				}
				writeLog(errLog)
			}
		})
	}
}

func Recoverer() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID, _ := r.Context().Value(RequestIDKey).(string)
					userID, _ := r.Context().Value(authmw.UserIDKey).(string)
					errLog := logRecord{
						"level":       "error",
						"event":       "panic_recovered",
						"request_id":  requestID,
						"method":      r.Method,
						"path":        r.URL.Path,
						"panic":       rec,
						"stack_trace": string(debug.Stack()),
					}
					if userID != "" {
						errLog["user_id"] = userID
					}
					writeLog(errLog)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func writeLog(entry logRecord) {
	entry["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}

func randomRequestID() string {
	return uuid.NewString()
}
