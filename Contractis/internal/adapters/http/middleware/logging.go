package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging es un middleware que registra las peticiones HTTP con más detalle
func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		rw := newResponseWriter(w)

		// Log request start
		log.Printf("→ [%s] %s %s | User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.Header.Get("User-Agent"))

		// Call next handler
		next(rw, r)

		// Log request completion with status and duration
		duration := time.Since(start)
		statusEmoji := getStatusEmoji(rw.statusCode)
		log.Printf("← [%s] %s | Status: %d %s | Duration: %v", r.Method, r.URL.Path, rw.statusCode, statusEmoji, duration)
	}
}

// getStatusEmoji returns an emoji based on HTTP status code
func getStatusEmoji(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "✅"
	case statusCode >= 300 && statusCode < 400:
		return "↪️"
	case statusCode >= 400 && statusCode < 500:
		return "⚠️"
	case statusCode >= 500:
		return "❌"
	default:
		return "❓"
	}
}
