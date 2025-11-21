package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type requestState int

const (
	requestInitialized requestState = iota
	requestDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf        = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	// buffer holding incoming bytes
	buf := make([]byte, bufferSize)

	// bytes we currently have
	readToIdx := 0
	request := &Request{
		state: requestInitialized,
	}

	for request.state != requestDone {
		// grow buffer if full
		if readToIdx >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// read new bytes into the end of buffer
		n, err := reader.Read(buf[readToIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = requestDone
				break
			}
			return nil, err
		}

		// update how many bytes we have for parsing
		readToIdx += n

		// parse the data we currently have
		numBytesParsed, err := request.parse(buf[:readToIdx])
		if err != nil {
			return nil, err
		}

		// shift the remaining (unparsed) bytes to the front
		copy(buf, buf[numBytesParsed:])
		readToIdx -= numBytesParsed

	}
	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineStr := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineStr)
	if err != nil {
		return nil, 0, err
	}
	// CRLF is 2 bytes long
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	lines := strings.Split(str, " ")

	if len(lines) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := lines[0]
	for _, ch := range method {
		if ch < 'A' || ch > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	target := lines[1]

	httpVersion := strings.Split(lines[2], "/")
	if len(httpVersion) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}
	http := httpVersion[0]
	if http != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", http)
	}
	version := httpVersion[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version, we only support 1.1: %s", version)
	}

	return &RequestLine{
		HttpVersion:   version,
		RequestTarget: target,
		Method:        method,
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {

	case requestInitialized:
		request, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 { // need more data
			return 0, nil
		}
		r.RequestLine = *request
		r.state = requestDone
		return n, nil

	case requestDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}
