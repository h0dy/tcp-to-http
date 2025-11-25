package response

import (
	"strconv"
	"time"

	"github.com/h0dy/tcp-to-http/internal/headers"
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Date", formateHTTPDate(time.Now()))
	headers.Set("Content-Type", "text/plain")
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	return headers
}

func formateHTTPDate(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}
