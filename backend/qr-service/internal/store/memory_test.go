package store

import "testing"

func TestMemoryStore_CRUD(t *testing.T) {

	s := NewMemoryStore()

	created, err := s.Create(CreateInput{Label: "A", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected id")
	}
	if created.URL != "https://example.com" {
		t.Fatalf("expected url")
	}
	if !created.Active {
		t.Fatalf("expected active=true by default")
	}

	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("expected same id")
	}

	newLabel := "B"
	updated, err := s.Update(created.ID, UpdateInput{Label: &newLabel})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Label != "B" {
		t.Fatalf("expected updated label")
	}
	if !updated.Active {
		t.Fatalf("expected active to remain true")
	}

	deactivate := false
	updated2, err := s.Update(created.ID, UpdateInput{Active: &deactivate})
	if err != nil {
		t.Fatalf("update active: %v", err)
	}
	if updated2.Active {
		t.Fatalf("expected active=false after update")
	}

	list := s.List()
	if len(list) != 1 {
		t.Fatalf("expected list size 1")
	}

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.Get(created.ID); err == nil {
		t.Fatalf("expected not found")
	}
}
