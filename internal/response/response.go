package response

import (
	"httpgo/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	Ok          StatusCode = 200
	BadRequest  StatusCode = 400
	InternalErr StatusCode = 500
)

var codeToMsg = map[StatusCode]string{Ok: "OK", BadRequest: "Bad Request", InternalErr: "Internal Server Error"}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {

	_, err := w.Write([]byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + codeToMsg[statusCode] + "\r\n"))
	if err != nil {
		return err
	}

	return nil

}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for h, v := range headers {
		_, err := w.Write([]byte(h + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
