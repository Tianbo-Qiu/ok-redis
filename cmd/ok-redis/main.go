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
)

func main() {
	ln, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("ok-redis listening on :6380")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept:", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
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
		default:
			_, _ = fmt.Fprintf(conn, "-ERR unknown command '%s'\r\n", args[0])
		}
	}
}
