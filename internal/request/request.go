package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jakeleesh/httpfromtcp/internal/headers"
)

type parserState string

const (
	StateInit    parserState = "initialized"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateHeaders parserState = "headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	State       parserState
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("request in error state")
var SEPERATOR = []byte("\r\n")

// Return RequestLine, rest of the string it doesn't know about
func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)
	// Haven't found first found new line yet
	// No SEPERATOR
	if idx == -1 {
		// Return nil, there's no error
		// Read 0 bytes
		return nil, 0, nil
	}

	startLine := b[:idx]
	// Skipping seperator, don't include
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))
	// If length does not equal 3
	// Meanining don't have method, path, HTTP Protocol
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	// State machine that keeps on running
	// Could get large piece of data that contains several state transitions
outer:
	for {
		currentData := data[read:]

		switch r.State {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE

		case StateInit:
			// Pass in data starting at read
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				// Error, didn't read anything
				r.State = StateError
				return 0, err
			}

			// Haven't read anything
			// Unable to move forward, keep on parsing
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			// Go to next state
			r.State = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}

			// Couldn't read anything
			if n == 0 {
				// Needs to return already read data
				break outer
			}

			// n is how much we have parsed
			read += n

			if done {
				r.State = StateDone
			}

		case StateDone:
			break outer

		default:
			panic("somehow we have programmed poorly")
		}
	}

	return read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone || r.State == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// Make buffer for reading
	// NOTE: buffer could get overrun
	buf := make([]byte, 1024)
	bufLen := 0
	// Instead of reading everything at once
	// Keep on reading while the request is not done
	for !request.done() {
		// Read out the buffer
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		// Read from offset of readN up until the length
		// Move all that to beginning
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
