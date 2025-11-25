package response

import (
	"fmt"
	"io"

	"github.com/h0dy/tcp-to-http/internal/headers"
)

// Writer provides methods to write HTTP responses
type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

// WriteStatusLine writes the HTTP status line for the given status code
func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	fmt.Fprint(w.writer, GetStatusLine(statusCode))
	return nil
}

// WriteHeaders writes the provided HTTP headers to the connection
func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if len(headers) < 1 {
		return fmt.Errorf("error: headers is empty")
	}

	for header, val := range headers {
		_, err := fmt.Fprintf(w.writer, "%v: %v\r\n", header, val)
		if err != nil {
			return err
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

// WriteBody writes the body content to the connection
func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.writer.Write(p)
}

// WriteChunkedBody writes a single chunk in HTTP chunked transfer encoding
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunkSize := len(p)

	nTotal := 0 // total bytes

	// Write the chunk size in hexadecimal followed by CRLF
	n, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	// Write the actual chunk data
	n, err = w.writer.Write(p)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

	// Write the trailing CRLF after the chunk data
	n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n
	return nTotal, nil
}

// WriteChunkedBodyDone writes the final chunk to indicate the end of a chunked HTTP message
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	return n, nil
}

// WriteTrailers writes HTTP trailer headers after the final chunk
func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n")) // end of trailers
	return err
}
