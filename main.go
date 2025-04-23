package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("can't open file")
		return
	}

	messages := getLinesChannel(file)
	for msg := range messages {
		fmt.Printf("read: %s\n", msg)
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
