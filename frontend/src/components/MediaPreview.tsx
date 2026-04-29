import { useEffect } from "react";
import { X, Download } from "lucide-react";
import { downloadFile } from "../api/client";
import { getMediaKind } from "../utils/media";

interface Props {
  filePath: string;
  onClose: () => void;
}

export function MediaPreview({ filePath, onClose }: Props) {
  const fileName = filePath.split("/").pop() || filePath;
  const kind = getMediaKind(fileName);
  const url = `/api/ec2/download?path=${encodeURIComponent(filePath)}&inline=1`;

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onClose]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 bg-slate-800 border-b border-slate-700">
        <span className="text-sm text-slate-300 truncate">{fileName}</span>
        <div className="flex items-center gap-1">
          <button
            onClick={() => downloadFile(filePath)}
            className="flex items-center gap-1 px-2 py-1 text-xs rounded bg-slate-700 hover:bg-slate-600 text-slate-200 transition-colors"
          >
            <Download size={12} />
            Download
          </button>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center bg-slate-950 overflow-auto p-4">
        {kind === "image" && (
          <img src={url} alt={fileName} className="max-w-full max-h-full object-contain" />
        )}
        {kind === "video" && (
          <video src={url} controls className="max-w-full max-h-full" />
        )}
        {kind === "audio" && (
          <audio src={url} controls className="w-full max-w-md" />
        )}
        {kind === "pdf" && (
          <iframe src={url} title={fileName} className="w-full h-full bg-white" />
        )}
        {!kind && (
          <span className="text-slate-500 text-sm">Preview not supported for this file type.</span>
        )}
      </div>
    </div>
  );
}
