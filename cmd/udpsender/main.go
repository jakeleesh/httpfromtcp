package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("error", "error", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println(">")
		s, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error", "error", err)
		}

		_, err = conn.Write([]byte(s))
		if err != nil {
			log.Fatal("error", "error", err)
		}
	}
}
