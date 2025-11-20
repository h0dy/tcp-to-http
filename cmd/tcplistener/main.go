package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

const PORT = ":42069"

func main() {
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("couldn't set up tcp listener: %v\n", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port%v\n", PORT)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("couldn't accept connection")
		}
		fmt.Println("connection has been accepted")

		for line := range getLinesChannel(conn) {
			fmt.Printf("%v\n", line)
		}
		fmt.Println("connection has been closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	strCh := make(chan string)
	go func() {
		currentLine := ""
		defer f.Close()
		defer close(strCh)

		for {
			buff := make([]byte, 8)
			n, err := f.Read(buff)
			if err != nil {
				// send the remaining text as the last line
				if currentLine != "" {
					strCh <- currentLine
				}
				if errors.Is(err, io.EOF) {
					return
				}
				log.Fatalf("Couldn't read the data into the buffer: %v\n", err)
			}

			str := string(buff[:n])
			parts := strings.Split(str, "\n")

			// iterate over  all the lines except the last
			for i := 0; i < len(parts)-1; i++ {
				//  complete the current line and send it
				completeLine := currentLine + parts[i]
				strCh <- completeLine
				currentLine = "" // reset for next line
			}

			// last line isn't complete, add it to the next "current" line
			currentLine += parts[len(parts)-1]
		}
	}()
	return strCh
}
