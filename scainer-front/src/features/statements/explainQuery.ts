import { queryOptions, type QueryClient } from "@tanstack/react-query";
import type { ProblemStatementExplainView } from "@/client/types.gen";
import { getContestProblemStatementExplainQueryKey } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { fetchProblemStatementExplain } from "@/features/statements/fetchProblemStatementExplain";

export function explainQueryKey(contestId: string, problemId: string) {
  return getContestProblemStatementExplainQueryKey({
    path: { id: contestId, problemId },
    headers: authHeaders(),
  });
}

export function explainQueryOptions(contestId: string, problemId: string) {
  return queryOptions({
    queryKey: explainQueryKey(contestId, problemId),
    queryFn: () => fetchProblemStatementExplain(contestId, problemId),
  });
}

export function getCachedProblemStatementExplain(
  queryClient: QueryClient,
  contestId: string,
  problemId: string,
): ProblemStatementExplainView | undefined {
  return queryClient.getQueryData(explainQueryKey(contestId, problemId));
}

export function prefetchProblemStatementExplain(
  queryClient: QueryClient,
  contestId: string,
  problemId: string,
): void {
  void queryClient.prefetchQuery(explainQueryOptions(contestId, problemId)).catch(() => undefined);
}

export function fetchProblemStatementExplainQuery(
  queryClient: QueryClient,
  contestId: string,
  problemId: string,
): Promise<ProblemStatementExplainView> {
  return queryClient.fetchQuery(explainQueryOptions(contestId, problemId));
}

export function setCachedProblemStatementExplain(
  queryClient: QueryClient,
  contestId: string,
  problemId: string,
  data: ProblemStatementExplainView,
): void {
  queryClient.setQueryData(explainQueryKey(contestId, problemId), data);
}

export function removeProblemStatementExplainQuery(
  queryClient: QueryClient,
  contestId: string,
  problemId: string,
): void {
  queryClient.removeQueries({ queryKey: explainQueryKey(contestId, problemId) });
}
