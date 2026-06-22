package command

import (
	"testing"

	"github.com/Tianbo-Qiu/ok-redis/internal/store"
)

func TestDispatch(t *testing.T) {
	s := store.New()
	s.Set("name", "alice")
	s.Set("word", "hello")
	s.Set("n1", "10")
	s.Set("n2", "10")
	s.Set("n3", "10")
	s.Set("n4", "10")

	tests := []struct {
		name string
		args []string
		want string
	}{
		{"ping", []string{"PING"}, "+PONG\r\n"},
		{"ping lowercase", []string{"ping"}, "+PONG\r\n"},
		{"set", []string{"SET", "city", "paris"}, "+OK\r\n"},
		{"get hit", []string{"GET", "name"}, "$5\r\nalice\r\n"},
		{"get miss", []string{"GET", "nope"}, "$-1\r\n"},
		{"incr", []string{"INCR", "n1"}, ":11\r\n"},
		{"decr", []string{"DECR", "n2"}, ":9\r\n"},
		{"incrby", []string{"INCRBY", "n3", "5"}, ":15\r\n"},
		{"decrby", []string{"DECRBY", "n4", "3"}, ":7\r\n"},
		{"incr non-integer value", []string{"INCR", "word"}, "-ERR value is not an integer or out of range\r\n"},
		{"incrby non-integer amount", []string{"INCRBY", "n1", "abc"}, "-ERR value is not an integer or out of range\r\n"},
		{"exists no", []string{"EXISTS", "nope"}, ":0\r\n"},
		{"unknown command", []string{"BOGUS"}, "-ERR unknown command\r\n"},
		{"wrong arg count", []string{"GET"}, "-ERR wrong number of arguments for 'get' command\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Dispatch(s, tt.args)
			if got != tt.want {
				t.Errorf("Dispatch(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestDispatchExpiry(t *testing.T) {
	s := store.New()
	s.Set("k", "v")

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"ttl no expiry", []string{"TTL", "k"}, ":-1\r\n"},
		{"ttl missing key", []string{"TTL", "missing"}, ":-2\r\n"},
		{"expire missing key", []string{"EXPIRE", "missing", "10"}, ":0\r\n"},
		{"expire ok", []string{"EXPIRE", "k", "100"}, ":1\r\n"},
		{"ttl after expire", []string{"TTL", "k"}, ":100\r\n"},
		{"persist ok", []string{"PERSIST", "k"}, ":1\r\n"},
		{"ttl after persist", []string{"TTL", "k"}, ":-1\r\n"},
		{"expire bad seconds", []string{"EXPIRE", "k", "abc"}, "-ERR value is not an integer or out of range\r\n"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dispatch(s, tt.args); got != tt.want {
				t.Errorf("Dispatch(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
