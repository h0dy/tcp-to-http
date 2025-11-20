package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const PORT = 42069

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%v", PORT))
	if err != nil {
		log.Fatalf("Couldn't set up UDP address: %v\n", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Couldn't dial up: %v\n", err)
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("error in ReadString: %v\n", err)
		}
		_, err = udpConn.Write([]byte(line))
		if err != nil {
			log.Fatalf("Couldn't write to udp connection: %v\n", err)
		}

	}
}
