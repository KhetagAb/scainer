import { useCallback, useState } from "react";

export const DEFAULT_SENSITIVITY = 0.75;
export const SENSITIVITY_MARKS: Record<number, string> = {
  50: "50%",
  75: "75%",
  90: "90%",
};

const STORAGE_KEY = "scainer.sensitivity";
const MARKS = [0.5, 0.75, 0.9] as const;

function snapMark(value: number): number {
  return MARKS.reduce((best, m) => (Math.abs(m - value) < Math.abs(best - value) ? m : best));
}

function parseStored(raw: string | null): number | null {
  if (raw == null) return null;
  const n = Number(raw);
  if (!Number.isFinite(n) || n < 0 || n > 1) return null;
  return snapMark(n);
}

function readStored(): number {
  try {
    const current = parseStored(localStorage.getItem(STORAGE_KEY));
    if (current != null) return current;
  } catch {
    /* ignore quota / private mode */
  }
  return DEFAULT_SENSITIVITY;
}

/** Чувствительность (доля 0..1) — один параметр пользователя на весь сайт. */
export function useUserSensitivity() {
  const [threshold, setThresholdState] = useState(readStored);

  const setThreshold = useCallback((value: number) => {
    const snapped = snapMark(value);
    setThresholdState(snapped);
    try {
      localStorage.setItem(STORAGE_KEY, String(snapped));
    } catch {
      /* ignore quota / private mode */
    }
  }, []);

  return { threshold, setThreshold } as const;
}
