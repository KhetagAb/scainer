import { useCallback, useEffect, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useLocation, useSearchParams } from "react-router-dom";
import {
  getContestFindingsOptions,
  getContestSubmissionsOptions,
  getContestsOptions,
} from "@/client/@tanstack/react-query.gen";
import type { ProblemInfo } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { problemSubmissionCountsMap } from "@/features/contests/shared/problemSignalStats";
import {
  indexSubmissionsByProblem,
  nextProblemWithMatches,
  queueForProblem,
} from "@/features/review/reviewModel";
import type { ReviewFiltersInput } from "@/features/review/reviewFilterUtils";
import { prefetchSubmissionComments } from "@/features/review/reviewPrefetch";

const PREFETCH_QUEUE_LIMIT = 12;

type Options = {
  contestId?: string;
  onUnauthorized: () => void;
  reviewFilters?: ReviewFiltersInput;
};

export function useContestPageData({ contestId, onUnauthorized, reviewFilters }: Options) {
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const queryClient = useQueryClient();

  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contest = contestId
    ? contestsQuery.data?.find((c) => c.id === contestId)
    : undefined;

  const problems = useMemo(
    () => (contest?.problems ?? []) as ProblemInfo[],
    [contest?.problems],
  );

  const findingsQuery = useQuery({
    ...getContestFindingsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const submissionsQuery = useQuery({
    ...getContestSubmissionsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const onReviewPage = location.pathname.includes("/review");

  const activeProblemId = useMemo(() => {
    const fromQuery = searchParams.get("problem");
    if (fromQuery) return fromQuery;
    const match = location.pathname.match(/\/review\/([^/]+)/);
    if (!match) return null;
    let sid = match[1];
    try {
      sid = decodeURIComponent(sid);
    } catch {
      /* keep */
    }
    const sub = (submissionsQuery.data ?? []).find((s) => s.id === sid);
    return sub?.problem ?? null;
  }, [searchParams, location.pathname, submissionsQuery.data]);

  useEffect(() => {
    if (!onReviewPage || !contestId || !submissionsQuery.data || !reviewFilters) return;

    const items = submissionsQuery.data;
    const itemsByProblem = indexSubmissionsByProblem(items);
    const activeProblem = searchParams.get("problem") ?? activeProblemId;
    const nextProblem =
      activeProblem && problems.length
        ? nextProblemWithMatches(problems, items, activeProblem, reviewFilters)
        : null;

    const prefetchQueue = (problemId: string) => {
      const problemItems = itemsByProblem.get(problemId) ?? [];
      const queue = queueForProblem(problemItems, reviewFilters).slice(0, PREFETCH_QUEUE_LIMIT);
      for (const s of queue) {
        prefetchSubmissionComments(queryClient, contestId, s.id);
      }
    };

    if (nextProblem) {
      prefetchQueue(nextProblem);
    }

    if (activeProblem && activeProblem !== nextProblem) {
      prefetchQueue(activeProblem);
    }
  }, [
    onReviewPage,
    contestId,
    submissionsQuery.data,
    problems,
    queryClient,
    searchParams,
    activeProblemId,
    reviewFilters,
  ]);

  const problemSubmissionCounts = useMemo(
    () => problemSubmissionCountsMap(problems),
    [problems],
  );

  const handleQueryError = useCallback(() => {
    if ((contestsQuery.error as { status?: number })?.status === 401) {
      onUnauthorized();
    }
  }, [contestsQuery.error, onUnauthorized]);

  return {
    contestsQuery,
    contest,
    findingsQuery,
    problems,
    submissionsQuery,
    problemSubmissionCounts,
    activeProblemId,
    onReviewPage,
    findingKey: searchParams.get("finding"),
    handleQueryError,
  };
}
