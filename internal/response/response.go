package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/dmytrochumakov/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

type WriterState int

const (
	stateInitial WriterState = iota
	stateStatusLineWritten
	stateHeadersWritten
)

type Writer struct {
	w     io.Writer
	state WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: stateInitial,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInitial {
		return fmt.Errorf("status line only can be written in initial state")
	}

	const httpMessage = "HTTP/1.1"

	responseMessage := fmt.Sprintf("%s %d %s \r\n", httpMessage, statusCode, reasonPhrase(statusCode))

	_, err := w.Write([]byte(responseMessage))
	if err != nil {
		return err
	}

	w.state = stateStatusLineWritten

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	contentLenStr := strconv.Itoa(contentLen)

	headers.Set("Content-Length", contentLenStr)
	headers.Set("Content-Type", "text/plain")
	headers.Set("Connection", "close")

	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != stateStatusLineWritten {
		return fmt.Errorf("headers only can be written after status line")
	}

	for key, value := range headers {
		headerStr := buildHeaderString(key, value)
		_, err := w.Write([]byte(headerStr))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = stateHeadersWritten

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, fmt.Errorf("body only can be written after headers")
	}
	return w.Write(p)
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.w.Write(p)
}

func buildHeaderString(key, value string) string {
	return fmt.Sprintf("%s: %s\r\n", key, value)
}

func reasonPhrase(statusCode StatusCode) string {
	switch statusCode {
	case OK:
		return "OK"
	case BadRequest:
		return "Bad Request"
	case InternalServerError:
		return "Internal Server Error"
	}
	return ""
}
