import { scoreLevel } from "@/features/findings/reportModel";

export const UNGROUPED_PARALLEL = "__ungrouped";

export const collator = new Intl.Collator("ru", { numeric: true, sensitivity: "base" });

/** Хвост имени после последней точки (пропуская пустые хвосты вроде trailing «.»). */
export function compactContestName(name: string): string {
  if (!name) return name;
  const parts = name.split(".").map((p) => p.trim()).filter(Boolean);
  if (parts.length <= 1) return name.trim();
  return parts[parts.length - 1];
}

export function contestsCountLabel(n: number): string {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod10 === 1 && mod100 !== 11) return `${n} контест`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return `${n} контеста`;
  return `${n} контестов`;
}

export type ContestStatsFields = {
  id?: string;
  submissionCount?: number | null;
  problemCount?: number | null;
  findingsCount?: number | null;
  weightedSuspicionPercent?: number | null;
};

/** Уровень «жара» по взвешенному % — те же пороги, что у score findings. */
export function suspicionLevel(percent?: number | null): string {
  if (percent == null || Number.isNaN(percent) || percent <= 0) return "level-0";
  return scoreLevel(percent / 100);
}

export function formatSuspicionPercent(value?: number | null): string {
  if (value == null || Number.isNaN(value)) return "—";
  return `${Math.round(value)}%`;
}

/** Сначала горячие: % desc → findings desc → id. */
export function compareContestsBySuspicion(a: ContestStatsFields, b: ContestStatsFields): number {
  const pa = a.weightedSuspicionPercent ?? -1;
  const pb = b.weightedSuspicionPercent ?? -1;
  if (pa !== pb) return pb - pa;
  const fa = a.findingsCount ?? 0;
  const fb = b.findingsCount ?? 0;
  if (fa !== fb) return fb - fa;
  return collator.compare(a.id ?? "", b.id ?? "");
}
