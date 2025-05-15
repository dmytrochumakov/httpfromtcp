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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	const httpMessage = "HTTP/1.1"

	responseMessage := fmt.Sprintf("%s %d %s \r\n", httpMessage, statusCode, reasonPhrase(statusCode))

	_, err := w.Write([]byte(responseMessage))
	if err != nil {
		return err
	}

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

func WriteHeaders(w io.Writer, headers headers.Headers) error {
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
	return nil
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
