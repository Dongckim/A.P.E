import { Cpu, MemoryStick, HardDrive, Clock } from "lucide-react";
import type { DashboardOverview } from "../../types";

interface Props {
  data: DashboardOverview;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(0) + " MB";
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + " GB";
}

function formatUptime(since: string): string {
  if (!since) return "--";
  const start = new Date(since.replace(" ", "T"));
  const diff = Date.now() - start.getTime();
  const days = Math.floor(diff / 86400000);
  const hours = Math.floor((diff % 86400000) / 3600000);
  if (days > 0) return `${days}d ${hours}h`;
  const mins = Math.floor((diff % 3600000) / 60000);
  return `${hours}h ${mins}m`;
}

function Bar({ percent, color }: { percent: number; color: string }) {
  return (
    <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
      <div className={`h-full rounded-full ${color}`} style={{ width: `${Math.min(percent, 100)}%` }} />
    </div>
  );
}

export function SystemOverview({ data }: Props) {
  const memPercent = data.memory.total > 0 ? (data.memory.used / data.memory.total) * 100 : 0;

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
        {data.hostname} — {data.os} ({data.arch})
      </div>

      <div className="space-y-4">
        {/* CPU */}
        <div>
          <div className="flex items-center justify-between mb-1">
            <span className="flex items-center gap-1.5 text-xs text-slate-400">
              <Cpu size={12} /> CPU ({data.cpu.cores} cores)
            </span>
            <span className="text-xs text-white">{data.cpu.usage_percent.toFixed(1)}%</span>
          </div>
          <Bar percent={data.cpu.usage_percent} color={data.cpu.usage_percent > 80 ? "bg-red-500" : data.cpu.usage_percent > 50 ? "bg-yellow-500" : "bg-cyan-500"} />
        </div>

        {/* Memory */}
        <div>
          <div className="flex items-center justify-between mb-1">
            <span className="flex items-center gap-1.5 text-xs text-slate-400">
              <MemoryStick size={12} /> Memory
            </span>
            <span className="text-xs text-white">{formatBytes(data.memory.used)} / {formatBytes(data.memory.total)}</span>
          </div>
          <Bar percent={memPercent} color={memPercent > 80 ? "bg-red-500" : memPercent > 50 ? "bg-yellow-500" : "bg-cyan-500"} />
        </div>

        {/* Disks */}
        {data.disks?.map((disk) => (
          <div key={disk.mount}>
            <div className="flex items-center justify-between mb-1">
              <span className="flex items-center gap-1.5 text-xs text-slate-400">
                <HardDrive size={12} /> {disk.mount}
              </span>
              <span className="text-xs text-white">{disk.percent}%</span>
            </div>
            <Bar percent={disk.percent} color={disk.percent > 90 ? "bg-red-500" : disk.percent > 70 ? "bg-yellow-500" : "bg-cyan-500"} />
          </div>
        ))}
      </div>
    </div>
  );
}
