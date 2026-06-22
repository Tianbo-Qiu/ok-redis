package store

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"
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

func TestExpiredKeyIsInvisible(t *testing.T) {
	s := New()

	s.data["temp"] = entry{value: "x", expireAt: time.Now().Add(-time.Second)}

	if _, ok := s.Get("temp"); ok {
		t.Errorf("expected expired key to be invisible to Get")
	}
}

func TestExpireAndTTL(t *testing.T) {
	s := New()

	if _, exists, _ := s.TTL("nope"); exists {
		t.Errorf("expected missing key to report exists=false")
	}

	s.Set("name", "alice")
	if _, exists, hasExpiry := s.TTL("name"); !exists || hasExpiry {
		t.Errorf("got exists=%v want true, hasExpiry=%v want false", exists, hasExpiry)
	}

	if s.Expire("nope", time.Minute) {
		t.Errorf("expected Expire on missing key to return false")
	}

	if !s.Expire("name", 100*time.Second) {
		t.Fatalf("expected Expire to succeed")
	}

	ttl, exists, hasExpiry := s.TTL("name")
	if !exists || !hasExpiry {
		t.Fatalf("got exists=%v hasExpiry=%v, want both true", exists, hasExpiry)
	}
	if ttl <= 99*time.Second || ttl > 100*time.Second {
		t.Errorf("ttl = %v, want ~100s", ttl)
	}
	// Persist removes the expiry
	if !s.Persist("name") {
		t.Errorf("expected Persist to remove an expiry")
	}
	if _, _, hasExpiry := s.TTL("name"); hasExpiry {
		t.Errorf("expected no expiry after Persist")
	}
}

func TestExpirePastDeletesOnAccess(t *testing.T) {
	s := New()
	s.Set("temp", "x")
	s.Expire("temp", -time.Second) // already in the past
	if _, ok := s.Get("temp"); ok {
		t.Errorf("expected key with a past expiry to be gone")
	}
}

func TestSweepExpired(t *testing.T) {
	s := New()
	s.Set("keep", "1")
	// plant an already-expired entry directly (same-package test):
	s.data["gone"] = entry{value: "2", expireAt: time.Now().Add(-time.Second)}

	if removed := s.sweepExpired(); removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
	if _, ok := s.data["gone"]; ok {
		t.Errorf("expected expired key to be swept")
	}
	if _, ok := s.data["keep"]; !ok {
		t.Errorf("expected live key to survive")
	}
}
