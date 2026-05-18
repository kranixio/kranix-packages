package sdk

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/kranix-io/kranix-packages/types"
)

// Config represents the client configuration.
type Config struct {
	ServerURL string
	APIKey    string
	Timeout   int // seconds; default 60
	// SkipAuth omits the Authorization header (for use with kranix-mock-api -skip-auth).
	SkipAuth bool
}

// Client is the public Go SDK client for kranix-api.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// New creates a new Kranix API client.
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("server URL is required")
	}
	if !cfg.SkipAuth && cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required unless SkipAuth is set")
	}
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: clientTimeout(cfg.Timeout),
		},
	}, nil
}

// Workloads provides workload-related operations.
func (c *Client) Workloads() *WorkloadsClient {
	return &WorkloadsClient{client: c}
}

// Pods provides pod-related operations.
func (c *Client) Pods() *PodsClient {
	return &PodsClient{client: c}
}

// Namespaces provides namespace-related operations.
func (c *Client) Namespaces() *NamespacesClient {
	return &NamespacesClient{client: c}
}

// WorkloadsClient handles workload operations.
type WorkloadsClient struct {
	client *Client
}

// Deploy deploys a new workload.
func (w *WorkloadsClient) Deploy(ctx context.Context, spec *types.WorkloadSpec) (*types.Workload, error) {
	var out types.Workload
	_, err := w.client.doJSON(ctx, http.MethodPost, "/api/v1/workloads", spec, &out, false)
	if err != nil {
		return nil, err
	}
	if out.ID == "" {
		return nil, fmt.Errorf("deploy response missing workload id (is the API still stubbed? use kranix-mock-api for tests)")
	}
	return &out, nil
}

// Get retrieves a workload by ID.
func (w *WorkloadsClient) Get(ctx context.Context, id string) (*types.Workload, error) {
	var out types.Workload
	_, err := w.client.doJSON(ctx, http.MethodGet, "/api/v1/workloads/"+urlPathEscape(id), nil, &out, false)
	if err != nil {
		return nil, err
	}
	if out.ID == "" && out.Name == "" {
		return nil, fmt.Errorf("api returned non-workload body for GET /workloads/%s", id)
	}
	return &out, nil
}

// List lists workloads, optionally filtered by namespace.
func (w *WorkloadsClient) List(ctx context.Context, namespace string) ([]*types.Workload, error) {
	path := "/api/v1/workloads"
	if namespace != "" {
		path += "?namespace=" + urlQueryVal(namespace)
	}
	body, err := w.client.doJSONRaw(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var asSlice []*types.Workload
	if err := jsonUnmarshal(body, &asSlice); err == nil {
		return asSlice, nil
	}
	var wrap struct {
		Workloads []*types.Workload `json:"workloads"`
	}
	if err := jsonUnmarshal(body, &wrap); err != nil {
		return nil, err
	}
	return wrap.Workloads, nil
}

// Update updates an existing workload.
func (w *WorkloadsClient) Update(ctx context.Context, id string, spec *types.WorkloadSpec) (*types.Workload, error) {
	var out types.Workload
	_, err := w.client.doJSON(ctx, http.MethodPatch, "/api/v1/workloads/"+urlPathEscape(id), spec, &out, false)
	return &out, err
}

// Delete deletes a workload.
func (w *WorkloadsClient) Delete(ctx context.Context, id string) error {
	_, err := w.client.doJSON(ctx, http.MethodDelete, "/api/v1/workloads/"+urlPathEscape(id), nil, nil, false)
	return err
}

// Restart requests a workload restart.
func (w *WorkloadsClient) Restart(ctx context.Context, id string) error {
	_, err := w.client.doJSON(ctx, http.MethodPost, "/api/v1/workloads/"+urlPathEscape(id)+"/restart", map[string]string{}, nil, false)
	return err
}

// ListPods returns pods belonging to a workload.
func (w *WorkloadsClient) ListPods(ctx context.Context, workloadID string) ([]*types.Pod, error) {
	body, err := w.client.doJSONRaw(ctx, http.MethodGet, "/api/v1/workloads/"+urlPathEscape(workloadID)+"/pods", nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Pods []*types.Pod `json:"pods"`
	}
	if err := jsonUnmarshal(body, &wrap); err != nil {
		return nil, err
	}
	return wrap.Pods, nil
}

// Analyze runs workload analysis.
func (w *WorkloadsClient) Analyze(ctx context.Context, id string) (*types.AnalysisResult, error) {
	var out types.AnalysisResult
	_, err := w.client.doJSON(ctx, http.MethodGet, "/api/v1/workloads/"+urlPathEscape(id)+"/analyze", nil, &out, false)
	return &out, err
}

// PodsClient handles pod operations.
type PodsClient struct {
	client *Client
}

// StreamLogs streams newline-delimited log payloads from GET /api/v1/pods/{id}/logs (SSE).
func (p *PodsClient) StreamLogs(ctx context.Context, podID string, opts *types.LogOptions) (<-chan string, error) {
	lc := logOptionsCompat{}
	if opts != nil {
		lc.Follow = opts.Follow
		if opts.Tail > 0 {
			lc.Tail = opts.Tail
		} else if opts.TailLines > 0 {
			lc.Tail = int64(opts.TailLines)
		}
		lc.Since = opts.Since
	}
	q := valuesFromLogOptions(&lc).Encode()
	path := "/api/v1/pods/" + urlPathEscape(podID) + "/logs"
	if q != "" {
		path += "?" + q
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.client.baseURL()+path, nil)
	if err != nil {
		return nil, err
	}
	p.client.authorize(req)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.httpClientFor(ctx, true).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("logs: %s", resp.Status)
	}

	out := make(chan string, 32)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		_ = readSSEStream(resp.Body, func(ev SSEEvent) error {
			if ev.Event != "log" {
				return nil
			}
			line := strings.TrimSpace(string(ev.Data))
			if line == "" {
				return nil
			}
			select {
			case out <- line:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()
	return out, nil
}

// Get retrieves a pod by ID (when supported by the API).
func (p *PodsClient) Get(ctx context.Context, id string) (*types.Pod, error) {
	var out types.Pod
	_, err := p.client.doJSON(ctx, http.MethodGet, "/api/v1/pods/"+urlPathEscape(id), nil, &out, false)
	return &out, err
}

// List lists pods in a namespace when the deployment exposes that route; may 404 on minimal APIs.
func (p *PodsClient) List(ctx context.Context, namespace string) ([]*types.Pod, error) {
	path := "/api/v1/pods"
	if namespace != "" {
		path += "?namespace=" + urlQueryVal(namespace)
	}
	body, err := p.client.doJSONRaw(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Pods []*types.Pod `json:"pods"`
	}
	if err := jsonUnmarshal(body, &wrap); err != nil {
		return nil, err
	}
	return wrap.Pods, nil
}

// NamespacesClient handles namespace operations.
type NamespacesClient struct {
	client *Client
}

// Get retrieves a namespace by name.
func (n *NamespacesClient) Get(ctx context.Context, name string) (*types.Namespace, error) {
	var out types.Namespace
	_, err := n.client.doJSON(ctx, http.MethodGet, "/api/v1/namespaces/"+urlPathEscape(name), nil, &out, false)
	return &out, err
}

// List lists all namespaces.
func (n *NamespacesClient) List(ctx context.Context) ([]*types.Namespace, error) {
	body, err := n.client.doJSONRaw(ctx, http.MethodGet, "/api/v1/namespaces", nil)
	if err != nil {
		return nil, err
	}
	var wrap struct {
		Namespaces []*types.Namespace `json:"namespaces"`
	}
	if err := jsonUnmarshal(body, &wrap); err != nil {
		return nil, err
	}
	return wrap.Namespaces, nil
}

// Create creates a new namespace.
func (n *NamespacesClient) Create(ctx context.Context, namespace *types.Namespace) (*types.Namespace, error) {
	var out types.Namespace
	_, err := n.client.doJSON(ctx, http.MethodPost, "/api/v1/namespaces", namespace, &out, false)
	return &out, err
}
