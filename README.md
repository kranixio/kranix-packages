# kranix-packages

> Shared SDK — types, utilities, and client libraries for the Kranix ecosystem.

`kranix-packages` is the common foundation imported by every other Kranix repo and by third-party tools building on top of Kranix. It contains shared domain types, the `RuntimeDriver` interface, error codes, logging utilities, config schemas, auth primitives, and public SDK clients for **Go**, **TypeScript**, and **Python** consumers. A **`kranix-mock-api`** binary provides a faithful in-memory API + SSE surface for unit and integration tests in any language.

All cross-cutting concerns that are needed by more than one repo live here. Nothing here contains business logic — that belongs in `kranix-core`.

---

## What it contains

| Package | Description |
|---|---|
| `types` | Core domain types (Workload, Pod, Namespace, Status, quotas, cron fields, …) |
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
| `sdk/go` | Public Go client for the kranix-api (REST + SSE event subscription) |
| `sdk/typescript` | Public TypeScript/Node.js client for the kranix-api (REST + SSE) |
| `sdk/python` | Public Python client (`kranix-io-sdk` on PyPI layout) for ML / data workflows |
| `sdk/rust` | Public Rust client for high-performance tooling on top of Kranix |
| `cmd/kranix-mock-api` | Local mock HTTP server mirroring core REST + `/api/sse` for tests |
| `cmd/cli-lib` | Shared CLI helper library for flag parsing and output formatting |
| `proto` | Versioned protobuf definitions with breaking change detection in CI |

---

## Architecture position

```
kranix-core    ──┐
kranix-api     ──┤
kranix-mcp     ──┼──►  kranix-packages
kranix-cli     ──┤
kranix-runtime ──┘

Third-party tools  ──►  kranix-packages (Go / TS / Python SDK + mock API)
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
    Backend   string            `json:"backend"`    // docker | kubernetes | podman | compose | remote
    RemoteHost string           `json:"remoteHost,omitempty"` // For remote SSH backend
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
    SkipAuth:  true, // for kranix-mock-api default
})

// Deploy a workload
workload, err := client.Workloads().Deploy(ctx, &types.WorkloadSpec{
    Name:      "my-app",
    Image:     "nginx:latest",
    Namespace: "staging",
    Replicas:  2,
    Backend:   "docker",
})

// Stream logs
logCh, err := client.Pods().StreamLogs(ctx, podID, &types.LogOptions{
    Follow: true,
    Tail:   100,
})
for line := range logCh {
    fmt.Println(line)
}

// Analyze a workload
analysis, err := client.Workloads().Analyze(ctx, workload.ID)
fmt.Println(analysis.ProbableFix)

// Live events (GET /api/sse) — workload.changed, workload.deleted, ...
_ = client.SubscribeSSE(ctx, &kraneclient.SubscribeOptions{
    Namespaces: []string{"staging"},
}, func(ev kraneclient.SSEEvent) error {
    fmt.Println(ev.Event, string(ev.Data))
    return nil
})
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
  skipAuth: true,
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

// Subscribe to workload / platform events (GET /api/sse)
for await (const frame of client.subscribeWorkloadEvents({
  namespaces: ["staging"],
})) {
  if (frame.event === "workload.changed") {
    console.log(frame.data);
  }
}
```

### Python SDK

```bash
cd sdk/python && pip install .
### Rust SDK

```bash
cd sdk/rust
cargo build
cargo test
```

Add to your `Cargo.toml`:

```toml
[dependencies]
kranix-sdk = { version = "0.1", git = "https://github.com/kranix-io/kranix-packages" }
tokio = { version = "1", features = ["full"] }
```

```

```python
from kranix_sdk import KraneClient
from kranix_sdk.events import subscribe_sse, workload_event_payload_json

client = KraneClient(
    "http://localhost:8080",
    api_key="krane_your_key",
    skip_auth=True,  # use with kranix-mock-api -skip-auth (default)
)

wl = client.workloads.deploy({
    "name": "trainer",
    "image": "pytorch/pytorch:latest",
    "namespace": "default",
    "replicas": 1,
    "backend": "docker",
})

for line in client.pods.stream_logs("pod-1", follow=True, tail=100):
    print(line)

for frame in subscribe_sse(
    "http://localhost:8080",
    None,
    skip_auth=True,
    namespaces=["default"],
):
    payload = workload_event_payload_json(frame)
    if payload:
        print(frame.event, payload)
```

### Mock API server (`kranix-mock-api`)

Run a local process that implements the same URL shapes and JSON models as [kranix-api](../kranix-api) for workloads, namespaces, pod log SSE, `/api/sse` broadcasts, **incident runbooks** (`/api/v1/incident/*`, including `POST .../runbooks/{id}/execute` with a seeded `rb-oncall-pagerduty` playbook), **analytics** (`POST /api/v1/analytics/metrics`, `GET .../analytics/workloads/{id}?type=latency` for latency percentiles), and **cost** (`GET /api/v1/cost/summary`, `GET /api/v1/workloads/{id}/cost` with mock **rightsizing** hints). Point any SDK at `http://localhost:8080` (or any `-addr`).

```bash
go run ./cmd/kranix-mock-api -addr :8080 -skip-auth=true
# Production-shaped auth check:
# KRANIX_MOCK_REQUIRE_AUTH=1 go run ./cmd/kranix-mock-api -skip-auth=false
# then pass a krane_* API key in the Authorization header from the SDK.
```

Environment hints:

| Variable | Effect |
|----------|--------|
| `KRANIX_MOCK_ADDR` | Overrides `-addr` listen address |
| `KRANIX_MOCK_REQUIRE_AUTH=1` | Forces `-skip-auth=false` (require `Bearer krane_*`) |

### Event subscription (SSE)

All first-class SDKs can consume `GET /api/sse` using the same query parameters as kranix-api (`client_id`, repeated `namespace`). Incoming frames use named events such as `connected`, `workload.changed`, and `workload.deleted`. Integration tests can also `POST /api/sse/broadcast` (see kranix-api docs) to inject synthetic events.

| SDK | API |
|-----|-----|
| Go | `client.SubscribeSSE(ctx, &SubscribeOptions{...}, handler)` |
| TypeScript | `for await (const f of client.subscribeWorkloadEvents({...}))` |
| Python | `subscribe_sse(...)` or `KraneClient.subscribe_workload_events()` |

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

```
kranix-packages/
├── cmd/
│   ├── kranix-mock-api/   # Mock HTTP API for local / CI tests
│   └── cli-lib/           # Shared CLI helper library
├── types/                  # Core domain types
│   ├── workload.go
│   ├── pod.go
│   └── ...
├── errors/
├── logging/
├── config/
├── auth/
├── runtime/
├── proto/                  # Versioned protobuf definitions
│   ├── v1/                # Proto version 1
│   ├── buf.yaml           # Buf configuration
│   └── Makefile           # Proto generation
└── sdk/
    ├── go/
    ├── typescript/
    ├── python/
    └── rust/              # Rust SDK for high-performance tooling
```

### Older reference (domain type files)

```
types/
├── ratelimit.go
├── sse.go
├── apiversion.go
├── analytics.go
├── version.go
├── webhook.go
└── ...
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

### Cron schedules (`types/workload.go`)

- **`CronSchedule`** on **`WorkloadSpec`** (`cronSchedule` in JSON): `schedule` (standard 5-field cron), `suspended`, `timeZone`, `concurrencyPolicy` (`allow` | `forbid` | `replace`)
- **`Cron`** on **`WorkloadStatus`**: optional `lastScheduleTime` — aligned with **`kranix-core`** reconcile triggers and **`kranix-runtime`** Kubernetes **CronJob** mapping

### Hard aggregate quotas (`types/quota.go`)

- **`HardResourceQuota`** — optional caps on total CPU/memory *requests*, workload count, and replica count keyed by **`namespace`** or **`teamId`** (matches team label `kranix.io/team` semantics in core). JSON field names are camelCase (`maxCpuRequests`, `maxReplicasTotal`). Use this type in API payloads and SDKs; **protobuf** equivalents live under `proto/v1/workload.proto` (`HardResourceQuota`).

### Scheduling — priority & preemption (`types/workload.go`)

- **`SchedulingConfig`** (`scheduling` on **`WorkloadSpec`**, camelCase JSON): **`workloadPriority`** — `critical` | `high` | `normal` | `low`; **`preemptionEnabled`** requests PriorityClasses that can preempt lower-priority pods (**runtime** selects `kranix-*` vs `kranix-*-np` name suffix); **`priorityClassName`** overrides automatic mapping.

### Spot / preemptible workloads (`types/workload.go`)

- **`SchedulingConfig.spot`**: **`enabled`**, **`rescheduleOnNodeTermination`** — surfaced to **`kranix-runtime`** Kubernetes driver for tolerations and faster reschedule after node interruptions.

### Cross-namespace traffic (`types/workload.go`)

- **`CrossNamespaceTrafficPolicy`** on **`WorkloadSpec`** (`crossNamespaceTraffic`): **`enabled`**, **`allowedIngressNamespaces`**, **`allowedEgressNamespaces`**, **`allowSameNamespace`**, **`blockClusterDNS`**, **`allowEgressInternet`** — consumed when creating **NetworkPolicies** on the Kubernetes backend.

### Secret rotation (`types/secrets.go`)

- **`SecretRotationSpec`** / **`SecretRotationStatus`** on workload spec/status — link secrets via **`secretRefs`**; core triggers rolling restarts on version change (`POST /api/v1/secrets/rotated` on core HTTP API).

### Pagination & changelog notifications (`pagination/cursor.go`, `types/pagination.go`, `types/changelog_notify.go`)

- **Cursor pagination** — shared `kranix-packages/pagination` package; `limit` + `cursor` on `GET /workloads` and `GET /changelog`.
- **`ChangelogSubscription`** — webhook/email alerts for breaking API releases; event `changelog.breaking` on the webhook system.

### Diff, search, and quota APIs (`types/diff.go`, `types/search.go`, `types/quota.go`)

- **`WorkloadDiffResult`** — desired vs live field changes for `GET /workloads/:id/diff`.
- **`WorkloadSearchQuery`** — filters for `GET /workloads` (`namespace`, `phase`, `image`, tags).
- **`ResourceQuotaUsage`** / **`HardResourceQuota`** — namespace limits and usage via `/api/v1/quotas`.

### Bulk & audit (`types/bulk.go`, `types/audit.go`)

- **`BulkWorkloadRequest`** / **`BulkWorkloadResponse`** — batch **deploy**, **restart**, or **delete** via **`POST /api/v1/workloads/bulk`** on kranix-api.
- **`AuditEntry`** / **`AuditQuery`** — API audit trail; combine with core domain events via **`GET /api/v1/audit/resources/{type}/{id}`**.

### Resilience — circuit breaker & warm standby (`types/resilience.go`)

- **`CircuitBreakerSpec`** / **`CircuitBreakerStatus`** on workload spec/status (`circuitBreaker`): failure/success thresholds, open duration, half-open probe limits, optional **`tripOnDegraded`**. States: `closed`, `open`, `half-open`.
- **`WarmStandbySpec`** / **`WarmStandbyStatus`** (`warmStandby`): cold standby replica count, **`autoPromote`**, optional **`standbyWorkloadId`**. Standby workloads use labels **`kranix.io/standby-for`** and **`kranix.io/role=standby`** (see `types/tags.go`).

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

### Versioned Proto Contracts (`proto/`)

Provides versioned protobuf definitions with automated breaking change detection:

- **v1/** - Proto version 1 definitions for workloads, namespaces, and API services
- **buf.yaml** - Configuration for buf (protobuf linter and breaking change detector)
- **buf.gen.yaml** - Configuration for generating Go and TypeScript stubs
- **Makefile** - Commands for generating, formatting, and linting proto files
- **CI Integration** - Automated breaking change detection on pull requests
, CLI helper library
Usage:

```bash, or Rust
cd proto
make generate    # Generate Go and TypeScript stubs
make format      # Format proto files
make lint        # Lint proto files
make check-breaking  # Check for breaking changes against main branch
```

### Rust SDK (`sdk/rust/`)

High-performance Rust SDK for teams building tooling on top of Kranix:

- **Async client** built on Tokio for high-performance operations
- **Type-safe** full type definitions for all Kranix API objects
- **Event streaming** support for SSE-based workload events
- **Comprehensive error handling** with retry support
- **Configurable** options for timeouts, retries, and authentication

See `sdk/rust/README.md` for detailed usage examples.

### CLI Helper Library (`cmd/cli-lib/`)

Shared CLI utilities for any Kranix CLI tool:

- **Common flag parsing** for server URL, API key, namespace, output format, etc.
- **Output formatting** support for table, JSON, and YAML
- **Configuration management** with context-based configuration
- **Interactive prompts** for confirmation dialogs
- **Progress tracking** for long-running operations
- **Table building** utilities for formatted output

See `cmd/cli-lib/README.md` for detailed usage examples.

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

### Remote SSH Backend (`types/workload.go`)

Provides support for remote SSH backend:

- `RemoteHost` field in `WorkloadSpec` - Specifies remote host for SSH-based deployment
- Enables agentless connections to bare metal servers
- Supports both Docker and Podman runtimes on remote hosts

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
| Third-party tools | Consume Go, TypeScript, or Python SDK; use `kranix-mock-api` in CI |

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Changes to any interface in `types/` or `runtime/` must go through a PR review with explicit sign-off from at least two maintainers. All public types must have godoc comments. The SDK must stay in sync with the proto definitions — run `make generate` after any proto change.

## License

Apache 2.0 — see [LICENSE](./LICENSE).
