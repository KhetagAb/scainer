import type { FindingView, ProblemInfo } from "@/client/types.gen";
import { collator } from "@/features/contests/contestHelpers";

export type ProblemSignalStat = {
  id: string;
  name: string;
  signalCount: number;
};

/** Findings с score >= threshold → сумма len(signals) по subject.problem. */
export function signalCountsByProblem(
  findings: FindingView[],
  threshold: number,
): Map<string, { count: number; name: string }> {
  const map = new Map<string, { count: number; name: string }>();
  for (const f of findings) {
    if (f.score < threshold) continue;
    const problemId = f.subject.problem;
    if (!problemId) continue;
    const n = f.signals?.length ?? 0;
    const prev = map.get(problemId);
    const name = f.subject.problem_name || prev?.name || "";
    map.set(problemId, { count: (prev?.count ?? 0) + n, name });
  }
  return map;
}

/** Полный список задач + signalCount (0 если нет сигналов выше порога). */
export function buildProblemSignalStats(
  problems: ProblemInfo[],
  findings: FindingView[],
  threshold: number,
): ProblemSignalStat[] {
  const counts = signalCountsByProblem(findings, threshold);
  const stats = problems.map((p) => ({
    id: p.id,
    name: p.name || counts.get(p.id)?.name || "",
    signalCount: counts.get(p.id)?.count ?? 0,
  }));
  stats.sort((a, b) => {
    const byName = collator.compare(a.name || a.id, b.name || b.id);
    if (byName !== 0) return byName;
    return collator.compare(a.id, b.id);
  });
  return stats;
}

/** Relative heat: count/max → 4 уровня (нет / мало / много / crit), как в легенде. */
export function problemSignalLevel(count: number, max: number): string {
  if (count <= 0 || max <= 0) return "level-0";
  const ratio = count / max;
  if (ratio < 1 / 3) return "level-low";
  if (ratio < 2 / 3) return "level-high";
  return "level-crit";
}

/** 100 * sum(score) / submissionCount для findings выше порога; null если посылок 0. */
export function weightedSuspicionFromFindings(
  findings: FindingView[],
  threshold: number,
  submissionCount: number,
): number | null {
  if (submissionCount <= 0) return null;
  let sum = 0;
  for (const f of findings) {
    if (f.score < threshold) continue;
    sum += f.score;
  }
  return (100 * sum) / submissionCount;
}

export function findingsCountAboveThreshold(findings: FindingView[], threshold: number): number {
  let n = 0;
  for (const f of findings) {
    if (f.score >= threshold) n += 1;
  }
  return n;
}
