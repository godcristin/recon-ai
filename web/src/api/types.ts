
export interface ParseError {
  line?: number;
  message?: string;
}

export interface CreateTaskResp {
  taskId?: string;
  totalCount?: number;
  diffCount?: number;
  parseErrors?: ParseError[];
}

export interface Task {
  id?: string;
  name?: string;
  status?: string;
  totalCount?: number;
  diffCount?: number;
  diffAmount?: string;
  createTime?: string;
  finishTime?: string;
}

export interface ReconRecord {
  id?: string;
  source?: number;
  tradeNo?: string;
  bizNo?: string;
  amount?: string;
  fee?: string;
  tradeTime?: string;
  status?: string;
  remark?: string;
}

export interface Diff {
  id?: string;
  taskId?: string;
  diffType?: string;
  amountDiff?: string;
  category?: string;
  confidence?: number;
  reason?: string;
  evidence?: string[];
  suggestion?: string;
  state?: string;
  fallback?: boolean;
  aiModel?: string;
  promptVersion?: string;
  createTime?: string;
}

export interface DiffDetail {
  diff?: Diff;
  channel?: ReconRecord;
  system?: ReconRecord;
  matchRule?: string;
  timeDiffHours?: string;
}

export interface ListDiffsResp {
  list?: Diff[];
  total?: number;
  totalAmount?: string;
}

export interface CategoryStat {
  category?: string;
  count?: number;
  amount?: string;
}

export interface TaskSummary {
  total?: number;
  diffCount?: number;
  diffAmount?: string;
  stats?: CategoryStat[];
  aiAcceptRate?: number;
}
