import { KeyboardShortcutsModal } from "./KeyboardShortcutsModal";
import { useState, useCallback, useEffect, useRef } from "react";
import type { FileInfo, ViewMode } from "../types";
import { Breadcrumb } from "./Breadcrumb";
import { Toolbar } from "./Toolbar";
import { GridView } from "./GridView";
import { ListView } from "./ListView";
import { ContextMenu } from "./ContextMenu";
import { ConfirmDialog } from "./ConfirmDialog";
import { UploadOverlay } from "./UploadOverlay";
import { TextEditor } from "./TextEditor";
import { useFileSystem } from "../hooks/useFileSystem";
import { uploadFile, downloadFile, deleteFile, renameFile, mkdir } from "../api/client";
import { Loader2, AlertCircle } from "lucide-react";

export function FileExplorer() {
  const { path, files, loading, error, navigate, refresh } = useFileSystem();
  const [viewMode, setViewMode] = useState<ViewMode>("grid");
  const [selected, setSelected] = useState<Set<string>>(new Set());

  // Context menu
  const [ctxMenu, setCtxMenu] = useState<{ x: number; y: number; file: FileInfo } | null>(null);

  // Delete confirm
  const [deleteTarget, setDeleteTarget] = useState<FileInfo | null>(null);

  // Rename
  const [renamingPath, setRenamingPath] = useState<string | null>(null);

  // Upload
  const [uploadProgress, setUploadProgress] = useState<number | null>(null);
  const [uploadName, setUploadName] = useState<string>("");
  const [isDragOver, setIsDragOver] = useState(false);

  // Editor
  const [editingFile, setEditingFile] = useState<string | null>(null);

  // Keyboard shortcuts modal
  const [showShortcuts, setShowShortcuts] = useState(false);

  const containerRef = useRef<HTMLDivElement>(null);

  // --- Selection ---
  const handleSelect = useCallback((file: FileInfo, e: React.MouseEvent) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (e.metaKey || e.ctrlKey) {
        if (next.has(file.path)) { next.delete(file.path); } else { next.add(file.path); }
      } else if (e.shiftKey && prev.size > 0) {
        const last = Array.from(prev).pop()!;
        const a = files.findIndex((f) => f.path === last);
        const b = files.findIndex((f) => f.path === file.path);
        for (let i = Math.min(a, b); i <= Math.max(a, b); i++) next.add(files[i].path);
      } else {
        next.clear();
        next.add(file.path);
      }
      return next;
    });
  }, [files]);

  // --- Open ---
  const handleOpen = useCallback((file: FileInfo) => {
    if (file.is_dir) {
      setSelected(new Set());
      setEditingFile(null);
      navigate(file.path);
    } else {
      setEditingFile(file.path);
    }
  }, [navigate]);

  // --- Context Menu ---
  const handleContextMenu = useCallback((file: FileInfo, e: React.MouseEvent) => {
    setSelected(new Set([file.path]));
    setCtxMenu({ x: e.clientX, y: e.clientY, file });
  }, []);

  // --- Delete ---
  const handleDeleteConfirm = useCallback(async () => {
    if (!deleteTarget) return;
    try {
      await deleteFile(deleteTarget.path);
      setDeleteTarget(null);
      setSelected(new Set());
      refresh();
    } catch {
      setDeleteTarget(null);
    }
  }, [deleteTarget, refresh]);

  // --- Rename ---
  const handleRenameConfirm = useCallback(async (file: FileInfo, newName: string) => {
    const dir = file.path.substring(0, file.path.lastIndexOf("/"));
    const newPath = dir + "/" + newName;
    try {
      await renameFile(file.path, newPath);
      setRenamingPath(null);
      refresh();
    } catch {
      setRenamingPath(null);
    }
  }, [refresh]);

  // --- New Folder ---
  const handleNewFolder = useCallback(async () => {
    const name = prompt("Folder name:");
    if (!name) return;
    try {
      await mkdir(path + "/" + name);
      refresh();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed to create folder");
    }
  }, [path, refresh]);

  // --- Upload (drag & drop) ---
  const handleDrop = useCallback(async (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    for (const file of droppedFiles) {
      setUploadName(file.name);
      setUploadProgress(0);
      try {
        await uploadFile(path, file, setUploadProgress);
      } catch { /* ignore individual errors */ }
    }
    setUploadProgress(null);
    setUploadName("");
    refresh();
  }, [path, refresh]);

  // --- Keyboard shortcuts ---
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      // Cmd+N: new folder
      if ((e.metaKey || e.ctrlKey) && e.key === "n") {
        e.preventDefault();
        handleNewFolder();
      }
      // Delete / Backspace: delete selected
      if ((e.key === "Delete" || e.key === "Backspace") && selected.size > 0 && !renamingPath && !editingFile) {
        const file = files.find((f) => selected.has(f.path));
        if (file) setDeleteTarget(file);
      }
      // Cmd+C: copy path
      if ((e.metaKey || e.ctrlKey) && e.key === "c" && selected.size === 1) {
        const sel = Array.from(selected)[0];
        navigator.clipboard.writeText(sel);
      }
      // Enter: open selected
      if (e.key === "Enter" && selected.size === 1 && !renamingPath) {
        const file = files.find((f) => selected.has(f.path));
        if (file) handleOpen(file);
      }
      // Escape: close editor / deselect
      if (e.key === "Escape") {
        if (showShortcuts) setShowShortcuts(false);
        else if (editingFile) setEditingFile(null);
        else setSelected(new Set());
      }
      // ?: show keyboard shortcuts
      if (e.key === "?" && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
        setShowShortcuts(true);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [selected, files, renamingPath, editingFile, handleNewFolder, handleOpen]);

  // Editor view
  if (editingFile) {
    return <TextEditor filePath={editingFile} onClose={() => setEditingFile(null)} />;
  }

  return (
    <div
      ref={containerRef}
      className="flex flex-col h-full relative"
      onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
      onDragLeave={() => setIsDragOver(false)}
      onDrop={handleDrop}
    >
      <Breadcrumb path={path} onNavigate={navigate} />
      <Toolbar
        viewMode={viewMode}
        onViewChange={setViewMode}
        onRefresh={refresh}
        onNewFolder={handleNewFolder}
        selectedCount={selected.size}
      />

      <div className="flex-1 overflow-auto" onClick={() => setSelected(new Set())}>
        {loading && (
          <div className="flex items-center justify-center h-64 text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}
        {error && (
          <div className="flex items-center justify-center h-64 gap-2 text-red-400">
            <AlertCircle size={18} /><span>{error}</span>
          </div>
        )}
        {!loading && !error && files.length === 0 && (
          <div className="flex items-center justify-center h-64 text-slate-500">
            This folder is empty — drag files here to upload
          </div>
        )}
        {!loading && !error && files.length > 0 && (
          viewMode === "grid" ? (
            <GridView
              files={files}
              selected={selected}
              renamingPath={renamingPath}
              onSelect={handleSelect}
              onOpen={handleOpen}
              onContextMenu={handleContextMenu}
              onRenameConfirm={handleRenameConfirm}
              onRenameCancel={() => setRenamingPath(null)}
            />
          ) : (
            <ListView
              files={files}
              selected={selected}
              renamingPath={renamingPath}
              onSelect={handleSelect}
              onOpen={handleOpen}
              onContextMenu={handleContextMenu}
              onRenameConfirm={handleRenameConfirm}
              onRenameCancel={() => setRenamingPath(null)}
            />
          )
        )}
      </div>

      <UploadOverlay progress={uploadProgress} fileName={uploadName} isDragOver={isDragOver} />

      {ctxMenu && (
        <ContextMenu
          x={ctxMenu.x}
          y={ctxMenu.y}
          file={ctxMenu.file}
          onClose={() => setCtxMenu(null)}
          onOpen={() => handleOpen(ctxMenu.file)}
          onDownload={() => downloadFile(ctxMenu.file.path)}
          onRename={() => setRenamingPath(ctxMenu.file.path)}
          onDelete={() => setDeleteTarget(ctxMenu.file)}
        />
      )}

      {deleteTarget && (
        <ConfirmDialog
          title="Delete"
          message={`Are you sure you want to delete "${deleteTarget.name}"?${deleteTarget.is_dir ? " This will delete all contents." : ""}`}
          confirmLabel="Delete"
          danger
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
      {showShortcuts && (
        <KeyboardShortcutsModal onClose={() => setShowShortcuts(false)} />
      )}
    </div>
  );
}
