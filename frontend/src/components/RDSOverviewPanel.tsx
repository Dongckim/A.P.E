import { useEffect, useState, useCallback } from "react";
import { AlertCircle, ArrowLeft, Database, Loader2, RefreshCw, Table } from "lucide-react";
import type { RDSOverview, RDSTablesResponse } from "../types";
import { fetchRDSOverview, fetchRDSTables } from "../api/client";

export function RDSOverviewPanel() {
  const [data, setData] = useState<RDSOverview | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // selectedDb is the database name the user clicked. Empty string = factory default.
  const [selectedDb, setSelectedDb] = useState<string>("");
  // selectedSchema drives the drill-down view. null = overview; non-null = tables list.
  const [selectedSchema, setSelectedSchema] = useState<string | null>(null);
  // tables data for drill-down view
  const [tables, setTables] = useState<RDSTablesResponse | null>(null);
  const [tablesLoading, setTablesLoading] = useState(false);
  const [tablesError, setTablesError] = useState<string | null>(null);

  const load = useCallback(async (db: string) => {
    setLoading(true);
    setError(null);
    try {
      const out = await fetchRDSOverview(db || undefined);
      setData(out);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load RDS overview");
    } finally {
      setLoading(false);
    }
  }, []);

  const loadTables = useCallback(async (schema: string, db: string) => {
    setTablesLoading(true);
    setTablesError(null);
    try {
      const out = await fetchRDSTables(schema, db || undefined);
      setTables(out);
    } catch (e) {
      setTablesError(e instanceof Error ? e.message : "Failed to load tables");
    } finally {
      setTablesLoading(false);
    }
  }, []);

  useEffect(() => {
    load(selectedDb);
  }, [load, selectedDb]);

  useEffect(() => {
    if (selectedSchema) {
      loadTables(selectedSchema, selectedDb);
    } else {
      setTables(null);
      setTablesError(null);
    }
  }, [selectedSchema, selectedDb, loadTables]);

  // Reset drill-down when DB changes.
  useEffect(() => {
    setSelectedSchema(null);
  }, [selectedDb]);

  const refreshing = loading || tablesLoading;
  const handleRefresh = () => {
    if (selectedSchema) {
      loadTables(selectedSchema, selectedDb);
    } else {
      load(selectedDb);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center px-4 py-2 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-2 text-sm text-slate-300 font-medium">
          <Database size={14} className="text-cyan-400" />
          RDS PostgreSQL Overview
        </div>
        <button
          onClick={handleRefresh}
          disabled={refreshing}
          className="ml-auto p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white disabled:opacity-60"
          title="Refresh"
        >
          <RefreshCw size={14} className={refreshing ? "animate-spin" : ""} />
        </button>
      </div>

      <div className="flex-1 overflow-auto p-4 space-y-4">
        {selectedSchema ? (
          <TablesDrilldown
            schema={selectedSchema}
            db={data?.current_db || selectedDb || ""}
            loading={tablesLoading}
            error={tablesError}
            data={tables}
            onBack={() => setSelectedSchema(null)}
          />
        ) : (
          <>
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
              <RDSOverviewLoaded
                data={data}
                onSelectDb={(name) => setSelectedDb(name)}
                onSelectSchema={(name) => setSelectedSchema(name)}
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}

function RDSOverviewLoaded({
  data,
  onSelectDb,
  onSelectSchema,
}: {
  data: RDSOverview;
  onSelectDb: (name: string) => void;
  onSelectSchema: (name: string) => void;
}) {
  const schemas = data.schemas ?? [];
  const databases = data.databases ?? [];

  return (
    <>
      {!data.connected && (
        <div className="flex items-start gap-2 rounded border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-200">
          <AlertCircle size={16} className="mt-0.5 shrink-0" />
          <span>{data.error || "RDS is not connected."}</span>
        </div>
      )}

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <StatCard label="Current DB" value={data.current_db || "-"} />
        <StatCard label="Databases" value={String(databases.length)} />
        <StatCard label="Schemas" value={String(data.schema_count)} />
        <StatCard label="Tables" value={String(data.table_count)} />
      </div>

      <div className="rounded border border-slate-700 bg-slate-900/40">
        <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700">
          PostgreSQL Version
        </div>
        <div className="px-3 py-2 text-sm text-slate-200 break-all">{data.version || "-"}</div>
      </div>

      <div className="rounded border border-slate-700 bg-slate-900/40 overflow-hidden">
        <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700">
          Databases on this Instance
        </div>
        {databases.length === 0 ? (
          <div className="px-3 py-6 text-sm text-slate-500 text-center">No accessible databases.</div>
        ) : (
          <table className="w-full text-sm">
            <thead className="text-slate-400 bg-slate-800/30">
              <tr>
                <th className="text-left px-3 py-2 font-medium">Database</th>
                <th className="text-right px-3 py-2 font-medium">Size</th>
              </tr>
            </thead>
            <tbody>
              {databases.map((db) => (
                <tr
                  key={db.name}
                  onClick={() => onSelectDb(db.name)}
                  className={`border-t border-slate-800 cursor-pointer transition-colors ${
                    db.is_current ? "bg-cyan-500/5" : "hover:bg-slate-800/50"
                  }`}
                  title={`Switch to ${db.name}`}
                >
                  <td className="px-3 py-2 text-slate-200">
                    {db.name}
                    {db.is_current && (
                      <span className="ml-2 text-xs px-1.5 py-0.5 rounded bg-cyan-500/20 text-cyan-300 border border-cyan-500/30">
                        connected
                      </span>
                    )}
                  </td>
                  <td className="px-3 py-2 text-right text-slate-300">{db.size_pretty}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      <div className="rounded border border-slate-700 bg-slate-900/40 overflow-hidden">
        <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700">
          Top Schemas in {data.current_db || "current DB"}
        </div>
        {schemas.length === 0 ? (
          <div className="px-3 py-6 text-sm text-slate-500 text-center">
            No user schemas in <code className="text-slate-400">{data.current_db || "this database"}</code>.
            {databases.length > 1 && " Try connecting to a different database from the list above."}
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead className="text-slate-400 bg-slate-800/30">
              <tr>
                <th className="text-left px-3 py-2 font-medium">Schema</th>
                <th className="text-right px-3 py-2 font-medium">Tables</th>
              </tr>
            </thead>
            <tbody>
              {schemas.map((schema) => (
                <tr
                  key={schema.name}
                  onClick={() => onSelectSchema(schema.name)}
                  className="border-t border-slate-800 cursor-pointer hover:bg-slate-800/50 transition-colors"
                  title={`View tables in ${schema.name}`}
                >
                  <td className="px-3 py-2 text-slate-200">{schema.name}</td>
                  <td className="px-3 py-2 text-right text-slate-300">{schema.table_count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </>
  );
}

function TablesDrilldown({
  schema,
  db,
  loading,
  error,
  data,
  onBack,
}: {
  schema: string;
  db: string;
  loading: boolean;
  error: string | null;
  data: RDSTablesResponse | null;
  onBack: () => void;
}) {
  return (
    <>
      <div className="flex items-center gap-2">
        <button
          onClick={onBack}
          className="flex items-center gap-1 text-sm text-slate-300 hover:text-white px-2 py-1 rounded hover:bg-slate-800/60"
        >
          <ArrowLeft size={14} />
          Back
        </button>
        <div className="text-sm text-slate-400">
          <span className="text-slate-500">{db || "current"}</span>
          <span className="mx-1 text-slate-600">/</span>
          <span className="text-slate-200 font-medium">{schema}</span>
        </div>
      </div>

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
        <div className="rounded border border-slate-700 bg-slate-900/40 overflow-hidden">
          <div className="px-3 py-2 text-xs uppercase tracking-wide text-slate-400 border-b border-slate-700 flex items-center gap-2">
            <Table size={12} />
            Tables in {schema} ({data.tables.length})
          </div>
          {data.tables.length === 0 ? (
            <div className="px-3 py-6 text-sm text-slate-500 text-center">
              No tables in <code className="text-slate-400">{schema}</code>.
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead className="text-slate-400 bg-slate-800/30">
                <tr>
                  <th className="text-left px-3 py-2 font-medium">Table</th>
                  <th className="text-right px-3 py-2 font-medium">Rows (est.)</th>
                  <th className="text-right px-3 py-2 font-medium">Size</th>
                </tr>
              </thead>
              <tbody>
                {data.tables.map((t) => (
                  <tr key={t.name} className="border-t border-slate-800">
                    <td className="px-3 py-2 text-slate-200">{t.name}</td>
                    <td className="px-3 py-2 text-right text-slate-300 tabular-nums">
                      {t.row_estimate < 0 ? "—" : t.row_estimate.toLocaleString()}
                    </td>
                    <td className="px-3 py-2 text-right text-slate-300 tabular-nums">{t.size_pretty}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
          <div className="px-3 py-2 text-xs text-slate-500 border-t border-slate-700 bg-slate-900/60">
            Row counts are estimates from <code className="text-slate-400">pg_class.reltuples</code> (updated by ANALYZE/autovacuum).
          </div>
        </div>
      )}
    </>
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
