package pulsetic

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
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

type IntBool bool

func (i *IntBool) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*i = false
		return nil
	}

	parsed, err := strconv.ParseBool(string(b))
	if err != nil {
		return err
	}

	*i = IntBool(parsed)
	return nil
}

type UnixOrTime time.Time

func (t *UnixOrTime) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		*t = UnixOrTime(time.Time{})
		return nil
	}

	if s, err := strconv.Unquote(string(b)); err == nil {
		parsed, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return err
		}

		*t = UnixOrTime(parsed)
		return nil
	}

	i, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return err
	}

	*t = UnixOrTime(time.Unix(i, 0))
	return nil
}
