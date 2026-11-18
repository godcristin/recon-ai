import { useState } from "react";
import type { FormEvent } from "react";
import { createTask } from "../api/client";
import type { CreateTaskResp } from "../api/types";

export default function UploadPage({
  onCreated,
}: {
  onCreated: (taskId: string) => void;
}) {
  const [name, setName] = useState("");
  const [channel, setChannel] = useState<File | null>(null);
  const [system, setSystem] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<CreateTaskResp | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (!channel || !system) {
      setError("请选择渠道账单和系统订单两个文件");
      return;
    }
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const rsp = await createTask(
        name || `task-${new Date().toISOString().slice(0, 19)}`,
        channel,
        system
      );
      setResult(rsp);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  const parseErrors = result?.parseErrors ?? [];

  return (
    <div className="mx-auto max-w-3xl">
      <h1 className="text-xl font-semibold">上传两份账单</h1>
      <p className="mt-1 text-sm text-slate-500">
        渠道账单金额单位为「元」，系统订单单位为「分」，系统会自动归一化。
      </p>

      <form onSubmit={submit} className="mt-6 card p-6">
        <label className="block text-sm font-medium text-slate-700">任务名称</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="例如：2026-09-01 对账"
          className="mt-1 w-full rounded-md border border-slate-300 px-3 py-2 text-sm outline-none focus:border-slate-500"
        />

        <div className="mt-5 grid gap-4 sm:grid-cols-2">
          <FilePicker
            label="渠道账单（微信/支付宝导出）"
            value={channel}
            onChange={setChannel}
          />
          <FilePicker
            label="系统订单（业务系统导出）"
            value={system}
            onChange={setSystem}
          />
        </div>

        {error && (
          <div className="mt-4 rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
            {error}
          </div>
        )}

        <div className="mt-6 flex items-center gap-3">
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? "对账中…" : "开始对账"}
          </button>
          {loading && (
            <span className="text-sm text-slate-500">
              正在解析、匹配，2 万行以内通常 1 秒内完成
            </span>
          )}
        </div>
      </form>

      {result && (
        <div className="mt-6 card p-6">
          <div className="flex items-baseline justify-between">
            <h2 className="text-base font-semibold">对账完成</h2>
            <span className="text-xs text-slate-500">
              任务 ID：{result.taskId}
            </span>
          </div>

          <div className="mt-4 grid grid-cols-3 gap-4 text-center">
            <Stat label="记录总数" value={String(result.totalCount ?? 0)} />
            <Stat label="差异条数" value={String(result.diffCount ?? 0)} />
            <Stat
              label="解析错误"
              value={String(parseErrors.length)}
              warn={parseErrors.length > 0}
            />
          </div>

          {parseErrors.length > 0 && (
            <div className="mt-4 rounded-md border border-amber-200 bg-amber-50 p-3">
              <div className="text-sm font-medium text-amber-800">
                以下行解析失败，已跳过（不影响其余数据）
              </div>
              <ul className="mt-2 max-h-40 space-y-1 overflow-auto text-xs text-amber-700">
                {parseErrors.map((e, i) => (
                  <li key={i}>
                    第 {e.line} 行：{e.message}
                  </li>
                ))}
              </ul>
            </div>
          )}

          <button
            className="btn-primary mt-5"
            onClick={() => result.taskId && onCreated(result.taskId)}
          >
            查看差异
          </button>
        </div>
      )}
    </div>
  );
}

function FilePicker({
  label,
  value,
  onChange,
}: {
  label: string;
  value: File | null;
  onChange: (f: File | null) => void;
}) {
  return (
    <label className="block cursor-pointer rounded-md border border-dashed border-slate-300 p-4 hover:border-slate-500">
      <div className="text-sm font-medium text-slate-700">{label}</div>
      <div className="mt-1 truncate text-xs text-slate-500">
        {value ? `${value.name}（${(value.size / 1024).toFixed(1)} KB）` : "点击选择 CSV 文件"}
      </div>
      <input
        type="file"
        accept=".csv,text/csv"
        className="hidden"
        onChange={(e) => onChange(e.target.files?.[0] ?? null)}
      />
    </label>
  );
}

function Stat({
  label,
  value,
  warn,
}: {
  label: string;
  value: string;
  warn?: boolean;
}) {
  return (
    <div>
      <div className={`text-2xl font-semibold ${warn ? "text-amber-600" : "text-slate-900"}`}>
        {value}
      </div>
      <div className="mt-1 text-xs text-slate-500">{label}</div>
    </div>
  );
}
