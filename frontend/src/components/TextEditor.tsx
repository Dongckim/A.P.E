import { useState, useEffect, useCallback } from "react";
import Editor from "@monaco-editor/react";
import { X, Save, Loader2 } from "lucide-react";
import { readFile, writeFile } from "../api/client";

interface Props {
  filePath: string;
  onClose: () => void;
}

function getLang(path: string): string {
  const ext = path.split(".").pop()?.toLowerCase() || "";
  const map: Record<string, string> = {
    ts: "typescript", tsx: "typescript", js: "javascript", jsx: "javascript",
    py: "python", go: "go", rs: "rust", java: "java", c: "c", cpp: "cpp", h: "c",
    json: "json", yaml: "yaml", yml: "yaml", toml: "toml", xml: "xml",
    html: "html", css: "css", scss: "scss", md: "markdown",
    sh: "shell", bash: "shell", zsh: "shell",
    sql: "sql", dockerfile: "dockerfile",
  };
  return map[ext] || "plaintext";
}

export function TextEditor({ filePath, onClose }: Props) {
  const [content, setContent] = useState("");
  const [original, setOriginal] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    readFile(filePath)
      .then((data) => {
        setContent(data.content);
        setOriginal(data.content);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, [filePath]);

  const handleSave = useCallback(async () => {
    setSaving(true);
    try {
      await writeFile(filePath, content);
      setOriginal(content);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }, [filePath, content]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "s") {
        e.preventDefault();
        handleSave();
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [handleSave]);

  const dirty = content !== original;
  const fileName = filePath.split("/").pop() || filePath;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 bg-slate-800 border-b border-slate-700">
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-sm text-slate-300 truncate">{fileName}</span>
          {dirty && <span className="text-[10px] text-yellow-400 shrink-0">Modified</span>}
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={handleSave}
            disabled={!dirty || saving}
            className="flex items-center gap-1 px-2 py-1 text-xs rounded bg-cyan-600 hover:bg-cyan-500 disabled:opacity-30 disabled:cursor-not-allowed text-white transition-colors"
          >
            {saving ? <Loader2 size={12} className="animate-spin" /> : <Save size={12} />}
            Save
          </button>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <div className="flex-1">
        {loading && (
          <div className="flex items-center justify-center h-full text-slate-400">
            <Loader2 size={24} className="animate-spin" />
          </div>
        )}
        {error && (
          <div className="flex items-center justify-center h-full text-red-400 text-sm">{error}</div>
        )}
        {!loading && !error && (
          <Editor
            height="100%"
            language={getLang(filePath)}
            value={content}
            onChange={(v) => setContent(v || "")}
            theme="vs-dark"
            options={{
              fontSize: 13,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              padding: { top: 8 },
              wordWrap: "on",
            }}
          />
        )}
      </div>
    </div>
  );
}
