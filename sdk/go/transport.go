package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func jsonUnmarshal(b []byte, v any) error {
	return json.Unmarshal(b, v)
}

func urlPathEscape(s string) string {
	return url.PathEscape(s)
}

func urlQueryVal(s string) string {
	return url.QueryEscape(s)
}

func (c *Client) doJSONRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL()+path, rdr)
	if err != nil {
		return nil, err
	}
	c.authorize(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kranix-api %s %s: %s: %s", method, path, resp.Status, string(raw))
	}
	return raw, nil
}

func (c *Client) baseURL() string {
	return strings.TrimRight(c.config.ServerURL, "/")
}

func (c *Client) authorize(req *http.Request) {
	if c.config.SkipAuth || c.config.APIKey == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
}

func (c *Client) httpClientFor(ctx context.Context, streaming bool) *http.Client {
	if streaming {
		return &http.Client{
			Transport:     c.httpClient.Transport,
			CheckRedirect: c.httpClient.CheckRedirect,
			Jar:           c.httpClient.Jar,
			Timeout:       0,
		}
	}
	return c.httpClient
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any, streaming bool) (*http.Response, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL()+path, rdr)
	if err != nil {
		return nil, err
	}
	c.authorize(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClientFor(ctx, streaming).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("kranix-api %s %s: %s: %s", method, path, resp.Status, string(b))
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp, fmt.Errorf("decode %s %s: %w", method, path, err)
		}
	} else if resp.StatusCode != http.StatusNoContent {
		_, _ = io.Copy(io.Discard, resp.Body)
	}
	return resp, nil
}

func clientTimeout(sec int) time.Duration {
	if sec <= 0 {
		return 60 * time.Second
	}
	return time.Duration(sec) * time.Second
}

func valuesFromLogOptions(opts *logOptionsCompat) url.Values {
	q := url.Values{}
	if opts == nil {
		return q
	}
	if opts.Follow {
		q.Set("follow", "true")
	}
	if opts.Tail > 0 {
		q.Set("tail", fmt.Sprintf("%d", opts.Tail))
	}
	if opts.Since > 0 {
		q.Set("since", fmt.Sprintf("%d", opts.Since))
	}
	return q
}

// logOptionsCompat avoids importing types in transport for tail type differences.
type logOptionsCompat struct {
	Follow bool
	Tail   int64
	Since  int64
}
