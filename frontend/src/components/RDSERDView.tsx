import { useEffect, useRef, useState } from "react";
import { AlertCircle, Copy, Loader2 } from "lucide-react";
import type { RDSERDResponse } from "../types";

/** Mermaid identifier rule: alphanumeric + underscore. Otherwise wrap in quotes. */
function safeIdent(s: string): string {
  if (/^[a-zA-Z_][a-zA-Z0-9_]*$/.test(s)) return s;
  return `"${s.replace(/"/g, '\\"')}"`;
}

/** Mermaid type token must match [a-zA-Z_][a-zA-Z0-9_]*. */
function safeType(s: string): string {
  return s.replace(/[^a-zA-Z0-9_]/g, "_") || "unknown";
}

/** Build mermaid erDiagram source from the API response. */
export function buildERDSource(data: RDSERDResponse): string {
  const lines: string[] = ["erDiagram"];

  for (const t of data.tables) {
    lines.push(`  ${safeIdent(t.name)} {`);
    for (const c of t.columns) {
      const marks = [
        c.is_primary_key ? "PK" : null,
        c.is_foreign_key ? "FK" : null,
      ]
        .filter(Boolean)
        .join(",");
      const tail = marks ? ` ${marks}` : "";
      lines.push(`    ${safeType(c.data_type)} ${safeIdent(c.name)}${tail}`);
    }
    lines.push(`  }`);
  }

  for (const e of data.edges) {
    // Child rows reference one parent row → child }o--|| parent.
    // Mermaid `: "label"` shows on the edge.
    lines.push(
      `  ${safeIdent(e.from_table)} }o--|| ${safeIdent(e.to_table)} : "${e.from_column}"`,
    );
  }

  return lines.join("\n");
}

export function RDSERDView({ data }: { data: RDSERDResponse }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [svg, setSvg] = useState<string | null>(null);
  const [renderError, setRenderError] = useState<string | null>(null);
  const [rendering, setRendering] = useState(true);
  const source = buildERDSource(data);

  useEffect(() => {
    let cancelled = false;
    setRendering(true);
    setRenderError(null);
    setSvg(null);

    // Lazy import to keep mermaid out of the initial bundle.
    import("mermaid")
      .then(async ({ default: mermaid }) => {
        mermaid.initialize({
          startOnLoad: false,
          theme: "dark",
          themeVariables: {
            background: "#0f172a",
            primaryColor: "#1e293b",
            primaryTextColor: "#e2e8f0",
            primaryBorderColor: "#334155",
            lineColor: "#64748b",
            secondaryColor: "#0f172a",
            tertiaryColor: "#1e293b",
          },
          er: {
            layoutDirection: "TB",
            entityPadding: 12,
            useMaxWidth: true,
          },
          securityLevel: "strict",
        });
        // Use a unique id per render to avoid mermaid's internal cache collisions.
        const id = `erd-${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
        try {
          const { svg } = await mermaid.render(id, source);
          if (!cancelled) {
            setSvg(svg);
            setRendering(false);
          }
        } catch (e) {
          if (!cancelled) {
            setRenderError(e instanceof Error ? e.message : String(e));
            setRendering(false);
          }
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setRenderError(`failed to load mermaid: ${e instanceof Error ? e.message : String(e)}`);
          setRendering(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [source]);

  const copySource = () => {
    navigator.clipboard?.writeText(source).catch(() => {});
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3 text-xs text-slate-400">
        <span>
          {data.tables.length} tables · {data.edges.length} foreign keys
          {data.truncated && (
            <span className="ml-2 px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30">
              truncated to {data.table_limit}
            </span>
          )}
        </span>
        <button
          onClick={copySource}
          className="ml-auto flex items-center gap-1 text-slate-400 hover:text-white px-2 py-1 rounded hover:bg-slate-800/60"
          title="Copy mermaid source"
        >
          <Copy size={12} />
          Copy mermaid
        </button>
      </div>

      <div className="rounded border border-slate-700 bg-slate-950/60 overflow-auto p-4">
        {rendering && (
          <div className="flex items-center justify-center h-64 text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}

        {!rendering && renderError && (
          <div className="space-y-2">
            <div className="flex items-start gap-2 rounded border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-300">
              <AlertCircle size={16} className="mt-0.5 shrink-0" />
              <span>Failed to render ERD: {renderError}</span>
            </div>
            <details className="text-xs text-slate-500">
              <summary className="cursor-pointer hover:text-slate-400">Show mermaid source</summary>
              <pre className="mt-2 p-2 bg-slate-900 rounded overflow-x-auto text-slate-400">
                {source}
              </pre>
            </details>
          </div>
        )}

        {!rendering && !renderError && svg && (
          <div
            ref={containerRef}
            className="erd-svg-host"
            // mermaid produces inline SVG that we trust (we built the source ourselves).
            dangerouslySetInnerHTML={{ __html: svg }}
          />
        )}
      </div>
    </div>
  );
}
