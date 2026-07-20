import type { FindingView, ProblemInfo } from "@/client/types.gen";
import { collator } from "@/features/contests/contestHelpers";

export type ProblemSignalStat = {
  id: string;
  name: string;
  /** Доля подозрительных посылок, %; null если посылок 0. */
  suspiciousSharePercent: number | null;
};

export const PROBLEM_SUSPICION_TOOLTIP =
  "Доля подозрительных посылок задачи: 100 × подозрительные / все посылки задачи";

/** >15% подозрительных посылок — жёлтый; ≤15% (и >0) — зелёный. */
export const PROBLEM_SUSPICION_LOW_MIN = 0.15;

/** Полный список задач + доля подозрительных посылок. */
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
      return "≤15% подозрительных — высокий сигнал";
    case "level-low":
      return ">15% подозрительных — низкий сигнал";
    default:
      return "нет подозрительных посылок";
  }
}

/** Уникальные подозрительные посылки задачи (findings ≥ threshold). */
export function countSuspiciousSubmissionsForProblem(
  findings: FindingView[],
  threshold: number,
  problemId: string,
): number {
  const ids = new Set<string>();
  for (const f of findings) {
    if (f.score < threshold) continue;
    if (f.subject.problem !== problemId) continue;
    if (f.subject.submission) ids.add(f.subject.submission);
    for (const sig of f.signals ?? []) {
      for (const ev of sig.evidence ?? []) {
        for (const sp of ev.spans ?? []) {
          if (sp.submission) ids.add(sp.submission);
        }
      }
    }
  }
  return ids.size;
}

/**
 * 100 × |подозрительные посылки| / все посылки задачи;
 * null если посылок 0.
 */
export function problemSuspiciousSharePercent(
  findings: FindingView[],
  threshold: number,
  problemId: string,
  submissionCount: number,
): number | null {
  if (submissionCount <= 0) return null;
  const n = countSuspiciousSubmissionsForProblem(findings, threshold, problemId);
  return (100 * n) / submissionCount;
}

/**
 * Цвет по абсолютной доле подозрительных посылок (чипы, % внутри, корешок):
 * 0 / null → серый; ≤15% → зелёный; >15% → жёлтый.
 */
export function problemSuspicionLevel(percent: number | null | undefined): string {
  if (percent == null || Number.isNaN(percent) || percent <= 0) return "level-0";
  const ratio = percent / 100;
  if (ratio <= PROBLEM_SUSPICION_LOW_MIN) return "level-high";
  return "level-low";
}

/**
 * Корешок карточки контеста:
 * есть зелёные задачи → green; иначе есть жёлтые → yellow; иначе без окраски.
 */
export function contestSpineFromProblemStats(
  problemStats: ProblemSignalStat[] | undefined,
): "green" | "yellow" | null {
  if (!problemStats?.length) return null;
  let hasYellow = false;
  for (const p of problemStats) {
    const heat = problemSuspicionLevel(p.suspiciousSharePercent);
    if (heat === "level-high") return "green";
    if (heat === "level-low") hasYellow = true;
  }
  return hasYellow ? "yellow" : null;
}

export function problemSubmissionCountsMap(problems: ProblemInfo[]): Record<string, number> {
  const out: Record<string, number> = {};
  for (const p of problems) {
    out[p.id] = p.submissionCount ?? 0;
  }
  return out;
}
