import { Folder, File, Image, Film, Music, FileText } from "lucide-react";
import type { FileInfo } from "../types";
import { RenameInput } from "./RenameInput";
import { getMediaKind } from "../utils/media";

interface Props {
  files: FileInfo[];
  selected: Set<string>;
  renamingPath: string | null;
  onSelect: (file: FileInfo, e: React.MouseEvent) => void;
  onOpen: (file: FileInfo) => void;
  onContextMenu: (file: FileInfo, e: React.MouseEvent) => void;
  onRenameConfirm: (file: FileInfo, newName: string) => void;
  onRenameCancel: () => void;
}

function formatSize(bytes: number): string {
  if (bytes === 0) return "--";
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
}

function formatDate(iso: string): string {
  if (!iso) return "--";
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
}

function getRowIcon(file: FileInfo) {
  if (file.is_dir) return <Folder size={16} className="text-cyan-400 shrink-0" />;
  const kind = getMediaKind(file.name);
  if (kind === "image") return <Image size={16} className="text-pink-400 shrink-0" />;
  if (kind === "video") return <Film size={16} className="text-purple-400 shrink-0" />;
  if (kind === "audio") return <Music size={16} className="text-indigo-400 shrink-0" />;
  if (kind === "pdf") return <FileText size={16} className="text-red-400 shrink-0" />;
  return <File size={16} className="text-slate-400 shrink-0" />;
}

export function ListView({ files, selected, renamingPath, onSelect, onOpen, onContextMenu, onRenameConfirm, onRenameCancel }: Props) {
  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="text-left text-slate-500 border-b border-slate-700">
          <th className="px-4 py-2 font-medium">Name</th>
          <th className="px-4 py-2 font-medium w-24">Size</th>
          <th className="px-4 py-2 font-medium w-48">Modified</th>
          <th className="px-4 py-2 font-medium w-28">Permissions</th>
        </tr>
      </thead>
      <tbody>
        {files.map((file) => (
          <tr
            key={file.path}
            className={`cursor-pointer transition-colors select-none border-b border-slate-800 ${
              selected.has(file.path) ? "bg-cyan-500/20" : "hover:bg-slate-700/30"
            }`}
            onClick={(e) => { e.stopPropagation(); onSelect(file, e); }}
            onDoubleClick={() => onOpen(file)}
            onContextMenu={(e) => { e.preventDefault(); e.stopPropagation(); onContextMenu(file, e); }}
          >
            <td className="px-4 py-1.5 flex items-center gap-2">
              {getRowIcon(file)}
              {renamingPath === file.path ? (
                <RenameInput
                  initialName={file.name}
                  onConfirm={(n) => onRenameConfirm(file, n)}
                  onCancel={onRenameCancel}
                />
              ) : (
                <span className="text-slate-200 truncate">{file.name}</span>
              )}
            </td>
            <td className="px-4 py-1.5 text-slate-400">{file.is_dir ? "--" : formatSize(file.size)}</td>
            <td className="px-4 py-1.5 text-slate-400">{formatDate(file.mod_time)}</td>
            <td className="px-4 py-1.5 text-slate-500 font-mono text-xs">{file.permissions}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
