import { useState } from "react";
import { Container, Layers, Server } from "lucide-react";
import type { ServicesData } from "../../types";

interface Props {
  data: ServicesData;
}

function StateDot({ state }: { state: string }) {
  const s = state.toLowerCase();
  const color = s === "running" || s === "online" || s === "active"
    ? "bg-green-400"
    : s === "exited" || s === "stopped" || s === "errored"
    ? "bg-red-400"
    : "bg-yellow-400";
  return <span className={`inline-block w-2 h-2 rounded-full ${color}`} />;
}

export function ServicesList({ data }: Props) {
  const tabs = data.available_runtimes;
  const [active, setActive] = useState(tabs[0] || "");

  if (tabs.length === 0) {
    return (
      <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
        <h3 className="text-sm font-medium text-white mb-3">Services</h3>
        <p className="text-xs text-slate-500 text-center py-6">No runtimes detected (Docker, PM2, systemd)</p>
      </div>
    );
  }

  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
      <h3 className="text-sm font-medium text-white mb-3">Services</h3>

      {/* Tabs */}
      <div className="flex gap-1 mb-3">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActive(tab)}
            className={`flex items-center gap-1 px-2.5 py-1 text-xs rounded transition-colors ${
              active === tab ? "bg-slate-600 text-white" : "text-slate-400 hover:text-white hover:bg-slate-700"
            }`}
          >
            {tab === "docker" && <Container size={12} />}
            {tab === "pm2" && <Layers size={12} />}
            {tab === "systemd" && <Server size={12} />}
            {tab}
          </button>
        ))}
      </div>

      {/* Docker */}
      {active === "docker" && data.docker && (
        data.docker.error ? (
          <p className="text-xs text-yellow-400 py-2">{data.docker.error}</p>
        ) : (
          <div className="space-y-1 max-h-48 overflow-y-auto">
            {data.docker.containers.map((c) => (
              <div key={c.id} className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-slate-700/50 text-xs">
                <StateDot state={c.state} />
                <span className="text-white font-medium w-28 truncate">{c.name}</span>
                <span className="text-slate-400 flex-1 truncate">{c.image}</span>
                <span className="text-slate-500 shrink-0">{c.status}</span>
              </div>
            ))}
            {data.docker.containers.length === 0 && (
              <p className="text-xs text-slate-500 text-center py-3">No containers</p>
            )}
          </div>
        )
      )}

      {/* PM2 */}
      {active === "pm2" && data.pm2 && (
        data.pm2.error ? (
          <p className="text-xs text-yellow-400 py-2">{data.pm2.error}</p>
        ) : (
          <div className="space-y-1 max-h-48 overflow-y-auto">
            {data.pm2.processes.map((p) => (
              <div key={p.id} className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-slate-700/50 text-xs">
                <StateDot state={p.status} />
                <span className="text-white font-medium w-28 truncate">{p.name}</span>
                <span className="text-slate-400 w-14">{p.mode}</span>
                <span className="text-slate-400 w-14">CPU {p.cpu}%</span>
                <span className="text-slate-500">restarts: {p.restarts}</span>
              </div>
            ))}
            {data.pm2.processes.length === 0 && (
              <p className="text-xs text-slate-500 text-center py-3">No processes</p>
            )}
          </div>
        )
      )}

      {/* Systemd */}
      {active === "systemd" && data.systemd && (
        <div className="space-y-1 max-h-48 overflow-y-auto">
          {data.systemd.services.map((s) => (
            <div key={s.unit} className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-slate-700/50 text-xs">
              <StateDot state={s.sub} />
              <span className="text-white font-medium w-40 truncate">{s.unit}</span>
              <span className="text-slate-400 flex-1 truncate">{s.description}</span>
            </div>
          ))}
          {data.systemd.services.length === 0 && (
            <p className="text-xs text-slate-500 text-center py-3">No running services</p>
          )}
        </div>
      )}
    </div>
  );
}
