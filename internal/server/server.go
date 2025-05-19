package server

import (
	"bytes"
	"fmt"
	"io"
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

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := HandlerError{
			StatusCode: response.BadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	b := buf.Bytes()

	response.WriteStatusLine(conn, response.OK)
	defaultHeaders := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, defaultHeaders)
	conn.Write(b)
	return
}

func (e *HandlerError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func (he *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}
