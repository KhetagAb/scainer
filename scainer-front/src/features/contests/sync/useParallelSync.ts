import { useCallback, useMemo } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { useParallelContestJobs } from "@/features/contests/jobs/useParallelContestJobs";
import {
  debounce,
  invalidateContestQueries,
  refetchContests,
} from "@/features/contests/sync/invalidateContestData";

type UseParallelSyncOptions = {
  contestIds: string[];
  onUnauthorized?: () => void;
};

export function useParallelSync({ contestIds, onUnauthorized }: UseParallelSyncOptions) {
  const queryClient = useQueryClient();
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const refreshContests = useCallback(() => {
    refetchContests(contestsQuery);
  }, [contestsQuery]);

  const refreshContestsDebounced = useMemo(
    () => debounce(() => refreshContests(), 400),
    [refreshContests],
  );

  const refreshContestCard = useCallback(
    async (id: string, ok: boolean) => {
      if (!ok) return;
      await Promise.all([
        contestsQuery.refetch(),
        invalidateContestQueries(queryClient, id),
      ]);
    },
    [contestsQuery, queryClient],
  );

  const syncJobs = useParallelContestJobs({
    contestIds,
    onUnauthorized,
    onImportDone: refreshContestsDebounced,
    onContestSettled: refreshContestCard,
    onSettled: refreshContests,
  });

  return {
    contestsQuery,
    syncJobs,
    refreshContests,
  };
}
