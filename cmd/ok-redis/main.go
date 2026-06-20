package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
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
		// first command loop
		// reads until the client sends a newline
		// this is not redis protocol yet
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println("read:", err)
			}
			return
		}

		command := strings.TrimSpace(line)

		switch strings.ToUpper(command) {
		case "PING":
			_, _ = conn.Write([]byte("PONG\r\n"))
		default:
			_, _ = conn.Write([]byte("ERR unknown command\r\n"))
		}
	}
}
