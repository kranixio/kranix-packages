// TypeScript SDK for Kranix API

export interface Config {
  serverUrl: string;
  apiKey?: string;
  timeout?: number; // in seconds
  /** When true, omit Authorization (use with kranix-mock-api -skip-auth). */
  skipAuth?: boolean;
}

export interface WorkloadSpec {
  name: string;
  image: string;
  namespace: string;
  replicas: number;
  env?: Record<string, string>;
  resources?: ResourceSpec;
  ports?: PortSpec[];
  backend: string;
  remoteHost?: string;
  [k: string]: unknown;
}

export interface ResourceSpec {
  cpuRequest?: string;
  cpuLimit?: string;
  memoryRequest?: string;
  memoryLimit?: string;
}

export interface PortSpec {
  name?: string;
  containerPort: number;
  protocol?: string;
}

export interface WorkloadStatus {
  phase: string;
  readyReplicas: number;
  message?: string;
  lastUpdated: string;
  state?: string;
}

export interface Workload {
  id: string;
  name: string;
  namespace: string;
  spec: WorkloadSpec;
  status: WorkloadStatus;
  createdAt: string;
  updatedAt: string;
  labels?: Record<string, string>;
}

export interface LogOptions {
  follow: boolean;
  tail?: number;
  since?: number;
}

export interface AnalysisResult {
  workloadId: string;
  status: string;
  issues?: Issue[];
  probableFix?: string;
  analyzedAt: string;
}

export interface Issue {
  severity: string;
  type: string;
  message: string;
  field?: string;
}

export interface Pod {
  id: string;
  name: string;
  namespace: string;
  workloadId: string;
  phase: string;
  createdAt?: string;
}

export interface Namespace {
  name: string;
  displayName?: string;
  description?: string;
  createdAt: string;
  updatedAt: string;
  labels?: Record<string, string>;
  status: { phase: string; message?: string };
}

export interface SSEFrame {
  id?: string;
  event: string;
  data: string;
}

export interface SubscribeOptions {
  clientId?: string;
  namespaces?: string[];
}

function trimBase(url: string): string {
  return url.replace(/\/+$/, "");
}

function authHeaders(cfg: Config): Record<string, string> {
  if (cfg.skipAuth || !cfg.apiKey) {
    return {};
  }
  return { Authorization: `Bearer ${cfg.apiKey}` };
}

async function parseSSE(
  body: ReadableStream<Uint8Array> | null,
  onFrame: (frame: SSEFrame) => Promise<void> | void
): Promise<void> {
  if (!body) {
    return;
  }
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buf = "";
  let id = "";
  let event = "";
  let dataLines: string[] = [];

  const flush = async () => {
    if (!dataLines.length && !event && !id) {
      return;
    }
    const data = dataLines.join("\n");
    await onFrame({ id, event, data });
    id = "";
    event = "";
    dataLines = [];
  };

  for (;;) {
    const { done, value } = await reader.read();
    if (value) {
      buf += decoder.decode(value, { stream: !done });
    }
    let idx: number;
    while ((idx = buf.indexOf("\n")) >= 0) {
      let line = buf.slice(0, idx);
      buf = buf.slice(idx + 1);
      if (line.endsWith("\r")) {
        line = line.slice(0, -1);
      }
      if (line === "") {
        await flush();
      } else if (line.startsWith(":")) {
        continue;
      } else if (line.startsWith("id:")) {
        id = line.slice(3).trimStart();
      } else if (line.startsWith("event:")) {
        event = line.slice(6).trimStart();
      } else if (line.startsWith("data:")) {
        dataLines.push(line.slice(5).trimStart());
      }
    }
    if (done) {
      await flush();
      break;
    }
  }
}

export class KraneClient {
  private readonly config: Config;

  constructor(config: Config) {
    if (!config.serverUrl) {
      throw new Error("serverUrl is required");
    }
    if (!config.skipAuth && !config.apiKey) {
      throw new Error("apiKey is required unless skipAuth is set");
    }
    this.config = config;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown
  ): Promise<T> {
    const url = `${trimBase(this.config.serverUrl)}${path}`;
    const init = {
      method,
      headers: {
        Accept: "application/json",
        ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
        ...authHeaders(this.config),
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    } as RequestInit;
    const res = await fetch(url, init);
    if (!res.ok) {
      const t = await res.text();
      throw new Error(`kranix-api ${method} ${path}: ${res.status} ${t}`);
    }
    if (res.status === 204) {
      return undefined as T;
    }
    return (await res.json()) as T;
  }

  private async *iterPodLogs(
    podId: string,
    options: LogOptions
  ): AsyncGenerator<string, void, unknown> {
    const q = new URLSearchParams();
    if (options.follow) {
      q.set("follow", "true");
    }
    if (options.tail != null) {
      q.set("tail", String(options.tail));
    }
    if (options.since != null) {
      q.set("since", String(options.since));
    }
    const qs = q.toString();
    const path = `/api/v1/pods/${encodeURIComponent(podId)}/logs${
      qs ? `?${qs}` : ""
    }`;
    const url = `${trimBase(this.config.serverUrl)}${path}`;
    const res = await fetch(url, {
      headers: { Accept: "text/event-stream", ...authHeaders(this.config) },
    });
    if (!res.ok || !res.body) {
      throw new Error(`logs: ${res.status}`);
    }

    const queue: string[] = [];
    let done = false;
    let notify: (() => void) | undefined;

    void parseSSE(res.body, async (frame) => {
      if (frame.event === "log" && frame.data) {
        queue.push(frame.data.trim());
        notify?.();
      }
    }).finally(() => {
      done = true;
      notify?.();
    });

    while (!done || queue.length) {
      if (!queue.length) {
        await new Promise<void>((r) => {
          notify = () => {
            notify = undefined;
            r();
          };
        });
      }
      while (queue.length) {
        yield queue.shift()!;
      }
    }
  }

  private async requestRaw(method: string, path: string): Promise<string> {
    const url = `${trimBase(this.config.serverUrl)}${path}`;
    const res = await fetch(url, {
      method,
      headers: { Accept: "application/json", ...authHeaders(this.config) },
    });
    if (!res.ok) {
      const t = await res.text();
      throw new Error(`kranix-api ${method} ${path}: ${res.status} ${t}`);
    }
    return res.text();
  }

  public readonly workloads = {
    deploy: async (spec: WorkloadSpec): Promise<Workload> => {
      const wl = await this.request<Workload>("POST", "/api/v1/workloads", spec);
      if (!wl?.id) {
        throw new Error(
          "deploy response missing id (API stubbed? use kranix-mock-api for tests)"
        );
      }
      return wl;
    },
    get: async (id: string): Promise<Workload> => {
      return this.request<Workload>(
        "GET",
        `/api/v1/workloads/${encodeURIComponent(id)}`
      );
    },
    list: async (namespace?: string): Promise<Workload[]> => {
      const q = namespace
        ? `?namespace=${encodeURIComponent(namespace)}`
        : "";
      const raw = await this.requestRaw("GET", `/api/v1/workloads${q}`);
      try {
        return JSON.parse(raw) as Workload[];
      } catch {
        const w = JSON.parse(raw) as { workloads?: Workload[] };
        return w.workloads ?? [];
      }
    },
    update: async (id: string, spec: WorkloadSpec): Promise<Workload> => {
      return this.request<Workload>(
        "PATCH",
        `/api/v1/workloads/${encodeURIComponent(id)}`,
        spec
      );
    },
    delete: async (id: string): Promise<void> => {
      await this.request<void>(
        "DELETE",
        `/api/v1/workloads/${encodeURIComponent(id)}`
      );
    },
    restart: async (id: string): Promise<void> => {
      await this.request<Record<string, unknown>>(
        "POST",
        `/api/v1/workloads/${encodeURIComponent(id)}/restart`,
        {}
      );
    },
    listPods: async (workloadId: string): Promise<Pod[]> => {
      const j = await this.request<{ pods: Pod[] }>(
        "GET",
        `/api/v1/workloads/${encodeURIComponent(workloadId)}/pods`
      );
      return j.pods ?? [];
    },
    analyze: async (id: string): Promise<AnalysisResult> => {
      return this.request<AnalysisResult>(
        "GET",
        `/api/v1/workloads/${encodeURIComponent(id)}/analyze`
      );
    },
  };

  public readonly pods = {
    streamLogs: (
      podId: string,
      options: LogOptions
    ): AsyncGenerator<string, void, unknown> =>
      this.iterPodLogs(podId, options),

    get: async (id: string): Promise<Pod> => {
      return this.request<Pod>(
        "GET",
        `/api/v1/pods/${encodeURIComponent(id)}`
      );
    },
    list: async (namespace?: string): Promise<Pod[]> => {
      const q = namespace
        ? `?namespace=${encodeURIComponent(namespace)}`
        : "";
      const j = await this.request<{ pods: Pod[] }>(
        "GET",
        `/api/v1/pods${q}`
      );
      return j.pods ?? [];
    },
  };

  public readonly namespaces = {
    get: async (name: string): Promise<Namespace> => {
      return this.request<Namespace>(
        "GET",
        `/api/v1/namespaces/${encodeURIComponent(name)}`
      );
    },
    list: async (): Promise<Namespace[]> => {
      const j = await this.request<{ namespaces: Namespace[] }>(
        "GET",
        "/api/v1/namespaces"
      );
      return j.namespaces ?? [];
    },
    create: async (namespace: Namespace): Promise<Namespace> => {
      return this.request<Namespace>(
        "POST",
        "/api/v1/namespaces",
        namespace
      );
    },
  };

  /**
   * Subscribe to platform SSE (/api/sse). Yields frames until the stream ends or breaks on error.
   */
  public async *subscribeWorkloadEvents(
    opts?: SubscribeOptions
  ): AsyncGenerator<SSEFrame, void, unknown> {
    const params = new URLSearchParams();
    if (opts?.clientId) {
      params.set("client_id", opts.clientId);
    }
    for (const ns of opts?.namespaces ?? []) {
      params.append("namespace", ns);
    }
    const qs = params.toString();
    const path = `/api/sse${qs ? `?${qs}` : ""}`;
    const url = `${trimBase(this.config.serverUrl)}${path}`;
    const res = await fetch(url, {
      headers: { Accept: "text/event-stream", ...authHeaders(this.config) },
    });
    if (!res.ok || !res.body) {
      throw new Error(`sse: ${res.status}`);
    }

    const queue: SSEFrame[] = [];
    let finished = false;
    let notify: (() => void) | undefined;

    void parseSSE(res.body, async (frame) => {
      queue.push(frame);
      notify?.();
    }).finally(() => {
      finished = true;
      notify?.();
    });

    while (!finished || queue.length) {
      if (!queue.length) {
        await new Promise<void>((r) => {
          notify = () => {
            notify = undefined;
            r();
          };
        });
      }
      while (queue.length) {
        yield queue.shift()!;
      }
    }
  }
}

export default KraneClient;
