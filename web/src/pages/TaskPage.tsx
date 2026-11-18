import { useCallback, useEffect, useState } from "react";
import { getSummary, getTask, listDiffs } from "../api/client";
import type { Diff, Task, TaskSummary } from "../api/types";
import DiffDrawer from "../components/DiffDrawer";
import DiffTable from "../components/DiffTable";
import { CATEGORY_OPTIONS, CATEGORY_LABELS, fenToYuan } from "../lib/format";

const PAGE_SIZE = 20;

export default function TaskPage({ taskId }: { taskId: string }) {
  const [task, setTask] = useState<Task | null>(null);
  const [summary, setSummary] = useState<TaskSummary | null>(null);
  const [diffs, setDiffs] = useState<Diff[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [category, setCategory] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [t, list, s] = await Promise.all([
        getTask(taskId),
        listDiffs(taskId, page, PAGE_SIZE, category || undefined),
        getSummary(taskId),
      ]);
      setTask(t);
      setDiffs(list.list ?? []);
      setTotal(list.total ?? 0);
      setSummary(s);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, [taskId, page, category]);

  useEffect(() => {
    void reload();
  }, [reload]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-xl font-semibold">{task?.name ?? "对账任务"}</h1>
          <p className="mt-1 text-sm text-slate-500">
            任务 ID {taskId} · 状态 {task?.status ?? "-"} · 记录{" "}
            {task?.totalCount ?? 0} 条
          </p>
        </div>

        <div className="flex gap-6 text-right">
          <Metric label="差异条数" value={String(summary?.diffCount ?? 0)} />
          <Metric
            label="差异金额"
            value={`¥${fenToYuan(summary?.diffAmount)}`}
          />
          <Metric
            label="AI 采纳率"
            value={`${Math.round((summary?.aiAcceptRate ?? 0) * 100)}%`}
          />
        </div>
      </div>

      {summary && (summary.stats?.length ?? 0) > 0 && (
        <div className="mt-4 flex flex-wrap gap-2">
          {summary.stats?.map((s) => (
            <span
              key={s.category || "none"}
              className="rounded-full border border-slate-200 bg-white px-3 py-1 text-xs text-slate-600"
            >
              {(s.category ? CATEGORY_LABELS[s.category] ?? s.category : "未归因") +
                ` ${s.count ?? 0} 条 · ¥${fenToYuan(s.amount)}`}
            </span>
          ))}
        </div>
      )}

      <div className="mt-5 flex items-center gap-3">
        <select
          value={category}
          onChange={(e) => {
            setCategory(e.target.value);
            setPage(1);
          }}
          className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm"
        >
          <option value="">全部分类</option>
          {CATEGORY_OPTIONS.map((c) => (
            <option key={c} value={c}>
              {CATEGORY_LABELS[c]}
            </option>
          ))}
        </select>

        <button className="btn-ghost" onClick={() => void reload()} disabled={loading}>
          {loading ? "刷新中…" : "刷新"}
        </button>

        <span className="text-sm text-slate-500">共 {total} 条差异</span>
      </div>

      {error && (
        <div className="mt-4 rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {error}
        </div>
      )}

      <div className="mt-4 card overflow-hidden">
        <DiffTable diffs={diffs} onSelect={setSelectedId} />
      </div>

      <div className="mt-4 flex items-center justify-between text-sm">
        <span className="text-slate-500">
          第 {page} / {totalPages} 页
        </span>
        <div className="flex gap-2">
          <button
            className="btn-ghost"
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            上一页
          </button>
          <button
            className="btn-ghost"
            disabled={page >= totalPages}
            onClick={() => setPage((p) => p + 1)}
          >
            下一页
          </button>
        </div>
      </div>

      {selectedId && (
        <DiffDrawer
          diffId={selectedId}
          onClose={() => setSelectedId(null)}
          onChanged={() => void reload()}
        />
      )}
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-lg font-semibold text-slate-900">{value}</div>
      <div className="text-xs text-slate-500">{label}</div>
    </div>
  );
}
