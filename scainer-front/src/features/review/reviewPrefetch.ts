import type { QueryClient } from "@tanstack/react-query";
import { getSubmissionCommentsOptions } from "@/client/@tanstack/react-query.gen";
import type { SubmissionListItem } from "@/client/types.gen";
import { authHeaders } from "@/features/auth/authStorage";

export async function ensureProblemComments(
  queryClient: QueryClient,
  contestId: string,
  submissions: SubmissionListItem[],
): Promise<void> {
  if (!submissions.length) return;

  await Promise.all(
    submissions.map((s) =>
      queryClient
        .ensureQueryData(
          getSubmissionCommentsOptions({
            path: { id: contestId, submissionId: s.id },
            headers: authHeaders(),
          }),
        )
        .catch(() => undefined),
    ),
  );
}

export function prefetchSubmissionComments(
  queryClient: QueryClient,
  contestId: string,
  submissionId: string,
): void {
  void queryClient.prefetchQuery(
    getSubmissionCommentsOptions({
      path: { id: contestId, submissionId },
      headers: authHeaders(),
    }),
  );
}
