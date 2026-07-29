import { useDeferredValue, useEffect, useMemo, useRef, useState } from "react";
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
  const deferredFilters = useDeferredValue(reviewFilters);
  const [sessionProblemId, setSessionProblemId] = useState<string | null>(null);
  const [sessionQueue, setSessionQueue] = useState<SubmissionListItem[]>([]);
  const prevFiltersRef = useRef(reviewFilters);

  useEffect(() => {
    if (!problemId) {
      setSessionProblemId(null);
      setSessionQueue([]);
      return;
    }
    if (isLoading) return;

    const problemChanged = sessionProblemId !== problemId;
    const filtersChanged = !filtersEqual(prevFiltersRef.current, deferredFilters);
    prevFiltersRef.current = deferredFilters;

    if (!problemChanged && !filtersChanged && sessionProblemId === problemId) return;

    setSessionQueue(queueForProblem(problemItems, deferredFilters));
    setSessionProblemId(problemId);
  }, [problemId, sessionProblemId, problemItems, isLoading, deferredFilters]);

  const queue = useMemo(() => {
    if (!problemId) return [];
    if (sessionProblemId === problemId) return sessionQueue;
    return queueForProblem(problemItems, deferredFilters);
  }, [problemId, sessionProblemId, sessionQueue, problemItems, deferredFilters]);

  const hiddenOnProblem = useMemo(
    () => hasHiddenSubmissions(problemItems, reviewFilters),
    [problemItems, reviewFilters],
  );

  return { queue, deferredFilters, hiddenOnProblem };
}
