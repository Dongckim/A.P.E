import { Folder, File, FileText, FileCode, Image, Film, Music } from "lucide-react";
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

function getIcon(file: FileInfo) {
  if (file.is_dir) return <Folder size={40} className="text-cyan-400" />;
  const kind = getMediaKind(file.name);
  if (kind === "image") return <Image size={40} className="text-pink-400" />;
  if (kind === "video") return <Film size={40} className="text-purple-400" />;
  if (kind === "audio") return <Music size={40} className="text-indigo-400" />;
  if (kind === "pdf") return <FileText size={40} className="text-red-400" />;
  const ext = file.name.split(".").pop()?.toLowerCase() || "";
  if (["ts", "tsx", "js", "jsx", "go", "py", "rs", "java", "c", "cpp", "h"].includes(ext))
    return <FileCode size={40} className="text-green-400" />;
  if (["txt", "md", "json", "yaml", "yml", "toml", "xml", "csv", "log"].includes(ext))
    return <FileText size={40} className="text-yellow-400" />;
  return <File size={40} className="text-slate-400" />;
}

export function GridView({ files, selected, renamingPath, onSelect, onOpen, onContextMenu, onRenameConfirm, onRenameCancel }: Props) {
  return (
    <div
      className="grid gap-1 p-4"
      style={{ gridTemplateColumns: "repeat(auto-fill, minmax(100px, 1fr))" }}
    >
      {files.map((file) => (
        <div
          key={file.path}
          className={`flex flex-col items-center gap-1.5 p-3 rounded-lg cursor-pointer transition-colors select-none ${
            selected.has(file.path)
              ? "bg-cyan-500/20 ring-1 ring-cyan-500/50"
              : "hover:bg-slate-700/50"
          }`}
          onClick={(e) => { e.stopPropagation(); onSelect(file, e); }}
          onDoubleClick={() => onOpen(file)}
          onContextMenu={(e) => { e.preventDefault(); e.stopPropagation(); onContextMenu(file, e); }}
        >
          {getIcon(file)}
          {renamingPath === file.path ? (
            <RenameInput
              initialName={file.name}
              onConfirm={(n) => onRenameConfirm(file, n)}
              onCancel={onRenameCancel}
            />
          ) : (
            <span className="text-xs text-center text-slate-300 w-full truncate">{file.name}</span>
          )}
        </div>
      ))}
    </div>
  );
}
