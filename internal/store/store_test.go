package store

import "testing"

func TestGetSet(t *testing.T) {
	s := New()

	if _, ok := s.Get("name"); ok {
		t.Fatalf("expected key to be missing, but it was present")
	}

	s.Set("name", "alice")

	got, ok := s.Get("name")
	if !ok {
		t.Fatalf("expected key to be present after Set")
	}
	if got != "alice" {
		t.Errorf("got %q, want %q", got, "alice")
	}
}

func TestDel(t *testing.T) {
	s := New()
	s.Set("name", "alice")

	if existed := s.Del("name"); !existed {
		t.Errorf("expected Del to report the key existed")
	}
	if _, ok := s.Get("name"); ok {
		t.Errorf("expected key to be gone after Del")
	}
	if existed := s.Del("missing"); existed {
		t.Errorf("expected Del on a missing key to report false")
	}
}
