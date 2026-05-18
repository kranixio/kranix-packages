"""Shared HTTP header helpers."""

from __future__ import annotations


def bearer_headers(api_key: str | None, *, skip_auth: bool) -> dict[str, str]:
    if skip_auth:
        return {}
    if not api_key:
        raise ValueError("api_key is required unless skip_auth is True")
    return {"Authorization": f"Bearer {api_key}"}
