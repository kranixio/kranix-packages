//! Configuration for the Kranix client

use crate::error::{KranixError, Result};
use serde::{Deserialize, Serialize};
use std::time::Duration;

/// Configuration for the Kranix client
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClientConfig {
    /// Server URL (e.g., "http://localhost:8080")
    pub server_url: String,

    /// API key for authentication
    pub api_key: Option<String>,

    /// Skip authentication (for testing with mock API)
    pub skip_auth: bool,

    /// Request timeout
    pub timeout: Duration,

    /// Maximum number of retries
    pub max_retries: u32,

    /// Retry delay
    pub retry_delay: Duration,

    /// User agent string
    pub user_agent: String,
}

impl Default for ClientConfig {
    fn default() -> Self {
        Self {
            server_url: "http://localhost:8080".to_string(),
            api_key: None,
            skip_auth: true,
            timeout: Duration::from_secs(30),
            max_retries: 3,
            retry_delay: Duration::from_millis(500),
            user_agent: format!("kranix-rust-sdk/{}", env!("CARGO_PKG_VERSION")),
        }
    }
}

impl ClientConfig {
    /// Create a new client configuration
    pub fn new(server_url: impl Into<String>) -> Self {
        Self {
            server_url: server_url.into(),
            ..Default::default()
        }
    }

    /// Set the API key
    pub fn with_api_key(mut self, api_key: impl Into<String>) -> Self {
        self.api_key = Some(api_key.into());
        self.skip_auth = false;
        self
    }

    /// Enable/disable authentication skip
    pub fn with_skip_auth(mut self, skip_auth: bool) -> Self {
        self.skip_auth = skip_auth;
        self
    }

    /// Set the request timeout
    pub fn with_timeout(mut self, timeout: Duration) -> Self {
        self.timeout = timeout;
        self
    }

    /// Set the maximum number of retries
    pub fn with_max_retries(mut self, max_retries: u32) -> Self {
        self.max_retries = max_retries;
        self
    }

    /// Set the retry delay
    pub fn with_retry_delay(mut self, retry_delay: Duration) -> Self {
        self.retry_delay = retry_delay;
        self
    }

    /// Validate the configuration
    pub fn validate(&self) -> Result<()> {
        if self.server_url.is_empty() {
            return Err(KranixError::InvalidConfig("server_url is required".into()));
        }

        if !self.skip_auth && self.api_key.is_none() {
            return Err(KranixError::InvalidConfig(
                "api_key is required when skip_auth is false".into(),
            ));
        }

        // Validate URL format
        if let Err(e) = url::Url::parse(&self.server_url) {
            return Err(KranixError::InvalidConfig(format!(
                "Invalid server_url: {}",
                e
            )));
        }

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = ClientConfig::default();
        assert_eq!(config.server_url, "http://localhost:8080");
        assert!(config.skip_auth);
    }

    #[test]
    fn test_config_builder() {
        let config = ClientConfig::new("http://example.com")
            .with_api_key("test-key")
            .with_timeout(Duration::from_secs(60))
            .with_max_retries(5);

        assert_eq!(config.server_url, "http://example.com");
        assert_eq!(config.api_key, Some("test-key".to_string()));
        assert!(!config.skip_auth);
        assert_eq!(config.timeout, Duration::from_secs(60));
        assert_eq!(config.max_retries, 5);
    }

    #[test]
    fn test_config_validation() {
        let config = ClientConfig::default();
        assert!(config.validate().is_ok());

        let invalid_config = ClientConfig {
            server_url: "".to_string(),
            ..Default::default()
        };
        assert!(invalid_config.validate().is_err());
    }
}
