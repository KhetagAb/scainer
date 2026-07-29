import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { QueryClient } from "@tanstack/react-query";
import type { ProblemInfo, SubmissionListItem } from "@/client/types.gen";
import {
  firstProblem,
  firstProblemWithMatches,
  indexSubmissionsByProblem,
  nextProblemWithMatches,
  queueForProblem,
  hasHiddenSubmissions,
} from "@/features/review/reviewModel";
import type { ReviewFiltersInput } from "@/features/review/reviewFilterUtils";
import { ensureProblemComments } from "@/features/review/reviewPrefetch";

type Options = {
  contestId: string | undefined;
  problemId: string | null;
  problems: ProblemInfo[];
  items: SubmissionListItem[];
  reviewFilters: ReviewFiltersInput;
  submissionsLoading: boolean;
  queryClient: QueryClient;
};

export function useReviewProblemNav({
  contestId,
  problemId,
  problems,
  items,
  reviewFilters,
  submissionsLoading,
  queryClient,
}: Options) {
  const navigate = useNavigate();
  const [isProblemTransitioning, setIsProblemTransitioning] = useState(false);

  const itemsByProblem = useMemo(() => indexSubmissionsByProblem(items), [items]);

  const nextProblemId = useMemo(
    () =>
      problemId
        ? nextProblemWithMatches(problems, items, problemId, reviewFilters)
        : null,
    [problemId, problems, items, reviewFilters],
  );

  useEffect(() => {
    if (problemId) return;
    if (submissionsLoading) return;
    if (!contestId || !problems.length) return;

    const target =
      firstProblemWithMatches(problems, items, reviewFilters) ?? firstProblem(problems);
    if (!target) return;

    navigate(
      `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(target)}`,
      { replace: true },
    );
  }, [
    problemId,
    submissionsLoading,
    problems,
    items,
    contestId,
    navigate,
    reviewFilters,
  ]);

  const goToNextProblem = useCallback(async () => {
    if (!nextProblemId || !contestId || !problemId || isProblemTransitioning) return;

    setIsProblemTransitioning(true);
    try {
      const targetItems = itemsByProblem.get(nextProblemId) ?? [];
      const targetQueue = queueForProblem(targetItems, reviewFilters);
      await ensureProblemComments(queryClient, contestId, targetQueue);

      navigate(
        `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(nextProblemId)}`,
        { replace: true },
      );
      if (window.location.hash) {
        window.history.replaceState(
          null,
          "",
          window.location.pathname + window.location.search,
        );
      }
      window.scrollTo(0, 0);
    } finally {
      setIsProblemTransitioning(false);
    }
  }, [
    nextProblemId,
    contestId,
    problemId,
    isProblemTransitioning,
    itemsByProblem,
    reviewFilters,
    queryClient,
    navigate,
  ]);

  const hasHiddenForProblem = useCallback(
    (pid: string) => {
      const bucket = itemsByProblem.get(pid) ?? [];
      return hasHiddenSubmissions(bucket, reviewFilters);
    },
    [itemsByProblem, reviewFilters],
  );

  return {
    nextProblemId,
    goToNextProblem,
    isProblemTransitioning,
    hasHiddenForProblem,
    itemsByProblem,
  };
}
