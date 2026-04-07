import { useEffect, useState } from "react";
import { AlertCircle, Database, Loader2, RefreshCw } from "lucide-react";
import type { RDSOverview } from "../types";
import { fetchRDSOverview } from "../api/client";

export function RDSOverviewPanel() {
  const [data, setData] = useState<RDSOverview | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      const out = await fetchRDSOverview();
      setData(out);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load RDS overview");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center px-4 py-2 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-2 text-sm text-slate-300 font-medium">
          <Database size={14} className="text-cyan-400" />
          RDS PostgreSQL Overview
        </div>
        <button
          onClick={load}
          disabled={loading}
          className="ml-auto p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white disabled:opacity-60"
          title="Refresh"
        >
          <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
        </button>
      </div>

      <div className="flex-1 overflow-auto p-4 space-y-4">
        {loading && (
          <div className="flex items-center justify-center h-64 text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}

        {!loading && error && (
          <div className="flex items-center gap-2 rounded border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300">
            <AlertCircle size={16} />
            <span>{error}</span>
          </div>
        )}

        {!loading && !error && data && (
          <>
            {!data.connected && (
              <div className="flex items-start gap-2 rounded border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-200">
                <AlertCircle size={16} className="mt-0.5 shrink-0" />
                <span>{data.error || "RDS is not connected."}</span>
              </div>
            )}

            <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
              <StatCard label="Database" value={data.current_db || "-"} />
              <StatCard label="Schemas" value={String(data.schema_count)} />
              <StatCard label="Tables" value={String(data.table_count)} />
              <StatCard label="Connected" value={data.connected ? "Yes" : "No"} />
            </div>

            <div className="rounded border border-slate-700 bg-slate-900/40">
              <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700">
                PostgreSQL Version
              </div>
              <div className="px-3 py-2 text-sm text-slate-200 break-all">{data.version || "-"}</div>
            </div>

            <div className="rounded border border-slate-700 bg-slate-900/40 overflow-hidden">
              <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700">
                Top Schemas by Table Count
              </div>
              {data.schemas.length === 0 ? (
                <div className="px-3 py-6 text-sm text-slate-500 text-center">No user schemas found.</div>
              ) : (
                <table className="w-full text-sm">
                  <thead className="text-slate-400 bg-slate-800/30">
                    <tr>
                      <th className="text-left px-3 py-2 font-medium">Schema</th>
                      <th className="text-right px-3 py-2 font-medium">Tables</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.schemas.map((schema) => (
                      <tr key={schema.name} className="border-t border-slate-800">
                        <td className="px-3 py-2 text-slate-200">{schema.name}</td>
                        <td className="px-3 py-2 text-right text-slate-300">{schema.table_count}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded border border-slate-700 bg-slate-900/40 px-3 py-2">
      <div className="text-xs uppercase tracking-wide text-slate-500">{label}</div>
      <div className="text-lg font-semibold text-slate-100 mt-1 truncate">{value}</div>
    </div>
  );
}
