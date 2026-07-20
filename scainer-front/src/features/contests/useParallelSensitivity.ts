import { useCallback, useEffect, useState } from "react";
import { UNGROUPED_PARALLEL } from "@/features/contests/contestHelpers";

export const DEFAULT_SENSITIVITY = 0.7;
export const SENSITIVITY_MARKS: Record<number, string> = { 50: "50%", 70: "70%", 95: "95%" };

const STORAGE_PREFIX = "scainer.sensitivity.";
const MARKS = [0.5, 0.7, 0.95] as const;

function storageKey(parallelId: string): string {
  return `${STORAGE_PREFIX}${parallelId || UNGROUPED_PARALLEL}`;
}

function snapMark(value: number): number {
  return MARKS.reduce((best, m) => (Math.abs(m - value) < Math.abs(best - value) ? m : best));
}

function readStored(parallelId: string): number {
  try {
    const raw = localStorage.getItem(storageKey(parallelId));
    if (raw == null) return DEFAULT_SENSITIVITY;
    const n = Number(raw);
    if (!Number.isFinite(n) || n < 0 || n > 1) return DEFAULT_SENSITIVITY;
    return snapMark(n);
  } catch {
    return DEFAULT_SENSITIVITY;
  }
}

/** Чувствительность (доля 0..1) на уровне параллели, в localStorage. */
export function useParallelSensitivity(parallelId: string | null | undefined) {
  const id = parallelId || UNGROUPED_PARALLEL;
  const [threshold, setThresholdState] = useState(() => readStored(id));

  useEffect(() => {
    setThresholdState(readStored(id));
  }, [id]);

  const setThreshold = useCallback(
    (value: number) => {
      const snapped = snapMark(value);
      setThresholdState(snapped);
      try {
        localStorage.setItem(storageKey(id), String(snapped));
      } catch {
        /* ignore quota / private mode */
      }
    },
    [id],
  );

  return { threshold, setThreshold } as const;
}
