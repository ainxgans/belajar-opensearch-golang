package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	base  string
	index string
	http  *http.Client
}

func NewClient(baseURL, index string) *Client {
	return &Client{base: baseURL, index: index, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) do(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return out, fmt.Errorf("opensearch %s %s: %d %s", method, path, resp.StatusCode, string(out))
	}
	return out, nil
}

func (c *Client) IndexDoc(ctx context.Context, id int64, doc any) error {
	b, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodPut, "/"+c.index+"/_doc/"+strconv.FormatInt(id, 10), b)
	return err
}

func (c *Client) DeleteDoc(ctx context.Context, id int64) error {
	_, err := c.do(ctx, http.MethodDelete, "/"+c.index+"/_doc/"+strconv.FormatInt(id, 10), nil)
	return err
}

// Bulk sends an NDJSON body to _bulk.
func (c *Client) Bulk(ctx context.Context, body []byte) error {
	out, err := c.do(ctx, http.MethodPost, "/_bulk", body)
	if err != nil {
		return err
	}
	var res struct {
		Errors bool `json:"errors"`
	}
	_ = json.Unmarshal(out, &res)
	if res.Errors {
		return fmt.Errorf("bulk had errors: %s", string(out))
	}
	return nil
}

func (c *Client) SearchRaw(ctx context.Context, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPost, "/"+c.index+"/_search", body)
}

func (c *Client) Count(ctx context.Context) (int64, error) {
	out, err := c.do(ctx, http.MethodGet, "/"+c.index+"/_count", nil)
	if err != nil {
		return 0, err
	}
	var res struct {
		Count int64 `json:"count"`
	}
	_ = json.Unmarshal(out, &res)
	return res.Count, nil
}
