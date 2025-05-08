package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

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
	if key == "" {
		return 0, false, errors.New("header key is empty")
	}
	value := strings.TrimSpace(string(line[colonIndex+1:]))
	h[key] = value

	return lineEnd + 2, false, nil
}
