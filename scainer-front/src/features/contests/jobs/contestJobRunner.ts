import type { JobState } from "@/client/types.gen";
import { getJob } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { formatApiError } from "@/lib/apiError";
import { subscribeJobEvents } from "@/features/contests/jobs/jobProgress";
import {
  clearStoredJobId,
  isAbortError,
  jobFailureMessage,
  progressFromState,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";

export type JobProgressCallback = (progress: JobProgress) => void;

export type ContestJobOutcome = {
  ok: boolean;
  error?: string;
};

export type WatchContestJobOptions = {
  contestId: string;
  jobId: string;
  storageKey: (id: string) => string;
  signal: AbortSignal;
  onProgress: JobProgressCallback;
  onImportDone?: () => void;
};

/** SSE-подписка на job; возвращает успех и текст ошибки job при failed. */
export async function watchContestJob({
  contestId,
  jobId,
  storageKey,
  signal,
  onProgress,
  onImportDone,
}: WatchContestJobOptions): Promise<ContestJobOutcome> {
  let notifiedImportDone = false;
  const final = await subscribeJobEvents(
    jobId,
    (state: JobState) => {
      onProgress(progressFromState(state));
      if (!notifiedImportDone && state.progress.phase === "analyzing") {
        notifiedImportDone = true;
        onImportDone?.();
      }
    },
    signal,
  );

  if (final.status === "succeeded" || final.status === "failed") {
    clearStoredJobId(contestId, storageKey);
  }

  if (final.status === "failed") {
    return { ok: false, error: jobFailureMessage(final) };
  }
  return { ok: true };
}

export type ResumeContestJobOptions = {
  contestId: string;
  jobId: string;
  storageKey: (id: string) => string;
  signal: AbortSignal;
  onUnauthorized?: () => void;
  onProgress: JobProgressCallback;
  onImportDone?: () => void;
};

export type ResumeContestJobResult =
  | { kind: "done"; ok: boolean; error?: string }
  | { kind: "still_running"; ok: boolean; error?: string }
  | { kind: "aborted" }
  | { kind: "cleared" };

/** Проверяет snapshot job и либо возобновляет SSE, либо возвращает финальный статус. */
export async function resumeContestJob({
  contestId,
  jobId,
  storageKey,
  signal,
  onUnauthorized,
  onProgress,
  onImportDone,
}: ResumeContestJobOptions): Promise<ResumeContestJobResult> {
  try {
    const snap = await getJob({
      path: { jobId },
      headers: authHeaders(),
    });

    if (signal.aborted) return { kind: "aborted" };

    if (snap.error) {
      if (snap.response.status === 401) onUnauthorized?.();
      if (snap.response.status === 404) clearStoredJobId(contestId, storageKey);
      return {
        kind: "done",
        ok: false,
        error: formatApiError(snap.error, snap.response.status),
      };
    }

    if (snap.response.status === 404 || !snap.data) {
      clearStoredJobId(contestId, storageKey);
      return { kind: "cleared" };
    }

    const status = snap.data.status;
    if (status === "succeeded" || status === "failed") {
      clearStoredJobId(contestId, storageKey);
      if (status === "failed") {
        return { kind: "done", ok: false, error: jobFailureMessage(snap.data) };
      }
      return { kind: "done", ok: true };
    }

    onProgress(progressFromState(snap.data));
    const outcome = await watchContestJob({
      contestId,
      jobId,
      storageKey,
      signal,
      onProgress,
      onImportDone,
    });
    if (signal.aborted) return { kind: "aborted" };
    return {
      kind: "still_running",
      ok: outcome.ok,
      error: outcome.error,
    };
  } catch (e) {
    if (isAbortError(e)) return { kind: "aborted" };
    throw e;
  }
}
