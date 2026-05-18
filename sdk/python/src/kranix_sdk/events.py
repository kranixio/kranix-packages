"""SSE helpers for /api/sse and other event streams."""

from __future__ import annotations

import json
from collections.abc import Callable, Iterator
from dataclasses import dataclass
from typing import Any

import httpx

from kranix_sdk.http_util import bearer_headers


@dataclass
class SSEFrame:
    """One Server-Sent Event frame."""

    event: str
    data: str
    id: str | None = None


def _flush_block(pending: list[str]) -> SSEFrame | None:
    sid = None
    event = ""
    data: list[str] = []
    for line in pending:
        if line.startswith("id:"):
            sid = line[3:].lstrip()
        elif line.startswith("event:"):
            event = line[6:].lstrip()
        elif line.startswith("data:"):
            data.append(line[5:].lstrip())
    if not data and not event and not sid:
        return None
    return SSEFrame(event=event, data="\n".join(data), id=sid)


def iter_sse(response: httpx.Response) -> Iterator[SSEFrame]:
    """Parse an httpx streaming response as SSE frames."""
    pending: list[str] = []
    for line in response.iter_lines():
        line = line.rstrip("\r")
        if line == "":
            frame = _flush_block(pending)
            pending = []
            if frame:
                yield frame
        elif not line.startswith(":"):
            pending.append(line)
    frame = _flush_block(pending)
    if frame:
        yield frame


def subscribe_sse(
    server_url: str,
    api_key: str | None,
    *,
    skip_auth: bool = False,
    client_id: str | None = None,
    namespaces: list[str] | None = None,
    timeout: float | None = None,
) -> Iterator[SSEFrame]:
    """
    Iterate SSE frames from GET /api/sse (blocking). For long-running consumers,
    run inside a thread or use async httpx separately.
    """
    base = server_url.rstrip("/")
    params: list[tuple[str, str]] = []
    if client_id:
        params.append(("client_id", client_id))
    for ns in namespaces or []:
        params.append(("namespace", ns))
    url = f"{base}/api/sse"
    headers = {"Accept": "text/event-stream", **bearer_headers(api_key, skip_auth=skip_auth)}
    with httpx.Client(timeout=timeout or 300.0) as client:
        with client.stream("GET", url, params=params or None, headers=headers) as r:
            r.raise_for_status()
            yield from iter_sse(r)


def workload_event_payload_json(frame: SSEFrame) -> dict[str, Any] | None:
    """If frame.data is JSON object, parse (e.g. workload.changed payloads)."""
    try:
        out = json.loads(frame.data)
        return out if isinstance(out, dict) else None
    except json.JSONDecodeError:
        return None


EventHandler = Callable[[SSEFrame], None]


def subscribe_workload_events_loop(
    server_url: str,
    api_key: str | None,
    handler: EventHandler,
    *,
    skip_auth: bool = False,
    client_id: str | None = None,
    namespaces: list[str] | None = None,
) -> None:
    """Run subscribe_sse until the server closes the connection."""
    for frame in subscribe_sse(
        server_url,
        api_key,
        skip_auth=skip_auth,
        client_id=client_id,
        namespaces=namespaces,
    ):
        handler(frame)
