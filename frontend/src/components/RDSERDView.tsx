import { useCallback, useMemo } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  type Node,
  type Edge,
  type NodeTypes,
  type NodeProps,
  Handle,
  Position,
  useNodesState,
  useEdgesState,
  MarkerType,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import dagre from "@dagrejs/dagre";
import { Copy, KeyRound, Link } from "lucide-react";
import type { RDSERDResponse, RDSERDTable, RDSERDEdge } from "../types";

const NODE_WIDTH = 260;
const COLUMN_ROW_HEIGHT = 24;
const HEADER_HEIGHT = 36;
const PADDING_BOTTOM = 8;

function nodeHeight(table: RDSERDTable): number {
  return HEADER_HEIGHT + table.columns.length * COLUMN_ROW_HEIGHT + PADDING_BOTTOM;
}

/** Dagre auto-layout: compute x/y for each node. */
function layoutNodes(
  tables: RDSERDTable[],
  edges: RDSERDEdge[],
): { nodes: Node[]; edges: Edge[] } {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: "TB", nodesep: 60, ranksep: 80, marginx: 40, marginy: 40 });

  for (const t of tables) {
    const h = nodeHeight(t);
    g.setNode(t.name, { width: NODE_WIDTH, height: h });
  }

  for (const e of edges) {
    g.setEdge(e.from_table, e.to_table);
  }

  dagre.layout(g);

  const flowNodes: Node[] = tables.map((t) => {
    const pos = g.node(t.name);
    const h = nodeHeight(t);
    return {
      id: t.name,
      type: "tableNode",
      position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - h / 2 },
      data: { table: t },
    };
  });

  const flowEdges: Edge[] = edges.map((e, i) => ({
    id: `edge-${i}`,
    source: e.from_table,
    target: e.to_table,
    sourceHandle: `${e.from_table}-${e.from_column}-source`,
    targetHandle: `${e.to_table}-${e.to_column}-target`,
    label: `${e.from_column} → ${e.to_column}`,
    type: "smoothstep",
    animated: true,
    markerEnd: { type: MarkerType.ArrowClosed, color: "#06b6d4" },
    style: { stroke: "#334155", strokeWidth: 1.5 },
    labelStyle: { fill: "#94a3b8", fontSize: 10, fontWeight: 500 },
    labelBgStyle: { fill: "#0f172a", fillOpacity: 0.9 },
    labelBgPadding: [6, 3] as [number, number],
    labelBgBorderRadius: 4,
  }));

  return { nodes: flowNodes, edges: flowEdges };
}

/** Custom node: renders a table card with columns. */
function TableNode({ data, selected }: NodeProps) {
  const table = (data as { table: RDSERDTable }).table;

  return (
    <div
      className={`rounded-lg border overflow-hidden shadow-lg transition-colors ${
        selected
          ? "border-cyan-400 shadow-cyan-500/20"
          : "border-slate-600 hover:border-slate-500"
      }`}
      style={{ width: NODE_WIDTH, background: "#1e293b" }}
    >
      {/* Header */}
      <div
        className="px-3 py-2 text-sm font-semibold text-slate-100 border-b border-slate-600 flex items-center gap-2"
        style={{ height: HEADER_HEIGHT, background: "#0f172a" }}
      >
        <span className="truncate">{table.name}</span>
        <span className="ml-auto text-xs text-slate-500">{table.columns.length} cols</span>
      </div>

      {/* Columns */}
      <div>
        {table.columns.map((col) => (
          <div
            key={col.name}
            className="relative flex items-center px-3 text-xs border-b border-slate-700/50 last:border-b-0 hover:bg-slate-700/30"
            style={{ height: COLUMN_ROW_HEIGHT }}
          >
            <Handle
              type="target"
              position={Position.Left}
              id={`${table.name}-${col.name}-target`}
              className="!w-2 !h-2 !bg-cyan-500 !border-slate-800"
              style={{ left: -4 }}
            />
            <div className="flex items-center gap-1.5 min-w-0 flex-1">
              {col.is_primary_key && <KeyRound size={10} className="text-amber-400 shrink-0" />}
              {col.is_foreign_key && <Link size={10} className="text-cyan-400 shrink-0" />}
              <span className="text-slate-200 truncate">{col.name}</span>
            </div>
            <span className="text-slate-500 font-mono ml-2 shrink-0">{col.data_type}</span>
            <Handle
              type="source"
              position={Position.Right}
              id={`${table.name}-${col.name}-source`}
              className="!w-2 !h-2 !bg-cyan-500 !border-slate-800"
              style={{ right: -4 }}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

const nodeTypes: NodeTypes = { tableNode: TableNode };

export function RDSERDView({
  data,
  onTableClick,
}: {
  data: RDSERDResponse;
  onTableClick?: (tableName: string) => void;
}) {
  const { nodes: initialNodes, edges: initialEdges } = useMemo(
    () => layoutNodes(data.tables, data.edges),
    [data],
  );

  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, , onEdgesChange] = useEdgesState(initialEdges);

  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      onTableClick?.(node.id);
    },
    [onTableClick],
  );

  const copySource = () => {
    const lines = ["erDiagram"];
    for (const t of data.tables) {
      lines.push(`  ${t.name} {`);
      for (const c of t.columns) {
        const marks = [c.is_primary_key ? "PK" : null, c.is_foreign_key ? "FK" : null]
          .filter(Boolean)
          .join(",");
        lines.push(`    ${c.data_type} ${c.name}${marks ? ` ${marks}` : ""}`);
      }
      lines.push("  }");
    }
    for (const e of data.edges) {
      lines.push(`  ${e.from_table} }o--|| ${e.to_table} : "${e.from_column}"`);
    }
    navigator.clipboard?.writeText(lines.join("\n")).catch(() => {});
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
        {onTableClick && (
          <span className="text-slate-500">click a table to view details</span>
        )}
        <button
          onClick={copySource}
          className="ml-auto flex items-center gap-1 text-slate-400 hover:text-white px-2 py-1 rounded hover:bg-slate-800/60"
          title="Copy mermaid source"
        >
          <Copy size={12} />
          Copy mermaid
        </button>
      </div>

      <div
        className="rounded border border-slate-700 bg-slate-950/60 overflow-hidden"
        style={{ height: "70vh", minHeight: 400 }}
      >
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={handleNodeClick}
          nodeTypes={nodeTypes}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          minZoom={0.1}
          maxZoom={2}
          proOptions={{ hideAttribution: true }}
          defaultEdgeOptions={{ type: "smoothstep" }}
        >
          <Background color="#1e293b" gap={20} size={1} />
          <Controls
            showInteractive={false}
            className="!bg-slate-800 !border-slate-700 !shadow-lg [&>button]:!bg-slate-800 [&>button]:!border-slate-600 [&>button]:!text-slate-300 [&>button:hover]:!bg-slate-700"
          />
          <MiniMap
            nodeColor="#1e293b"
            nodeStrokeColor="#334155"
            maskColor="rgba(15, 23, 42, 0.7)"
            className="!bg-slate-900 !border-slate-700"
          />
        </ReactFlow>
      </div>
    </div>
  );
}
