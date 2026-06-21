package command

import (
	"strconv"

	"github.com/Tianbo-Qiu/ok-redis/internal/resp"
	"github.com/Tianbo-Qiu/ok-redis/internal/store"
)

func cmdPing(s *store.Store, args []string) string {
	return resp.SimpleString("PONG")
}

func cmdSet(s *store.Store, args []string) string {
	s.Set(args[1], args[2])
	return resp.SimpleString("OK")
}

func cmdGet(s *store.Store, args []string) string {
	value, ok := s.Get(args[1])
	if !ok {
		return resp.NilBulk
	}
	return resp.BulkString(value)
}

func cmdDel(s *store.Store, args []string) string {
	deleted := 0
	if s.Del(args[1]) {
		deleted = 1
	}
	return resp.Integer(int64(deleted))
}

func cmdIncr(s *store.Store, args []string) string {
	return incrBy(s, args[1], 1)
}

func cmdDecr(s *store.Store, args []string) string {
	return incrBy(s, args[1], -1)
}

func cmdIncrBy(s *store.Store, args []string) string {
	delta, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.Error("ERR " + store.ErrNotInteger.Error())
	}
	return incrBy(s, args[1], delta)
}

func cmdDecrBy(s *store.Store, args []string) string {
	delta, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return resp.Error("ERR " + store.ErrNotInteger.Error())
	}
	return incrBy(s, args[1], -delta)
}

func cmdExists(s *store.Store, args []string) string {
	var exists int64
	if _, ok := s.Get(args[1]); ok {
		exists = 1
	}
	return resp.Integer(exists)
}

func incrBy(s *store.Store, key string, delta int64) string {
	n, err := s.Incr(key, delta)
	if err != nil {
		return resp.Error("ERR " + err.Error())
	}
	return resp.Integer(n)
}
