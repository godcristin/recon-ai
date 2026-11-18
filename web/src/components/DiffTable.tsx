import type { Diff } from "../api/types";
import {
  DIFF_TYPE_LABELS,
  categoryLabel,
  confidenceClass,
  signedYuan,
  stateLabel,
} from "../lib/format";

export default function DiffTable({
  diffs,
  onSelect,
}: {
  diffs: Diff[];
  onSelect: (id: string) => void;
}) {
  if (diffs.length === 0) {
    return (
      <div className="px-6 py-10 text-center text-sm text-slate-500">
        没有差异，或当前筛选条件下无数据
      </div>
    );
  }

  return (
    <table className="w-full text-sm">
      <thead className="bg-slate-50 text-left text-xs uppercase tracking-wide text-slate-500">
        <tr>
          <th className="px-4 py-2 font-medium">ID</th>
          <th className="px-4 py-2 font-medium">类型</th>
          <th className="px-4 py-2 text-right font-medium">差额（元）</th>
          <th className="px-4 py-2 font-medium">AI 分类</th>
          <th className="px-4 py-2 font-medium">置信度</th>
          <th className="px-4 py-2 font-medium">状态</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-slate-100">
        {diffs.map((d, i) => (
          <tr
            key={d.id ?? i}
            onClick={() => d.id && onSelect(d.id)}
            className="cursor-pointer hover:bg-slate-50"
          >
            <td className="px-4 py-2 font-mono text-xs text-slate-500">{d.id}</td>
            <td className="px-4 py-2">
              {DIFF_TYPE_LABELS[d.diffType ?? ""] ?? d.diffType ?? "-"}
            </td>
            <td
              className={`px-4 py-2 text-right font-mono ${
                Number(d.amountDiff ?? 0) < 0 ? "text-rose-600" : "text-emerald-600"
              }`}
            >
              {signedYuan(d.amountDiff)}
            </td>
            <td className="px-4 py-2">
              {d.category ? (
                <span
                  className={`inline-block rounded border px-2 py-0.5 text-xs ${confidenceClass(
                    d.confidence
                  )}`}
                >
                  {categoryLabel(d.category)}
                </span>
              ) : (
                <span className="text-xs text-slate-400">待归因</span>
              )}
              {d.fallback && (
                <span className="ml-1 text-[10px] text-slate-400">(兜底)</span>
              )}
            </td>
            <td className="px-4 py-2 text-xs text-slate-600">
              {d.confidence ? `${Math.round(d.confidence * 100)}%` : "-"}
            </td>
            <td className="px-4 py-2 text-xs text-slate-600">
              {stateLabel(d.state)}
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
