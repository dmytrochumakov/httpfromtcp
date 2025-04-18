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
	defer file.Close()

	res := make([]byte, 8)
	var currentLine string

	for {
		n, err := file.Read(res)
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
			fmt.Printf("read: %s\n", parts[i])
		}

		currentLine = parts[len(parts)-1]
	}

	if currentLine != "" {
		fmt.Println(currentLine)
	}
}
