package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

// n = the number of bytes consumed
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		// end of the headers
		// consume the CRLF
		return 2, true, nil
	}
	headersSlice := bytes.SplitN(data[:idx], []byte(":"), 2)
	header, value, err := parseAndValidateHeader(headersSlice)
	if err != nil {
		return 0, false, fmt.Errorf("poorly formatted header: %s", err.Error())
	}
	h.Set(header, value)
	return idx + 2, false, nil
}

func parseAndValidateHeader(line [][]byte) (string, string, error) {
	header := line[0]
	if bytes.HasSuffix(header, []byte(" ")) {
		return "", "", fmt.Errorf("poorly formatted headers: %s", header)
	}
	if len(header) < 1 {
		return "", "", fmt.Errorf("the length of the header must be at least 1: %s", header)
	}

	parsedHeader := strings.TrimSpace(string(header))
	// Validate each character/token against RFC token rules
	for _, c := range parsedHeader {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '!' || c == '#' || c == '$' || c == '%' ||
			c == '&' || c == '\'' || c == '*' || c == '+' ||
			c == '-' || c == '.' || c == '^' || c == '_' ||
			c == '`' || c == '|' || c == '~':
		default:
			return "", "", fmt.Errorf("invalid header: %s", header)
		}
	}
	value := strings.TrimSpace(string(line[1]))

	return strings.ToLower(parsedHeader), value, nil
}

func (h Headers) Set(key, value string) {
	if old, ok := h[key]; ok {
		value = old + ", " + value
		h[key] = value
	}
	h[key] = value
}

func NewHeaders() Headers {
	return Headers{}
}
