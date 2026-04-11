import { useEffect, useRef, useState, useCallback } from "react";
import { Minus, Plus, X, TerminalSquare, ChevronUp, ChevronDown } from "lucide-react";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import "@xterm/xterm/css/xterm.css";

const WS_PROTO = window.location.protocol === "https:" ? "wss:" : "ws:";
const WS_BASE = `${WS_PROTO}//${window.location.host}/api/ec2/terminal`;

const PANEL_HEIGHTS = [200, 320, 480];

interface TermSession {
  id: number;
  label: string;
  term: XTerm;
  fitAddon: FitAddon;
  ws: WebSocket;
  status: "connecting" | "connected" | "disconnected";
}

let nextId = 1;

export function TerminalPanel({ onClose }: { onClose: () => void }) {
  const [sessions, setSessions] = useState<TermSession[]>([]);
  const [activeId, setActiveId] = useState<number | null>(null);
  const [heightIdx, setHeightIdx] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);
  const panelHeight = PANEL_HEIGHTS[heightIdx];
  const activeSession = sessions.find((s) => s.id === activeId) ?? null;

  const addSession = useCallback((connId?: string) => {
    const id = nextId++;

    const term = new XTerm({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, monospace",
      theme: {
        background: "#0f172a",
        foreground: "#e2e8f0",
        cursor: "#22d3ee",
        selectionBackground: "#334155",
        black: "#0f172a",
        red: "#f87171",
        green: "#4ade80",
        yellow: "#facc15",
        blue: "#60a5fa",
        magenta: "#c084fc",
        cyan: "#22d3ee",
        white: "#e2e8f0",
        brightBlack: "#475569",
        brightRed: "#fca5a5",
        brightGreen: "#86efac",
        brightYellow: "#fde68a",
        brightBlue: "#93c5fd",
        brightMagenta: "#d8b4fe",
        brightCyan: "#67e8f9",
        brightWhite: "#f8fafc",
      },
      scrollback: 5000,
      allowProposedApi: true,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.loadAddon(new WebLinksAddon());

    const params = connId ? `?conn=${encodeURIComponent(connId)}` : "";
    const ws = new WebSocket(`${WS_BASE}${params}`);
    ws.binaryType = "arraybuffer";

    ws.onopen = () => {
      setSessions((prev) =>
        prev.map((s) => (s.id === id ? { ...s, status: "connected" as const } : s)),
      );
      const msg = JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows });
      ws.send(msg);
    };

    ws.onmessage = (ev) => {
      if (ev.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(ev.data));
      } else {
        term.write(ev.data as string);
      }
    };

    ws.onclose = () => {
      setSessions((prev) =>
        prev.map((s) => (s.id === id ? { ...s, status: "disconnected" as const } : s)),
      );
      term.write("\r\n\x1b[90m--- session ended ---\x1b[0m\r\n");
    };

    ws.onerror = () => {
      setSessions((prev) =>
        prev.map((s) => (s.id === id ? { ...s, status: "disconnected" as const } : s)),
      );
    };

    const encoder = new TextEncoder();
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode(data));
      }
    });

    const session: TermSession = { id, label: `Terminal ${id}`, term, fitAddon, ws, status: "connecting" };
    setSessions((prev) => [...prev, session]);
    setActiveId(id);
  }, []);

  const removeSession = useCallback(
    (id: number) => {
      setSessions((prev) => {
        const session = prev.find((s) => s.id === id);
        if (session) {
          session.ws.close();
          session.term.dispose();
        }
        const next = prev.filter((s) => s.id !== id);
        if (activeId === id) {
          setActiveId(next.length > 0 ? next[next.length - 1].id : null);
        }
        return next;
      });
    },
    [activeId],
  );

  // Auto-create first session on mount
  useEffect(() => {
    addSession();
    return () => {
      setSessions((prev) => {
        for (const s of prev) {
          s.ws.close();
          s.term.dispose();
        }
        return [];
      });
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Attach active terminal to DOM — key fix: wait for container to have dimensions
  useEffect(() => {
    const el = containerRef.current;
    if (!el || !activeSession) return;

    el.replaceChildren();
    activeSession.term.open(el);

    // xterm needs the container to have pixel dimensions before fit
    const raf = requestAnimationFrame(() => {
      try {
        activeSession.fitAddon.fit();
        if (activeSession.ws.readyState === WebSocket.OPEN) {
          const msg = JSON.stringify({
            type: "resize",
            cols: activeSession.term.cols,
            rows: activeSession.term.rows,
          });
          activeSession.ws.send(msg);
        }
      } catch {
        // fit can throw if container has zero dimensions
      }
    });
    return () => cancelAnimationFrame(raf);
  }, [activeSession]);

  // Re-fit when panel height changes
  useEffect(() => {
    if (!activeSession) return;
    const timer = setTimeout(() => {
      try {
        activeSession.fitAddon.fit();
        if (activeSession.ws.readyState === WebSocket.OPEN) {
          const msg = JSON.stringify({
            type: "resize",
            cols: activeSession.term.cols,
            rows: activeSession.term.rows,
          });
          activeSession.ws.send(msg);
        }
      } catch {
        // ignore
      }
    }, 50);
    return () => clearTimeout(timer);
  }, [panelHeight, activeSession]);

  // Resize observer for width changes
  useEffect(() => {
    const el = containerRef.current;
    if (!el || !activeSession) return;

    const observer = new ResizeObserver(() => {
      try {
        activeSession.fitAddon.fit();
        if (activeSession.ws.readyState === WebSocket.OPEN) {
          const msg = JSON.stringify({
            type: "resize",
            cols: activeSession.term.cols,
            rows: activeSession.term.rows,
          });
          activeSession.ws.send(msg);
        }
      } catch {
        // ignore
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [activeSession]);

  return (
    <div
      className="absolute bottom-0 left-0 right-0 border-t border-slate-700 bg-slate-900 flex flex-col z-40"
      style={{ height: panelHeight }}
    >
      {/* Header bar */}
      <div className="flex items-center px-2 bg-slate-800/80 border-b border-slate-700 shrink-0">
        <div className="flex items-center gap-1 overflow-x-auto flex-1 py-1">
          {sessions.map((s) => (
            <button
              key={s.id}
              onClick={() => setActiveId(s.id)}
              className={`flex items-center gap-1.5 px-2.5 py-0.5 rounded text-xs whitespace-nowrap transition-colors ${
                s.id === activeId
                  ? "bg-slate-700 text-white"
                  : "text-slate-400 hover:text-white hover:bg-slate-800"
              }`}
            >
              <TerminalSquare size={11} />
              <span>{s.label}</span>
              <span
                className={`w-1.5 h-1.5 rounded-full ${
                  s.status === "connecting"
                    ? "bg-amber-400 animate-pulse"
                    : s.status === "connected"
                      ? "bg-emerald-400"
                      : "bg-red-400"
                }`}
              />
              <span
                onClick={(e) => {
                  e.stopPropagation();
                  removeSession(s.id);
                }}
                className="p-0.5 rounded hover:bg-slate-600 text-slate-500 hover:text-white"
              >
                <X size={9} />
              </span>
            </button>
          ))}
          <button
            onClick={() => addSession()}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white"
            title="New terminal"
          >
            <Plus size={12} />
          </button>
        </div>

        <div className="flex items-center gap-0.5 ml-2 shrink-0">
          <button
            onClick={() => setHeightIdx((i) => Math.min(i + 1, PANEL_HEIGHTS.length - 1))}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white"
            title="Expand"
          >
            <ChevronUp size={12} />
          </button>
          <button
            onClick={() => setHeightIdx((i) => Math.max(i - 1, 0))}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white"
            title="Shrink"
          >
            <ChevronDown size={12} />
          </button>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white"
            title="Close terminal"
          >
            <Minus size={12} />
          </button>
        </div>
      </div>

      {/* Terminal render area — explicit pixel height so xterm can measure */}
      <div
        ref={containerRef}
        className="flex-1 min-h-0 overflow-hidden"
        style={{ background: "#0f172a", padding: 4 }}
      />
    </div>
  );
}
