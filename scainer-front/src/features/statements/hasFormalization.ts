import type { ProblemStatementExplainView } from "@/client/types.gen";

export function hasFormalization(data: ProblemStatementExplainView | null | undefined): boolean {
  return Boolean(data?.statement?.trim());
}
