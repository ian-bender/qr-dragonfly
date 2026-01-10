package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"qr-service/internal/store"
)

type errResp struct {
	Error string `json:"error"`
}

func TestQuota_Free_TotalExceeded(t *testing.T) {
	s := store.NewMemoryStore()
	r := NewRouter(Server{Store: s})

	// Free max total = 20
	for i := 0; i < 20; i++ {
		body, _ := json.Marshal(map[string]any{"label": "x", "url": "https://example.com", "active": false})
		req := httptest.NewRequest(http.MethodPost, "/api/qr-codes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Type", "free")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	}

	// 21st should fail
	body, _ := json.Marshal(map[string]any{"label": "x", "url": "https://example.com", "active": false})
	req := httptest.NewRequest(http.MethodPost, "/api/qr-codes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Type", "free")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, w.Code)
	}
	var resp errResp
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "quota_total_exceeded" {
		t.Fatalf("expected quota_total_exceeded, got %q", resp.Error)
	}
}

func TestQuota_Free_ActiveExceededOnActivate(t *testing.T) {
	s := store.NewMemoryStore()
	r := NewRouter(Server{Store: s})

	// Create 5 active (max active for free)
	for i := 0; i < 5; i++ {
		body, _ := json.Marshal(map[string]any{"label": "x", "url": "https://example.com"})
		req := httptest.NewRequest(http.MethodPost, "/api/qr-codes", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-Type", "free")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d", http.StatusCreated, w.Code)
		}
	}

	// Create one inactive (should be allowed since total < 20)
	body, _ := json.Marshal(map[string]any{"label": "inactive", "url": "https://example.com", "active": false})
	createReq := httptest.NewRequest(http.MethodPost, "/api/qr-codes", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("X-User-Type", "free")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, createW.Code)
	}

	// Find its ID
	listReq := httptest.NewRequest(http.MethodGet, "/api/qr-codes", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, listW.Code)
	}
	var items []map[string]any
	_ = json.NewDecoder(listW.Body).Decode(&items)
	var inactiveID string
	for _, it := range items {
		if it["label"] == "inactive" {
			if v, ok := it["id"].(string); ok {
				inactiveID = v
			}
		}
	}
	if inactiveID == "" {
		t.Fatalf("expected to find inactive code id")
	}

	// Attempt to activate (should be blocked at 5 active)
	patchBody, _ := json.Marshal(map[string]any{"active": true})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/qr-codes/"+inactiveID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("X-User-Type", "free")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	if patchW.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, patchW.Code)
	}
	var resp errResp
	_ = json.NewDecoder(patchW.Body).Decode(&resp)
	if resp.Error != "quota_active_exceeded" {
		t.Fatalf("expected quota_active_exceeded, got %q", resp.Error)
	}
}
