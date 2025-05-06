package middlewares

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// statusResponseWriter wraps http.ResponseWriter to capture status codes
// and ensures a default of 200 if WriteHeader is not explicitly called.
type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write ensures that a status code of 200 is captured if WriteHeader
// was never called, then delegates to the underlying Write.
func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Logging middleware reads and logs request payloads, then wraps
// the ResponseWriter to capture and log the response status code.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and log request body without consuming r.Body permanently
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("{04} - [Logging] erro ao ler body: %v", err)
		} else {
			log.Printf("{05} - [Logging] payload recebido: %s", string(bodyBytes))
		}
		// Reset r.Body for further handlers
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Wrap the ResponseWriter to capture status code
		srw := &statusResponseWriter{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(srw, r)

		// Final log: method, URI, remote address, status code, and duration
		log.Printf(
			"%s %s %s %d %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			srw.statusCode,
			time.Since(start),
		)
	})
}
