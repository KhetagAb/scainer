import { useCallback, useEffect, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useLocation, useSearchParams } from "react-router-dom";
import {
  getContestFindingsOptions,
  getContestProblemsOptions,
  getContestSubmissionsOptions,
  getContestsOptions,
  getSubmissionCommentsOptions,
} from "@/client/@tanstack/react-query.gen";
import type { ProblemInfo } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { problemSubmissionCountsMap } from "@/features/contests/shared/problemSignalStats";

type Options = {
  contestId?: string;
  onUnauthorized: () => void;
};

export function useContestPageData({ contestId, onUnauthorized }: Options) {
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const queryClient = useQueryClient();

  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contest = contestId
    ? contestsQuery.data?.find((c) => c.id === contestId)
    : undefined;

  const findingsQuery = useQuery({
    ...getContestFindingsOptions({
      path: { id: contestId ?? "" },
      headers: authHeaders(),
    }),
    enabled: Boolean(contestId),
  });

  const problemsQuery = useQuery({
    ...getContestProblemsOptions({
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

  useEffect(() => {
    if (!onReviewPage || !contestId || !submissionsQuery.data) return;
    for (const s of submissionsQuery.data) {
      if (s.verdict !== "PR") continue;
      void queryClient.prefetchQuery(
        getSubmissionCommentsOptions({
          path: { id: contestId, submissionId: s.id },
          headers: authHeaders(),
        }),
      );
    }
  }, [onReviewPage, contestId, submissionsQuery.data, queryClient]);

  const problemSubmissionCounts = useMemo(
    () => problemSubmissionCountsMap((problemsQuery.data ?? []) as ProblemInfo[]),
    [problemsQuery.data],
  );

  const headerStats = useMemo(() => {
    if (!contest) return null;
    return {
      id: contest.id,
      submissionCount: contest.submissionCount ?? 0,
      problemCount: contest.problemCount,
    };
  }, [contest]);

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

  const handleQueryError = useCallback(() => {
    if ((contestsQuery.error as { status?: number })?.status === 401) {
      onUnauthorized();
    }
  }, [contestsQuery.error, onUnauthorized]);

  return {
    contestsQuery,
    contest,
    findingsQuery,
    problemsQuery,
    submissionsQuery,
    problemSubmissionCounts,
    headerStats,
    activeProblemId,
    onReviewPage,
    findingKey: searchParams.get("finding"),
    handleQueryError,
  };
}

export type ContestPageOutletContext = {
  findingsQuery: ReturnType<typeof useContestPageData>["findingsQuery"];
  submissionsQuery: ReturnType<typeof useContestPageData>["submissionsQuery"];
  problems: ProblemInfo[];
  problemSubmissionCounts: Record<string, number>;
  onUnauthorized: () => void;
  findingKey: string | null;
  reviewPrOnly: boolean;
  setReviewPrOnly: (value: boolean) => void;
};
