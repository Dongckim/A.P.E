import { Monitor, HardDrive, Plus, Settings, LayoutDashboard, Database, TerminalSquare } from "lucide-react";

type View = "dashboard" | "ec2" | "s3" | "rds";

interface Props {
  activeView: View;
  onViewChange: (view: View) => void;
  onOpenConnections: () => void;
  terminalOpen: boolean;
  onToggleTerminal: () => void;
}

function NavButton({ active, icon: Icon, label, onClick }: { active: boolean; icon: typeof Monitor; label: string; onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className={`w-full flex items-center gap-2 px-2 py-1.5 rounded text-sm transition-colors ${
        active ? "bg-cyan-500/10 text-cyan-400" : "text-slate-400 hover:bg-slate-800"
      }`}
    >
      <Icon size={14} />
      <span className="truncate">{label}</span>
    </button>
  );
}

export function Sidebar({ activeView, onViewChange, onOpenConnections, terminalOpen, onToggleTerminal }: Props) {
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

      <div className="p-3 flex-1 overflow-y-auto space-y-1">
        <NavButton active={activeView === "dashboard"} icon={LayoutDashboard} label="Dashboard" onClick={() => onViewChange("dashboard")} />
        <NavButton active={activeView === "ec2"} icon={Monitor} label="File Explorer" onClick={() => onViewChange("ec2")} />
        <NavButton active={activeView === "s3"} icon={HardDrive} label="S3 Browser" onClick={() => onViewChange("s3")} />
        <NavButton active={activeView === "rds"} icon={Database} label="RDS PostgreSQL" onClick={() => onViewChange("rds")} />
      </div>

      <div className="p-3 border-t border-slate-700 space-y-2">
        <button
          onClick={onToggleTerminal}
          className={`w-full flex items-center justify-center gap-1.5 px-3 py-1.5 rounded text-sm transition-colors ${
            terminalOpen
              ? "bg-cyan-500/10 text-cyan-400 border border-cyan-500/30"
              : "bg-slate-800 hover:bg-slate-700 text-slate-300"
          }`}
        >
          <TerminalSquare size={14} />
          Terminal
        </button>
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
