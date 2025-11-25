package response

import (
	"fmt"
)

type StatusCode int

const (
	Successful  StatusCode = (iota + 2) * 100 // 200
	Redirect                                  // 300
	ClientError                               // 400
	ServerError                               // 500
)

func GetStatusLine(statusCode StatusCode) string {
	var res string
	switch statusCode {
	case Successful:
		res = "OK"

	case ClientError:
		res = "Bad Request"

	case ServerError:
		res = "Internal Server Error"

	default:
		res = ""
	}
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, res)
}
