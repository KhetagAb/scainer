import type { JobState } from "@/client/types.gen";

export type JobProgress = {
  phase: string;
  done: number;
  total: number;
  label?: string;
} | null;

export function formatJobProgress(progress: JobProgress): string {
  if (!progress) return "Загрузка…";
  if (progress.label?.trim()) {
    if (progress.total > 0) {
      return `${progress.label.trim()} (${progress.done}/${progress.total})`;
    }
    return progress.label.trim();
  }
  const phaseLabel = progress.phase === "analyzing" ? "Анализируем" : "Выгружаем";
  if (progress.total > 0) {
    return `${phaseLabel}: ${progress.done}/${progress.total}`;
  }
  return `${phaseLabel}…`;
}

export function jobProgressPercent(progress: JobProgress): number | null {
  if (!progress || progress.total <= 0) return null;

  if (progress.phase === "analyzing" && progress.done === 0) return null;
  return Math.min(100, Math.round((100 * progress.done) / progress.total));
}

export function readStoredJobId(
  contestId: string,
  keyFn: (id: string) => string,
): string | null {
  try {
    return localStorage.getItem(keyFn(contestId));
  } catch {
    return null;
  }
}

export function writeStoredJobId(
  contestId: string,
  jobId: string,
  keyFn: (id: string) => string,
): void {
  try {
    localStorage.setItem(keyFn(contestId), jobId);
  } catch {
    /* ignore quota / private mode */
  }
}

export function clearStoredJobId(
  contestId: string,
  keyFn: (id: string) => string,
): void {
  try {
    localStorage.removeItem(keyFn(contestId));
  } catch {
    /* ignore */
  }
}

export function isAbortError(e: unknown): boolean {
  if (e instanceof DOMException && e.name === "AbortError") return true;
  return e instanceof Error && e.name === "AbortError";
}

export const SYNC_JOB_FAIL_MESSAGE = "обновление завершилось с ошибкой";

export const SYNC_PROGRESS_SIZER_LABEL = "Анализируем: 9999/9999";

export function jobFailureMessage(
  state: JobState,
  fallback = SYNC_JOB_FAIL_MESSAGE,
): string {
  const trimmed = state.error?.trim();
  return trimmed || fallback;
}

export function progressFromState(state: JobState): NonNullable<JobProgress> {
  return {
    phase: state.progress.phase,
    done: state.progress.done,
    total: state.progress.total,
    label: state.progress.label || undefined,
  };
}

export function extractJobId(data: unknown): string | null {
  const jobId = (data as { jobId?: string } | undefined)?.jobId;
  return jobId ?? null;
}

export const syncJobStorageKey = (contestId: string): string =>
  `scainer.syncJob.${contestId}`;

export const resyncJobStorageKey = (contestId: string): string =>
  `scainer.resyncJob.${contestId}`;

export type ContestJobKind = "sync" | "resync";
