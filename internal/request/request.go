package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

var BUF_SIZE = 8
var ERROR_EMPTY_TARGET = fmt.Errorf("request target cannot be empty")
var ERROR_INVALID_HTTP_VERSION = fmt.Errorf("HTTP version is invalid")
var ERROR_INVALID_METHOD = fmt.Errorf("request method is invalid")
var ERROR_MALFORMED_REQUEST = fmt.Errorf("invalid number of parts in the request.")
var ERROR_PARSED_REQUEST = fmt.Errorf("the request is already parsed.")
var ERROR_UNKNOWN_STATE = fmt.Errorf("Unknown request state.")
var ERROR_EOF_B4_END = fmt.Errorf("Got io.EOF with unparsable data in buffer.")

type parsedState int

const (
	StateInit parsedState = iota
	StateDone
)

type Request struct {
	RequestLine RequestLine
	State       parsedState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func newRequest() *Request {
	return &Request{
		State: StateInit,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case StateInit:
		rl, n, err := parseRequestLine(data)
		if err != nil {
			return n, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *rl
		r.State = StateDone

		return n, nil
	case StateDone:
		return 0, ERROR_PARSED_REQUEST

	default:
		return 0, ERROR_UNKNOWN_STATE
	}

}

func (rl RequestLine) isValid() error {
	if rl.RequestTarget == "" {
		return ERROR_EMPTY_TARGET
	}

	if !isAlphaUpper(rl.Method) { // could also check if in pre-defined allowed methods set "GET", "POST", etc.
		return ERROR_INVALID_METHOD
	}

	if rl.HttpVersion != "1.1" {
		return ERROR_INVALID_HTTP_VERSION
	}
	return nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	b := make([]byte, BUF_SIZE)
	bufIdx := 0

	for req.State != StateDone {
		if len(b) == bufIdx {
			newBuf := make([]byte, len(b)*2)
			copy(newBuf, b)
			b = newBuf
		}

		readN, err := reader.Read(b[bufIdx:])

		if err != nil {
			// check if its EOF
			if errors.Is(err, io.EOF) {
				req.State = StateDone
				break
			}
			return nil, err

		}

		bufIdx += readN

		parsedN, err := req.parse(b[:bufIdx])
		if err != nil {
			return nil, fmt.Errorf("error while parsing: %w", err)
		}
		copy(b, b[parsedN:bufIdx])
		bufIdx -= parsedN

	}

	return req, nil
}

func parseRequestLine(request []byte) (*RequestLine, int, error) {
	//lines := strings.Split(string(request), "\r\n")
	idx := bytes.Index(request, []byte("\r\n"))
	/* if len(lines) == 1 {
		fmt.Println("Not enough!")
		// Can't find the \r\n
		return RequestLine{}, 0, nil
	}*/

	if idx == -1 {
		fmt.Println("Not enough!")
		// Can't find the \r\n
		return nil, 0, nil
	}

	parts := strings.Split(string(request[:idx]), " ")

	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST
	}

	//parts[0] is the method, parts[1] is target, parts[2] the version

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, 0, ERROR_INVALID_HTTP_VERSION
	}

	rl := &RequestLine{Method: parts[0], RequestTarget: parts[1], HttpVersion: httpParts[1]}

	if err := rl.isValid(); err != nil {
		return nil, 0, fmt.Errorf("Invalid HTTP request-line: %w", err)
	}
	return rl, idx + 2, nil

}

func isAlphaUpper(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		char := s[i]
		if !(char >= 'A' && char <= 'Z') {
			return false
		}
	}
	return true
}
