export type CiteRange = { from: number; to: number };

export function citeLinesLabel(from: number, to: number): string {
  return from === to ? `Строка ${from}: ` : `Строки ${from}–${to}: `;
}

function addUnique(nums: number[], n: number): void {
  if (!nums.includes(n)) nums.push(n);
}

function parseCiteNumbers(lineBeforeCaret: string): number[] | null {
  const m = lineBeforeCaret.match(/^(?:Строка|Строки) (.+): $/);
  if (!m) return null;

  const nums: number[] = [];
  for (const part of m[1].split(", ")) {
    const dash = part.indexOf("–");
    if (dash > 0) {
      const a = Number(part.slice(0, dash));
      const b = Number(part.slice(dash + 1));
      if (Number.isNaN(a) || Number.isNaN(b)) return null;
      for (let n = Math.min(a, b); n <= Math.max(a, b); n++) addUnique(nums, n);
    } else {
      const n = Number(part);
      if (Number.isNaN(n)) return null;
      addUnique(nums, n);
    }
  }

  return nums.length ? nums : null;
}

function addLineNumbers(nums: number[], from: number, to: number): void {
  const lo = Math.min(from, to);
  const hi = Math.max(from, to);
  for (let n = lo; n <= hi; n++) addUnique(nums, n);
}

function compressLineNumbers(nums: number[]): Array<number | CiteRange> {
  if (!nums.length) return [];

  const sorted = [...nums].sort((a, b) => a - b);
  const parts: Array<number | CiteRange> = [];
  let start = sorted[0];
  let prev = sorted[0];

  for (let i = 1; i < sorted.length; i++) {
    const n = sorted[i];
    if (n === prev + 1) {
      prev = n;
      continue;
    }
    if (start === prev) parts.push(start);
    else parts.push({ from: start, to: prev });
    start = prev = n;
  }

  if (start === prev) parts.push(start);
  else parts.push({ from: start, to: prev });

  return parts;
}

function formatCompressed(parts: Array<number | CiteRange>): string {
  if (!parts.length) return "";

  if (parts.length === 1) {
    const p = parts[0];
    if (typeof p === "number") return `Строка ${p}: `;
    if (p.from === p.to) return `Строка ${p.from}: `;
    return `Строки ${p.from}–${p.to}: `;
  }

  const allSingles = parts.every((p) => typeof p === "number");
  if (allSingles) {
    return `Строка ${(parts as number[]).join(", ")}: `;
  }

  const chunks = parts.map((p) => {
    if (typeof p === "number") return String(p);
    if (p.from === p.to) return String(p.from);
    return `${p.from}–${p.to}`;
  });
  return `Строки ${chunks.join(", ")}: `;
}

function lineBounds(value: string, index: number): { start: number; end: number } {
  const start = value.lastIndexOf("\n", index - 1) + 1;
  const endIdx = value.indexOf("\n", index);
  const end = endIdx === -1 ? value.length : endIdx;
  return { start, end };
}

export function tryMergeCiteAt(
  value: string,
  caret: number,
  from: number,
  to: number,
): { newValue: string; caret: number } | null {
  const { start: lineStart, end: lineEnd } = lineBounds(value, caret);
  const lineText = value.slice(lineStart, lineEnd);
  const caretInLine = caret - lineStart;
  const beforeCaret = lineText.slice(0, caretInLine);
  const afterCaret = lineText.slice(caretInLine);

  if (afterCaret.trim() !== "") return null;

  const existing = parseCiteNumbers(beforeCaret);
  if (!existing) return null;

  addLineNumbers(existing, from, to);
  const newLine = formatCompressed(compressLineNumbers(existing));
  const newValue = value.slice(0, lineStart) + newLine + value.slice(lineEnd);
  const newCaret = lineStart + newLine.length;
  return { newValue, caret: newCaret };
}
