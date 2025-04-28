package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	network := "udp"
	addr, err := net.ResolveUDPAddr(network, "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP(network, nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal(err)
		}
	}

}
