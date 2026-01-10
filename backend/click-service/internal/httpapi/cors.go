package httpapi

import (
	"net/http"
	"strings"
)

type CorsOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

func NewCorsMiddleware(opts CorsOptions) func(http.Handler) http.Handler {
	allowedOrigins := normalizeList(opts.AllowedOrigins)
	allowedMethods := opts.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	}
	allowedHeaders := opts.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = []string{"Content-Type", "Authorization"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if opts.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Add("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				if origin != "" && originAllowed(origin, allowedOrigins) {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
					w.WriteHeader(http.StatusNoContent)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func normalizeList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func originAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, a := range allowed {
		if a == "*" {
			return true
		}
		if a == origin {
			return true
		}
	}
	return false
}
