import { Cpu, MemoryStick, HardDrive, Clock } from "lucide-react";
import type { DashboardOverview } from "../../types";

interface Props {
  data: DashboardOverview;
}

function formatBytes(bytes: number): string {
  if (!bytes) return "0";
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(0) + " MB";
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
}

function formatUptime(since: string): string {
  if (!since) return "--";
  const start = new Date(since.replace(" ", "T"));
  if (isNaN(start.getTime())) return "--";
  const diff = Date.now() - start.getTime();
  const days = Math.floor(diff / 86400000);
  const hours = Math.floor((diff % 86400000) / 3600000);
  if (days > 0) return `${days}d ${hours}h`;
  const mins = Math.floor((diff % 3600000) / 60000);
  return `${hours}h ${mins}m`;
}

function Bar({ percent, color }: { percent: number; color: string }) {
  const p = isNaN(percent) ? 0 : Math.min(percent, 100);
  return (
    <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
      <div className={`h-full rounded-full ${color}`} style={{ width: `${p}%` }} />
    </div>
  );
}

function barColor(pct: number, high: number, mid: number): string {
  if (pct > high) return "bg-red-500";
  if (pct > mid) return "bg-yellow-500";
  return "bg-cyan-500";
}

export function SystemOverview({ data }: Props) {
  const cpu = data.cpu || { usage_percent: 0, cores: 0 };
  const mem = data.memory || { total: 0, used: 0, available: 0 };
  const memPercent = mem.total > 0 ? (mem.used / mem.total) * 100 : 0;
  const disks = data.disks || [];

  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-white">System</h3>
        <div className="flex items-center gap-1.5 text-xs text-slate-400">
          <Clock size={12} />
          <span>Up {formatUptime(data.uptime_since)}</span>
        </div>
      </div>

      <div className="text-xs text-slate-500 mb-4">
        {data.hostname || "unknown"} — {data.os || "unknown"} ({data.arch || "unknown"})
      </div>

      <div className="space-y-4">
        <div>
          <div className="flex items-center justify-between mb-1">
            <span className="flex items-center gap-1.5 text-xs text-slate-400">
              <Cpu size={12} /> CPU ({cpu.cores} cores)
            </span>
            <span className="text-xs text-white">{(cpu.usage_percent || 0).toFixed(1)}%</span>
          </div>
          <Bar percent={cpu.usage_percent || 0} color={barColor(cpu.usage_percent || 0, 80, 50)} />
        </div>

        <div>
          <div className="flex items-center justify-between mb-1">
            <span className="flex items-center gap-1.5 text-xs text-slate-400">
              <MemoryStick size={12} /> Memory
            </span>
            <span className="text-xs text-white">{formatBytes(mem.used)} / {formatBytes(mem.total)}</span>
          </div>
          <Bar percent={memPercent} color={barColor(memPercent, 80, 50)} />
        </div>

        {disks.map((disk) => (
          <div key={disk.mount}>
            <div className="flex items-center justify-between mb-1">
              <span className="flex items-center gap-1.5 text-xs text-slate-400">
                <HardDrive size={12} /> {disk.mount}
              </span>
              <span className="text-xs text-white">{disk.percent || 0}%</span>
            </div>
            <Bar percent={disk.percent || 0} color={barColor(disk.percent || 0, 90, 70)} />
          </div>
        ))}
      </div>
    </div>
  );
}
