package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/dmytrochumakov/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine    RequestLine
	ParserState    ParserState
	Headers        headers.Headers
	Body           []byte
	BodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

type ParserState int

const (
	StateInitialized ParserState = iota
	StateParsingHeaders
	StateParsingBody
	StateDone
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		ParserState: StateInitialized,
		Headers:     headers.NewHeaders(),
		Body:        make([]byte, 0),
	}

	for request.ParserState != StateDone {
		bufLen := len(buf)
		if readToIndex == bufLen {
			newBuf := make([]byte, bufLen*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			if request.ParserState != StateDone {
				return nil, fmt.Errorf("incomplete request in state: %d", request.ParserState)
			}

			break
		}
		if err != nil {
			return nil, err
		}

		readToIndex += n
		numberOfParsedBytes, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if numberOfParsedBytes == 0 {
			continue
		}
		copy(buf, buf[numberOfParsedBytes:readToIndex])
		readToIndex -= numberOfParsedBytes
	}

	if request.ParserState != StateDone {
		return nil, errors.New("incomplete request")
	}

	return request, nil
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	idx := strings.Index(line, "\r\n")
	if idx == -1 {
		return nil, 0, nil
	}
	parts := strings.Split(line, "\r\n")
	requestLineParts := strings.Split(parts[0], " ")
	if len(requestLineParts) <= 2 {
		return nil, 0, errors.New("missing part")
	}
	httpVersionParts := requestLineParts[2]
	httpVersion := strings.Split(httpVersionParts, "/")
	return &RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   httpVersion[1],
	}, idx + 2, nil
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.ParserState != StateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case StateInitialized:
		parsedRequestLine, numberOfBytes, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if numberOfBytes == 0 {
			return 0, nil
		}
		r.RequestLine = *parsedRequestLine
		r.ParserState = StateParsingHeaders
		return numberOfBytes, nil
	case StateParsingHeaders:
		numberOfBytes, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = StateParsingBody
		}
		return numberOfBytes, nil
	case StateParsingBody:
		contentLengthStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.ParserState = StateDone
			return len(data), nil
		}
		contentLengthInt, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length value: %s", contentLengthStr)
		}

		r.Body = append(r.Body, data...)
		r.BodyLengthRead += len(data)

		if r.BodyLengthRead > contentLengthInt {
			return 0, errors.New("body is greater than the Content-Length header")
		}
		if r.BodyLengthRead == contentLengthInt {
			r.ParserState = StateDone
		}
		return len(data), nil
	case StateDone:
		return 0, errors.New("trying read data in done state")
	default:
		return 0, errors.New("unknown state")
	}
}
