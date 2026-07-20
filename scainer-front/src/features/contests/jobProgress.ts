import type { JobState } from "@/client/types.gen";
import { getJob } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";

function isTerminal(status: JobState["status"]): boolean {
  return status === "succeeded" || status === "failed";
}

/**
 * subscribeJobEvents читает SSE-поток GET /api/jobs/{jobId}/events через fetch+ReadableStream —
 * не нативный EventSource, потому что EventSource не умеет слать Authorization-заголовок (JWT),
 * а этот эндпоинт защищён тем же JWT-мидлваром, что и весь остальной /api.
 * Вызывает onEvent на каждое полученное состояние job'а и резолвится финальным состоянием, когда
 * поток закрывается (сервер сам закрывает его на succeeded/failed).
 *
 * Если поток оборвался без терминального статуса (прокси, remount) — добираем снимок через getJob.
 */
export async function subscribeJobEvents(
  jobId: string,
  onEvent: (state: JobState) => void,
  signal?: AbortSignal,
): Promise<JobState> {
  const res = await fetch(`/api/jobs/${encodeURIComponent(jobId)}/events`, {
    headers: authHeaders(),
    signal,
  });
  if (!res.ok || !res.body) {
    const err = new Error(`не удалось открыть поток прогресса: HTTP ${res.status}`) as Error & {
      status?: number;
    };
    err.status = res.status;
    throw err;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let last: JobState | null = null;

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    let sep: number;
    while ((sep = buffer.indexOf("\n\n")) !== -1) {
      const rawEvent = buffer.slice(0, sep);
      buffer = buffer.slice(sep + 2);
      const state = parseProgressEvent(rawEvent);
      if (state) {
        last = state;
        onEvent(state);
      }
    }
  }

  if (last && isTerminal(last.status)) {
    return last;
  }

  // Поток закрылся без финала — уточняем через snapshot (частый артефакт при remount/прокси).
  const snap = await getJob({
    path: { jobId },
    headers: authHeaders(),
  });
  if (snap.data && isTerminal(snap.data.status)) {
    onEvent(snap.data);
    return snap.data;
  }
  if (last) return last;
  throw new Error("поток прогресса закрылся, не дождавшись финального состояния");
}

function parseProgressEvent(raw: string): JobState | null {
  let eventName = "message";
  const dataLines: string[] = [];
  for (const line of raw.split("\n")) {
    if (line.startsWith(":")) continue; // heartbeat-комментарий, игнорируем
    if (line.startsWith("event:")) eventName = line.slice("event:".length).trim();
    else if (line.startsWith("data:")) dataLines.push(line.slice("data:".length).trim());
  }
  if (eventName !== "progress" || dataLines.length === 0) return null;
  try {
    return JSON.parse(dataLines.join("\n")) as JobState;
  } catch {
    return null;
  }
}
