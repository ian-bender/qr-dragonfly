package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAdmin_Disabled(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)

	h := requireAdmin("", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h(w, r)

	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected %d, got %d", http.StatusNotImplemented, w.Code)
	}
}

func TestRequireAdmin_BadKey(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	r.Header.Set("X-Admin-Key", "wrong")

	h := requireAdmin("right", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
