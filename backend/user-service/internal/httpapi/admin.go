package httpapi

import "net/http"

func requireAdmin(apiKey string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "admin_disabled"})
			return
		}
		if r.Header.Get("X-Admin-Key") != apiKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}
