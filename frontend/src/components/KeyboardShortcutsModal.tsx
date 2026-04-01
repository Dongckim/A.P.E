import { X, Keyboard } from "lucide-react";

interface Props {
  onClose: () => void;
}

const shortcuts = [
  { keys: ["Cmd/Ctrl", "N"], description: "Create new folder" },
  { keys: ["Cmd/Ctrl", "C"], description: "Copy file path" },
  { keys: ["Cmd/Ctrl", "S"], description: "Save file in editor" },
  { keys: ["Enter"], description: "Open selected file or folder" },
  { keys: ["Delete / Backspace"], description: "Delete selected file" },
  { keys: ["Escape"], description: "Close editor or deselect" },
  { keys: ["Shift", "Click"], description: "Select a range of files" },
  { keys: ["Cmd/Ctrl", "Click"], description: "Select multiple files" },
  { keys: ["?"], description: "Show keyboard shortcuts" },
];

export function KeyboardShortcutsModal({ onClose }: Props) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="bg-slate-800 border border-slate-600 rounded-lg w-96 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
          <div className="flex items-center gap-2">
            <Keyboard size={15} className="text-slate-400" />
            <h3 className="text-white font-medium text-sm">Keyboard Shortcuts</h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white"
          >
            <X size={16} />
          </button>
        </div>

        {/* Shortcuts list */}
        <div className="p-4 space-y-1">
          {shortcuts.map((shortcut, i) => (
            <div
              key={i}
              className="flex items-center justify-between py-2 px-2 rounded hover:bg-slate-700/50"
            >
              <span className="text-slate-300 text-sm">{shortcut.description}</span>
              <div className="flex items-center gap-1">
                {shortcut.keys.map((key, j) => (
                  <span key={j} className="flex items-center gap-1">
                    <kbd className="bg-slate-700 border border-slate-500 rounded px-2 py-0.5 text-[11px] text-slate-300 font-mono">
                      {key}
                    </kbd>
                    {j < shortcut.keys.length - 1 && (
                      <span className="text-slate-500 text-[11px]">+</span>
                    )}
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Footer */}
        <div className="px-4 py-3 border-t border-slate-700">
          <p className="text-slate-500 text-[11px] text-center">
            Press <kbd className="bg-slate-700 border border-slate-500 rounded px-1 text-slate-300 font-mono">?</kbd> or <kbd className="bg-slate-700 border border-slate-500 rounded px-1 text-slate-300 font-mono">Esc</kbd> to close
          </p>
        </div>
      </div>
    </div>
  );
}
