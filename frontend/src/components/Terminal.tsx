import { useEffect, useRef, useState, useCallback } from "react";
import { AlertCircle, Plus, X, TerminalSquare } from "lucide-react";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from "@xterm/addon-web-links";
import "@xterm/xterm/css/xterm.css";

const WS_BASE = `ws://${window.location.host}/api/ec2/terminal`;

interface TermSession {
  id: number;
  label: string;
  term: XTerm;
  fitAddon: FitAddon;
  ws: WebSocket;
  status: "connecting" | "connected" | "disconnected";
}

let nextId = 1;

export function Terminal() {
  const [sessions, setSessions] = useState<TermSession[]>([]);
  const [activeId, setActiveId] = useState<number | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
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

    // Wire WebSocket events immediately (before adding to state)
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

    // Terminal input → WebSocket (binary)
    const encoder = new TextEncoder();
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode(data));
      }
    });

    const session: TermSession = {
      id,
      label: `Terminal ${id}`,
      term,
      fitAddon,
      ws,
      status: "connecting",
    };

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
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Attach active terminal to DOM
  useEffect(() => {
    const el = containerRef.current;
    if (!el || !activeSession) return;

    el.innerHTML = "";
    activeSession.term.open(el);

    requestAnimationFrame(() => {
      activeSession.fitAddon.fit();
    });
  }, [activeSession]);

  // Resize observer
  useEffect(() => {
    if (!containerRef.current || !activeSession) return;

    const observer = new ResizeObserver(() => {
      activeSession.fitAddon.fit();
      if (activeSession.ws.readyState === WebSocket.OPEN) {
        const msg = JSON.stringify({
          type: "resize",
          cols: activeSession.term.cols,
          rows: activeSession.term.rows,
        });
        activeSession.ws.send(msg);
      }
    });
    observer.observe(containerRef.current);
    return () => observer.disconnect();
  }, [activeSession]);

  return (
    <div className="flex flex-col h-full">
      {/* Tab bar */}
      <div className="flex items-center px-2 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-1 overflow-x-auto flex-1 py-1">
          {sessions.map((s) => (
            <button
              key={s.id}
              onClick={() => setActiveId(s.id)}
              className={`flex items-center gap-1.5 px-3 py-1 rounded text-xs whitespace-nowrap transition-colors ${
                s.id === activeId
                  ? "bg-slate-700 text-white"
                  : "text-slate-400 hover:text-white hover:bg-slate-800"
              }`}
            >
              <TerminalSquare size={12} />
              <span>{s.label}</span>
              {s.status === "connecting" && (
                <span className="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
              )}
              {s.status === "connected" && (
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
              )}
              {s.status === "disconnected" && (
                <span className="w-1.5 h-1.5 rounded-full bg-red-400" />
              )}
              <span
                onClick={(e) => {
                  e.stopPropagation();
                  removeSession(s.id);
                }}
                className="ml-1 p-0.5 rounded hover:bg-slate-600 text-slate-500 hover:text-white"
              >
                <X size={10} />
              </span>
            </button>
          ))}
        </div>
        <button
          onClick={() => addSession()}
          className="p-1.5 rounded hover:bg-slate-700 text-slate-400 hover:text-white ml-1 shrink-0"
          title="New terminal"
        >
          <Plus size={14} />
        </button>
      </div>

      {/* Terminal area */}
      {sessions.length === 0 ? (
        <div className="flex-1 flex items-center justify-center text-slate-500">
          <div className="text-center space-y-2">
            <AlertCircle size={32} className="mx-auto text-slate-600" />
            <p className="text-sm">No terminal sessions</p>
            <button
              onClick={() => addSession()}
              className="px-3 py-1.5 rounded bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 text-sm"
            >
              Open Terminal
            </button>
          </div>
        </div>
      ) : (
        <div ref={containerRef} className="flex-1 min-h-0 bg-[#0f172a] p-1" />
      )}
    </div>
  );
}
