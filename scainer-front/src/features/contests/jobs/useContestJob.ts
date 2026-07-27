import { useCallback, useEffect, useRef, useState } from "react";
import { postContestSync } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { resumeContestJob, watchContestJob } from "@/features/contests/jobs/contestJobRunner";
import {
  clearStoredJobId,
  extractJobId,
  isAbortError,
  readStoredJobId,
  syncJobStorageKey,
  writeStoredJobId,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";
import { formatApiError } from "@/lib/apiError";

export type { JobProgress as SyncProgress } from "@/features/contests/jobs/jobShared";

const FAIL_MESSAGE = "обновление завершилось с ошибкой";

type UseContestJobOptions = {
  contestId?: string | null;
  onSettled?: (ok: boolean) => void | Promise<void>;
  onImportDone?: () => void;
};

/** POST /sync → jobId → SSE → localStorage resume. */
export function useContestJob(
  onUnauthorized?: () => void,
  options?: UseContestJobOptions,
) {
  const contestId = options?.contestId ?? null;

  const onSettledRef = useRef(options?.onSettled);
  onSettledRef.current = options?.onSettled;
  const onImportDoneRef = useRef(options?.onImportDone);
  onImportDoneRef.current = options?.onImportDone;

  const [isRunning, setIsRunning] = useState(false);
  const [isSettling, setIsSettling] = useState(false);
  const [progress, setProgress] = useState<JobProgress>(null);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  const activeContestRef = useRef<string | null>(null);

  const isBusy = isRunning || isSettling;

  const settle = useCallback(async (ok: boolean) => {
    setIsSettling(true);
    try {
      await onSettledRef.current?.(ok);
    } finally {
      setIsSettling(false);
    }
  }, []);

  const watchJob = useCallback(
    async (id: string, jobId: string, signal: AbortSignal): Promise<boolean> => {
      const ok = await watchContestJob({
        contestId: id,
        jobId,
        storageKey: syncJobStorageKey,
        signal,
        onProgress: setProgress,
        onImportDone: () => onImportDoneRef.current?.(),
      });
      if (!ok) {
        setError(FAIL_MESSAGE);
      }
      return ok;
    },
    [],
  );

  const start = useCallback(
    async (id: string): Promise<boolean> => {
      if (activeContestRef.current) return false;

      setIsRunning(true);
      setError(null);
      setProgress(null);
      activeContestRef.current = id;

      const controller = new AbortController();
      abortRef.current = controller;
      let jobIdWritten = false;
      let ok = false;
      try {
        const res = await postContestSync({
          path: { id },
          headers: authHeaders(),
        });

        if (res.error) {
          if (res.response.status === 401) onUnauthorized?.();
          if (res.response.status === 404) clearStoredJobId(id, syncJobStorageKey);
          setError(formatApiError(res.error, res.response.status));
          await settle(false);
          return false;
        }

        const jobId = extractJobId(res.data);
        if (!jobId) {
          setError("сервер не вернул jobId");
          await settle(false);
          return false;
        }

        writeStoredJobId(id, jobId, syncJobStorageKey);
        jobIdWritten = true;
        ok = await watchJob(id, jobId, controller.signal);
        await settle(ok);
        return ok;
      } catch (e) {
        if (isAbortError(e)) return false;
        if (!jobIdWritten) clearStoredJobId(id, syncJobStorageKey);
        setError(formatApiError(e));
        await settle(false);
        return false;
      } finally {
        setIsRunning(false);
        abortRef.current = null;
        activeContestRef.current = null;
      }
    },
    [onUnauthorized, settle, watchJob],
  );

  useEffect(() => {
    if (!contestId) return;
    if (activeContestRef.current) return;

    const jobId = readStoredJobId(contestId, syncJobStorageKey);
    if (!jobId) return;

    let cancelled = false;
    const controller = new AbortController();
    abortRef.current = controller;
    activeContestRef.current = contestId;
    setIsRunning(true);
    setError(null);

    void (async () => {
      try {
        const result = await resumeContestJob({
          contestId,
          jobId,
          storageKey: syncJobStorageKey,
          signal: controller.signal,
          onUnauthorized,
          onProgress: setProgress,
          onImportDone: () => onImportDoneRef.current?.(),
        });

        if (cancelled) return;

        if (result.kind === "aborted" || result.kind === "cleared") return;

        if (result.kind === "done") {
          if (!result.ok) setError(FAIL_MESSAGE);
          await settle(result.ok);
          return;
        }

        if (!result.ok) setError(FAIL_MESSAGE);
        if (!cancelled) await settle(result.ok);
      } catch (e) {
        if (cancelled || isAbortError(e)) return;
        setError(formatApiError(e));
        await settle(false);
      } finally {
        if (!cancelled) {
          setIsRunning(false);
          abortRef.current = null;
          activeContestRef.current = null;
        }
      }
    })();

    return () => {
      cancelled = true;
      controller.abort();
      if (abortRef.current === controller) abortRef.current = null;
      if (activeContestRef.current === contestId) activeContestRef.current = null;
    };
  }, [contestId, onUnauthorized, settle]);

  const cancel = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { start, cancel, isRunning, isSettling, isBusy, progress, error };
}
