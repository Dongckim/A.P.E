import { useState, useEffect, useCallback } from "react";
import { ChevronRight, HardDrive, Folder, File, ArrowLeft, Loader2, AlertCircle, Download, Trash2, Upload } from "lucide-react";
import type { BucketInfo, S3Object } from "../types";
import { listBuckets, listObjects, downloadObject, deleteObject, uploadObject } from "../api/client";
import { ConfirmDialog } from "./ConfirmDialog";

export function S3Browser() {
  const [buckets, setBuckets] = useState<BucketInfo[]>([]);
  const [selectedBucket, setSelectedBucket] = useState<string | null>(null);
  const [prefix, setPrefix] = useState("");
  const [objects, setObjects] = useState<S3Object[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<S3Object | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);

  useEffect(() => {
    setLoading(true);
    listBuckets()
      .then(setBuckets)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const loadObjects = useCallback(async (bucket: string, pfx: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await listObjects(bucket, pfx);
      setObjects(data.objects);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load objects");
    } finally {
      setLoading(false);
    }
  }, []);

  const openBucket = (name: string) => {
    setSelectedBucket(name);
    setPrefix("");
    loadObjects(name, "");
  };

  const navigatePrefix = (pfx: string) => {
    if (!selectedBucket) return;
    setPrefix(pfx);
    loadObjects(selectedBucket, pfx);
  };

  const goBack = () => {
    if (!prefix) {
      setSelectedBucket(null);
      setObjects([]);
      return;
    }
    const parts = prefix.replace(/\/$/, "").split("/");
    parts.pop();
    const newPrefix = parts.length > 0 ? parts.join("/") + "/" : "";
    setPrefix(newPrefix);
    if (selectedBucket) loadObjects(selectedBucket, newPrefix);
  };

  const handleDelete = async () => {
    if (!selectedBucket || !deleteTarget) return;
    try {
      await deleteObject(selectedBucket, deleteTarget.key);
      setDeleteTarget(null);
      loadObjects(selectedBucket, prefix);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Delete failed");
      setDeleteTarget(null);
    }
  };

  const handleUpload = async () => {
    if (!selectedBucket) return;
    const input = document.createElement("input");
    input.type = "file";
    input.onchange = async () => {
      const file = input.files?.[0];
      if (!file) return;
      const key = prefix + file.name;
      try {
        setUploadProgress(0);
        await uploadObject(selectedBucket, key, file, setUploadProgress);
        setUploadProgress(null);
        loadObjects(selectedBucket, prefix);
      } catch {
        setUploadProgress(null);
        setError("Upload failed");
      }
    };
    input.click();
  };

  // Bucket list
  if (!selectedBucket) {
    return (
      <div className="flex flex-col h-full">
        <div className="px-4 py-2 bg-slate-800/50 border-b border-slate-700 text-sm text-slate-300 font-medium">
          S3 Buckets
        </div>
        <div className="flex-1 overflow-auto">
          {loading && (
            <div className="flex items-center justify-center h-64 text-slate-400">
              <Loader2 size={24} className="animate-spin" />
            </div>
          )}
          {error && (
            <div className="flex items-center justify-center h-64 gap-2 text-red-400">
              <AlertCircle size={16} /><span className="text-sm">{error}</span>
            </div>
          )}
          {!loading && !error && buckets.length === 0 && (
            <div className="flex items-center justify-center h-64 text-slate-500 text-sm">No buckets found</div>
          )}
          {!loading && buckets.map((b) => (
            <button
              key={b.name}
              onClick={() => openBucket(b.name)}
              className="w-full flex items-center gap-2 px-4 py-2 text-sm text-slate-300 hover:bg-slate-800 transition-colors border-b border-slate-800"
            >
              <HardDrive size={16} className="text-cyan-400 shrink-0" />
              <span className="truncate">{b.name}</span>
              <ChevronRight size={14} className="text-slate-600 ml-auto shrink-0" />
            </button>
          ))}
        </div>
      </div>
    );
  }

  // Object list
  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 px-4 py-2 bg-slate-800/50 border-b border-slate-700 text-sm">
        <button onClick={goBack} className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white">
          <ArrowLeft size={16} />
        </button>
        <HardDrive size={14} className="text-cyan-400 shrink-0" />
        <span className="text-slate-300 font-medium">{selectedBucket}</span>
        {prefix && (
          <>
            <ChevronRight size={14} className="text-slate-600" />
            <span className="text-slate-400 truncate">{prefix}</span>
          </>
        )}
        <div className="ml-auto flex gap-1">
          <button onClick={handleUpload} className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white" title="Upload">
            <Upload size={16} />
          </button>
        </div>
      </div>

      {uploadProgress !== null && (
        <div className="px-4 py-1 bg-slate-800">
          <div className="h-1 bg-slate-700 rounded-full overflow-hidden">
            <div className="h-full bg-cyan-500 transition-all" style={{ width: `${uploadProgress}%` }} />
          </div>
        </div>
      )}

      <div className="flex-1 overflow-auto">
        {loading && (
          <div className="flex items-center justify-center h-64 text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}
        {error && (
          <div className="flex items-center justify-center h-64 gap-2 text-red-400">
            <AlertCircle size={16} /><span className="text-sm">{error}</span>
          </div>
        )}
        {!loading && !error && objects.length === 0 && (
          <div className="flex items-center justify-center h-64 text-slate-500 text-sm">Empty</div>
        )}
        {!loading && !error && objects.map((obj) => (
          <div
            key={obj.key}
            className="flex items-center gap-2 px-4 py-1.5 text-sm hover:bg-slate-800/50 border-b border-slate-800 cursor-pointer group"
            onClick={() => obj.is_dir && navigatePrefix(obj.key)}
          >
            {obj.is_dir ? (
              <Folder size={16} className="text-cyan-400 shrink-0" />
            ) : (
              <File size={16} className="text-slate-400 shrink-0" />
            )}
            <span className="text-slate-300 truncate flex-1">
              {obj.is_dir ? obj.key.replace(prefix, "").replace(/\/$/, "") : obj.key.replace(prefix, "")}
            </span>
            {!obj.is_dir && (
              <div className="hidden group-hover:flex items-center gap-1">
                <button
                  onClick={(e) => { e.stopPropagation(); downloadObject(selectedBucket, obj.key); }}
                  className="p-1 rounded hover:bg-slate-700 text-slate-400"
                  title="Download"
                >
                  <Download size={14} />
                </button>
                <button
                  onClick={(e) => { e.stopPropagation(); setDeleteTarget(obj); }}
                  className="p-1 rounded hover:bg-red-500/10 text-slate-400 hover:text-red-400"
                  title="Delete"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            )}
          </div>
        ))}
      </div>

      {deleteTarget && (
        <ConfirmDialog
          title="Delete Object"
          message={`Delete "${deleteTarget.key}"?`}
          confirmLabel="Delete"
          danger
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}
