// Package store provides an in-memory, concurrency-safe key-value store
package store

import (
	"errors"
	"strconv"
	"sync"
)

var ErrNotInteger = errors.New("value is not an integer or out of range")

type Store struct {
	mu   sync.Mutex
	data map[string]string
}

func New() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get returns the value for key and whether it was present.
func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	return value, ok
}

// Del removes key and reports whether it was present.
func (s *Store) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.data[key]
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
	if cur, ok := s.data[key]; ok {
		parsed, err := strconv.ParseInt(cur, 10, 64)
		if err != nil {
			return 0, ErrNotInteger
		}
		n = parsed
	}

	n += delta
	s.data[key] = strconv.FormatInt(n, 10)
	return n, nil
}
