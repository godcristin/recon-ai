import type {
  CreateTaskResp,
  DiffDetail,
  ListDiffsResp,
  Task,
  TaskSummary,
} from "./types";

const BASE = "/api";

async function http<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, init);
  const text = await res.text();

  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = null;
    }
  }

  if (!res.ok) {
    const anyBody = body as { message?: string; error?: string; reason?: string } | null;
    const msg =
      anyBody?.message ?? anyBody?.error ?? anyBody?.reason ?? `HTTP ${res.status}`;
    throw new Error(msg);
  }

  return body as T;
}

export async function createTask(
  name: string,
  channel: File,
  system: File
): Promise<CreateTaskResp> {
  const fd = new FormData();
  fd.append("name", name);
  fd.append("channel", channel);
  fd.append("system", system);
  return http<CreateTaskResp>("/tasks", { method: "POST", body: fd });
}

export function getTask(id: string): Promise<Task> {
  return http<Task>(`/tasks/${id}`);
}

export function listDiffs(
  taskId: string,
  page: number,
  pageSize: number,
  category?: string,
  state?: string
): Promise<ListDiffsResp> {
  const q = new URLSearchParams();
  q.set("page", String(page));
  q.set("page_size", String(pageSize));
  if (category) q.set("category", category);
  if (state) q.set("state", state);
  return http<ListDiffsResp>(`/tasks/${taskId}/diffs?${q.toString()}`);
}

export function getDiff(id: string): Promise<DiffDetail> {
  return http<DiffDetail>(`/diffs/${id}`);
}

export function confirmDiff(
  id: string,
  action: "confirm" | "correct",
  humanCategory?: string
): Promise<unknown> {
  return http(`/diffs/${id}/confirm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      action,
      humanCategory: humanCategory ?? "",
      operator: "web",
    }),
  });
}

export function getSummary(taskId: string): Promise<TaskSummary> {
  return http<TaskSummary>(`/tasks/${taskId}/summary`);
}
