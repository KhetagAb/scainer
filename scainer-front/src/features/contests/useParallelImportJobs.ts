import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getJob, postContestImport } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { subscribeJobEvents } from "@/features/contests/jobProgress";
import {
  clearStoredJobId,
  extractJobId,
  isAbortError,
  progressFromState,
  readStoredJobId,
  writeStoredJobId,
  type ImportProgress,
} from "@/features/contests/importJobShared";

export type { ImportProgress } from "@/features/contests/importJobShared";

type UseParallelImportJobsOptions = {
  contestIds: string[];
  onUnauthorized?: () => void;
  /** Когда все job'ы текущей пачки (start + resume) завершились. */
  onSettled?: () => void;
  /** Один раз на контест при переходе в analyzing — посылки уже в store. */
  onImportDone?: (contestId: string) => void;
  /**
   * После завершения job'а одного контеста (до снятия прогресс-бара).
   * Можно дождаться refetch findings/problems, чтобы чипы появились без «дыры».
   */
  onContestSettled?: (contestId: string, ok: boolean) => void | Promise<void>;
};

/**
 * Пачка import-job'ов для всех контестов параллели.
 * Отдельный POST …/import на каждый id; прогресс — Record по contestId.
 */
export function useParallelImportJobs({
  contestIds,
  onUnauthorized,
  onSettled,
  onImportDone,
  onContestSettled,
}: UseParallelImportJobsOptions) {
  const onSettledRef = useRef(onSettled);
  onSettledRef.current = onSettled;
  const onImportDoneRef = useRef(onImportDone);
  onImportDoneRef.current = onImportDone;
  const onContestSettledRef = useRef(onContestSettled);
  onContestSettledRef.current = onContestSettled;
  const onUnauthorizedRef = useRef(onUnauthorized);
  onUnauthorizedRef.current = onUnauthorized;

  const [progressById, setProgressById] = useState<Record<string, ImportProgress>>({});
  const [errorById, setErrorById] = useState<Record<string, string>>({});
  const [activeIds, setActiveIds] = useState<Set<string>>(() => new Set());

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

  const watchJob = useCallback(
    async (id: string, jobId: string, signal: AbortSignal): Promise<boolean> => {
      let notifiedImportDone = false;
      const final = await subscribeJobEvents(
        jobId,
        (state) => {
          setProgressById((prev) => ({ ...prev, [id]: progressFromState(state) }));
          if (!notifiedImportDone && state.progress.phase === "analyzing") {
            notifiedImportDone = true;
            onImportDoneRef.current?.(id);
          }
        },
        signal,
      );

      clearStoredJobId(id);
      // Прогресс-бар снимаем после onContestSettled — см. finishOne.

      if (final.status === "failed") {
        setErrorById((prev) => ({
          ...prev,
          [id]: final.error || "импорт завершился с ошибкой",
        }));
        return false;
      }
      return true;
    },
    [],
  );

  const finishOne = useCallback(
    async (id: string, ok: boolean) => {
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
      try {
        setErrorById((prev) => {
          if (!(id in prev)) return prev;
          const next = { ...prev };
          delete next[id];
          return next;
        });
        setProgressById((prev) => ({ ...prev, [id]: null }));

        const res = await postContestImport({
          path: { id },
          headers: authHeaders(),
          throwOnError: true,
        });
        const jobId = extractJobId(res.data);
        if (!jobId) {
          throw new Error("сервер не вернул jobId");
        }
        writeStoredJobId(id, jobId);
        settledOk = await watchJob(id, jobId, controller.signal);
      } catch (e) {
        if (isAbortError(e)) return;
        clearStoredJobId(id);
        if ((e as { status?: number })?.status === 401) onUnauthorizedRef.current?.();
        setErrorById((prev) => ({
          ...prev,
          [id]: e instanceof Error ? e.message : String(e),
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
    [bumpPending, clearContestProgress, finishOne, markActive, watchJob],
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
        const snap = await getJob({
          path: { jobId },
          headers: authHeaders(),
        });

        if (controller.signal.aborted) return;

        if (snap.response.status === 404 || !snap.data) {
          clearStoredJobId(id);
          return;
        }

        const status = snap.data.status;
        if (status === "succeeded" || status === "failed") {
          clearStoredJobId(id);
          // Уже завершённый job после remount: при success — дотянем карточку;
          // при fail — не показываем устаревший баннер, storage просто чистим.
          if (status === "succeeded") settledOk = true;
          return;
        }

        setProgressById((prev) => ({ ...prev, [id]: progressFromState(snap.data!) }));
        settledOk = await watchJob(id, jobId, controller.signal);
      } catch (e) {
        if (isAbortError(e)) return;
        clearStoredJobId(id);
        if ((e as { status?: number })?.status === 401) onUnauthorizedRef.current?.();
        if ((e as { status?: number })?.status === 404) return;
        setErrorById((prev) => ({
          ...prev,
          [id]: e instanceof Error ? e.message : String(e),
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
    [bumpPending, clearContestProgress, finishOne, markActive, watchJob],
  );

  const startAll = useCallback(() => {
    setErrorById({});
    for (const id of contestIdsRef.current) {
      if (activeRef.current.has(id)) continue;
      void runOne(id);
    }
  }, [runOne]);

  // Resume сохранённых job'ов при появлении contestIds.
  useEffect(() => {
    const ids = contestIdsKey ? contestIdsKey.split("\0").filter(Boolean) : [];
    for (const id of ids) {
      if (activeRef.current.has(id)) continue;
      const jobId = readStoredJobId(id);
      if (!jobId) continue;
      void resumeOne(id, jobId);
    }
  }, [contestIdsKey, resumeOne]);

  // Abort job'ов контестов, ушедших из списка (смена параллели).
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

  const batchProgress = useMemo(() => {
    if (!isRunning) return null;
    const total = contestIds.length;
    if (total === 0) return null;
    return { done: Math.max(0, total - activeIds.size), total };
  }, [activeIds.size, contestIds.length, isRunning]);

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
    batchProgress,
  };
}
