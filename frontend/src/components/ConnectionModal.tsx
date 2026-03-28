import { useState, useEffect } from "react";
import { X, Monitor, Trash2, Loader2 } from "lucide-react";
import type { ConnectionInfo } from "../types";
import { listConnections, removeConnection } from "../api/client";

interface Props {
  onClose: () => void;
}

export function ConnectionModal({ onClose }: Props) {
  const [connections, setConnections] = useState<ConnectionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const data = await listConnections();
      setConnections(data);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, []);

  const handleRemove = async (id: string) => {
    try {
      await removeConnection(id);
      load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to disconnect");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div className="bg-slate-800 border border-slate-600 rounded-lg w-96 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
          <h3 className="text-white font-medium text-sm">Connections</h3>
          <button onClick={onClose} className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white">
            <X size={16} />
          </button>
        </div>

        <div className="p-4">
          {loading && (
            <div className="flex justify-center py-6 text-slate-400">
              <Loader2 size={20} className="animate-spin" />
            </div>
          )}

          {error && <p className="text-red-400 text-sm text-center py-2">{error}</p>}

          {!loading && connections.length === 0 && (
            <p className="text-slate-500 text-sm text-center py-6">No active connections</p>
          )}

          {!loading && connections.map((c) => (
            <div key={c.id} className="flex items-center gap-3 py-2 px-2 rounded hover:bg-slate-700/50 group">
              <Monitor size={16} className="text-cyan-400 shrink-0" />
              <div className="flex-1 min-w-0">
                <p className="text-sm text-slate-200 truncate">{c.host}</p>
                <p className="text-[11px] text-slate-500">{c.username}:{c.port}</p>
              </div>
              <span className="text-[10px] text-green-400 shrink-0">Connected</span>
              <button
                onClick={() => handleRemove(c.id)}
                className="p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-red-500/10 text-slate-400 hover:text-red-400 transition-all"
                title="Disconnect"
              >
                <Trash2 size={14} />
              </button>
            </div>
          ))}

          <p className="text-slate-500 text-[11px] mt-4 text-center">
            Use <code className="bg-slate-700 px-1 rounded text-slate-300">/add</code> in the CLI terminal to add new connections
          </p>
        </div>
      </div>
    </div>
  );
}
