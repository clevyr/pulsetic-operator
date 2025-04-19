package pulsetic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"path"
	"strconv"
)

var ErrMonitorNotFound = errors.New("monitor not found")

type MonitorClient struct {
	client Client
}

const endpointMonitors = "monitors"

type CreateMonitorRequest struct {
	URLs []string `json:"urls"`
}

func (m MonitorClient) Create(ctx context.Context, monitor Monitor) (Monitor, error) {
	req := CreateMonitorRequest{
		URLs: []string{monitor.URL},
	}
	b, err := json.Marshal(req)
	if err != nil {
		return Monitor{}, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	res, err := m.client.Do(ctx, http.MethodPost, endpointMonitors, bytes.NewReader(b))
	if err != nil {
		return Monitor{}, err
	}
	defer consumeAndClose(res.Body)

	for v, err := range iterMonitors(res.Body) {
		if err != nil {
			return v, err
		}

		if v.URL == monitor.URL {
			cancel()
			monitor = v
			break
		}
	}

	return m.Update(ctx, monitor.ID, monitor)
}

func (m MonitorClient) List(ctx context.Context) iter.Seq2[Monitor, error] {
	return func(yield func(Monitor, error) bool) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		res, err := m.client.Do(ctx, http.MethodGet, endpointMonitors, nil)
		if err != nil {
			yield(Monitor{}, err)
			return
		}
		defer consumeAndClose(res.Body)

		for monitor, err := range iterMonitors(res.Body) {
			if !yield(monitor, err) {
				cancel()
				return
			}
		}
	}
}

func (m MonitorClient) Get(ctx context.Context, opts ...FindOption) (Monitor, error) {
	var findBy FindRequest
	for _, opt := range opts {
		opt(&findBy)
	}
	for monitor, err := range m.List(ctx) {
		if err != nil || findBy.Matches(monitor) {
			return monitor, err
		}
	}
	return Monitor{}, ErrMonitorNotFound
}

func (m MonitorClient) Update(ctx context.Context, id int64, monitor Monitor) (Monitor, error) {
	b, err := json.Marshal(monitor.EditParams())
	if err != nil {
		return Monitor{}, err
	}

	u := path.Join(endpointMonitors, strconv.FormatInt(id, 10))
	res, err := m.client.Do(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return Monitor{}, err
	}
	defer consumeAndClose(res.Body)

	return m.Get(ctx, FindByID(id))
}

func (m MonitorClient) Delete(ctx context.Context, id int64) error {
	p := path.Join(endpointMonitors, strconv.FormatInt(id, 10))
	res, err := m.client.Do(ctx, http.MethodDelete, p, nil)
	if err != nil {
		return err
	}
	defer consumeAndClose(res.Body)
	return nil
}
