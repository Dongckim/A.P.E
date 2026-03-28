import { ChevronRight } from "lucide-react";

interface Props {
  path: string;
  onNavigate: (path: string) => void;
}

export function Breadcrumb({ path, onNavigate }: Props) {
  const parts = path.split("/").filter(Boolean);

  const crumbs = parts.map((part, i) => ({
    label: part,
    path: "/" + parts.slice(0, i + 1).join("/"),
  }));

  return (
    <nav className="flex items-center gap-1 px-4 py-2 bg-slate-800/50 border-b border-slate-700 text-sm overflow-x-auto">
      <button
        onClick={() => onNavigate("/")}
        className="text-slate-400 hover:text-white shrink-0"
      >
        /
      </button>
      {crumbs.map((crumb) => (
        <span key={crumb.path} className="flex items-center gap-1 shrink-0">
          <ChevronRight size={14} className="text-slate-600" />
          <button
            onClick={() => onNavigate(crumb.path)}
            className="text-slate-300 hover:text-white"
          >
            {crumb.label}
          </button>
        </span>
      ))}
    </nav>
  );
}
