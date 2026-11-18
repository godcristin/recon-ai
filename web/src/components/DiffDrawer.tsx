import { useCallback, useEffect, useState } from "react";
import { confirmDiff, getDiff } from "../api/client";
import type { DiffDetail, ReconRecord } from "../api/types";
import {
  CATEGORY_LABELS,
  CATEGORY_OPTIONS,
  DIFF_TYPE_LABELS,
  SUGGESTION_LABELS,
  categoryLabel,
  confidenceClass,
  fenToYuan,
  signedYuan,
  stateLabel,
} from "../lib/format";

export default function DiffDrawer({
  diffId,
  onClose,
  onChanged,
}: {
  diffId: string;
  onClose: () => void;
  onChanged: () => void;
}) {
  const [detail, setDetail] = useState<DiffDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [correctTo, setCorrectTo] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const d = await getDiff(diffId);
      setDetail(d);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }, [diffId]);

  useEffect(() => {
    void load();
  }, [load]);

  async function act(action: "confirm" | "correct") {
    if (action === "correct" && !correctTo) {
      setError("请先选择正确的分类");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await confirmDiff(diffId, action, action === "correct" ? correctTo : undefined);
      await load();
      onChanged();
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  const diff = detail?.diff;
  const settled = diff?.state === "CONFIRMED" || diff?.state === "CORRECTED";

  return (
    <div className="fixed inset-0 z-40 flex justify-end bg-black/30">
      <div className="h-full w-full max-w-xl overflow-y-auto bg-white shadow-xl">
        <div className="sticky top-0 flex items-center justify-between border-b border-slate-200 bg-white px-5 py-3">
          <div>
            <div className="text-sm font-semibold">差异 #{diffId}</div>
            <div className="text-xs text-slate-500">
              {diff?.diffType ? DIFF_TYPE_LABELS[diff.diffType] ?? diff.diffType : "-"}
            </div>
          </div>
          <button className="btn-ghost" onClick={onClose}>
            关闭
          </button>
        </div>

        <div className="space-y-5 px-5 py-5">
          {loading && (
            <div className="rounded-md border border-slate-200 bg-slate-50 px-4 py-6 text-center text-sm text-slate-600">
              正在加载差异详情并请求 AI 归因（通常 1~2 秒）…
            </div>
          )}

          {error && (
            <div className="rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
              {error}
            </div>
          )}

          {detail && (
            <>
              <section className="grid grid-cols-3 gap-3">
                <Info label="差额" value={`¥${signedYuan(diff?.amountDiff)}`} />
                <Info label="匹配规则" value={detail.matchRule ?? "-"} />
                <Info
                  label="时间差"
                  value={`${detail.timeDiffHours ?? "0"} 小时`}
                />
              </section>

              <section>
                <h3 className="mb-2 text-sm font-semibold">两侧原始记录</h3>
                <div className="grid gap-3 sm:grid-cols-2">
                  <RecordCard title="渠道账单" record={detail.channel} />
                  <RecordCard title="系统订单" record={detail.system} />
                </div>
              </section>

              <section>
                <div className="mb-2 flex items-center gap-2">
                  <h3 className="text-sm font-semibold">AI 归因</h3>
                  {diff?.fallback && (
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[11px] text-slate-500">
                      规则兜底
                    </span>
                  )}
                  {diff?.state && (
                    <span className="rounded bg-slate-100 px-2 py-0.5 text-[11px] text-slate-500">
                      {stateLabel(diff.state)}
                    </span>
                  )}
                </div>

                {diff?.category ? (
                  <div className="card p-4">
                    <div className="flex flex-wrap items-center gap-2">
                      <span
                        className={`rounded border px-2 py-0.5 text-xs font-medium ${confidenceClass(
                          diff.confidence
                        )}`}
                      >
                        {categoryLabel(diff.category)}
                      </span>
                      <span className="text-xs text-slate-500">
                        置信度 {Math.round((diff.confidence ?? 0) * 100)}%
                      </span>
                      {diff.suggestion && (
                        <span className="text-xs text-slate-500">
                          · {SUGGESTION_LABELS[diff.suggestion] ?? diff.suggestion}
                        </span>
                      )}
                    </div>

                    <p className="mt-3 text-sm leading-relaxed text-slate-700">
                      {diff.reason || "模型未给出理由"}
                    </p>

                    {(diff.evidence?.length ?? 0) > 0 && (
                      <div className="mt-3 flex flex-wrap gap-1">
                        {diff.evidence?.map((e) => (
                          <span
                            key={e}
                            className="rounded bg-slate-100 px-2 py-0.5 font-mono text-[11px] text-slate-600"
                          >
                            {e}
                          </span>
                        ))}
                      </div>
                    )}

                    {(diff.promptVersion || diff.aiModel) && (
                      <div className="mt-3 text-[11px] text-slate-400">
                        {diff.aiModel} · {diff.promptVersion}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-700">
                    AI 暂未给出结论（模型不可用或该差异尚未归因）
                  </div>
                )}
              </section>

              <section className="border-t border-slate-200 pt-4">
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    className="btn-primary"
                    disabled={busy || settled}
                    onClick={() => void act("confirm")}
                  >
                    确认无误
                  </button>

                  <select
                    value={correctTo}
                    onChange={(e) => setCorrectTo(e.target.value)}
                    className="rounded-md border border-slate-300 px-2 py-1.5 text-sm"
                  >
                    <option value="">选择正确分类…</option>
                    {CATEGORY_OPTIONS.map((c) => (
                      <option key={c} value={c}>
                        {CATEGORY_LABELS[c]}
                      </option>
                    ))}
                  </select>

                  <button
                    className="btn-ghost"
                    disabled={busy || settled}
                    onClick={() => void act("correct")}
                  >
                    纠正
                  </button>

                  {settled && (
                    <span className="text-xs text-slate-500">
                      该差异已处理，不能再修改
                    </span>
                  )}
                </div>
              </section>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md bg-slate-50 px-3 py-2">
      <div className="text-[11px] uppercase tracking-wide text-slate-500">{label}</div>
      <div className="mt-0.5 font-mono text-sm text-slate-800">{value}</div>
    </div>
  );
}

function RecordCard({
  title,
  record,
}: {
  title: string;
  record?: ReconRecord;
}) {
  if (!record) {
    return (
      <div className="rounded-md border border-dashed border-slate-300 p-3 text-sm text-slate-400">
        {title}：无对应记录
      </div>
    );
  }

  const rows: Array<[string, string]> = [
    ["商户单号", record.bizNo ?? "-"],
    ["渠道交易号", record.tradeNo ?? "-"],
    ["金额", `¥${fenToYuan(record.amount)}`],
    ["手续费", `¥${fenToYuan(record.fee)}`],
    ["交易时间", record.tradeTime ?? "-"],
    ["状态", record.status ?? "-"],
    ["备注", record.remark || "-"],
  ];

  return (
    <div className="rounded-md border border-slate-200 p-3">
      <div className="mb-2 text-xs font-semibold text-slate-600">{title}</div>
      <dl className="space-y-1 text-xs">
        {rows.map(([k, v]) => (
          <div key={k} className="flex gap-2">
            <dt className="w-20 shrink-0 text-slate-400">{k}</dt>
            <dd className="break-all text-slate-700">{v}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}
