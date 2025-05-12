package main

import (
	"fmt"
	"log"
	"net"

	"github.com/dmytrochumakov/httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Println("can't open file")
		return
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("connection has been accepted")
		requestLine, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", requestLine.RequestLine.Method)
		fmt.Printf("- Target: %s\n", requestLine.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", requestLine.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range requestLine.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

	}

}
