import { useState } from "react";
import TaskPage from "./pages/TaskPage";
import UploadPage from "./pages/UploadPage";

export default function App() {
  const [taskId, setTaskId] = useState<string | null>(null);

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-3">
          <div>
            <div className="text-base font-semibold tracking-tight">recon-ai</div>
            <div className="text-xs text-slate-500">对账 AI 助手</div>
          </div>
          {taskId && (
            <button className="btn-ghost" onClick={() => setTaskId(null)}>
              新建对账
            </button>
          )}
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-6">
        {taskId ? (
          <TaskPage taskId={taskId} />
        ) : (
          <UploadPage onCreated={setTaskId} />
        )}
      </main>
    </div>
  );
}
