package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if bytes.HasPrefix(data, []byte("\r\n")) {
		return 2, true, nil
	}

	lineEnd := bytes.Index(data, []byte("\r\n"))

	if lineEnd == -1 {
		return 0, false, nil
	}

	line := data[:lineEnd]
	colonIndex := bytes.IndexByte(line, ':')

	if colonIndex <= 0 || line[colonIndex-1] == ' ' || line[colonIndex-1] == '\n' {
		return 0, false, errors.New("invalid header format (space before colon or no key)")
	}

	key := strings.TrimSpace(string(line[:colonIndex]))
	value := strings.TrimSpace(string(line[colonIndex+1:]))

	if !keyIsValid(key) {
		return 0, false, errors.New("invalid key")
	}

	if key == "" {
		return 0, false, errors.New("header key is empty")
	}

	h.Set(key, value)

	return lineEnd + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, exist := h[key]
	if exist {
		value = strings.Join([]string{
			v,
			value,
		}, ", ")
	}
	h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	v, ok := h[key]
	return v, ok
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func keyIsValid(s string) bool {
	for _, char := range s {
		if !isAllowedKeyChar(char) {
			return false
		}
	}
	return true
}

func isAllowedKeyChar(r rune) bool {
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}

	switch r {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}

	return false
}
