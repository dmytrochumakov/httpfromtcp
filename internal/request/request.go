package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
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
	StateDone
)

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{ParserState: StateInitialized}

	for request.ParserState != StateDone {
		bufLen := len(buf)
		if readToIndex == bufLen {
			newBuf := make([]byte, bufLen*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			request.ParserState = StateDone
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
		r.ParserState = StateDone
		return numberOfBytes, nil
	case StateDone:
		return 0, errors.New("trying read data in done state")
	default:
		return 0, errors.New("unknown state")
	}
}
