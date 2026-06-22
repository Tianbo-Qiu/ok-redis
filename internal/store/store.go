// Package store provides an in-memory, concurrency-safe key-value store
package store

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"
)

var ErrNotInteger = errors.New("value is not an integer or out of range")

// entry is a stored value plus an optional expiration time.
// a zero expireAt (time.Time{}) means the key never expires.
type entry struct {
	value    string
	expireAt time.Time
}

func (e entry) expired() bool {
	return !e.expireAt.IsZero() && time.Now().After(e.expireAt)
}

type Store struct {
	mu   sync.Mutex
	data map[string]entry
}

func New() *Store {
	return &Store{
		data: make(map[string]entry),
	}
}

// getLive returns the live entry for key, deleting it first if it has expired.
// the caller MUST already hold s.mu
func (s *Store) getLive(key string) (entry, bool) {
	e, ok := s.data[key]
	if !ok {
		return entry{}, false
	}
	if e.expired() {
		delete(s.data, key)
		return entry{}, false
	}
	return e, true
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = entry{value: value}
}

// Get returns the value for key and whether it was present.
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.getLive(key)
	if !ok {
		return "", false
	}
	return e.value, true
}

// Del removes key and reports whether it was present.
func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.getLive(key)
	delete(s.data, key)
	return ok
}

// Incr adds delta to the integer value at key and returns the new value.
// A missing key is treated as 0. If the current value isn't a valid integer,
// it returns ErrNotInteger and leaves the value unchanged.
func (s *Store) Incr(key string, delta int64) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var n int64
	e, ok := s.getLive(key)
	if ok {
		parsed, err := strconv.ParseInt(e.value, 10, 64)
		if err != nil {
			return 0, ErrNotInteger
		}
		n = parsed
	}

	n += delta
	e.value = strconv.FormatInt(n, 10)
	s.data[key] = e
	return n, nil
}

// Expire sets a time-to-live on key.
// It reports whether the key existed and whether the expiry
// was actually applied.
func (s *Store) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.getLive(key)
	if !ok {
		return false
	}

	e.expireAt = time.Now().Add(ttl)
	s.data[key] = e
	return true
}

// TTL reports the remaining time-to-live for key.
func (s *Store) TTL(key string) (ttl time.Duration, exists bool, hasExpiry bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.getLive(key)
	if !ok {
		return 0, false, false
	}
	if e.expireAt.IsZero() {
		return 0, true, false
	}
	return time.Until(e.expireAt), true, true
}

// Persist removes any expiry from key.
// It reports whether an expiry was removed.
func (s *Store) Persist(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.getLive(key)
	if !ok || e.expireAt.IsZero() {
		return false
	}
	e.expireAt = time.Time{}
	s.data[key] = e
	return true
}

// TODO: sampling instead of full scan
func (s *Store) sweepExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	for k, e := range s.data {
		if e.expired() {
			delete(s.data, k)
			removed++
		}
	}
	return removed
}

// StartExpiryWorker launches a background goroutine that sweeps expired keys
// every interval, until ctx is cancelled.
func (s *Store) StartExpiryWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.sweepExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}
