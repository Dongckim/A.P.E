export interface FileInfo {
  name: string;
  path: string;
  size: number;
  is_dir: boolean;
  permissions: string;
  mod_time: string;
}

export interface BucketInfo {
  name: string;
  creation_date: string;
}

export interface S3Object {
  key: string;
  size: number;
  is_dir: boolean;
  last_modified?: string;
  storage_class?: string;
}

export interface S3ObjectList {
  prefix: string;
  objects: S3Object[];
  is_truncated: boolean;
}

export interface ConnectionInfo {
  id: string;
  host: string;
  username: string;
  port: string;
}

export interface ApiResponse<T> {
  data: T;
  error: string;
}

export type ViewMode = "grid" | "list";

// --- Dashboard ---

export interface DashboardOverview {
  hostname: string;
  os: string;
  kernel: string;
  arch: string;
  uptime_since: string;
  cpu: { usage_percent: number; cores: number };
  memory: { total: number; used: number; available: number };
  disks: { mount: string; total: number; used: number; avail: number; percent: number }[];
}

export interface DockerContainer {
  id: string;
  name: string;
  image: string;
  status: string;
  ports: string;
  state: string;
}

export interface PM2Process {
  name: string;
  id: number;
  mode: string;
  status: string;
  cpu: number;
  memory: number;
  restarts: number;
  uptime: number;
}

export interface SystemdService {
  unit: string;
  active: string;
  sub: string;
  description: string;
}

export interface ServicesData {
  available_runtimes: string[];
  docker: { containers: DockerContainer[]; error?: string } | null;
  pm2: { processes: PM2Process[]; error?: string } | null;
  systemd: { services: SystemdService[] } | null;
}

export interface GitCommit {
  hash: string;
  message: string;
  author: string;
  relative_date?: string;
  date?: string;
}

export interface GitInfo {
  path: string;
  branch: string;
  last_commit: GitCommit | null;
  commits: GitCommit[];
  error?: string;
}

export interface ProcessInfo {
  pid: string;
  user: string;
  cpu: number;
  memory: number;
  command: string;
}

export interface RDSSchemaSummary {
  name: string;
  table_count: number;
}

export interface RDSOverview {
  version: string;
  current_db: string;
  schema_count: number;
  table_count: number;
  schemas: RDSSchemaSummary[];
  connected: boolean;
  error?: string;
}
