// 展示层的小工具：金额换算、枚举中文名、颜色

export function fenToYuan(v?: string | number | null): string {
  if (v === undefined || v === null || v === "") return "0.00";
  const n = Number(v);
  if (!Number.isFinite(n)) return "0.00";
  const sign = n < 0 ? "-" : "";
  const abs = Math.abs(n);
  return `${sign}${Math.floor(abs / 100)}.${String(abs % 100).padStart(2, "0")}`;
}

export function signedYuan(v?: string | number | null): string {
  const n = Number(v ?? 0);
  const s = fenToYuan(n);
  return n > 0 ? `+${s}` : s;
}

export const CATEGORY_LABELS: Record<string, string> = {
  cross_day: "跨日入账",
  fee_diff: "手续费差异",
  refund_unsync: "退款未同步",
  missing_in_system: "系统漏记",
  missing_in_channel: "渠道漏记",
  duplicate: "重复记账",
};

export const CATEGORY_OPTIONS = Object.keys(CATEGORY_LABELS);

export const STATE_LABELS: Record<string, string> = {
  PENDING: "待归因",
  CLASSIFIED: "已归因",
  CONFIRMED: "已确认",
  CORRECTED: "已纠正",
};

export const DIFF_TYPE_LABELS: Record<string, string> = {
  channel_only: "渠道单边",
  system_only: "系统单边",
  amount_mismatch: "金额不一致",
};

export const SUGGESTION_LABELS: Record<string, string> = {
  confirm: "可确认",
  investigate: "需核查",
  ignore: "可忽略",
};

export function categoryLabel(c?: string): string {
  if (!c) return "未归因";
  return CATEGORY_LABELS[c] ?? c;
}

export function stateLabel(s?: string): string {
  if (!s) return "-";
  return STATE_LABELS[s] ?? s;
}

// 置信度分档：高/中/低，用于颜色与“是否需要人工”
export function confidenceLevel(c?: number): "high" | "mid" | "low" {
  const v = c ?? 0;
  if (v >= 0.85) return "high";
  if (v >= 0.6) return "mid";
  return "low";
}

export function confidenceClass(c?: number): string {
  switch (confidenceLevel(c)) {
    case "high":
      return "bg-emerald-50 text-emerald-700 border-emerald-200";
    case "mid":
      return "bg-amber-50 text-amber-700 border-amber-200";
    default:
      return "bg-rose-50 text-rose-700 border-rose-200";
  }
}
