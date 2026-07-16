package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string
type parsedState int

const (
	StateInit parsedState = iota
	StateDone
)

var BUF_SIZE = 8
var ERROR_INVALID_HEADER_SPACES = fmt.Errorf("field-name must not have spaces before or after.")
var ERROR_INCOMPLETE_HEADER_LINE = fmt.Errorf("header line incomplete, missing crlf.")

// var ERROR_INVALID_HEADER_FORMAT = fmt.Errorf("incorrect header format: should be:\n\t'field-name: field-value\\r\\n\\r\\n'\n")

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idxSep := bytes.Index(data, []byte("\r\n"))

	if idxSep == -1 {
		// Can't find the \r\n
		return 0, false, nil
	}
	if idxSep == 0 {
		return 2, true, nil
	}

	idxCol := bytes.Index(data, []byte(":"))

	n = len(data[:idxSep]) + len("\r\n")

	parts := make([]string, 2)
	parts[0] = string(data[:idxCol])
	parts[1] = string(data[idxCol+1 : idxSep])

	// field-name part
	if strings.ContainsAny(parts[0], " \t") {
		return 0, false, ERROR_INVALID_HEADER_SPACES
	}

	fieldValue := strings.TrimSpace(parts[1])
	h[parts[0]] = fieldValue

	return n, false, nil
}
