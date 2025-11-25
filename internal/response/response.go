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
