import { LayoutGrid, List, RefreshCw, FolderPlus } from "lucide-react";
import type { ViewMode } from "../types";

interface Props {
  viewMode: ViewMode;
  onViewChange: (mode: ViewMode) => void;
  onRefresh: () => void;
  onNewFolder: () => void;
  selectedCount: number;
}

export function Toolbar({ viewMode, onViewChange, onRefresh, onNewFolder, selectedCount }: Props) {
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-slate-800/30 border-b border-slate-700">
      <div className="flex items-center gap-2">
        <button
          onClick={onRefresh}
          className="p-1.5 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
          title="Refresh"
        >
          <RefreshCw size={16} />
        </button>
        <button
          onClick={onNewFolder}
          className="p-1.5 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
          title="New Folder (Cmd+N)"
        >
          <FolderPlus size={16} />
        </button>
        {selectedCount > 0 && (
          <span className="text-xs text-slate-400">{selectedCount} selected</span>
        )}
      </div>
      <div className="flex items-center gap-1 bg-slate-800 rounded p-0.5">
        <button
          onClick={() => onViewChange("grid")}
          className={`p-1.5 rounded transition-colors ${
            viewMode === "grid" ? "bg-slate-600 text-white" : "text-slate-400 hover:text-white"
          }`}
          title="Grid view"
        >
          <LayoutGrid size={16} />
        </button>
        <button
          onClick={() => onViewChange("list")}
          className={`p-1.5 rounded transition-colors ${
            viewMode === "list" ? "bg-slate-600 text-white" : "text-slate-400 hover:text-white"
          }`}
          title="List view"
        >
          <List size={16} />
        </button>
      </div>
    </div>
  );
}
