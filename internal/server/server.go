package server

import (
	"fmt"
	"io"
	"net"

	"github.com/jakeleesh/httpfromtcp/internal/response"
)

type Server struct {
	Closed bool
}

func runConnection(_s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	headers := response.GetDefaultHeaders(0)
	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, headers)
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

func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{Closed: false}
	// Run the server in the background
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.Closed = true
	return nil
}
