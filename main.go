package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("error", "error", err)
	}

	lines := getLinesChannel(f)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
