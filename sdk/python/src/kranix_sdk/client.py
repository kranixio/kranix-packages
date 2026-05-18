"""Sync HTTP client for kranix-api."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any, cast
from urllib.parse import quote

import httpx

from kranix_sdk.events import SSEFrame, iter_sse
from kranix_sdk.http_util import bearer_headers


@dataclass
class Config:
    server_url: str
    api_key: str | None = None
    timeout: float = 60.0
    skip_auth: bool = False

    def headers_json(self) -> dict[str, str]:
        h = {"Accept": "application/json", **bearer_headers(self.api_key, skip_auth=self.skip_auth)}
        return h


class KraneClient:
    def __init__(
        self,
        server_url: str | None = None,
        api_key: str | None = None,
        *,
        config: Config | None = None,
        timeout: float = 60.0,
        skip_auth: bool = False,
    ) -> None:
        if config is not None:
            self._cfg = config
        elif server_url:
            self._cfg = Config(
                server_url=server_url,
                api_key=api_key,
                timeout=timeout,
                skip_auth=skip_auth,
            )
        else:
            raise ValueError("server_url is required (or pass config=)")
        if not self._cfg.server_url:
            raise ValueError("server_url is required")
        if not self._cfg.skip_auth and not self._cfg.api_key:
            raise ValueError("api_key is required unless skip_auth is True")
        self._base = self._cfg.server_url.rstrip("/")

    def _url(self, path: str) -> str:
        return f"{self._base}{path}"

    def _request(
        self,
        method: str,
        path: str,
        *,
        json_body: Any | None = None,
        params: list[tuple[str, str]] | None = None,
    ) -> Any:
        headers = self._cfg.headers_json()
        if json_body is not None:
            headers = {**headers, "Content-Type": "application/json"}
        with httpx.Client(timeout=self._cfg.timeout) as client:
            r = client.request(
                method,
                self._url(path),
                headers=headers,
                json=json_body,
                params=params,
            )
            if r.status_code == 204:
                return None
            r.raise_for_status()
            return r.json()

    @property
    def workloads(self) -> Workloads:
        return Workloads(self)

    @property
    def pods(self) -> Pods:
        return Pods(self)

    @property
    def namespaces(self) -> Namespaces:
        return Namespaces(self)

    def subscribe_workload_events(
        self,
        *,
        client_id: str | None = None,
        namespaces: list[str] | None = None,
    ) -> Iterator[SSEFrame]:
        params: list[tuple[str, str]] = []
        if client_id:
            params.append(("client_id", client_id))
        for ns in namespaces or []:
            params.append(("namespace", ns))
        headers = {"Accept": "text/event-stream", **bearer_headers(self._cfg.api_key, skip_auth=self._cfg.skip_auth)}
        with httpx.Client(timeout=300.0) as client:
            with client.stream(
                "GET",
                self._url("/api/sse"),
                params=params or None,
                headers=headers,
            ) as r:
                r.raise_for_status()
                yield from iter_sse(r)


class Workloads:
    def __init__(self, c: KraneClient) -> None:
        self._c = c

    def deploy(self, spec: dict[str, Any]) -> dict[str, Any]:
        body = self._c._request("POST", "/api/v1/workloads", json_body=spec)
        b = cast(dict[str, Any], body)
        if not b.get("id"):
            raise ValueError(
                "deploy response missing id (is kranix-api stubbed? use kranix-mock-api for tests)"
            )
        return b

    def get(self, workload_id: str) -> dict[str, Any]:
        return cast(dict[str, Any], self._c._request("GET", f"/api/v1/workloads/{quote(workload_id, safe='')}"))

    def list(self, namespace: str | None = None) -> list[dict[str, Any]]:
        params: list[tuple[str, str]] = []
        if namespace:
            params.append(("namespace", namespace))
        raw = self._c._request("GET", "/api/v1/workloads", params=params or None)
        if isinstance(raw, list):
            return cast(list[dict[str, Any]], raw)
        if isinstance(raw, dict) and "workloads" in raw:
            return cast(list[dict[str, Any]], raw["workloads"])
        raise ValueError(f"unexpected list workloads response: {raw!r}")

    def update(self, workload_id: str, spec: dict[str, Any]) -> dict[str, Any]:
        return cast(
            dict[str, Any],
            self._c._request("PATCH", f"/api/v1/workloads/{quote(workload_id, safe='')}", json_body=spec),
        )

    def delete(self, workload_id: str) -> None:
        self._c._request("DELETE", f"/api/v1/workloads/{quote(workload_id, safe='')}")

    def restart(self, workload_id: str) -> None:
        self._c._request(
            "POST",
            f"/api/v1/workloads/{quote(workload_id, safe='')}/restart",
            json_body={},
        )

    def list_pods(self, workload_id: str) -> list[dict[str, Any]]:
        j = cast(
            dict[str, Any],
            self._c._request("GET", f"/api/v1/workloads/{quote(workload_id, safe='')}/pods"),
        )
        return cast(list[dict[str, Any]], j.get("pods") or [])

    def analyze(self, workload_id: str) -> dict[str, Any]:
        return cast(
            dict[str, Any],
            self._c._request("GET", f"/api/v1/workloads/{quote(workload_id, safe='')}/analyze"),
        )


class Pods:
    def __init__(self, c: KraneClient) -> None:
        self._c = c

    def stream_logs(
        self,
        pod_id: str,
        *,
        follow: bool = True,
        tail: int | None = None,
        since: int | None = None,
    ) -> Iterator[str]:
        q: list[tuple[str, str]] = []
        if follow:
            q.append(("follow", "true"))
        if tail is not None:
            q.append(("tail", str(tail)))
        if since is not None:
            q.append(("since", str(since)))
        path = f"/api/v1/pods/{quote(pod_id, safe='')}/logs"
        headers = {"Accept": "text/event-stream", **bearer_headers(self._c._cfg.api_key, skip_auth=self._c._cfg.skip_auth)}
        with httpx.Client(timeout=300.0) as client:
            with client.stream("GET", self._c._url(path), params=q or None, headers=headers) as r:
                r.raise_for_status()
                for frame in iter_sse(r):
                    if frame.event == "log" and frame.data.strip():
                        yield frame.data.strip()


class Namespaces:
    def __init__(self, c: KraneClient) -> None:
        self._c = c

    def get(self, name: str) -> dict[str, Any]:
        return cast(dict[str, Any], self._c._request("GET", f"/api/v1/namespaces/{quote(name, safe='')}"))

    def list(self) -> list[dict[str, Any]]:
        j = cast(dict[str, Any], self._c._request("GET", "/api/v1/namespaces"))
        return cast(list[dict[str, Any]], j.get("namespaces") or [])

    def create(self, body: dict[str, Any]) -> dict[str, Any]:
        return cast(dict[str, Any], self._c._request("POST", "/api/v1/namespaces", json_body=body))
