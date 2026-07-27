import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { postContestSync } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { resumeContestJob, watchContestJob } from "@/features/contests/jobs/contestJobRunner";
import {
  clearStoredJobId,
  extractJobId,
  isAbortError,
  readStoredJobId,
  SYNC_JOB_FAIL_MESSAGE,
  syncJobStorageKey,
  writeStoredJobId,
  type JobProgress,
} from "@/features/contests/jobs/jobShared";
import { formatApiError } from "@/lib/apiError";

export type { JobProgress } from "@/features/contests/jobs/jobShared";

type UseParallelContestJobsOptions = {
  contestIds: string[];
  onUnauthorized?: () => void;
  onSettled?: () => void;
  onImportDone?: (contestId: string) => void;
  onContestSettled?: (contestId: string, ok: boolean) => void | Promise<void>;
};

/**
 * Пачка sync-job'ов для всех контестов параллели.
 */
export function useParallelContestJobs({
  contestIds,
  onUnauthorized,
  onSettled,
  onImportDone,
  onContestSettled,
}: UseParallelContestJobsOptions) {
  const onSettledRef = useRef(onSettled);
  onSettledRef.current = onSettled;
  const onImportDoneRef = useRef(onImportDone);
  onImportDoneRef.current = onImportDone;
  const onContestSettledRef = useRef(onContestSettled);
  onContestSettledRef.current = onContestSettled;
  const onUnauthorizedRef = useRef(onUnauthorized);
  onUnauthorizedRef.current = onUnauthorized;

  const [progressById, setProgressById] = useState<Record<string, JobProgress>>({});
  const [errorById, setErrorById] = useState<Record<string, string>>({});
  const [activeIds, setActiveIds] = useState<Set<string>>(() => new Set());
  const [settlingCount, setSettlingCount] = useState(0);

  const abortsRef = useRef<Map<string, AbortController>>(new Map());
  const activeRef = useRef<Set<string>>(new Set());
  const pendingRef = useRef(0);
  const mountedRef = useRef(true);
  const contestIdsKey = contestIds.join("\0");
  const contestIdsRef = useRef(contestIds);
  contestIdsRef.current = contestIds;

  const syncActive = useCallback(() => {
    setActiveIds(new Set(activeRef.current));
  }, []);

  const markActive = useCallback(
    (id: string, on: boolean) => {
      if (on) activeRef.current.add(id);
      else activeRef.current.delete(id);
      syncActive();
    },
    [syncActive],
  );

  const bumpPending = useCallback((delta: number) => {
    pendingRef.current += delta;
    if (pendingRef.current <= 0) {
      pendingRef.current = 0;
      if (mountedRef.current) onSettledRef.current?.();
    }
  }, []);

  const clearContestProgress = useCallback((id: string) => {
    setProgressById((prev) => {
      if (!(id in prev)) return prev;
      const next = { ...prev };
      delete next[id];
      return next;
    });
  }, []);

  const finishOne = useCallback(
    async (id: string, ok: boolean) => {
      setSettlingCount((n) => n + 1);
      try {
        if (ok) {
          setErrorById((prev) => {
            if (!(id in prev)) return prev;
            const next = { ...prev };
            delete next[id];
            return next;
          });
        }
        await onContestSettledRef.current?.(id, ok);
      } finally {
        clearContestProgress(id);
        setSettlingCount((n) => Math.max(0, n - 1));
      }
    },
    [clearContestProgress],
  );

  const runOne = useCallback(
    async (id: string): Promise<void> => {
      if (activeRef.current.has(id)) return;

      markActive(id, true);
      bumpPending(1);

      const controller = new AbortController();
      abortsRef.current.get(id)?.abort();
      abortsRef.current.set(id, controller);

      let settledOk: boolean | null = null;
      let jobIdWritten = false;
      try {
        setErrorById((prev) => {
          if (!(id in prev)) return prev;
          const next = { ...prev };
          delete next[id];
          return next;
        });
        setProgressById((prev) => ({ ...prev, [id]: null }));

        const res = await postContestSync({
          path: { id },
          headers: authHeaders(),
        });

        if (res.error) {
          if (res.response.status === 401) onUnauthorizedRef.current?.();
          if (res.response.status === 404) clearStoredJobId(id, syncJobStorageKey);
          setErrorById((prev) => ({
            ...prev,
            [id]: formatApiError(res.error, res.response.status),
          }));
          settledOk = false;
          return;
        }

        const jobId = extractJobId(res.data);
        if (!jobId) {
          setErrorById((prev) => ({ ...prev, [id]: "сервер не вернул jobId" }));
          settledOk = false;
          return;
        }
        writeStoredJobId(id, jobId, syncJobStorageKey);
        jobIdWritten = true;
        const outcome = await watchContestJob({
          contestId: id,
          jobId,
          storageKey: syncJobStorageKey,
          signal: controller.signal,
          onProgress: (p) => setProgressById((prev) => ({ ...prev, [id]: p })),
          onImportDone: () => onImportDoneRef.current?.(id),
        });
        settledOk = outcome.ok;
        if (!outcome.ok) {
          setErrorById((prev) => ({
            ...prev,
            [id]: outcome.error ?? SYNC_JOB_FAIL_MESSAGE,
          }));
        }
      } catch (e) {
        if (isAbortError(e)) return;
        if (!jobIdWritten) clearStoredJobId(id, syncJobStorageKey);
        setErrorById((prev) => ({
          ...prev,
          [id]: formatApiError(e),
        }));
        settledOk = false;
      } finally {
        if (settledOk !== null) {
          await finishOne(id, settledOk);
        } else {
          clearContestProgress(id);
        }
        if (abortsRef.current.get(id) === controller) abortsRef.current.delete(id);
        markActive(id, false);
        bumpPending(-1);
      }
    },
    [bumpPending, clearContestProgress, finishOne, markActive],
  );

  const resumeOne = useCallback(
    async (id: string, jobId: string): Promise<void> => {
      if (activeRef.current.has(id)) return;

      markActive(id, true);
      bumpPending(1);

      const controller = new AbortController();
      abortsRef.current.get(id)?.abort();
      abortsRef.current.set(id, controller);

      let settledOk: boolean | null = null;
      try {
        const result = await resumeContestJob({
          contestId: id,
          jobId,
          storageKey: syncJobStorageKey,
          signal: controller.signal,
          onUnauthorized: () => onUnauthorizedRef.current?.(),
          onProgress: (p) => setProgressById((prev) => ({ ...prev, [id]: p })),
          onImportDone: () => onImportDoneRef.current?.(id),
        });

        if (controller.signal.aborted) return;

        if (result.kind === "aborted" || result.kind === "cleared") return;

        if (result.kind === "done") {
          settledOk = result.ok;
          if (!result.ok) {
            setErrorById((prev) => ({
              ...prev,
              [id]: result.error ?? SYNC_JOB_FAIL_MESSAGE,
            }));
          }
          return;
        }

        settledOk = result.ok;
        if (!result.ok) {
          setErrorById((prev) => ({
            ...prev,
            [id]: result.error ?? SYNC_JOB_FAIL_MESSAGE,
          }));
        }
      } catch (e) {
        if (isAbortError(e)) return;
        setErrorById((prev) => ({
          ...prev,
          [id]: formatApiError(e),
        }));
        settledOk = false;
      } finally {
        if (settledOk !== null) {
          await finishOne(id, settledOk);
        } else {
          clearContestProgress(id);
        }
        if (abortsRef.current.get(id) === controller) abortsRef.current.delete(id);
        markActive(id, false);
        bumpPending(-1);
      }
    },
    [bumpPending, clearContestProgress, finishOne, markActive],
  );

  const startAll = useCallback(() => {
    setErrorById({});
    for (const id of contestIdsRef.current) {
      if (activeRef.current.has(id)) continue;
      void runOne(id);
    }
  }, [runOne]);

  useEffect(() => {
    const ids = contestIdsKey ? contestIdsKey.split("\0").filter(Boolean) : [];
    for (const id of ids) {
      if (activeRef.current.has(id)) continue;
      const jobId = readStoredJobId(id, syncJobStorageKey);
      if (!jobId) continue;
      void resumeOne(id, jobId);
    }
  }, [contestIdsKey, resumeOne]);

  useEffect(() => {
    const ids = new Set(contestIdsKey ? contestIdsKey.split("\0").filter(Boolean) : []);
    let changed = false;
    for (const [id, controller] of abortsRef.current) {
      if (ids.has(id)) continue;
      controller.abort();
      abortsRef.current.delete(id);
      activeRef.current.delete(id);
      changed = true;
    }
    if (changed) syncActive();
  }, [contestIdsKey, syncActive]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      for (const controller of abortsRef.current.values()) {
        controller.abort();
      }
      abortsRef.current.clear();
      activeRef.current.clear();
      pendingRef.current = 0;
    };
  }, []);

  const isRunning = activeIds.size > 0;
  const isBusy = isRunning || settlingCount > 0;

  const batchProgress = useMemo(() => {
    if (!isBusy) return null;
    const total = contestIds.length;
    if (total === 0) return null;
    return { done: Math.max(0, total - activeIds.size), total };
  }, [activeIds.size, contestIds.length, isBusy]);

  const error = useMemo(() => {
    const entries = Object.entries(errorById);
    if (entries.length === 0) return null;
    if (entries.length === 1) {
      const [id, msg] = entries[0]!;
      return `${id}: ${msg}`;
    }
    return entries.map(([id, msg]) => `${id}: ${msg}`).join("; ");
  }, [errorById]);

  return {
    startAll,
    progressById,
    errorById,
    error,
    isRunning,
    isBusy,
    batchProgress,
  };
}
