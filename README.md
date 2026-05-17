# kranix-packages

> Shared SDK — types, utilities, and client libraries for the Kranix ecosystem.

`kranix-packages` is the common foundation imported by every other Kranix repo and by third-party tools building on top of Kranix. It contains shared domain types, the `RuntimeDriver` interface, error codes, logging utilities, config schemas, auth primitives, and the public SDK client for Go and TypeScript consumers.

All cross-cutting concerns that are needed by more than one repo live here. Nothing here contains business logic — that belongs in `kranix-core`.

---

## What it contains

| Package | Description |
|---|---|
| `types` | Core domain types (Workload, Pod, Namespace, Status, etc.) |
| `types/ratelimit` | Rate limiting and quota types |
| `types/sse` | Server-Sent Events streaming types |
| `types/apiversion` | API versioning and routing types |
| `types/analytics` | Usage analytics and metrics types |
| `types/version` | Semantic versioning and changelog types |
| `types/webhook` | Webhook configuration and event types |
| `errors` | Typed error codes and wrapping utilities |
| `logging` | Structured logger (zap-based) with consistent field conventions |
| `config` | Config schema definitions and loader |
| `auth` | Token types, validation helpers, RBAC primitives, OIDC support |
| `runtime` | The `RuntimeDriver` interface (implemented by kranix-runtime) |
| `sdk/go` | Public Go client for the kranix-api |
| `sdk/typescript` | Public TypeScript/Node.js client for the kranix-api |
| `proto` | Shared protobuf definitions and generated Go/TS stubs |

---

## Architecture position

```
kranix-core    ──┐
kranix-api     ──┤
kranix-mcp     ──┼──►  kranix-packages
kranix-cli     ──┤
kranix-runtime ──┘

Third-party tools  ──►  kranix-packages (SDK)
```

`kranix-packages` has no dependencies on any other Kranix repo. The dependency arrow always points *toward* packages, never away.

---

## Domain types

### `Workload`

```go
type Workload struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Namespace string            `json:"namespace"`
    Spec      WorkloadSpec      `json:"spec"`
    Status    WorkloadStatus    `json:"status"`
    CreatedAt time.Time         `json:"createdAt"`
    UpdatedAt time.Time         `json:"updatedAt"`
}

type WorkloadSpec struct {
    Image     string            `json:"image"`
    Replicas  int               `json:"replicas"`
    Env       map[string]string `json:"env,omitempty"`
    Resources ResourceSpec      `json:"resources,omitempty"`
    Ports     []PortSpec        `json:"ports,omitempty"`
    Backend   string            `json:"backend"`    // docker | kubernetes
}

type ResourceSpec struct {
    CPURequest    string `json:"cpuRequest,omitempty"`
    CPULimit      string `json:"cpuLimit,omitempty"`
    MemoryRequest string `json:"memoryRequest,omitempty"`
    MemoryLimit   string `json:"memoryLimit,omitempty"`
    GPU           *GPUSpec `json:"gpu,omitempty"`
}

type GPUSpec struct {
    Vendor   string `json:"vendor"`   // nvidia | amd
    Count    int32  `json:"count"`    // Number of GPUs
    Type     string `json:"type,omitempty"`     // GPU type (e.g., "A100", "V100", "MI250")
    SKU      string `json:"sku,omitempty"`      // GPU SKU for specific models
    Memory   string `json:"memory,omitempty"`   // GPU memory requirement (e.g., "16Gi")
}

type WorkloadStatus struct {
    Phase         WorkloadPhase `json:"phase"`
    ReadyReplicas int           `json:"readyReplicas"`
    Message       string        `json:"message,omitempty"`
    LastUpdated   time.Time     `json:"lastUpdated"`
}

type WorkloadPhase string

const (
    WorkloadPhasePending   WorkloadPhase = "Pending"
    WorkloadPhaseDeploying WorkloadPhase = "Deploying"
    WorkloadPhaseRunning   WorkloadPhase = "Running"
    WorkloadPhaseDegraded  WorkloadPhase = "Degraded"
    WorkloadPhaseFailed    WorkloadPhase = "Failed"
)
```

### `RuntimeDriver` interface

```go
// Implemented by kranix-runtime backends.
type RuntimeDriver interface {
    Deploy(ctx context.Context, spec *WorkloadSpec) (*WorkloadStatus, error)
    Destroy(ctx context.Context, workloadID string) error
    Restart(ctx context.Context, workloadID string) error
    GetStatus(ctx context.Context, workloadID string) (*WorkloadStatus, error)
    ListWorkloads(ctx context.Context, namespace string) ([]*WorkloadStatus, error)
    StreamLogs(ctx context.Context, podID string, opts *LogOptions) (<-chan string, error)
    Ping(ctx context.Context) error
    Backend() string
}
```

---

## Error codes

```go
var (
    ErrWorkloadNotFound     = &KraneError{Code: "WORKLOAD_NOT_FOUND", HTTP: 404}
    ErrNamespaceNotFound    = &KraneError{Code: "NAMESPACE_NOT_FOUND", HTTP: 404}
    ErrInvalidSpec          = &KraneError{Code: "INVALID_SPEC", HTTP: 400}
    ErrBackendUnavailable   = &KraneError{Code: "BACKEND_UNAVAILABLE", HTTP: 503}
    ErrReconcileFailed      = &KraneError{Code: "RECONCILE_FAILED", HTTP: 500}
    ErrUnauthorized         = &KraneError{Code: "UNAUTHORIZED", HTTP: 401}
    ErrForbidden            = &KraneError{Code: "FORBIDDEN", HTTP: 403}
)

// Wrapping
return fmt.Errorf("deploy %s: %w", name, errors.ErrInvalidSpec)
```

---

## Public SDK

### Go SDK

```bash
go get github.com/kranix-io/kranix-packages/sdk/go
```

```go
import kraneclient "github.com/kranix-io/kranix-packages/sdk/go"

client, err := kraneclient.New(&kraneclient.Config{
    ServerURL: "http://localhost:8080",
    APIKey:    "krane_your_key",
})

// Deploy a workload
workload, err := client.Workloads.Deploy(ctx, &types.WorkloadSpec{
    Name:      "my-app",
    Image:     "nginx:latest",
    Namespace: "staging",
    Replicas:  2,
})

// Stream logs
logCh, err := client.Pods.StreamLogs(ctx, podID, &types.LogOptions{
    Follow: true,
    Tail:   100,
})
for line := range logCh {
    fmt.Println(line)
}

// Analyze a workload
analysis, err := client.Workloads.Analyze(ctx, workload.ID)
fmt.Println(analysis.ProbableFix)
```

### TypeScript SDK

```bash
npm install @kranix-io/sdk
```

```typescript
import { KraneClient } from "@kranix-io/sdk";

const client = new KraneClient({
  serverUrl: "http://localhost:8080",
  apiKey: "krane_your_key",
});

// Deploy
const workload = await client.workloads.deploy({
  name: "my-app",
  image: "nginx:latest",
  namespace: "staging",
  replicas: 2,
});

// Stream logs
for await (const line of client.pods.streamLogs(podId, { follow: true })) {
  console.log(line);
}

// Analyze
const analysis = await client.workloads.analyze(workload.id);
console.log(analysis.probableFix);
```

---

## Logging conventions

All Kranix repos use the shared logger from `kranix-packages/logging`:

```go
import "github.com/kranix-io/kranix-packages/logging"

log := logging.New("kranix-core")

log.Info("reconciling workload",
    "workload_id", id,
    "namespace", ns,
    "backend", backend,
)

log.Error("deploy failed",
    "workload_id", id,
    "error", err,
)
```

Standard fields used across all repos:

| Field | Description |
|---|---|
| `workload_id` | Workload identifier |
| `namespace` | Kubernetes/Kranix namespace |
| `backend` | Runtime backend (docker, kubernetes) |
| `agent_id` | AI agent identifier (kranix-mcp) |
| `request_id` | Inbound request ID (kranix-api) |

---

## Project structure

```├s.go
│   ├── ratelimit.go         # Rate limiting & quota types
│   ├── sse.go               # SSE streaming types
│   ├── apiversion.go        # API versioning types
│   ├── analytics.go         # Analytics & metrics types
│   ├── veriono           # Semantic versining types
│   ├── webhook.go           # Webhook types
│   └── auth.go              # Authentication types
kranix-packages/
├── types/                  # Core domain types
│   ├── workload.go
│   ├── pod.go
│   ├── namespace.go
│   └── status.go
├── errors/                 # Typed error codes
├── logging/                # Shared logger
├── config/                 # Config schema and loader
├── auth/                   # Token types and validation
├── runtime/                # RuntimeDriver interface
├── proto/                  # Protobuf definitions + generated stubs
│   ├── *.proto
│   └── gen/
│       ├── go/
│       └── ts/
└── sdk/
    ├── go/                 # Public Go SDK
    └── typescript/         # Public TypeScript SDK
```

---

## Versioning

`kranix-packages` follows semver strictly:

- **Patch** — bug fixes, no interface changes
- **Minor** — new types or fields (backward-compatible)
- **Major** — breaking changes to interfaces, types, or error codes

All other Kranix repos pin to a specific minor version of `kranix-packages`. Breaking changes require a coordinated release across the ecosystem.

The `RuntimeDriver` interface is considered **stable** after v1.0.0 — changes will only happen in major versions with a deprecation period.

---

## New Feature Types

### Rate Limiting & Quotas (`types/ratelimit.go`)

Provides types for rate limiting and namespace quota management:

- `RateLimitConfig` - Configuration for rate limiting (requests per second, burst size)
- `NamespaceQuota` - Resource quotas per namespace (workloads, CPU, memory, storage)
- `NamespaceQuotaUsage` - Current quota usage with percentages
- `RateLimitInfo` - Rate limit information for clients
- `QuotaRequest` / `QuotaResponse` - Quota management types

### SSE Streaming (`types/sse.go`)

Provides types for Server-Sent Events:

- `SSEEvent` - Server-Sent Event structure with ID, event type, data, timestamp
- `WorkloadStateChange` - Workload state change event with old/new state
- `SSESubscription` - Client subscription with namespace and event filters
- `SSEClient` - SSE client connection information
- `BroadcastMessage` - Message to broadcast to connected clients

### API Versioning (`types/apiversion.go`)

Provides types for API versioning:

- `APIRouteVersion` - API version information (v1, v2) with status and deprecation
- `APIEndpoint` - API endpoint with version mappings
- `APIVersionConfig` - API versioning configuration
- `CompatibilityRule` - Compatibility rules between versions
- `VersionMigration` - Migration guidance between versions

### Analytics (`types/analytics.go`)

Provides types for usage analytics:

- `AnalyticsMetrics` - Time-series metrics for workloads
- `DeployMetrics` - Deployment success/failure metrics
- `ErrorMetrics` - Error rates and types
- `LatencyMetrics` - Performance metrics with percentiles
- `UsageSummary` - Aggregated usage across namespaces/tenants
- `NamespaceUsage` / `TenantUsage` - Usage by namespace or tenant

### Version Management (`types/version.go`)

Provides types for semantic versioning:

- `SemanticVersion` - Semantic version with major, minor, patch
- `DeprecationInfo` - Deprecation details with sunset dates
- `ChangelogEntry` - Changelog entries with change types
- `MigrationInfo` - Migration guidance for breaking changes
- `ChangeType` - Enum of change types (added, changed, deprecated, etc.)

### Webhooks (`types/webhook.go`)

Provides types for webhook configuration:

- `Webhook` - Webhook configuration with provider-specific settings
- `WebhookEvent` - Webhook event types and payloads
- `WebhookDelivery` - Webhook delivery status and retries

### GPU Resources (`types/workload.go`)

Provides types for GPU workload scheduling:

- `GPUSpec` - GPU resource requirements with vendor (nvidia/amd), count, type, SKU, and memory
- Integrated into `ResourceSpec` for workload specifications

### Ephemeral Environments (`types/workload.go`)

Provides types for ephemeral environment lifecycle management:

- `EphemeralEnvironmentSpec` - Configuration for ephemeral environments (trigger type, TTL, auto-teardown)
- `EphemeralEnvironmentStatus` - Status of ephemeral environments (phase, expiration, termination)
- Supports PR/branch triggers with automatic cleanup

### Edge Node Agent (`types/workload.go`)

Provides types for edge node agent configuration:

- `EdgeNodeSpec` - Edge node specification (node ID, IP, capabilities, resources, auth)
- `EdgeNodeStatus` - Edge node status (phase, heartbeat, resource availability, running workloads)
- Enables lightweight binary connections for remote nodes

### Image Caching (`types/workload.go`)

Provides types for image caching layer:

- `ImageCacheConfig` - Image cache configuration (size, limits, TTL, prepull images, mirrors)
- `ImageCacheStatus` - Image cache status (total size, cached images, hit rate, cleanup info)
- Enables faster image pulls by caching across nodes

### Resource Metrics (`types/workload.go`)

Provides types for resource usage metrics:

- `ResourceMetrics` - Comprehensive metrics for workloads (CPU, memory, GPU, network, storage)
- `CPUMetrics` - CPU usage metrics (cores, percentage, requests, limits)
- `MemoryMetrics` - Memory usage metrics (bytes, percentage, cache)
- `GPUMetrics` - GPU metrics (utilization, memory, temperature, power)
- `NetworkMetrics` - Network metrics (throughput, packets, errors)
- `StorageMetrics` - Storage metrics (I/O, disk usage)

---

## Connectivity

| Repo | Relationship |
|---|---|
| `kranix-core` | Imports types, errors, logging, RuntimeDriver interface |
| `kranix-api` | Imports types, errors, auth, proto stubs |
| `kranix-mcp` | Imports types, errors, API client |
| `kranix-cli` | Imports types, errors, Go SDK |
| `kranix-runtime` | Implements RuntimeDriver interface |
| `kranix-operator` | Imports CRD types (re-exported from types package) |
| Third-party tools | Consume Go SDK or TypeScript SDK |

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Changes to any interface in `types/` or `runtime/` must go through a PR review with explicit sign-off from at least two maintainers. All public types must have godoc comments. The SDK must stay in sync with the proto definitions — run `make generate` after any proto change.

## License

Apache 2.0 — see [LICENSE](./LICENSE).
