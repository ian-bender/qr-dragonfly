package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadCookie_Missing(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := readCookie(r, "missing")
	if ok {
		t.Fatalf("expected ok=false")
	}
}
