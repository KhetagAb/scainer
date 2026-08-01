import type { ProblemStatementExplainView } from "@/client/types.gen";

export type FormalizationView =
  | { phase: "hidden" }
  | { phase: "fetching" }
  | { phase: "typing"; data: ProblemStatementExplainView; text: string }
  | { phase: "ready"; data: ProblemStatementExplainView }
  | { phase: "error"; message: string };
