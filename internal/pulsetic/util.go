package pulsetic

import (
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"reflect"
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

func iterMonitors(r io.Reader) iter.Seq2[Monitor, error] {
	return func(yield func(Monitor, error) bool) {
		decoder := json.NewDecoder(r)
		if t, err := decoder.Token(); err != nil {
			yield(Monitor{}, err)
			return
		} else if t != json.Delim('[') {
			yield(Monitor{}, &json.UnmarshalTypeError{Value: "object", Type: reflect.TypeOf([]Monitor{})})
			return
		}

		for decoder.More() {
			var monitor Monitor
			if err := decoder.Decode(&monitor); err != nil {
				yield(Monitor{}, err)
				return
			}

			if !yield(monitor, nil) {
				return
			}
		}
	}
}
