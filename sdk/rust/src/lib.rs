//! # Kranix Rust SDK
//!
//! High-performance Rust SDK for interacting with the Kranix platform.
//! Provides a type-safe, async client for workload management, pod operations,
//! and event streaming.

pub mod client;
pub mod types;
pub mod error;
pub mod config;

pub use client::KranixClient;
pub use config::ClientConfig;
pub use error::{KranixError, Result};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_version() {
        assert_eq!(env!("CARGO_PKG_VERSION"), env!("CARGO_PKG_VERSION"));
    }
}
