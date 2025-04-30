package request

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	parsedRequestLine, err := parseRequestLine(string(b))
	if err != nil {
		return nil, err
	}

	return &Request{RequestLine: *parsedRequestLine}, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, "\r\n")
	requestLineParts := strings.Split(parts[0], " ")
	if len(requestLineParts) <= 2 {
		return nil, errors.New("missing part")
	}
	httpVersionParts := requestLineParts[2]
	httpVersion := strings.Split(httpVersionParts, "/")
	return &RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   httpVersion[1],
	}, nil
}
