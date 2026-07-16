package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
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
var ERROR_INVALID_CHAR = fmt.Errorf("field name must contain only: A-Z, a-z, 0-9 and special chars: !, #, $, %%, &, ', *, +, -, ., ^, _, `, |, ~")

// var ERROR_INVALID_HEADER_FORMAT = fmt.Errorf("incorrect header format: should be:\n\t'field-name: field-value\\r\\n\\r\\n'\n")

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idxSep := bytes.Index(data, []byte("\r\n"))

	if idxSep == -1 {
		// Can't find the \r\n, incomplete data
		return 0, false, nil
	}
	if idxSep == 0 { // rn in the beggining so whatever was parsed until now, are the headers
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

	err = isHeaderValid(parts[0])
	if err != nil {
		return 0, true, err
	}

	key := strings.ToLower(parts[0])
	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + fieldValue
	} else {
		h[key] = fieldValue
	}

	return n, false, nil
}

// Don't touch helpers

func isHeaderValid(header string) error {
	// nil if header only contians aA-zZ, 0-9, special allowed
	for _, ch := range header {
		if isAlphaNum(ch) {
			continue
		} else if isSpecialAllowed(ch) {
			continue
		}
		return ERROR_INVALID_CHAR

	}
	return nil
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isSpecialAllowed(ch rune) bool {
	var allowed = map[rune]struct{}{
		'!':  {},
		'#':  {},
		'$':  {},
		'%':  {},
		'&':  {},
		'\'': {},
		'*':  {},
		'+':  {},
		'-':  {},
		'.':  {},
		'^':  {},
		'_':  {},
		'`':  {},
		'|':  {},
		'~':  {},
	}
	if _, ok := allowed[ch]; !ok {
		return false
	}
	return true

}
