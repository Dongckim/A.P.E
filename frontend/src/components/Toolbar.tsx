import { LayoutGrid, List, RefreshCw, FolderPlus, Search, ArrowUp, ArrowDown } from "lucide-react";
import type { ViewMode, SortKey, SortOrder } from "../types";

interface Props {
  viewMode: ViewMode;
  onViewChange: (mode: ViewMode) => void;
  onRefresh: () => void;
  onNewFolder: () => void;
  selectedCount: number;
  search: string;
  onSearchChange: (q: string) => void;
  sortKey: SortKey;
  sortOrder: SortOrder;
  onSortChange: (key: SortKey, order: SortOrder) => void;
}

export function Toolbar({
  viewMode,
  onViewChange,
  onRefresh,
  onNewFolder,
  selectedCount,
  search,
  onSearchChange,
  sortKey,
  sortOrder,
  onSortChange,
}: Props) {
  const handleSortKeyChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onSortChange(e.target.value as SortKey, sortOrder);
  };

  const toggleOrder = () => {
    onSortChange(sortKey, sortOrder === "asc" ? "desc" : "asc");
  };

  return (
    <div className="flex items-center justify-between px-4 py-2 bg-slate-800/30 border-b border-slate-700 gap-3">
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

      <div className="flex items-center gap-2 flex-1 max-w-md">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none" />
          <input
            type="text"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Search in this folder..."
            className="w-full pl-7 pr-2 py-1 text-xs bg-slate-800 border border-slate-700 rounded text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>
      </div>

      <div className="flex items-center gap-2">
        <div className="flex items-center gap-1">
          <select
            value={sortKey}
            onChange={handleSortKeyChange}
            className="text-xs bg-slate-800 border border-slate-700 rounded px-1.5 py-1 text-slate-200 focus:outline-none focus:border-cyan-500"
            title="Sort by"
          >
            <option value="name">Name</option>
            <option value="size">Size</option>
            <option value="mod_time">Modified</option>
          </select>
          <button
            onClick={toggleOrder}
            className="p-1.5 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
            title={sortOrder === "asc" ? "Ascending" : "Descending"}
          >
            {sortOrder === "asc" ? <ArrowUp size={14} /> : <ArrowDown size={14} />}
          </button>
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
    </div>
  );
}
