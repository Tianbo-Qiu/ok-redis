// Package command implements the command registry and dispatcher.
package command

import (
	"fmt"
	"strings"

	"github.com/Tianbo-Qiu/ok-redis/internal/resp"
	"github.com/Tianbo-Qiu/ok-redis/internal/store"
)

// Handler executes one command and returns the RESP-encoded reply.
type Handler func(s *store.Store, args []string) string

type command struct {
	arity   int
	handler Handler
}

var registry = map[string]command{
	"PING":   {arity: 1, handler: cmdPing},
	"SET":    {arity: 3, handler: cmdSet},
	"GET":    {arity: 2, handler: cmdGet},
	"DEL":    {arity: 2, handler: cmdDel},
	"INCR":   {arity: 2, handler: cmdIncr},
	"DECR":   {arity: 2, handler: cmdDecr},
	"INCRBY": {arity: 3, handler: cmdIncrBy},
	"DECRBY": {arity: 3, handler: cmdDecrBy},
	"EXISTS": {arity: 2, handler: cmdExists},
}

func Dispatch(s *store.Store, args []string) string {
	if len(args) == 0 {
		return ""
	}

	name := strings.ToUpper(args[0])
	cmd, ok := registry[name]
	if !ok {
		return resp.Error("ERR unknown command")
	}

	if len(args) != cmd.arity {
		return resp.Error(fmt.Sprintf("ERR wrong number of arguments for '%s' command", strings.ToLower(name)))
	}

	return cmd.handler(s, args)
}
