//nolint:bodyclose
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

	res, err := m.client.Do(ctx, http.MethodPost, endpointMonitors, bytes.NewReader(b))
	if err != nil {
		return Monitor{}, err
	}
	defer consumeAndClose(res.Body)

	var monitors []Monitor
	if err := json.NewDecoder(res.Body).Decode(&monitors); err != nil {
		return Monitor{}, err
	}
	consumeAndClose(res.Body)

	if len(monitors) == 0 {
		return Monitor{}, ErrMonitorNotFound
	}
	monitor.ID = monitors[0].ID

	return m.Update(ctx, monitor.ID, monitor)
}

type ListResponse struct {
	CurrentPage int       `json:"current_page"`
	LastPage    int       `json:"last_page"`
	Data        []Monitor `json:"data"`
}

func (m MonitorClient) List(ctx context.Context) iter.Seq2[Monitor, error] {
	return func(yield func(Monitor, error) bool) {
		page := 1
		for {
			res, err := m.ListPage(ctx, page)
			if err != nil {
				yield(Monitor{}, err)
				return
			}

			for _, monitor := range res.Data {
				if !yield(monitor, nil) {
					return
				}
			}

			if page >= res.LastPage {
				break
			}
			page++
		}
	}
}

func (m MonitorClient) ListPage(ctx context.Context, page int) (*ListResponse, error) {
	u := endpointMonitors + "?page=" + strconv.Itoa(page)

	res, err := m.client.Do(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	defer consumeAndClose(res.Body)

	decoder := json.NewDecoder(res.Body)

	var data ListResponse
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (m MonitorClient) Get(ctx context.Context, opts ...FindOption) (Monitor, error) {
	var findBy FindRequest
	for _, opt := range opts {
		opt(&findBy)
	}

	if findBy.ID != nil && *findBy.ID != 0 {
		if m, err := m.FindByID(ctx, *findBy.ID); err == nil {
			return m, nil
		}
	}

	if findBy.URL != nil && *findBy.URL != "" {
		return m.FindByURL(ctx, *findBy.URL)
	}

	return Monitor{}, ErrMonitorNotFound
}

func (m MonitorClient) FindByID(ctx context.Context, id int64) (Monitor, error) {
	u := path.Join(endpointMonitors, strconv.FormatInt(id, 10))

	res, err := m.client.Do(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Monitor{}, err
	}
	defer consumeAndClose(res.Body)

	var parsed UpdateResponse
	err = json.NewDecoder(res.Body).Decode(&parsed)

	if parsed.Data.ID == 0 {
		return Monitor{}, ErrMonitorNotFound
	}
	return parsed.Data, err
}

func (m MonitorClient) FindByURL(ctx context.Context, url string) (Monitor, error) {
	for monitor, err := range m.List(ctx) {
		if err != nil || monitor.URL == url || monitor.URL+"/" == url {
			return monitor, err
		}
	}
	return Monitor{}, ErrMonitorNotFound
}

type UpdateResponse struct {
	Data Monitor `json:"data"`
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

	var parsed UpdateResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return Monitor{}, err
	}

	return parsed.Data, nil
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
