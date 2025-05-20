package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/dmytrochumakov/httpfromtcp/internal/request"
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
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	server := Server{
		state:    initialized,
		listener: listener,
		handler:  handler,
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
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.BadRequest)
		body := "error parsing request"
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody([]byte(body))
		return
	}

	s.handler(w, req)
}
