import { useState, useEffect, useCallback } from "react";
import type { FileInfo } from "../types";
import { listFiles } from "../api/client";

export function useFileSystem(initialPath = "/home/ubuntu") {
  const [path, setPath] = useState(initialPath);
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await listFiles(path);
      setFiles(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load files");
      setFiles([]);
    } finally {
      setLoading(false);
    }
  }, [path]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const navigate = useCallback((newPath: string) => {
    setPath(newPath);
  }, []);

  return { path, files, loading, error, navigate, refresh };
}
