import { useState } from "react";
import { Sidebar } from "./components/Sidebar";
import { Dashboard } from "./components/Dashboard";
import { FileExplorer } from "./components/FileExplorer";
import { S3Browser } from "./components/S3Browser";
import { RDSOverviewPanel } from "./components/RDSOverviewPanel";
import { TerminalPanel } from "./components/Terminal";
import { ConnectionModal } from "./components/ConnectionModal";

function App() {
  const [activeView, setActiveView] = useState<"dashboard" | "ec2" | "s3" | "rds">("dashboard");
  const [showConnections, setShowConnections] = useState(false);
  const [terminalOpen, setTerminalOpen] = useState(false);

  return (
    <div className="flex h-screen bg-slate-950 text-slate-200">
      <Sidebar
        activeView={activeView}
        onViewChange={setActiveView}
        onOpenConnections={() => setShowConnections(true)}
        terminalOpen={terminalOpen}
        onToggleTerminal={() => setTerminalOpen((v) => !v)}
      />
      <div className="flex-1 flex flex-col min-w-0 relative">
        <main className="flex-1 flex flex-col min-h-0">
          {activeView === "dashboard" && <Dashboard />}
          {activeView === "ec2" && <FileExplorer />}
          {activeView === "s3" && <S3Browser />}
          {activeView === "rds" && <RDSOverviewPanel />}
        </main>

        {terminalOpen && (
          <TerminalPanel onClose={() => setTerminalOpen(false)} />
        )}
      </div>

      {showConnections && <ConnectionModal onClose={() => setShowConnections(false)} />}
    </div>
  );
}

export default App;
