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
