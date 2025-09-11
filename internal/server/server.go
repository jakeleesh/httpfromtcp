package server

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/jakeleesh/httpfromtcp/internal/request"
	"github.com/jakeleesh/httpfromtcp/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	Closed  bool
	Handler Handler
}

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	headers := response.GetDefaultHeaders(0)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		response.WriteStatusLine(conn, response.StatusBadRequest)
		response.WriteHeaders(conn, headers)
		return
	}

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.Handler(writer, r)

	var body []byte = nil
	var status response.StatusCode = response.StatusOk
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
		// return
	} else {
		body = writer.Bytes()
	}

	// Reset Content-length
	headers.Replace("Content-length", fmt.Sprintf("%d", len(body)))

	// Write status line, headers, and then body
	response.WriteStatusLine(conn, status)
	response.WriteHeaders(conn, headers)
	conn.Write(body)
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.Closed {
			return
		}

		if err != nil {
			return
		}
		// Handle multiple requests
		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		Closed:  false,
		Handler: handler,
	}
	// Run the server in the background
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
