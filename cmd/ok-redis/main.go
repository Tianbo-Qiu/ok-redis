package main

import (
	"log"
	"net"
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
}
