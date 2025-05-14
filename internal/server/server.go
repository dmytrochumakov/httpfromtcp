package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type ServerState int

const (
	initialized ServerState = iota
)

type Server struct {
	state    ServerState
	listener net.Listener
	wg       sync.WaitGroup
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := Server{
		state:    initialized,
		listener: listener,
	}
	server.closed.Store(false)

	server.wg.Add(1)
	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	if err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *Server) listen() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 12\r\n" + // Length of "Hello World!"
		"\r\n" +
		"Hello World!"

	_, _ = io.WriteString(conn, response)
}
