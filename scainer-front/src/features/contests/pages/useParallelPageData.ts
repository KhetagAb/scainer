import { useMemo } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import {
  getContestFindingsOptions,
  getContestProblemsOptions,
  getContestsOptions,
} from "@/client/@tanstack/react-query.gen";
import type { FindingView, ProblemInfo, ReportData } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  UNGROUPED_PARALLEL,
  compareContestId,
  compactContestName,
} from "@/features/contests/shared/contestHelpers";
import {
  buildProblemSignalStats,
  contestHasStrongSignals,
  contestPendingCount,
} from "@/features/contests/shared/problemSignalStats";

type Options = {
  parallelId: string;
  threshold: number;
  onUnauthorized: () => void;
};

export function useParallelPageData({ parallelId, threshold, onUnauthorized }: Options) {
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contests = contestsQuery.data ?? [];
  const parallelContests = useMemo(
    () =>
      contests.filter((c) =>
        parallelId === UNGROUPED_PARALLEL ? !c.parallelId : c.parallelId === parallelId,
      ),
    [contests, parallelId],
  );

  const contestIds = useMemo(
    () => parallelContests.map((c) => c.id),
    [parallelContests],
  );

  const headers = authHeaders();
  const findingsQueries = useQueries({
    queries: parallelContests.map((c) => ({
      ...getContestFindingsOptions({ path: { id: c.id }, headers }),
    })),
  });
  const problemsQueries = useQueries({
    queries: parallelContests.map((c) => ({
      ...getContestProblemsOptions({ path: { id: c.id }, headers }),
    })),
  });

  const checkUnauthorized = () => {
    if (contestsQuery.isError && (contestsQuery.error as { status?: number })?.status === 401) {
      onUnauthorized();
      return true;
    }
    for (const q of [...findingsQueries, ...problemsQueries]) {
      if (q.isError && (q.error as { status?: number })?.status === 401) {
        onUnauthorized();
        return true;
      }
    }
    return false;
  };

  const rows = useMemo(
    () =>
      parallelContests
        .map((c, i) => {
          const report = findingsQueries[i]?.data as ReportData | undefined;
          const findings = (report?.findings ?? []) as FindingView[];
          const problems = (problemsQueries[i]?.data ?? []) as ProblemInfo[];
          const findingsQ = findingsQueries[i];
          const problemsQ = problemsQueries[i];
          const ready = Boolean(findingsQ?.isSuccess && problemsQ?.isSuccess);
          const submissionCount = c.submissionCount ?? 0;
          const problemStats = ready
            ? buildProblemSignalStats(problems, findings, threshold)
            : undefined;
          const statsLoading =
            !ready &&
            !findingsQ?.isError &&
            !problemsQ?.isError &&
            (findingsQ?.isPending ||
              problemsQ?.isPending ||
              findingsQ?.isFetching ||
              problemsQ?.isFetching);

          return {
            id: c.id,
            name: c.name,
            displayName: compactContestName(c.name),
            freshness: {
              lastImportedAt: c.lastImportedAt,
              computedAt: c.computedAt,
            },
            hasStrongSignals: contestHasStrongSignals(problemStats),
            pendingCount: contestPendingCount(problemStats),
            statsLoading,
            stats: {
              id: c.id,
              submissionCount,
              problemCount: c.problemCount,
            },
            problemStats,
          };
        })
        .sort((a, b) => compareContestId(a.id, b.id)),
    [parallelContests, findingsQueries, problemsQueries, threshold],
  );

  return {
    contestsQuery,
    parallelContests,
    contestIds,
    rows,
    checkUnauthorized,
  };
}
