interface Props {
  progress: number | null; // null = not uploading, 0-100 = progress
  fileName?: string;
  isDragOver: boolean;
}

export function UploadOverlay({ progress, fileName, isDragOver }: Props) {
  if (!isDragOver && progress === null) return null;

  return (
    <div className="absolute inset-0 z-40 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm">
      {isDragOver && progress === null && (
        <div className="border-2 border-dashed border-cyan-500 rounded-xl p-10 text-center">
          <p className="text-cyan-400 text-lg font-medium">Drop files to upload</p>
          <p className="text-slate-400 text-sm mt-1">Files will be uploaded to the current directory</p>
        </div>
      )}
      {progress !== null && (
        <div className="text-center w-64">
          <p className="text-slate-300 text-sm mb-2 truncate">{fileName || "Uploading..."}</p>
          <div className="h-2 bg-slate-700 rounded-full overflow-hidden">
            <div
              className="h-full bg-cyan-500 rounded-full transition-all duration-200"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-slate-400 text-xs mt-1">{progress}%</p>
        </div>
      )}
    </div>
  );
}
