import { Monitor, HardDrive, Plus, Settings } from "lucide-react";

interface Props {
  activeView: "ec2" | "s3";
  onViewChange: (view: "ec2" | "s3") => void;
  onOpenConnections: () => void;
}

export function Sidebar({ activeView, onViewChange, onOpenConnections }: Props) {
  return (
    <aside className="w-56 bg-slate-900 border-r border-slate-700 flex flex-col shrink-0">
      <div className="px-4 py-3 border-b border-slate-700 flex items-center justify-between">
        <div>
          <h1 className="text-sm font-semibold text-white tracking-wide">A.P.E</h1>
          <p className="text-[10px] text-slate-500">AWS Platform Explorer</p>
        </div>
        <button
          onClick={onOpenConnections}
          className="p-1.5 rounded hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
          title="Manage Connections"
        >
          <Settings size={14} />
        </button>
      </div>

      <div className="p-3 flex-1 overflow-y-auto">
        <p className="text-[10px] uppercase tracking-wider text-slate-500 mb-2 px-1">
          EC2 Connections
        </p>
        <button
          onClick={() => onViewChange("ec2")}
          className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm transition-colors ${
            activeView === "ec2"
              ? "bg-cyan-500/10 text-cyan-400"
              : "text-slate-400 hover:bg-slate-800"
          }`}
        >
          <Monitor size={14} />
          <span className="truncate">File Explorer</span>
        </button>

        <p className="text-[10px] uppercase tracking-wider text-slate-500 mt-4 mb-2 px-1">
          S3 Buckets
        </p>
        <button
          onClick={() => onViewChange("s3")}
          className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm transition-colors ${
            activeView === "s3"
              ? "bg-cyan-500/10 text-cyan-400"
              : "text-slate-400 hover:bg-slate-800"
          }`}
        >
          <HardDrive size={14} />
          <span className="truncate">S3 Browser</span>
        </button>
      </div>

      <div className="p-3 border-t border-slate-700">
        <button
          onClick={onOpenConnections}
          className="w-full flex items-center justify-center gap-1.5 px-3 py-1.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm transition-colors"
        >
          <Plus size={14} />
          Add Connection
        </button>
      </div>
    </aside>
  );
}
