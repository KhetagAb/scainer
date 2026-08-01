import { useEffect, useMemo, useRef, useState } from "react";
import type { SubmissionListItem } from "@/client/types.gen";
import {
  hasHiddenSubmissions,
  queueForProblem,
} from "@/features/review/reviewModel";
import {
  filtersEqual,
  type ReviewFiltersInput,
} from "@/features/review/reviewFilterUtils";

type Options = {
  problemId: string | null;
  problemItems: SubmissionListItem[];
  reviewFilters: ReviewFiltersInput;
  isLoading: boolean;
};

/**
 * Очередь посылок задачи с session stickiness: после OK/RJ панель остаётся
 * в стеке до смены задачи или фильтра (не пересчитываем при обновлении items).
 */
export function useReviewProblemQueue({
  problemId,
  problemItems,
  reviewFilters,
  isLoading,
}: Options) {
  const [sessionProblemId, setSessionProblemId] = useState<string | null>(null);
  const [sessionQueue, setSessionQueue] = useState<SubmissionListItem[]>([]);
  const prevFiltersRef = useRef(reviewFilters);
  const prevProblemIdRef = useRef<string | null>(null);

  useEffect(() => {
    if (!problemId) {
      setSessionProblemId(null);
      setSessionQueue([]);
      prevProblemIdRef.current = null;
      return;
    }
    if (isLoading) return;

    const filtersChanged = !filtersEqual(prevFiltersRef.current, reviewFilters);
    const problemChanged = prevProblemIdRef.current !== problemId;

    if (!problemChanged && !filtersChanged) return;

    setSessionQueue(queueForProblem(problemItems, reviewFilters));
    setSessionProblemId(problemId);
    prevFiltersRef.current = reviewFilters;
    prevProblemIdRef.current = problemId;
  }, [problemId, problemItems, isLoading, reviewFilters]);

  const queue = useMemo(() => {
    if (!problemId) return [];
    if (sessionProblemId === problemId) return sessionQueue;
    return queueForProblem(problemItems, reviewFilters);
  }, [problemId, sessionProblemId, sessionQueue, problemItems, reviewFilters]);

  const hiddenOnProblem = useMemo(
    () => hasHiddenSubmissions(problemItems, reviewFilters),
    [problemItems, reviewFilters],
  );

  return { queue, hiddenOnProblem };
}
