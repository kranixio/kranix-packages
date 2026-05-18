//! Error types for the Kranix SDK

use thiserror::Error;

/// Result type alias for Kranix operations
pub type Result<T> = std::result::Result<T, KranixError>;

/// Main error type for the Kranix SDK
#[derive(Error, Debug)]
pub enum KranixError {
    /// HTTP request failed
    #[error("HTTP request failed: {0}")]
    HttpError(#[from] reqwest::Error),

    /// JSON serialization/deserialization failed
    #[error("JSON error: {0}")]
    JsonError(#[from] serde_json::Error),

    /// API returned an error
    #[error("API error (status {status}): {message}")]
    ApiError { status: u16, message: String },

    /// Workload not found
    #[error("Workload not found: {0}")]
    WorkloadNotFound(String),

    /// Namespace not found
    #[error("Namespace not found: {0}")]
    NamespaceNotFound(String),

    /// Invalid configuration
    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    /// Authentication failed
    #[error("Authentication failed: {0}")]
    AuthError(String),

    /// Timeout error
    #[error("Operation timed out")]
    Timeout,

    /// Connection error
    #[error("Connection error: {0}")]
    ConnectionError(String),

    /// Invalid input
    #[error("Invalid input: {0}")]
    InvalidInput(String),

    /// SSE stream error
    #[error("SSE stream error: {0}")]
    StreamError(String),
}

impl KranixError {
    /// Create an API error from status code and message
    pub fn api_error(status: u16, message: impl Into<String>) -> Self {
        KranixError::ApiError {
            status,
            message: message.into(),
        }
    }

    /// Check if this is a retryable error
    pub fn is_retryable(&self) -> bool {
        matches!(
            self,
            KranixError::HttpError(_) | KranixError::Timeout | KranixError::ConnectionError(_)
        )
    }
}
