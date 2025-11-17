package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatalf("Error in opening messages.txt file: %v\n", err)
	}

	for line := range getLinesChannel(f) {
		fmt.Printf("read: %v\n", line)
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
