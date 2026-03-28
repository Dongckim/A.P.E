import { useState, useRef, useEffect } from "react";

interface Props {
  initialName: string;
  onConfirm: (newName: string) => void;
  onCancel: () => void;
}

export function RenameInput({ initialName, onConfirm, onCancel }: Props) {
  const [value, setValue] = useState(initialName);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
      // Select name without extension
      const dot = initialName.lastIndexOf(".");
      inputRef.current.setSelectionRange(0, dot > 0 ? dot : initialName.length);
    }
  }, [initialName]);

  return (
    <input
      ref={inputRef}
      value={value}
      onChange={(e) => setValue(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === "Enter" && value.trim() && value !== initialName) {
          onConfirm(value.trim());
        } else if (e.key === "Escape") {
          onCancel();
        }
        e.stopPropagation();
      }}
      onBlur={onCancel}
      className="bg-slate-700 text-white text-xs px-1 py-0.5 rounded border border-cyan-500 outline-none w-full text-center"
      onClick={(e) => e.stopPropagation()}
      onDoubleClick={(e) => e.stopPropagation()}
    />
  );
}
