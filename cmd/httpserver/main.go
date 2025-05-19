package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmytrochumakov/httpfromtcp/internal/request"
	"github.com/dmytrochumakov/httpfromtcp/internal/response"
	"github.com/dmytrochumakov/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		hErr := server.HandlerError{
			StatusCode: response.BadRequest,
			Message:    "Your problem is not my problem\n",
		}
		hErr.Write(w)
		return &hErr
	}

	if req.RequestLine.RequestTarget == "/myproblem" {
		hErr := server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "Woopsie, my bad\n",
		}
		hErr.Write(w)
		return &hErr
	}

	w.Write([]byte("All good, frfr\n"))
	return nil
}
