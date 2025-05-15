package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/dmytrochumakov/httpfromtcp/internal/response"
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

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		fmt.Printf("unable to write status line, %s\n", err)
		return
	}

	defaultHeaders := response.GetDefaultHeaders(0)

	response.WriteHeaders(conn, defaultHeaders)
}
