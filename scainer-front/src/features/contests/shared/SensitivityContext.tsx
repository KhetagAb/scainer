import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { UNGROUPED_PARALLEL } from "@/features/contests/shared/contestHelpers";
import {
  DEFAULT_SENSITIVITY,
  useUserSensitivity,
} from "@/features/contests/shared/useUserSensitivity";
import type { GroupBy } from "@/features/findings/reportModel";
import {
  DEFAULT_GROUP_BY,
  useParallelFindingsGroupBy,
} from "@/features/findings/useParallelFindingsGroupBy";

type SensitivityContextValue = {
  /** null — вне параллели/контеста (главная и пр.). */
  scopeParallelId: string | null;
  /** true на /contests/:id/(findings|review) — табы в шапке контеста. */
  onContestPage: boolean;
  /** true только на findings — табы группировки. */
  onFindingsPage: boolean;
  threshold: number;
  setThreshold: (value: number) => void;
  groupBy: GroupBy;
  setGroupBy: (value: GroupBy) => void;
};

const SensitivityContext = createContext<SensitivityContextValue>({
  scopeParallelId: null,
  onContestPage: false,
  onFindingsPage: false,
  threshold: DEFAULT_SENSITIVITY,
  setThreshold: () => {},
  groupBy: DEFAULT_GROUP_BY,
  setGroupBy: () => {},
});

/** parallelId из URL: /parallels/:id или /contests/:id/… (для groupBy). */
function useSensitivityScope(): {
  parallelId: string | null;
  onFindingsPage: boolean;
  onReviewPage: boolean;
} {
  const { pathname } = useLocation();
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  return useMemo(() => {
    const parallelMatch = pathname.match(/^\/parallels\/([^/]+)\/?$/);
    if (parallelMatch) {
      try {
        return {
          parallelId: decodeURIComponent(parallelMatch[1]),
          onFindingsPage: false,
          onReviewPage: false,
        };
      } catch {
        return {
          parallelId: parallelMatch[1],
          onFindingsPage: false,
          onReviewPage: false,
        };
      }
    }

    const contestMatch = pathname.match(/^\/contests\/([^/]+)/);
    if (contestMatch) {
      let contestId = contestMatch[1];
      try {
        contestId = decodeURIComponent(contestId);
      } catch {
        /* keep raw */
      }
      const contest = contestsQuery.data?.find((c) => c.id === contestId);
      const onFindingsPage = /\/contests\/[^/]+\/findings/.test(pathname);
      const onReviewPage = /\/contests\/[^/]+\/review/.test(pathname);
      return {
        parallelId: contest ? contest.parallelId || UNGROUPED_PARALLEL : UNGROUPED_PARALLEL,
        onFindingsPage,
        onReviewPage,
      };
    }

    return { parallelId: null, onFindingsPage: false, onReviewPage: false };
  }, [pathname, contestsQuery.data]);
}

export function SensitivityProvider({ children }: { children: ReactNode }) {
  const { parallelId: scopeParallelId, onFindingsPage, onReviewPage } =
    useSensitivityScope();
  const { threshold, setThreshold } = useUserSensitivity();
  const { groupBy, setGroupBy } = useParallelFindingsGroupBy(
    scopeParallelId ?? UNGROUPED_PARALLEL,
  );

  const value = useMemo(
    () => ({
      // threshold общий; groupBy — только на findings.
      scopeParallelId,
      onContestPage: onFindingsPage || onReviewPage,
      onFindingsPage,
      threshold,
      setThreshold,
      groupBy,
      setGroupBy,
    }),
    [
      scopeParallelId,
      onFindingsPage,
      onReviewPage,
      threshold,
      setThreshold,
      groupBy,
      setGroupBy,
    ],
  );

  return (
    <SensitivityContext.Provider value={value}>{children}</SensitivityContext.Provider>
  );
}

export function useSensitivity(): SensitivityContextValue {
  return useContext(SensitivityContext);
}
