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
}

let nextId = 1;

function sendResize(ws: WebSocket, term: XTerm) {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }));
  }
}

export function TerminalPanel({ onClose }: { onClose: () => void }) {
  // Sessions are stored in a ref to avoid re-renders on mutable object changes.
  // Only the ID list and status map live in state (for UI rendering).
  const sessionsRef = useRef<Map<number, TermSession>>(new Map());
  const [sessionIds, setSessionIds] = useState<number[]>([]);
  const [statuses, setStatuses] = useState<Record<number, string>>({});
  const [activeId, setActiveId] = useState<number | null>(null);
  const [heightIdx, setHeightIdx] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);
  const attachedIdRef = useRef<number | null>(null);
  const panelHeight = PANEL_HEIGHTS[heightIdx];

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
      setStatuses((prev) => ({ ...prev, [id]: "connected" }));
      sendResize(ws, term);
    };

    ws.onmessage = (ev) => {
      if (ev.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(ev.data));
      } else {
        term.write(ev.data as string);
      }
    };

    ws.onclose = () => {
      setStatuses((prev) => ({ ...prev, [id]: "disconnected" }));
      term.write("\r\n\x1b[90m--- session ended ---\x1b[0m\r\n");
    };

    ws.onerror = () => {
      setStatuses((prev) => ({ ...prev, [id]: "disconnected" }));
    };

    const encoder = new TextEncoder();
    term.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(encoder.encode(data));
      }
    });

    sessionsRef.current.set(id, { id, label: `Terminal ${id}`, term, fitAddon, ws });
    setStatuses((prev) => ({ ...prev, [id]: "connecting" }));
    setSessionIds((prev) => [...prev, id]);
    setActiveId(id);
  }, []);

  const removeSession = useCallback((id: number) => {
    const session = sessionsRef.current.get(id);
    if (session) {
      session.ws.close();
      session.term.dispose();
      sessionsRef.current.delete(id);
    }
    if (attachedIdRef.current === id) {
      attachedIdRef.current = null;
    }
    setSessionIds((prev) => {
      const next = prev.filter((sid) => sid !== id);
      setActiveId((curActive) => {
        if (curActive === id) return next.length > 0 ? next[next.length - 1] : null;
        return curActive;
      });
      return next;
    });
    setStatuses((prev) => {
      const next = { ...prev };
      delete next[id];
      return next;
    });
  }, []);

  // Auto-create first session on mount
  useEffect(() => {
    addSession();
    return () => {
      for (const s of sessionsRef.current.values()) {
        s.ws.close();
        s.term.dispose();
      }
      sessionsRef.current.clear();
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Attach active terminal to DOM — only when activeId changes
  useEffect(() => {
    const el = containerRef.current;
    if (!el || activeId === null) return;
    if (attachedIdRef.current === activeId) return;

    const session = sessionsRef.current.get(activeId);
    if (!session) return;

    attachedIdRef.current = activeId;
    el.replaceChildren();
    session.term.open(el);
    session.term.focus();

    const raf = requestAnimationFrame(() => {
      try {
        session.fitAddon.fit();
        sendResize(session.ws, session.term);
      } catch {
        // ignore if container not ready
      }
    });
    return () => cancelAnimationFrame(raf);
  }, [activeId]);

  // Re-fit when panel height changes
  useEffect(() => {
    if (activeId === null) return;
    const session = sessionsRef.current.get(activeId);
    if (!session) return;
    const timer = setTimeout(() => {
      try {
        session.fitAddon.fit();
        sendResize(session.ws, session.term);
      } catch {
        // ignore
      }
    }, 50);
    return () => clearTimeout(timer);
  }, [panelHeight, activeId]);

  // Resize observer for width changes
  useEffect(() => {
    const el = containerRef.current;
    if (!el || activeId === null) return;

    const observer = new ResizeObserver(() => {
      const session = sessionsRef.current.get(activeId);
      if (!session) return;
      try {
        session.fitAddon.fit();
        sendResize(session.ws, session.term);
      } catch {
        // ignore
      }
    });
    observer.observe(el);
    return () => observer.disconnect();
  }, [activeId]);

  return (
    <div
      className="absolute bottom-0 left-0 right-0 border-t border-slate-700 bg-slate-900 flex flex-col z-40"
      style={{ height: panelHeight }}
    >
      {/* Header bar */}
      <div className="flex items-center px-2 bg-slate-800/80 border-b border-slate-700 shrink-0">
        <div className="flex items-center gap-1 overflow-x-auto flex-1 py-1">
          {sessionIds.map((id) => {
            const status = statuses[id] ?? "connecting";
            const label = sessionsRef.current.get(id)?.label ?? `Terminal ${id}`;
            return (
              <button
                key={id}
                onClick={() => setActiveId(id)}
                className={`flex items-center gap-1.5 px-2.5 py-0.5 rounded text-xs whitespace-nowrap transition-colors ${
                  id === activeId
                    ? "bg-slate-700 text-white"
                    : "text-slate-400 hover:text-white hover:bg-slate-800"
                }`}
              >
                <TerminalSquare size={11} />
                <span>{label}</span>
                <span
                  className={`w-1.5 h-1.5 rounded-full ${
                    status === "connecting"
                      ? "bg-amber-400 animate-pulse"
                      : status === "connected"
                        ? "bg-emerald-400"
                        : "bg-red-400"
                  }`}
                />
                <span
                  onClick={(e) => {
                    e.stopPropagation();
                    removeSession(id);
                  }}
                  className="p-0.5 rounded hover:bg-slate-600 text-slate-500 hover:text-white"
                >
                  <X size={9} />
                </span>
              </button>
            );
          })}
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

      {/* Terminal render area */}
      <div
        ref={containerRef}
        className="flex-1 min-h-0 overflow-hidden"
        style={{ background: "#0f172a", padding: 4 }}
      />
    </div>
  );
}
