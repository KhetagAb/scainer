import type { JobState } from "@/client/types.gen";

export type ImportProgress = {
  phase: string;
  done: number;
  total: number;
  label?: string;
} | null;

export function formatImportProgress(progress: ImportProgress): string {
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

export function importProgressPercent(progress: ImportProgress): number | null {
  if (!progress || progress.total <= 0) return null;
  
  if (progress.phase === "analyzing" && progress.done === 0) return null;
  return Math.min(100, Math.round((100 * progress.done) / progress.total));
}

export function storageKey(contestId: string): string {
  return `scainer.importJob.${contestId}`;
}

/** localStorage — переживает F5; чистим только при финале job / 404. */
export function readStoredJobId(contestId: string): string | null {
  try {
    return localStorage.getItem(storageKey(contestId));
  } catch {
    return null;
  }
}

export function writeStoredJobId(contestId: string, jobId: string): void {
  try {
    localStorage.setItem(storageKey(contestId), jobId);
  } catch {
    /* ignore quota / private mode */
  }
}

export function clearStoredJobId(contestId: string): void {
  try {
    localStorage.removeItem(storageKey(contestId));
  } catch {
    /* ignore */
  }
}

export function isAbortError(e: unknown): boolean {
  if (e instanceof DOMException && e.name === "AbortError") return true;
  return e instanceof Error && e.name === "AbortError";
}

export function progressFromState(state: JobState): NonNullable<ImportProgress> {
  return {
    phase: state.progress.phase,
    done: state.progress.done,
    total: state.progress.total,
    label: state.progress.label || undefined,
  };
}

export function extractJobId(data: unknown): string | null {
  const payload = data as { jobId?: string; importedCount?: number } | undefined;
  if (payload?.jobId) return payload.jobId;
  if (payload && "importedCount" in payload) {
    throw new Error("бэкенд вернул синхронный импорт без jobId — пересоберите контейнер scainer");
  }
  return null;
}
