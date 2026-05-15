// TypeScript SDK for Kranix API

export interface Config {
  serverUrl: string;
  apiKey: string;
  timeout?: number; // in seconds
}

export interface WorkloadSpec {
  name: string;
  image: string;
  namespace: string;
  replicas: number;
  env?: Record<string, string>;
  resources?: ResourceSpec;
  ports?: PortSpec[];
  backend: string; // docker | kubernetes
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
  protocol?: string; // tcp | udp
}

export interface WorkloadStatus {
  phase: string;
  readyReplicas: number;
  message?: string;
  lastUpdated: string;
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
  severity: string; // error, warning, info
  type: string;
  message: string;
  field?: string;
}

export class KraneClient {
  private config: Config;

  constructor(config: Config) {
    if (!config.serverUrl) {
      throw new Error("serverUrl is required");
    }
    if (!config.apiKey) {
      throw new Error("apiKey is required");
    }
    this.config = config;
  }

  public readonly workloads = {
    deploy: async (spec: WorkloadSpec): Promise<Workload> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    get: async (id: string): Promise<Workload> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    list: async (namespace?: string): Promise<Workload[]> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    update: async (id: string, spec: WorkloadSpec): Promise<Workload> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    delete: async (id: string): Promise<void> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    analyze: async (id: string): Promise<AnalysisResult> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
  };

  public readonly pods = {
    streamLogs: async function* (
      podId: string,
      options: LogOptions
    ): AsyncGenerator<string, void, unknown> {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    get: async (id: string): Promise<any> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    list: async (namespace?: string): Promise<any[]> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
  };

  public readonly namespaces = {
    get: async (name: string): Promise<any> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    list: async (): Promise<any[]> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
    create: async (namespace: any): Promise<any> => {
      // TODO: Implement HTTP client to call kranix-api
      throw new Error("not implemented");
    },
  };
}

export default KraneClient;
