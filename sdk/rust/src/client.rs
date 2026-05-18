//! Kranix HTTP client for interacting with the Kranix API

use crate::config::ClientConfig;
use crate::error::{KranixError, Result};
use crate::types::*;
use eventsource_client as es;
use futures::stream::StreamExt;
use reqwest::header::{AUTHORIZATION, CONTENT_TYPE, USER_AGENT};
use serde::de::DeserializeOwned;
use std::time::Duration;

/// Main client for interacting with the Kranix API
pub struct KranixClient {
    config: ClientConfig,
    http_client: reqwest::Client,
}

impl KranixClient {
    /// Create a new Kranix client with the given configuration
    pub fn new(config: ClientConfig) -> Result<Self> {
        config.validate()?;

        let http_client = reqwest::Client::builder()
            .timeout(config.timeout)
            .user_agent(&config.user_agent)
            .build()?;

        Ok(Self { config, http_client })
    }

    /// Create a new client with default configuration
    pub fn with_url(server_url: impl Into<String>) -> Result<Self> {
        Self::new(ClientConfig::new(server_url))
    }

    /// Build the base URL for API requests
    fn base_url(&self) -> String {
        self.config.server_url.clone()
    }

    /// Add authentication headers to the request
    fn add_auth_headers(&self, builder: reqwest::RequestBuilder) -> reqwest::RequestBuilder {
        if !self.config.skip_auth {
            if let Some(api_key) = &self.config.api_key {
                return builder.header(AUTHORIZATION, format!("Bearer {}", api_key));
            }
        }
        builder
    }

    /// Make a GET request
    async fn get<T: DeserializeOwned>(&self, path: &str) -> Result<T> {
        let url = format!("{}{}", self.base_url(), path);
        let response = self
            .add_auth_headers(self.http_client.get(&url))
            .header(CONTENT_TYPE, "application/json")
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Make a POST request
    async fn post<T: DeserializeOwned>(&self, path: &str, body: impl serde::Serialize) -> Result<T> {
        let url = format!("{}{}", self.base_url(), path);
        let response = self
            .add_auth_headers(self.http_client.post(&url))
            .header(CONTENT_TYPE, "application/json")
            .json(&body)
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Make a DELETE request
    async fn delete(&self, path: &str) -> Result<()> {
        let url = format!("{}{}", self.base_url(), path);
        let response = self
            .add_auth_headers(self.http_client.delete(&url))
            .header(CONTENT_TYPE, "application/json")
            .send()
            .await?;

        self.handle_response(response).await
    }

    /// Handle HTTP response
    async fn handle_response<T: DeserializeOwned>(&self, response: reqwest::Response) -> Result<T> {
        let status = response.status();

        if status.is_success() {
            response.json::<T>().await.map_err(KranixError::from)
        } else {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            Err(KranixError::api_error(status.as_u16(), error_text))
        }
    }
}

/// Workload operations
impl KranixClient {
    /// Deploy a new workload
    pub async fn deploy_workload(
        &self,
        spec: &WorkloadSpec,
        namespace: &str,
    ) -> Result<Workload> {
        let request = serde_json::json!({
            "spec": spec,
            "namespace": namespace,
        });

        #[derive(Deserialize)]
        struct Response {
            workload: Workload,
        }

        let response: Response = self.post("/api/v1/workloads", request).await?;
        Ok(response.workload)
    }

    /// Get a workload by ID
    pub async fn get_workload(&self, id: &str, namespace: &str) -> Result<Workload> {
        let path = format!("/api/v1/workloads/{}?namespace={}", id, namespace);
        self.get(&path).await
    }

    /// List workloads in a namespace
    pub async fn list_workloads(
        &self,
        namespace: &str,
        label_selector: Option<&str>,
        limit: Option<u32>,
    ) -> Result<Vec<Workload>> {
        let mut path = format!("/api/v1/workloads?namespace={}", namespace);
        if let Some(selector) = label_selector {
            path.push_str(&format!("&labelSelector={}", selector));
        }
        if let Some(lim) = limit {
            path.push_str(&format!("&limit={}", lim));
        }

        #[derive(Deserialize)]
        struct Response {
            workloads: Vec<Workload>,
        }

        let response: Response = self.get(&path).await?;
        Ok(response.workloads)
    }

    /// Delete a workload
    pub async fn delete_workload(&self, id: &str, namespace: &str) -> Result<()> {
        let path = format!("/api/v1/workloads/{}?namespace={}", id, namespace);
        self.delete(&path).await
    }

    /// Restart a workload
    pub async fn restart_workload(&self, id: &str, namespace: &str) -> Result<()> {
        let path = format!("/api/v1/workloads/{}/restart?namespace={}", id, namespace);
        self.post(&path, serde_json::json!({})).await
    }

    /// Subscribe to workload events via SSE
    pub async fn subscribe_workload_events(
        &self,
        namespaces: Vec<String>,
        workload_ids: Vec<String>,
    ) -> Result<impl futures::Stream<Item = Result<WorkloadEvent>>> {
        let mut url = format!("{}/api/v1/events", self.base_url());
        let mut params = Vec::new();

        for ns in &namespaces {
            params.push(format!("namespace={}", ns));
        }
        for wid in &workload_ids {
            params.push(format!("workloadId={}", wid));
        }

        if !params.is_empty() {
            url.push('?');
            url.push_str(&params.join("&"));
        }

        let client = es::ClientBuilder::for_url(&url)?
            .build()
            .map_err(|e| KranixError::StreamError(e.to_string()))?;

        Ok(client
            .stream()
            .map(|event| match event {
                Ok(es_event) => {
                    let event_type = es_event.event_type;
                    let data = es_event.data;

                    // Parse the workload from the event data
                    let workload: Workload = serde_json::from_str(&data)
                        .map_err(|e| KranixError::JsonError(e))?;

                    Ok(WorkloadEvent {
                        event_type,
                        workload,
                        timestamp: chrono::Utc::now(),
                    })
                }
                Err(e) => Err(KranixError::StreamError(e.to_string())),
            }))
    }
}

/// Namespace operations
impl KranixClient {
    /// Create a namespace
    pub async fn create_namespace(&self, name: &str) -> Result<Namespace> {
        let request = serde_json::json!({
            "name": name,
        });

        #[derive(Deserialize)]
        struct Response {
            namespace: Namespace,
        }

        let response: Response = self.post("/api/v1/namespaces", request).await?;
        Ok(response.namespace)
    }

    /// Get a namespace by name
    pub async fn get_namespace(&self, name: &str) -> Result<Namespace> {
        let path = format!("/api/v1/namespaces/{}", name);
        self.get(&path).await
    }

    /// List all namespaces
    pub async fn list_namespaces(&self) -> Result<Vec<Namespace>> {
        #[derive(Deserialize)]
        struct Response {
            namespaces: Vec<Namespace>,
        }

        let response: Response = self.get("/api/v1/namespaces").await?;
        Ok(response.namespaces)
    }

    /// Delete a namespace
    pub async fn delete_namespace(&self, name: &str) -> Result<()> {
        let path = format!("/api/v1/namespaces/{}", name);
        self.delete(&path).await
    }
}

/// Pod operations
impl KranixClient {
    /// Stream pod logs
    pub async fn stream_pod_logs(
        &self,
        pod_id: &str,
        namespace: &str,
        follow: bool,
        tail_lines: Option<u32>,
    ) -> Result<impl futures::Stream<Item = Result<String>>> {
        let mut path = format!(
            "/api/v1/pods/{}/logs?namespace={}&follow={}",
            pod_id, namespace, follow
        );

        if let Some(tail) = tail_lines {
            path.push_str(&format!("&tailLines={}", tail));
        }

        let response = self
            .add_auth_headers(self.http_client.get(&path))
            .header(CONTENT_TYPE, "application/json")
            .send()
            .await?;

        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            return Err(KranixError::api_error(response.status().as_u16(), error_text));
        }

        let stream = response.bytes_stream();
        Ok(stream
            .map(|result| result.map_err(KranixError::from))
            .map(|bytes_result| {
                bytes_result.and_then(|bytes| {
                    String::from_utf8(bytes.to_vec())
                        .map_err(|e| KranixError::StreamError(e.to_string()))
                })
            }))
    }

    /// List pods in a namespace
    pub async fn list_pods(&self, namespace: &str, workload_id: Option<&str>) -> Result<Vec<Pod>> {
        let mut path = format!("/api/v1/pods?namespace={}", namespace);
        if let Some(wid) = workload_id {
            path.push_str(&format!("&workloadId={}", wid));
        }

        #[derive(Deserialize)]
        struct Response {
            pods: Vec<Pod>,
        }

        let response: Response = self.get(&path).await?;
        Ok(response.pods)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_client_creation() {
        let config = ClientConfig::new("http://localhost:8080").with_skip_auth(true);
        let client = KranixClient::new(config);
        assert!(client.is_ok());
    }

    #[test]
    fn test_invalid_config() {
        let config = ClientConfig {
            server_url: "".to_string(),
            ..Default::default()
        };
        let client = KranixClient::new(config);
        assert!(client.is_err());
    }
}
