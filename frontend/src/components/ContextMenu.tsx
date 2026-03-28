import { useEffect, useRef } from "react";
import { FolderOpen, Download, Pencil, Trash2 } from "lucide-react";
import type { FileInfo } from "../types";

interface Props {
  x: number;
  y: number;
  file: FileInfo;
  onClose: () => void;
  onOpen: () => void;
  onDownload: () => void;
  onRename: () => void;
  onDelete: () => void;
}

export function ContextMenu({ x, y, file, onClose, onOpen, onDownload, onRename, onDelete }: Props) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [onClose]);

  const items = [
    { icon: FolderOpen, label: file.is_dir ? "Open" : "Open in Editor", action: onOpen },
    ...(!file.is_dir ? [{ icon: Download, label: "Download", action: onDownload }] : []),
    { icon: Pencil, label: "Rename", action: onRename },
    { icon: Trash2, label: "Delete", action: onDelete, danger: true },
  ];

  return (
    <div
      ref={ref}
      className="fixed z-50 bg-slate-800 border border-slate-600 rounded-lg shadow-xl py-1 min-w-[160px]"
      style={{ left: x, top: y }}
    >
      {items.map((item) => (
        <button
          key={item.label}
          onClick={() => { item.action(); onClose(); }}
          className={`w-full flex items-center gap-2 px-3 py-1.5 text-sm text-left transition-colors ${
            "danger" in item && item.danger
              ? "text-red-400 hover:bg-red-500/10"
              : "text-slate-300 hover:bg-slate-700"
          }`}
        >
          <item.icon size={14} />
          {item.label}
        </button>
      ))}
    </div>
  );
}
