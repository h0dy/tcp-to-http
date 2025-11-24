package main

import (
	"fmt"
	"log"
	"net"

	"github.com/h0dy/tcp-to-http/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("couldn't set up tcp listener: %v\n", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port%v\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("couldn't accept connection")
		}
		fmt.Printf("connection has been accepted from: %v\n", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error parsing request: %v\n", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for h, v := range req.Headers {
			fmt.Printf("- %s: %s\n", h, v)
		}
		fmt.Printf("Body:\n%s", string(req.Body))
	}
}
