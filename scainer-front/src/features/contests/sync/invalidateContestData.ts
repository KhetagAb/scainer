import type { QueryClient, UseQueryResult } from "@tanstack/react-query";
import {
  getContestFindingsQueryKey,
  getContestProblemsQueryKey,
  getContestSubmissionsQueryKey,
} from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";

export function refetchContests(contestsQuery: { refetch: () => unknown }): void {
  void contestsQuery.refetch();
}

export async function invalidateContestQueries(
  queryClient: QueryClient,
  contestId: string,
  opts?: { includeSubmissions?: boolean },
): Promise<void> {
  const tasks = [
    queryClient.invalidateQueries({
      queryKey: getContestFindingsQueryKey({
        path: { id: contestId },
        headers: authHeaders(),
      }),
    }),
    queryClient.invalidateQueries({
      queryKey: getContestProblemsQueryKey({
        path: { id: contestId },
        headers: authHeaders(),
      }),
    }),
  ];

  if (opts?.includeSubmissions) {
    tasks.push(
      queryClient.invalidateQueries({
        queryKey: getContestSubmissionsQueryKey({
          path: { id: contestId },
          headers: authHeaders(),
        }),
      }),
    );
  }

  await Promise.all(tasks);
}

export async function refreshContestAfterSync(
  queryClient: QueryClient,
  contestsQuery: UseQueryResult<unknown>,
  contestId: string,
  ok: boolean,
): Promise<void> {
  if (!ok) return;
  await Promise.all([
    contestsQuery.refetch(),
    invalidateContestQueries(queryClient, contestId, { includeSubmissions: true }),
  ]);
}

export function debounce<T extends (...args: never[]) => void>(
  fn: T,
  ms: number,
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return (...args: Parameters<T>) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      timer = null;
      fn(...args);
    }, ms);
  };
}
