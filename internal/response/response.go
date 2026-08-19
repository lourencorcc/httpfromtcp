package response

import (
	"fmt"
	"httpgo/internal/headers"
	"io"
	"strconv"
	"strings"
)

type StatusCode int
type WriterState int

const BadRequestHtmlMsg = "<html>\n <head>\n <title>400 Bad Request</title>\n </head>\n <body>\n <h1>Bad Request</h1>\n <p>Your request honestly kinda sucked.</p>\n </body>\n</html>\n"
const InternalErrHtmlMsg = "<html>\n <head>\n <title>500 Internal Server Error</title>\n </head>\n <body>\n <h1>Internal Server Error</h1>\n <p>Okay, you know what? This one is on me.</p>\n </body>\n</html>\n"
const OkHtmlMsg = "<html>\n <head>\n <title>200 OK</title>\n </head>\n <body>\n <h1>Success!</h1>\n <p>Your request was an absolute banger.</p>\n </body>\n</html>\n"

var ERR_INVALID_WRITE_ORDER error = fmt.Errorf("wrote to response in invalid order. correct order is: StatusLine > Headers > Body.")

const (
	Ok          StatusCode = 200
	BadRequest  StatusCode = 400
	InternalErr StatusCode = 500
)

const (
	StatusLine WriterState = iota
	Headers
	Body
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
	headers[strings.ToLower("Content-Length")] = strconv.Itoa(contentLen)
	headers[strings.ToLower("Connection")] = "close"
	headers[strings.ToLower("Content-Type")] = "text/plain"

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

// Refactor for handling HTML response and let users customize their response
type Writer struct {
	NextWrite WriterState // enforce order of writes: Line > Headers > Body
	/* headers    headers.Headers
	statusCode StatusCode
	body       []byte */
	ResponseType string
	Writer       io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.NextWrite != StatusLine {
		return ERR_INVALID_WRITE_ORDER
	}
	_, err := w.Writer.Write([]byte("HTTP/1.1 " + strconv.Itoa(int(statusCode)) + " " + codeToMsg[statusCode] + "\r\n"))
	if err != nil {
		return err
	}
	w.NextWrite = Headers
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.NextWrite != Headers {
		return ERR_INVALID_WRITE_ORDER
	}
	cType, err := headers.Get("Content-Type")
	if err != nil {
		return fmt.Errorf("content-type header not found: %w", err)
	}
	if cType != w.ResponseType {
		headers.Set("Content-Type", w.ResponseType)
	}
	for h, v := range headers {
		_, err := w.Writer.Write([]byte(h + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err = w.Writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.NextWrite = Body
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.NextWrite != Body {
		return 0, ERR_INVALID_WRITE_ORDER
	}
	n, err := w.Writer.Write(p)
	if err != nil {
		return n, err
	}
	return n, nil // should i return new bytes or cur lenght?
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {

	_, err := w.WriteBody([]byte(fmt.Sprintf("%x\r\n", len(p)))) // can be optimized
	if err != nil {
		return 0, err
	}

	n, err := w.WriteBody(p)
	if err != nil {
		return 0, err
	}

	_, err = w.WriteBody([]byte("\r\n"))
	if err != nil {
		return 0, err
	}
	fmt.Printf("responded with %d bytes chunked\n", n)

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.WriteBody([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, err
	}
	return n, nil
}
