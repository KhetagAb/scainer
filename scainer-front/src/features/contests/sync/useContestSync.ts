import { useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getContestsOptions } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { useContestJob } from "@/features/contests/jobs/useContestJob";
import { buildSyncAction } from "@/features/contests/sync/contestDataStatus";
import {
  refetchContests,
  refreshContestAfterSync,
} from "@/features/contests/sync/invalidateContestData";

type UseContestSyncOptions = {
  contestId?: string | null;
  onUnauthorized?: () => void;
};

export function useContestSync({ contestId, onUnauthorized }: UseContestSyncOptions) {
  const queryClient = useQueryClient();
  const contestsQuery = useQuery({
    ...getContestsOptions({ headers: authHeaders() }),
  });

  const contest = contestId
    ? contestsQuery.data?.find((c) => c.id === contestId)
    : undefined;

  const refreshContests = useCallback(() => {
    refetchContests(contestsQuery);
  }, [contestsQuery]);

  const refreshAfterJob = useCallback(
    async (ok: boolean) => {
      if (!contestId) return;
      await refreshContestAfterSync(queryClient, contestsQuery, contestId, ok);
    },
    [contestId, contestsQuery, queryClient],
  );

  const syncJob = useContestJob(onUnauthorized, {
    contestId,
    onImportDone: refreshContests,
    onSettled: refreshAfterJob,
  });

  const resyncJob = useContestJob(onUnauthorized, {
    contestId,
    kind: "resync",
    onImportDone: refreshContests,
    onSettled: refreshAfterJob,
  });

  const jobBusy = syncJob.isBusy || resyncJob.isBusy;

  const syncAction = buildSyncAction(
    {
      lastImportedAt: contest?.lastImportedAt,
      computedAt: contest?.computedAt,
    },
    { busy: jobBusy },
  );

  const startSync = useCallback(() => {
    if (!contestId) return;
    void syncJob.start(contestId);
  }, [contestId, syncJob]);

  const startResync = useCallback(() => {
    if (!contestId) return;
    void resyncJob.start(contestId);
  }, [contestId, resyncJob]);

  const progress = syncJob.isBusy
    ? syncJob.progress
    : resyncJob.isBusy
      ? resyncJob.progress
      : null;

  return {
    contestsQuery,
    contest,
    syncAction,
    progress,
    isBusy: jobBusy,
    error: syncJob.error ?? resyncJob.error,
    startSync,
    startResync,
  };
}
