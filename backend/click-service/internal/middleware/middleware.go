package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
)

// EnforceJSONHandler sets the response Content-Type and, for requests with a body,
// enforces Content-Type: application/json.
func EnforceJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		ct := r.Header.Get("Content-Type")
		if ct != "" && !strings.HasPrefix(ct, "application/json") {
			http.Error(w, `{"error":"content_type_must_be_json"}`, http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ExposeResponseHeaders configures CORS Access-Control-Expose-Headers for clients.
func ExposeResponseHeaders(next http.Handler, headers ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(headers) == 0 {
			headers = []string{"X-Request-Id"}
		}
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(headers, ", "))
		next.ServeHTTP(w, r)
	})
}

// Recoverer converts panics into a 500 JSON response.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestID ensures each request has an X-Request-Id header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-Id"))
		if id == "" {
			id = newHexID(8)
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r)
	})
}

func newHexID(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
