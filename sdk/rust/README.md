# Kranix Rust SDK

High-performance Rust SDK for interacting with the Kranix platform. This SDK provides a type-safe, async client for workload management, pod operations, and event streaming.

## Installation

Add this to your `Cargo.toml`:

```toml
[dependencies]
kranix-sdk = { version = "0.1", git = "https://github.com/kranix-io/kranix-packages" }
tokio = { version = "1", features = ["full"] }
```

## Usage

### Basic Setup

```rust
use kranix_sdk::{KranixClient, ClientConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = ClientConfig::new("http://localhost:8080")
        .with_api_key("your-api-key")
        .with_timeout(std::time::Duration::from_secs(30));

    let client = KranixClient::new(config)?;

    // Use the client...
    Ok(())
}
```

### Deploying a Workload

```rust
use kranix_sdk::types::WorkloadSpec;

let spec = WorkloadSpec {
    name: "my-app".to_string(),
    image: "nginx:latest".to_string(),
    replicas: 2,
    backend: "docker".to_string(),
    namespace: Some("default".to_string()),
    ..Default::default()
};

let workload = client.deploy_workload(&spec, "default").await?;
println!("Deployed workload: {}", workload.id);
```

### Listing Workloads

```rust
let workloads = client.list_workloads("default", None, Some(100)).await?;
for workload in workloads {
    println!("{}: {}", workload.name, workload.phase);
}
```

### Streaming Pod Logs

```rust
use futures::StreamExt;

let mut log_stream = client.stream_pod_logs("pod-123", "default", true, Some(100)).await?;

while let Some(log_line) = log_stream.next().await {
    match log_line {
        Ok(line) => println!("{}", line),
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

### Subscribing to Workload Events

```rust
use futures::StreamExt;

let mut event_stream = client
    .subscribe_workload_events(vec!["default".to_string()], vec![])
    .await?;

while let Some(event) = event_stream.next().await {
    match event {
        Ok(event) => println!("Event: {} - Workload: {}", event.event_type, event.workload.name),
        Err(e) => eprintln!("Error: {}", e),
    }
}
```

### Namespace Operations

```rust
// Create a namespace
let namespace = client.create_namespace("my-namespace").await?;

// List namespaces
let namespaces = client.list_namespaces().await?;

// Get namespace details
let namespace = client.get_namespace("my-namespace").await?;

// Delete a namespace
client.delete_namespace("my-namespace").await?;
```

## Features

- **Type-safe**: Full type definitions for all Kranix API objects
- **Async**: Built on Tokio for high-performance async operations
- **Event Streaming**: Support for SSE-based event streaming
- **Error Handling**: Comprehensive error types with retry support
- **Configurable**: Flexible configuration options for timeouts, retries, etc.

## Configuration Options

```rust
let config = ClientConfig::new("http://localhost:8080")
    .with_api_key("your-api-key")           // Set API key
    .with_skip_auth(false)                  // Enable/disable auth skip
    .with_timeout(Duration::from_secs(30))  // Set request timeout
    .with_max_retries(3)                    // Set max retries
    .with_retry_delay(Duration::from_millis(500)); // Set retry delay
```

## Error Handling

```rust
use kranix_sdk::KranixError;

match client.get_workload("invalid-id", "default").await {
    Ok(workload) => println!("{:?}", workload),
    Err(KranixError::WorkloadNotFound(id)) => eprintln!("Workload not found: {}", id),
    Err(KranixError::ApiError { status, message }) => {
        eprintln!("API error {}: {}", status, message);
    }
    Err(e) => eprintln!("Error: {}", e),
}
```

## Development

Run tests:

```bash
cd sdk/rust
cargo test
```

Run examples:

```bash
cargo run --example basic
```

## License

Apache 2.0
