export const DEFAULT_PARALLEL_IDS = [
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  "10",
  "R",
  "F",
  "X",
] as const;

export function parallelLabel(id: string, ungroupedId: string): string {
  if (id === ungroupedId) return "Без параллели";
  return `Параллель ${id}`;
}

export type ParsedImportRow = {
  id: string;
  name: string;
  parallelId: string;
};

export type SkippedImportRow = {
  id: string;
  name: string;
  reason: string;
};

export type ImportParseResult = {
  rows: ParsedImportRow[];
  skipped: SkippedImportRow[];
};

const PARALLEL_PATTERN = /Параллель\s+(10|[3-9]|X|F|R)\b/i;

// parseEjudgeContestTable разбирает вставленную табличку списка контестов ejudge
// (строки вида "<row>\t<id>\t<name>\t...", TSV из копипасты со страницы списка контестов).
export function parseEjudgeContestTable(raw: string): ImportParseResult {
  const rows: ParsedImportRow[] = [];
  const skipped: SkippedImportRow[] = [];

  for (const line of raw.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed) continue;

    const cols = trimmed.split("\t");
    const id = (cols[1] ?? "").trim();
    const name = (cols[2] ?? "").trim();
    if (!id || !name) continue;

    if (/template/i.test(name)) {
      skipped.push({ id, name, reason: "шаблон" });
      continue;
    }

    const match = name.match(PARALLEL_PATTERN);
    if (!match) {
      skipped.push({ id, name, reason: "параллель не распознана" });
      continue;
    }

    const key = /^[0-9]+$/.test(match[1]) ? match[1] : match[1].toUpperCase();
    rows.push({ id, name, parallelId: key });
  }

  return { rows, skipped };
}
