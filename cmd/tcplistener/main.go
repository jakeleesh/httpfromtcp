package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/jakeleesh/httpfromtcp/internal/request"
)

// ReadCloser an interface with Read and Close
func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""
		// Reading 8 at a time
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			// Slicing things out
			data = data[:n]
			// Seeing if there's a \n
			// If we found an index
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				// Starting at i + 1 because don't want new line
				data = data[i+1:]
				out <- str
				// Resetting the string
				str = ""
			}

			// Adding the rest of the data
			str += string(data)
		}

		if len(str) != 0 {
			out <- str
		}
	}()

	return out
}

func main() {
	// Create a Server
	// Going to get TCP connections coming in
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}

		// // Can read 8 bytes from a connection/file
		// // Same principal, working with ordered data
		// for line := range getLinesChannel(conn) {
		// 	fmt.Printf("read: %s\n", line)
		// }

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
		r.Headers.ForEach(func(n, v string) {
			fmt.Printf("- %s: %s\n", n, v)
		})
	}
}
