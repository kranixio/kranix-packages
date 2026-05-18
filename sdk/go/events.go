package sdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SSEEvent is a single event frame from GET /api/sse ( workload lifecycle, or custom broadcasts ).
type SSEEvent struct {
	ID    string
	Event string
	Data  json.RawMessage
}

// SubscribeOptions configures workload / platform event streaming.
type SubscribeOptions struct {
	// ClientID is sent as client_id; mock and API default when empty.
	ClientID string
	// Namespaces limits delivery when the server applies filters ( repeat Namespace query param ).
	Namespaces []string
}

// SubscribeSSE streams Server-Sent Events until ctx is cancelled.
// Handler is invoked for each event; return an error to stop with that error.
func (c *Client) SubscribeSSE(ctx context.Context, opts *SubscribeOptions, handler func(SSEEvent) error) error {
	q := ""
	if opts != nil {
		v := url.Values{}
		if opts.ClientID != "" {
			v.Set("client_id", opts.ClientID)
		}
		for _, ns := range opts.Namespaces {
			v.Add("namespace", ns)
		}
		q = v.Encode()
	}
	path := "/api/sse"
	if q != "" {
		path += "?" + q
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL()+path, nil)
	if err != nil {
		return err
	}
	c.authorize(req)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClientFor(ctx, true).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sse: %s: %s", resp.Status, string(body))
	}
	return readSSEStream(resp.Body, handler)
}

// readSSEStream parses SSE frames ( id / event / data / retry ), joining multi-line data.
func readSSEStream(r io.Reader, handler func(SSEEvent) error) error {
	sc := bufio.NewScanner(r)
	// Large data lines
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)

	var ev SSEEvent
	var dataBuf bytes.Buffer
	flush := func() error {
		if dataBuf.Len() == 0 && ev.Event == "" && ev.ID == "" {
			return nil
		}
		ev.Data = append(json.RawMessage(nil), dataBuf.Bytes()...)
		if err := handler(ev); err != nil {
			return err
		}
		ev = SSEEvent{}
		dataBuf.Reset()
		return nil
	}

	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "id:") {
			ev.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			continue
		}
		if strings.HasPrefix(line, "event:") {
			ev.Event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			d := strings.TrimPrefix(line, "data:")
			if len(d) > 0 && d[0] == ' ' {
				d = d[1:]
			}
			if dataBuf.Len() > 0 {
				dataBuf.WriteByte('\n')
			}
			dataBuf.WriteString(d)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return flush()
}