package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/h0dy/tcp-to-http/internal/request"
	"github.com/h0dy/tcp-to-http/internal/response"
)

// Server is an HTTP 1.1 server
type Server struct {
	closed   atomic.Bool
	listener net.Listener
	port     int
	handler  Handler
}

// Handler processes an HTTP request
type Handler func(w *response.Writer, req *request.Request)

// Serve starts a TCP server on the given port with the provided handler
func Serve(port int, handler Handler) (*Server, error) {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, fmt.Errorf("error: failed to bind to port: %v", port)
	}
	server := &Server{
		port:     port,
		listener: listen,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

// s.Close() closes the connection of the server
func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

// s.listen() starts listing and accepts incoming requests
func (s *Server) listen() {
	fmt.Printf("serving on port: %v\n", s.port)

	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Fatalf("error: couldn't accept connection: %v", err.Error())
			continue
		}
		go s.handle(conn)
	}
}

// handle handles incoming connection
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.ClientError)
		body := fmt.Appendf(nil, "error parsing request: %v", err)
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
	}

	s.handler(w, req)
}
