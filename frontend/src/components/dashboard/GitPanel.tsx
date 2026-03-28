import { useState } from "react";
import { GitBranch, GitCommit as GitCommitIcon, Search } from "lucide-react";
import type { GitInfo } from "../../types";

interface Props {
  data: GitInfo | null;
  gitPath: string;
  onPathChange: (path: string) => void;
}

export function GitPanel({ data, gitPath, onPathChange }: Props) {
  const [input, setInput] = useState(gitPath);

  const handleSubmit = () => {
    if (input.trim()) onPathChange(input.trim());
  };

  return (
    <div className="bg-slate-800/50 rounded-lg border border-slate-700 p-4">
      <h3 className="text-sm font-medium text-white mb-3">Git Deploys</h3>

      {/* Path input */}
      <div className="flex gap-1 mb-3">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
          placeholder="/home/ubuntu/myapp"
          className="flex-1 bg-slate-700 text-white text-xs px-2 py-1.5 rounded border border-slate-600 outline-none focus:border-cyan-500"
        />
        <button
          onClick={handleSubmit}
          className="px-2 py-1.5 bg-slate-700 hover:bg-slate-600 rounded text-slate-300 transition-colors"
        >
          <Search size={12} />
        </button>
      </div>

      {!data && (
        <p className="text-xs text-slate-500 text-center py-4">Enter a repo path on the server</p>
      )}

      {data?.error && (
        <p className="text-xs text-yellow-400 text-center py-4">{data.error}</p>
      )}

      {data && !data.error && (
        <>
          {/* Branch + last commit */}
          <div className="flex items-center gap-2 mb-3 text-xs">
            <GitBranch size={12} className="text-cyan-400" />
            <span className="text-cyan-400">{data.branch}</span>
            {data.last_commit && (
              <span className="text-slate-500 ml-auto">{data.last_commit.date?.split("T")[0]}</span>
            )}
          </div>

          {/* Commit log */}
          <div className="space-y-0.5 max-h-52 overflow-y-auto">
            {data.commits.map((c, i) => (
              <div key={c.hash + i} className="flex items-start gap-2 px-1 py-1 text-xs group">
                <GitCommitIcon size={12} className={`shrink-0 mt-0.5 ${i === 0 ? "text-cyan-400" : "text-slate-600"}`} />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <code
                      className="text-slate-400 cursor-pointer hover:text-white"
                      onClick={() => navigator.clipboard.writeText(c.hash)}
                      title="Click to copy"
                    >
                      {c.hash}
                    </code>
                    <span className="text-white truncate">{c.message}</span>
                  </div>
                  <div className="text-slate-500">
                    {c.author} — {c.relative_date}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
