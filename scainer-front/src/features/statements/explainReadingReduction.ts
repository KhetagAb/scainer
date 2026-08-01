const PROMPT_RAW_START = "===== НАЧАЛО ТЕКСТА ДЛЯ РЕДАКТИРОВАНИЯ =====";
const PROMPT_RAW_END = "===== КОНЕЦ ТЕКСТА ДЛЯ РЕДАКТИРОВАНИЯ =====";
const OUTPUT_FORMAT_HEADER_RE = /формат\s*выходных\s*данных/i;

export function trimStatementForExplain(text: string): string {
  const trimmed = text.trim();
  const m = OUTPUT_FORMAT_HEADER_RE.exec(trimmed);
  if (!m) return trimmed;
  return trimmed.slice(0, m.index).trim();
}

export type ReadingReduction = {
  rawChars: number;
  formalChars: number;
  savedChars: number;
  percentLess: number;
};

export function extractRawStatementFromPrompt(prompt: string): string | null {
  const start = prompt.indexOf(PROMPT_RAW_START);
  if (start < 0) return null;
  const contentStart = start + PROMPT_RAW_START.length;
  const end = prompt.indexOf(PROMPT_RAW_END, contentStart);
  if (end < 0) return null;
  const raw = prompt.slice(contentStart, end).trim();
  return raw || null;
}

/** Убирает markdown-разметку для сравнения объёма читаемого текста. */
export function stripMarkdownForMeasure(text: string): string {
  return text
    .replace(/```[\s\S]*?```/g, " ")
    .replace(/`[^`]*`/g, " ")
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/\*([^*]+)\*/g, "$1")
    .replace(/__([^_]+)__/g, "$1")
    .replace(/_([^_]+)_/g, "$1")
    .replace(/^#{1,6}\s+/gm, "")
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1")
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, "$1")
    .replace(/^\s*[-*+]\s+/gm, "")
    .replace(/\s+/g, " ")
    .trim();
}

export function readableCharCount(text: string): number {
  return stripMarkdownForMeasure(text).length;
}

export function computeReadingReduction(
  rawStatement: string,
  formalStatement: string,
): ReadingReduction | null {
  const rawChars = readableCharCount(rawStatement);
  const formalChars = readableCharCount(formalStatement);
  if (rawChars === 0) return null;

  const savedChars = rawChars - formalChars;
  const percentLess = Math.round((savedChars / rawChars) * 100);

  return { rawChars, formalChars, savedChars, percentLess };
}

export function readingReductionChipLabel(reduction: ReadingReduction): string {
  if (reduction.percentLess > 0) {
    return `−${reduction.percentLess}% текста`;
  }
  if (reduction.percentLess < 0) {
    return `+${Math.abs(reduction.percentLess)}% текста`;
  }
  return "без сокращения";
}

export function readingReductionChipTitle(reduction: ReadingReduction): string {
  return `Формализация: ${reduction.formalChars} симв. · оригинал: ${reduction.rawChars} симв.`;
}
