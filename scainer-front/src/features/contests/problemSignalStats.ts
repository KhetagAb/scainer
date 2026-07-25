import type { FindingView, ProblemInfo } from "@/client/types.gen";
import { collator } from "@/features/contests/contestHelpers";

export type ProblemSignalStat = {
  id: string;
  name: string;
  /** Доля сигналов, %; null если посылок 0. */
  suspiciousSharePercent: number | null;
  /** Посылки со статусом PR (pending review). */
  pendingCount: number;
};

export const PROBLEM_SUSPICION_TOOLTIP =
  "Доля сигналов задачи: 100 × сигналы / все посылки задачи";

/** >10% сигналов — жёлтый; ≤10% (и >0) — зелёный. */
export const PROBLEM_SUSPICION_LOW_MIN = 0.1;

/** Полный список задач + доля сигналов. */
export function buildProblemSignalStats(
  problems: ProblemInfo[],
  findings: FindingView[],
  threshold: number,
): ProblemSignalStat[] {
  const nameById = new Map<string, string>();
  for (const f of findings) {
    const id = f.subject.problem;
    if (!id || nameById.has(id)) continue;
    if (f.subject.problem_name) nameById.set(id, f.subject.problem_name);
  }

  const stats = problems.map((p) => ({
    id: p.id,
    name: p.name || nameById.get(p.id) || "",
    suspiciousSharePercent: problemSuspiciousSharePercent(
      findings,
      threshold,
      p.id,
      p.submissionCount ?? 0,
    ),
    pendingCount: p.pendingCount ?? 0,
  }));
  stats.sort((a, b) => {
    const byName = collator.compare(a.name || a.id, b.name || b.id);
    if (byName !== 0) return byName;
    return collator.compare(a.id, b.id);
  });
  return stats;
}

/** Краткая интерпретация уровня для title/tooltip. */
export function problemSuspicionLevelHint(level: string): string {
  switch (level) {
    case "level-high":
      return "≤10% сигналов — высокий сигнал";
    case "level-low":
      return ">10% сигналов — низкий сигнал";
    default:
      return "нет сигналов";
  }
}

/** Число findings задачи со score ≥ threshold (любой kind субъекта). */
export function countSignalsForProblem(
  findings: FindingView[],
  threshold: number,
  problemId: string,
): number {
  let n = 0;
  for (const f of findings) {
    if (f.score < threshold) continue;
    if (f.subject.problem !== problemId) continue;
    n += 1;
  }
  return n;
}

/**
 * 100 × |сигналы| / все посылки задачи;
 * null если посылок 0.
 */
export function problemSuspiciousSharePercent(
  findings: FindingView[],
  threshold: number,
  problemId: string,
  submissionCount: number,
): number | null {
  if (submissionCount <= 0) return null;
  const n = countSignalsForProblem(findings, threshold, problemId);
  return (100 * n) / submissionCount;
}

/**
 * Цвет чипа по абсолютной доле сигналов:
 * 0 / null → серый; ≤10% → зелёный; >10% → жёлтый.
 */
export function problemSuspicionLevel(percent: number | null | undefined): string {
  if (percent == null || Number.isNaN(percent) || percent <= 0) return "level-0";
  const ratio = percent / 100;
  if (ratio <= PROBLEM_SUSPICION_LOW_MIN) return "level-high";
  return "level-low";
}

/**
 * Есть ли у контеста «сильные» (зелёные) задачи — для подсветки иконки детекта.
 */
export function contestHasStrongSignals(
  problemStats: ProblemSignalStat[] | undefined,
): boolean {
  if (!problemStats?.length) return false;
  for (const p of problemStats) {
    if (problemSuspicionLevel(p.suspiciousSharePercent) === "level-high") return true;
  }
  return false;
}

/** Сумма PR по задачам контеста. */
export function contestPendingCount(problemStats: ProblemSignalStat[] | undefined): number {
  if (!problemStats?.length) return 0;
  let n = 0;
  for (const p of problemStats) n += p.pendingCount;
  return n;
}

export function problemSubmissionCountsMap(problems: ProblemInfo[]): Record<string, number> {
  const out: Record<string, number> = {};
  for (const p of problems) {
    out[p.id] = p.submissionCount ?? 0;
  }
  return out;
}
