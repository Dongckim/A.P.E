import { Activity } from "lucide-react";
import type { ProcessInfo } from "../../types";

interface Props {
  data: ProcessInfo[];
}

export function ProcessTable({ data }: Props) {
  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
      <h3 className="text-sm font-medium text-white mb-3 flex items-center gap-1.5">
        <Activity size={14} /> Top Processes
      </h3>

      <div className="overflow-x-auto max-h-52 overflow-y-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-slate-500 border-b border-slate-700">
              <th className="text-left py-1 pr-2 font-medium">PID</th>
              <th className="text-left py-1 pr-2 font-medium">User</th>
              <th className="text-right py-1 pr-2 font-medium">CPU%</th>
              <th className="text-right py-1 pr-2 font-medium">MEM%</th>
              <th className="text-left py-1 font-medium">Command</th>
            </tr>
          </thead>
          <tbody>
            {data.map((p, i) => (
              <tr key={p.pid + i} className="border-b border-slate-800 hover:bg-slate-700/30">
                <td className="py-1 pr-2 text-slate-400 font-mono">{p.pid}</td>
                <td className="py-1 pr-2 text-slate-400">{p.user}</td>
                <td className="py-1 pr-2 text-right">
                  <span className={p.cpu > 50 ? "text-red-400" : p.cpu > 20 ? "text-yellow-400" : "text-slate-300"}>
                    {p.cpu.toFixed(1)}
                  </span>
                </td>
                <td className="py-1 pr-2 text-right">
                  <span className={p.memory > 50 ? "text-red-400" : p.memory > 20 ? "text-yellow-400" : "text-slate-300"}>
                    {p.memory.toFixed(1)}
                  </span>
                </td>
                <td className="py-1 text-slate-300 font-mono truncate max-w-[200px]">{p.command}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {data.length === 0 && (
          <p className="text-xs text-slate-500 text-center py-4">No process data</p>
        )}
      </div>
    </div>
  );
}
