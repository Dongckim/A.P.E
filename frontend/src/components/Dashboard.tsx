import { RefreshCw, Pause, Play, Loader2, AlertCircle } from "lucide-react";
import { useDashboard } from "../hooks/useDashboard";
import { SystemOverview } from "./dashboard/SystemOverview";
import { ServicesList } from "./dashboard/ServicesList";
import { GitPanel } from "./dashboard/GitPanel";
import { ProcessTable } from "./dashboard/ProcessTable";

export function Dashboard() {
  const {
    overview, services, gitInfo, processes,
    loading, error, lastRefresh,
    paused, setPaused,
    gitPath, updateGitPath,
    refresh,
  } = useDashboard();

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <h2 className="text-sm font-medium text-white">Dashboard</h2>
          {overview?.hostname && (
            <span className="text-xs text-slate-400">{overview.hostname}</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-slate-500">
            {lastRefresh.toLocaleTimeString()}
          </span>
          <button
            onClick={() => setPaused(!paused)}
            className={`p-1.5 rounded transition-colors ${
              paused ? "bg-yellow-500/10 text-yellow-400" : "text-slate-400 hover:text-white hover:bg-slate-700"
            }`}
            title={paused ? "Resume auto-refresh" : "Pause auto-refresh"}
          >
            {paused ? <Play size={14} /> : <Pause size={14} />}
          </button>
          <button
            onClick={refresh}
            className="p-1.5 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
            title="Refresh now"
          >
            <RefreshCw size={14} />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4">
        {/* Logo */}
        <div className="flex flex-col items-center pt-6 pb-4">
          <pre className="text-cyan-400 leading-[1.1] text-[clamp(0.45rem,1.8vw,0.85rem)] font-bold select-none">{[
"    ▄▄██████████▄▄",
"  ▄████████████████▄",
" ████████████████████",
" ███ (◕)    (◕) ███",
" ████     ▄▄     ████",
"  ████ ┌──────┐ ████",
"   ████│ ━━━━ │████",
"    ▀██└──────┘██▀",
"      ▀████████▀",
].join("\n")}</pre>
          <pre className="text-white leading-[1.15] text-[clamp(0.5rem,2vw,0.95rem)] font-bold mt-3 select-none">{[
" ██████  ██████  ██████",
" ██  ██  ██  ██  ██",
" ██████  ██████  ████",
" ██  ██  ██      ██",
" ██  ██  ██      ██████",
].join("\n")}</pre>
          <p className="text-slate-400 text-sm mt-3 tracking-widest">AWS Platform Explorer</p>
        </div>

        {/* Loading / Error */}
        {loading && !overview && (
          <div className="flex items-center justify-center py-12 text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}

        {error && !overview && (
          <div className="flex items-center justify-center py-12 gap-2 text-red-400">
            <AlertCircle size={18} />
            <span>{error}</span>
          </div>
        )}

        {/* Dashboard grid */}
        {overview && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <SystemOverview data={overview} />
            <ServicesList data={services || { available_runtimes: [], docker: null, pm2: null, systemd: null }} />
            <GitPanel data={gitInfo} gitPath={gitPath} onPathChange={updateGitPath} />
            <ProcessTable data={processes || []} />
          </div>
        )}
      </div>
    </div>
  );
}
