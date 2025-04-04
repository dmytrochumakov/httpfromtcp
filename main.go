package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("can't open file")
		return
	}
	defer file.Close()

	res := make([]byte, 8)
	for {
		n, err := file.Read(res)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("can't read from file")
			return
		}
		fmt.Printf("read: %s\n", res[:n])
	}
}
