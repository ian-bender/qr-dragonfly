package store

import (
	"testing"
	"time"
)

func TestMemoryStore_RecordAndStats(t *testing.T) {
	s := NewMemoryStore()

	if _, err := s.GetStats("abc"); err == nil {
		t.Fatalf("expected not found")
	}

	e := ClickEvent{QrCodeID: "abc", At: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), Country: "US"}
	if err := s.RecordClick(e); err != nil {
		t.Fatalf("record: %v", err)
	}

	st, err := s.GetStats("abc")
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if st.Total != 1 {
		t.Fatalf("expected total=1, got %d", st.Total)
	}
	if st.LastCountry != "US" {
		t.Fatalf("expected lastCountry=US, got %q", st.LastCountry)
	}
	if st.LastAtIso == "" {
		t.Fatalf("expected lastAtIso")
	}
}
