import type { ApiResponse, FileInfo, BucketInfo, S3ObjectList, ConnectionInfo, DashboardOverview, ServicesData, GitInfo, ProcessInfo, RDSOverview, RDSTablesResponse } from "../types";

const BASE = "/api";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + url, init);
  const text = await res.text();
  let json: ApiResponse<T>;
  try {
    json = JSON.parse(text);
  } catch {
    throw new Error(`Invalid response from ${url}: ${text.slice(0, 100)}`);
  }
  if (json.error) throw new Error(json.error);
  return json.data;
}

// --- EC2 File Operations ---

export async function listFiles(path: string): Promise<FileInfo[]> {
  return request<FileInfo[]>(`/ec2/files?path=${encodeURIComponent(path)}`);
}

export async function readFile(path: string): Promise<{ path: string; content: string; size: number }> {
  return request(`/ec2/file?path=${encodeURIComponent(path)}`);
}

export async function writeFile(path: string, content: string): Promise<void> {
  await request(`/ec2/file?path=${encodeURIComponent(path)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ content }),
  });
}

export async function deleteFile(path: string): Promise<void> {
  await request(`/ec2/file?path=${encodeURIComponent(path)}`, { method: "DELETE" });
}

export async function renameFile(oldPath: string, newPath: string): Promise<void> {
  await request(`/ec2/file?path=${encodeURIComponent(oldPath)}&dest=${encodeURIComponent(newPath)}`, {
    method: "PATCH",
  });
}

export async function mkdir(path: string): Promise<void> {
  await request(`/ec2/mkdir?path=${encodeURIComponent(path)}`, { method: "POST" });
}

export function uploadFile(
  path: string,
  file: File,
  onProgress?: (pct: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${BASE}/ec2/upload?path=${encodeURIComponent(path)}`);

    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) {
        onProgress(Math.round((e.loaded / e.total) * 100));
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve();
      } else {
        try {
          const json = JSON.parse(xhr.responseText);
          reject(new Error(json.error || "Upload failed"));
        } catch {
          reject(new Error("Upload failed"));
        }
      }
    };

    xhr.onerror = () => reject(new Error("Network error"));

    const form = new FormData();
    form.append("file", file);
    xhr.send(form);
  });
}

export function downloadFile(path: string): void {
  window.open(`${BASE}/ec2/download?path=${encodeURIComponent(path)}`, "_blank");
}

// --- S3 Operations ---

export async function listBuckets(): Promise<BucketInfo[]> {
  return request<BucketInfo[]>("/s3/buckets");
}

export async function listObjects(bucket: string, prefix: string): Promise<S3ObjectList> {
  return request<S3ObjectList>(`/s3/objects?bucket=${encodeURIComponent(bucket)}&prefix=${encodeURIComponent(prefix)}`);
}

export async function deleteObject(bucket: string, key: string): Promise<void> {
  await request(`/s3/object?bucket=${encodeURIComponent(bucket)}&key=${encodeURIComponent(key)}`, {
    method: "DELETE",
  });
}

export function downloadObject(bucket: string, key: string): void {
  window.open(`${BASE}/s3/download?bucket=${encodeURIComponent(bucket)}&key=${encodeURIComponent(key)}`, "_blank");
}

export function uploadObject(
  bucket: string,
  key: string,
  file: File,
  onProgress?: (pct: number) => void
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${BASE}/s3/upload?bucket=${encodeURIComponent(bucket)}&key=${encodeURIComponent(key)}`);
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable && onProgress) onProgress(Math.round((e.loaded / e.total) * 100));
    };
    xhr.onload = () => (xhr.status < 300 ? resolve() : reject(new Error("Upload failed")));
    xhr.onerror = () => reject(new Error("Network error"));
    const form = new FormData();
    form.append("file", file);
    xhr.send(form);
  });
}

// --- Connections ---

export async function listConnections(): Promise<ConnectionInfo[]> {
  return request<ConnectionInfo[]>("/connections");
}

export async function removeConnection(id: string): Promise<void> {
  await request(`/connections?id=${encodeURIComponent(id)}`, { method: "DELETE" });
}

// --- Dashboard ---

export async function fetchOverview(): Promise<DashboardOverview> {
  return request<DashboardOverview>("/dashboard/overview");
}

export async function fetchServices(): Promise<ServicesData> {
  return request<ServicesData>("/dashboard/services");
}

export async function fetchGitLog(path: string): Promise<GitInfo> {
  return request<GitInfo>(`/dashboard/git?path=${encodeURIComponent(path)}`);
}

export async function fetchProcesses(): Promise<ProcessInfo[]> {
  return request<ProcessInfo[]>("/dashboard/processes");
}

// --- RDS / PostgreSQL ---

export async function fetchRDSOverview(db?: string): Promise<RDSOverview> {
  const qs = db ? `?db=${encodeURIComponent(db)}` : "";
  return request<RDSOverview>(`/rds/overview${qs}`);
}

export async function fetchRDSTables(schema: string, db?: string): Promise<RDSTablesResponse> {
  const params = new URLSearchParams({ schema });
  if (db) params.set("db", db);
  return request<RDSTablesResponse>(`/rds/tables?${params.toString()}`);
}
