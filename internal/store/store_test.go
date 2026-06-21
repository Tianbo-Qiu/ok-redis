package store

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

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

func TestIncr(t *testing.T) {
	s := New()

	n, err := s.Incr("counter", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 1 {
		t.Errorf("got %d, want 1", n)
	}

	n, _ = s.Incr("counter", 5)
	if n != 6 {
		t.Errorf("got %d, want 6", n)
	}

	s.Set("word", "hello")
	if _, err := s.Incr("word", 1); !errors.Is(err, ErrNotInteger) {
		t.Errorf("expected ErrNotInteger, got %v", err)
	}
}

func TestIncrConcurrent(t *testing.T) {
	s := New()

	const goroutines = 100
	const perGoroutine = 100

	var wg sync.WaitGroup
	for range goroutines {
		wg.Go(func() {
			for range perGoroutine {
				if _, err := s.Incr("counter", 1); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
	wg.Wait()

	got, _ := s.Get("counter")
	want := strconv.Itoa(goroutines * perGoroutine)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
