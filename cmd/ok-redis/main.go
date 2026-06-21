package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/Tianbo-Qiu/ok-redis/internal/resp"
	"github.com/Tianbo-Qiu/ok-redis/internal/store"
)

func main() {
	ln, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	st := store.New()

	log.Println("ok-redis listening on :6380")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}

		go handleConn(conn, st)
	}
}

func handleConn(conn net.Conn, st *store.Store) {
	defer conn.Close()

	log.Println("client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {
		args, err := resp.ReadCommand(reader)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("read:", err)
			}
			return
		}

		if len(args) == 0 {
			continue
		}

		command := strings.ToUpper(args[0])

		switch command {
		case "PING":
			_, _ = conn.Write([]byte("+PONG\r\n"))

		case "SET":
			if len(args) != 3 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'set' command\r\n")
				continue
			}
			st.Set(args[1], args[2])
			_, _ = conn.Write([]byte("+OK\r\n"))

		case "GET":
			if len(args) != 2 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'get' command\r\n")
				continue
			}
			value, ok := st.Get(args[1])
			if !ok {
				_, _ = conn.Write([]byte("$-1\r\n")) // null bulk string = nil
				continue
			}
			_, _ = fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(value), value)

		case "DEL":
			if len(args) != 2 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'del' command\r\n")
				continue
			}
			deleted := 0
			if st.Del(args[1]) {
				deleted = 1
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", deleted)

		case "INCR":
			if len(args) != 2 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'incr' command\r\n")
				continue
			}
			n, err := st.Incr(args[1], 1)
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-ERR %s\r\n", err)
				continue
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", n)

		case "DECR":
			if len(args) != 2 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'decr' command\r\n")
				continue
			}
			n, err := st.Incr(args[1], -1)
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-ERR %s\r\n", err)
				continue
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", n)

		case "EXISTS":
			if len(args) != 2 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'exists' command\r\n")
				continue
			}
			exists := 0
			if _, ok := st.Get(args[1]); ok {
				exists = 1
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", exists)

		case "INCRBY":
			if len(args) != 3 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'incrby' command\r\n")
				continue
			}
			delta, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				_, _ = fmt.Fprint(conn, "-ERR value is not an integer or out of range\r\n")
				continue
			}
			n, err := st.Incr(args[1], delta)
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-ERR %s\r\n", err)
				continue
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", n)

		case "DECRBY":
			if len(args) != 3 {
				_, _ = fmt.Fprint(conn, "-ERR wrong number of arguments for 'decrby' command\r\n")
				continue
			}
			delta, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				_, _ = fmt.Fprint(conn, "-ERR value is not an integer or out of range\r\n")
				continue
			}
			n, err := st.Incr(args[1], -delta)
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-ERR %s\r\n", err)
				continue
			}
			_, _ = fmt.Fprintf(conn, ":%d\r\n", n)

		default:
			_, _ = fmt.Fprint(conn, "-ERR unknown command\r\n")
		}
	}
}
