package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		messages := getLinesChannel(conn)
		for msg := range messages {
			fmt.Printf("read: %s\n", msg)
		}
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	channelOfStrings := make(chan string)

	go func() {
		defer f.Close()
		defer close(channelOfStrings)

		res := make([]byte, 8)
		var currentLine string

		for {
			n, err := f.Read(res)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("can't read from file")
				return
			}
			data := currentLine + string(res[:n])
			parts := strings.Split(data, "\n")
			for i := 0; i < len(parts)-1; i++ {
				channelOfStrings <- parts[i]
			}

			currentLine = parts[len(parts)-1]
		}

		if currentLine != "" {
			channelOfStrings <- currentLine
		}
	}()

	return channelOfStrings
}
