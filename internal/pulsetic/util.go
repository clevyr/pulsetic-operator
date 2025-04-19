package pulsetic

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ResponseError struct {
	Response *http.Response      `json:"-"`
	Message  string              `json:"message"`
	Errors   map[string][]string `json:"errors"`
}

func (e ResponseError) Error() string {
	err := "Pulsetic API error"
	if e.Response != nil {
		err += " " + strconv.Itoa(e.Response.StatusCode)
	}
	if e.Message != "" {
		err += ": " + e.Message
	}
	for k, v := range e.Errors {
		err += "\n" + k + ": " + strings.Join(v, " ")
	}
	return err
}

func consumeAndClose(r io.ReadCloser) {
	_, _ = io.Copy(io.Discard, r)
	_ = r.Close()
}

