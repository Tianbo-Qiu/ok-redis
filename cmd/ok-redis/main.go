package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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

		default:
			_, _ = fmt.Fprint(conn, "-ERR unknown command\r\n")
		}
	}
}
