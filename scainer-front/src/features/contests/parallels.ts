export type Parallel = { id: string; name: string };

// Дефолтный набор параллелей ЛКШ — скелет на странице параллелей
// и сопоставление при массовом импорте из вставленной таблицы ejudge.
export const DEFAULT_PARALLELS: Parallel[] = [
  { id: "2", name: "Параллель 2" },
  { id: "3", name: "Параллель 3" },
  { id: "4", name: "Параллель 4" },
  { id: "5", name: "Параллель 5" },
  { id: "6", name: "Параллель 6" },
  { id: "7", name: "Параллель 7" },
  { id: "8", name: "Параллель 8" },
  { id: "9", name: "Параллель 9" },
  { id: "X", name: "Параллель X" },
  { id: "F", name: "Параллель F" },
  { id: "R", name: "Параллель R" },
];

export type ParsedImportRow = {
  id: string;
  name: string;
  parallelId: string;
  parallelName: string;
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

const PARALLEL_PATTERN = /Параллель\s+([2-9]|X|F|R)\b/i;

// parseEjudgeContestTable разбирает вставленную табличку списка контестов ejudge
// (строки вида "<row>\t<id>\t<name>\t...", TSV из копипасты со страницы списка контестов).
// Пропускает шаблоны ("...Template" в имени) и строки, где не удалось распознать параллель
// по паттерну "Параллель <2-9|X|F|R>" — но возвращает их отдельно (с причиной), чтобы учитель
// видел в превью, что именно отфильтровано, а не только итоговый счётчик.
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

    const parallelKey = match[1].toUpperCase();
    const parallel = DEFAULT_PARALLELS.find((p) => p.id === parallelKey);
    rows.push({
      id,
      name,
      parallelId: parallelKey,
      parallelName: parallel?.name ?? `Параллель ${parallelKey}`,
    });
  }

  return { rows, skipped };
}
