package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/Tianbo-Qiu/ok-redis/internal/command"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st.StartExpiryWorker(ctx, time.Second)

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

		reply := command.Dispatch(st, args)
		if reply == "" {
			continue
		}

		if _, err := io.WriteString(conn, reply); err != nil {
			log.Println("write:", err)
			return
		}
	}
}
