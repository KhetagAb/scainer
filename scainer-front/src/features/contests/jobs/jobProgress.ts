import type { JobState } from "@/client/types.gen";
import { getJob } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";

function isTerminal(status: JobState["status"]): boolean {
  return status === "succeeded" || status === "failed";
}

function isAbortError(e: unknown): boolean {
  if (e instanceof DOMException && e.name === "AbortError") return true;
  return e instanceof Error && e.name === "AbortError";
}

function abortError(): DOMException {
  return new DOMException("Aborted", "AbortError");
}

/**
 * subscribeJobEvents читает SSE-поток GET /api/jobs/{jobId}/events через fetch+ReadableStream —
 * не нативный EventSource, потому что EventSource не умеет слать Authorization-заголовок (JWT),
 * а этот эндпоинт защищён тем же JWT-мидлваром, что и весь остальной /api.
 * Вызывает onEvent на каждое полученное состояние job'а и резолвится финальным состоянием, когда
 * job переходит в succeeded/failed.
 *
 * Remount/F5/прокси часто рвут SSE без финала: тогда переподключаемся, пока job жив.
 * Нельзя резолвить «успехом» на незавершённом state — caller сотрёт jobId из localStorage
 * и после перезагрузки страницы resume сломается.
 */
export async function subscribeJobEvents(
  jobId: string,
  onEvent: (state: JobState) => void,
  signal?: AbortSignal,
): Promise<JobState> {
  for (;;) {
    if (signal?.aborted) throw abortError();

    const terminal = await readProgressStream(jobId, onEvent, signal);
    if (terminal) return terminal;

    // Поток оборвался, job ещё running — короткая пауза и снова SSE.
    await wait(300, signal);
  }
}

async function readProgressStream(
  jobId: string,
  onEvent: (state: JobState) => void,
  signal?: AbortSignal,
): Promise<JobState | null> {
  const res = await fetch(`/api/jobs/${encodeURIComponent(jobId)}/events`, {
    headers: authHeaders(),
    signal,
  });
  if (!res.ok || !res.body) {
    if (res.status === 404) {
      const err = new Error(`не удалось открыть поток прогресса: HTTP 404`) as Error & {
        status?: number;
      };
      err.status = 404;
      throw err;
    }
    // Временный сбой прокси — пусть верхний цикл переподключится через snapshot.
    const snap = await snapshotJob(jobId, onEvent, signal);
    return snap && isTerminal(snap.status) ? snap : null;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  let last: JobState | null = null;

  try {
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
          if (isTerminal(state.status)) {
            return state;
          }
        }
      }
    }
  } catch (e) {
    if (isAbortError(e) || signal?.aborted) throw abortError();
    throw e;
  }

  if (signal?.aborted) throw abortError();

  if (last && isTerminal(last.status)) {
    return last;
  }

  const snap = await snapshotJob(jobId, onEvent, signal);
  if (snap && isTerminal(snap.status)) {
    return snap;
  }
  // Job ещё идёт (или исчез) — null = переподключиться / выйти по 404 на следующем круге.
  return null;
}

async function snapshotJob(
  jobId: string,
  onEvent: (state: JobState) => void,
  signal?: AbortSignal,
): Promise<JobState | null> {
  if (signal?.aborted) throw abortError();
  const snap = await getJob({
    path: { jobId },
    headers: authHeaders(),
  });
  if (signal?.aborted) throw abortError();
  if (snap.response.status === 404 || !snap.data) {
    const err = new Error("job not found") as Error & { status?: number };
    err.status = 404;
    throw err;
  }
  onEvent(snap.data);
  return snap.data;
}

function wait(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(abortError());
      return;
    }
    const t = window.setTimeout(() => {
      signal?.removeEventListener("abort", onAbort);
      resolve();
    }, ms);
    const onAbort = () => {
      window.clearTimeout(t);
      reject(abortError());
    };
    signal?.addEventListener("abort", onAbort, { once: true });
  });
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
