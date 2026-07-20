import { useCallback, useEffect, useState } from "react";
import { UNGROUPED_PARALLEL } from "@/features/contests/contestHelpers";
import type { GroupBy } from "@/features/findings/reportModel";

export const DEFAULT_GROUP_BY: GroupBy = "problem";

const STORAGE_PREFIX = "scainer.findingsGroupBy.";

function storageKey(parallelId: string): string {
  return `${STORAGE_PREFIX}${parallelId || UNGROUPED_PARALLEL}`;
}

function readStored(parallelId: string): GroupBy {
  try {
    const raw = localStorage.getItem(storageKey(parallelId));
    if (raw === "problem" || raw === "participant") return raw;
  } catch {
    /* ignore */
  }
  return DEFAULT_GROUP_BY;
}

/** Группировка findings на уровне параллели, в localStorage (как чувствительность). */
export function useParallelFindingsGroupBy(parallelId: string | null | undefined) {
  const id = parallelId || UNGROUPED_PARALLEL;
  const [groupBy, setGroupByState] = useState<GroupBy>(() => readStored(id));

  useEffect(() => {
    setGroupByState(readStored(id));
  }, [id]);

  const setGroupBy = useCallback(
    (value: GroupBy) => {
      setGroupByState(value);
      try {
        localStorage.setItem(storageKey(id), value);
      } catch {
        /* ignore quota / private mode */
      }
    },
    [id],
  );

  return { groupBy, setGroupBy } as const;
}
