import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { UNGROUPED_PARALLEL } from "@/features/contests/contestHelpers";
import {
  DEFAULT_SENSITIVITY,
  useParallelSensitivity,
} from "@/features/contests/useParallelSensitivity";
import type { GroupBy } from "@/features/findings/reportModel";
import {
  DEFAULT_GROUP_BY,
  useParallelFindingsGroupBy,
} from "@/features/findings/useParallelFindingsGroupBy";

type SensitivityContextValue = {
  /** null — слайдер не показываем (главная и пр.). */
  scopeParallelId: string | null;
  /** true на странице контеста — там есть findings и группировка. */
  onContestPage: boolean;
  threshold: number;
  setThreshold: (value: number) => void;
  groupBy: GroupBy;
  setGroupBy: (value: GroupBy) => void;
};

const SensitivityContext = createContext<SensitivityContextValue>({
  scopeParallelId: null,
  onContestPage: false,
  threshold: DEFAULT_SENSITIVITY,
  setThreshold: () => {},
  groupBy: DEFAULT_GROUP_BY,
  setGroupBy: () => {},
});

/** parallelId из URL: /parallels/:id или parallel контеста /contests/:id. */
function useSensitivityScope(): { parallelId: string | null; onContestPage: boolean } {
  const { pathname } = useLocation();
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  return useMemo(() => {
    const parallelMatch = pathname.match(/^\/parallels\/([^/]+)\/?$/);
    if (parallelMatch) {
      try {
        return { parallelId: decodeURIComponent(parallelMatch[1]), onContestPage: false };
      } catch {
        return { parallelId: parallelMatch[1], onContestPage: false };
      }
    }

    const contestMatch = pathname.match(/^\/contests\/([^/]+)\/?$/);
    if (contestMatch) {
      let contestId = contestMatch[1];
      try {
        contestId = decodeURIComponent(contestId);
      } catch {
        /* keep raw */
      }
      const contest = contestsQuery.data?.find((c) => c.id === contestId);
      return {
        parallelId: contest ? contest.parallelId || UNGROUPED_PARALLEL : UNGROUPED_PARALLEL,
        onContestPage: true,
      };
    }

    return { parallelId: null, onContestPage: false };
  }, [pathname, contestsQuery.data]);
}

export function SensitivityProvider({ children }: { children: ReactNode }) {
  const { parallelId: scopeParallelId, onContestPage } = useSensitivityScope();
  const { threshold, setThreshold } = useParallelSensitivity(
    scopeParallelId ?? UNGROUPED_PARALLEL,
  );
  const { groupBy, setGroupBy } = useParallelFindingsGroupBy(
    scopeParallelId ?? UNGROUPED_PARALLEL,
  );

  const value = useMemo(
    () => ({
      scopeParallelId,
      onContestPage,
      threshold,
      setThreshold,
      groupBy,
      setGroupBy,
    }),
    [scopeParallelId, onContestPage, threshold, setThreshold, groupBy, setGroupBy],
  );

  return (
    <SensitivityContext.Provider value={value}>{children}</SensitivityContext.Provider>
  );
}

export function useSensitivity(): SensitivityContextValue {
  return useContext(SensitivityContext);
}
