import { useCallback } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import { getContestEjudgeLogin } from "@/client/sdk.gen";
import { authHeaders } from "@/features/auth/authStorage";
import { openEjudgeContest } from "@/features/ejudge/openEjudge";

export function useEjudgeContest() {
  const { id: contestId } = useParams<{ id: string }>();
  const { data, isLoading } = useQuery({
    queryKey: ["ejudge-contest-login", contestId],
    queryFn: async () => {
      const res = await getContestEjudgeLogin({
        path: { id: contestId! },
        headers: authHeaders(),
      });
      if (res.response.status === 503 || res.response.status === 404) {
        return null;
      }
      if (!res.data) {
        throw new Error(`ejudge login: HTTP ${res.response.status}`);
      }
      return res.data;
    },
    enabled: Boolean(contestId),
    staleTime: 60_000,
    retry: false,
  });

  const openContest = useCallback(() => {
    if (!data?.login) return;
    openEjudgeContest(data);
  }, [data]);

  return {
    available: Boolean(data?.login),
    isLoading,
    openContest,
  };
}
