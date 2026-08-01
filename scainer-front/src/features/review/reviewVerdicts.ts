export function isPendingReview(verdict: string): boolean {
  const v = verdict.trim().toUpperCase();
  return v === "PR" || v === "PD";
}

export function isPrVerdict(verdict: string): boolean {
  return verdict.trim().toUpperCase() === "PR";
}

const VERDICT_LABELS: Record<string, string> = {
  CE: "Compilation error",
  WA: "Wrong answer",
  TL: "Time limit",
  ML: "Memory limit",
  CF: "Check failed",
  DQ: "Disqualified",
  RT: "Run-time error",
  IG: "Ignored",
  PE: "Presentation error",
  RU: "Running",
};

export function formatVerdictLabel(verdict: string): string {
  const v = verdict.trim().toUpperCase();
  if (v === "PR" || v === "PD") return "Pending Review";
  if (v === "OK") return "Accepted";
  if (v === "RJ") return "Rejected";
  if (VERDICT_LABELS[v]) return VERDICT_LABELS[v];
  if (verdict.trim().toLowerCase() === "compilation error") return "Compilation error";
  return verdict.trim() || "—";
}

export function verdictChipTone(verdict: string): "ok" | "pr" | "fail" | "neutral" | "dq" {
  const v = verdict.trim().toUpperCase();
  if (!v || v === "—" || v === "-") return "neutral";
  if (v === "OK") return "ok";
  if (v === "PR" || v === "PD") return "pr";
  if (v === "RU" || v === "CG" || v === "CD") return "neutral";
  if (v === "CE" || v === "CF") return "neutral";
  if (v === "DQ") return "dq";
  return "fail";
}
