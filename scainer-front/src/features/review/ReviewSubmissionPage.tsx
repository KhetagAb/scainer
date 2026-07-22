import { useEffect } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import type { UseQueryResult } from "@tanstack/react-query";
import type { SubmissionListItem } from "@/client/types.gen";
import { submissionPanelId } from "@/features/review/reviewFindings";

/** Старый URL /review/:submissionId → стек задачи с якорем на посылку. */
export default function ReviewSubmissionPage({
  submissionsQuery,
}: {
  submissionsQuery: UseQueryResult<SubmissionListItem[]>;
  findingsQuery: UseQueryResult<unknown>;
  onUnauthorized: () => void;
}) {
  const { id: contestId, submissionId: rawSubmissionId } = useParams<{
    id: string;
    submissionId: string;
  }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const submissionId = rawSubmissionId ? decodeURIComponent(rawSubmissionId) : "";
  const items = submissionsQuery.data ?? [];
  const fromList = items.find((s) => s.id === submissionId);
  const problemId = searchParams.get("problem") || fromList?.problem || "";

  useEffect(() => {
    if (!contestId || !submissionId) return;
    if (submissionsQuery.isLoading) return;
    if (!problemId) {
      navigate(`/contests/${encodeURIComponent(contestId)}/review`, { replace: true });
      return;
    }
    const hash = submissionPanelId(submissionId);
    navigate(
      `/contests/${encodeURIComponent(contestId)}/review?problem=${encodeURIComponent(problemId)}#${hash}`,
      { replace: true },
    );
  }, [contestId, submissionId, problemId, submissionsQuery.isLoading, navigate]);

  return <div className="page-center">Переход к ревью…</div>;
}
