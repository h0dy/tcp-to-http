package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/h0dy/tcp-to-http/internal/headers"
)

type requestState int

const (
	requestInitialized requestState = iota
	requestDone
	requestParsingHeader
	requestParsingBody
)

// Request represents an HTTP request
type Request struct {
	state          requestState    // current parsing state
	RequestLine    RequestLine     // HTTP method, target path, and HTTP version
	Headers        headers.Headers // HTTP headers
	Body           []byte          // request body
	bodyLengthRead int             // track of body bytes already read (parsed)
}

// RequestLine represents the start line in HTTP request
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	crlf       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	// buffer holding incoming bytes
	buf := make([]byte, bufferSize)

	// bytes we currently have
	readToIdx := 0
	request := &Request{
		state:   requestInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
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
				if request.state != requestDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", request.state, n)
				}
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

// parse process the incoming raw data
func (r *Request) parse(data []byte) (int, error) {
	totalParsed := 0
	for r.state != requestDone {
		n, err := r.parseSingle(data[totalParsed:])
		if err != nil {
			return 0, err
		}
		totalParsed += n
		if n == 0 {
			break
		}
	}
	return totalParsed, nil
}

// parseSingle reads/parses the next part of the HTTP request based on the current state
func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {

	// requestInitialized case handles the start line part
	case requestInitialized:
		request, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 { // need more data
			return 0, nil
		}
		r.RequestLine = *request
		r.state = requestParsingHeader
		return n, nil

	// requestParsingHeader case handles the headers
	case requestParsingHeader:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestParsingBody
		}
		return n, nil

	// requestParsingBody case handles the body (if any)
	case requestParsingBody:
		lengthVal, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestDone
			return len(data), nil
		}

		length, err := strconv.Atoi(lengthVal)
		if err != nil {
			return 0, fmt.Errorf("error: Content-Length header contains invalid data (non-numeric)")
		}

		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)

		if r.bodyLengthRead > length {
			return 0, fmt.Errorf("error: body's length doesn't match Content-Length header\nContent-Length: %v, body's length: %v", length, len(data))
		}
		if length == len(r.Body) {
			r.state = requestDone
		}
		return len(data), nil

	case requestDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")

	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

// parseRequestLine reads/parses the start-line
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

	return requestLine, idx + 2, nil
}

// requestLineFromString parses the start line from string to RequestLine
func requestLineFromString(str string) (*RequestLine, error) {
	lines := strings.Split(str, " ")

	if len(lines) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := lines[0]
	// check if method is uppercase
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
