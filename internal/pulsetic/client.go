package pulsetic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func NewClient(apiKey string) Client {
	api := "https://api.pulsetic.com/api/public"
	if env := os.Getenv("PULSETIC_API"); env != "" {
		api = strings.TrimSuffix(env, "/")
	}

	return Client{url: api, apiKey: apiKey}
}

type Client struct {
	url    string
	apiKey string
}

func (c Client) NewRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	u := c.url + "/" + endpoint

	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", c.apiKey)

	return req, nil
}

func (c Client) Do(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := c.NewRequest(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 400 {
		errRes := ResponseError{Response: res}
		if b, err := io.ReadAll(res.Body); err == nil {
			if err := json.Unmarshal(b, &errRes); err != nil {
				return nil, fmt.Errorf("%w: %s", errRes, b)
			}
		}
		return nil, errRes
	}

	return res, nil
}

func (c Client) Monitors() MonitorClient {
	return MonitorClient{client: c}
}
