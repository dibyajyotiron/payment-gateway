package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs details of each incoming HTTP request and response status.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the start time
		startTime := time.Now()

		// Create a response writer to capture the status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler in the chain
		next.ServeHTTP(lrw, r)

		// Log the details
		// Ideally in production, we will have our own logger package with json format
		log.Printf(
			"Method: %s | Path: %s | Status: %d | Duration: %s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			time.Since(startTime),
		)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader Overrides WriteHeader to capture the status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
