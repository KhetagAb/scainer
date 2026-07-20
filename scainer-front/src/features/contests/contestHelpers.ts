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
};

export function formatSuspicionPercent(value?: number | null): string {
  if (value == null || Number.isNaN(value)) return "—";
  return `${Math.round(value)}%`;
}

/** Числовой id, если оба — целые; иначе строковое сравнение (как на бэке). */
export function compareContestId(a: string, b: string): number {
  if (/^\d+$/.test(a) && /^\d+$/.test(b)) {
    return Number(a) - Number(b);
  }
  return collator.compare(a, b);
}
