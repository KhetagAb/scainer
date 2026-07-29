import type * as pdfjs from "pdfjs-dist";

/** Заголовок условия в PDF ЛКШ: «Задача A. …» */
export const PROBLEM_HEADER_RE = /Задача\s+([A-Za-zА-Яа-яЁё0-9]+)\s*\./giu;

export type ProblemSection = {
  label: string;
  startPage: number;
};

export function normalizeProblemLabel(label: string): string {
  return label.trim().toLocaleUpperCase("ru");
}

export async function extractPdfPageText(page: pdfjs.PDFPageProxy): Promise<string> {
  const content = await page.getTextContent();
  return content.items
    .map((item) => ("str" in item ? item.str : ""))
    .join(" ")
    .replace(/\s+/g, " ")
    .trim();
}

export function findProblemSections(pageTexts: readonly string[]): ProblemSection[] {
  const sections: ProblemSection[] = [];
  pageTexts.forEach((text, index) => {
    const page = index + 1;
    const re = new RegExp(PROBLEM_HEADER_RE.source, PROBLEM_HEADER_RE.flags);
    const seenOnPage = new Set<string>();
    let match: RegExpExecArray | null;
    while ((match = re.exec(text)) !== null) {
      const label = normalizeProblemLabel(match[1]);
      if (seenOnPage.has(label)) continue;
      seenOnPage.add(label);
      sections.push({ label, startPage: page });
    }
  });
  return sections;
}

/** null — не удалось сопоставить задачу, показываем весь PDF. */
export function pagesForProblem(
  numPages: number,
  sections: readonly ProblemSection[],
  problemLabel: string | null | undefined,
): number[] | null {
  const trimmed = problemLabel?.trim();
  if (!trimmed) return null;

  const key = normalizeProblemLabel(trimmed);
  let idx = sections.findIndex((section) => section.label === key);
  if (idx < 0) {
    const dashIdx = trimmed.indexOf(" - ");
    if (dashIdx > 0) {
      const shortKey = normalizeProblemLabel(trimmed.slice(0, dashIdx));
      idx = sections.findIndex((section) => section.label === shortKey);
    }
  }
  if (idx < 0) return null;

  const start = sections[idx].startPage;
  const end =
    idx + 1 < sections.length ? sections[idx + 1].startPage - 1 : numPages;
  if (start > end || start < 1 || end > numPages) return null;

  return Array.from({ length: end - start + 1 }, (_, i) => start + i);
}

export async function resolveStatementPages(
  pdf: pdfjs.PDFDocumentProxy,
  problemLabel: string | null | undefined,
): Promise<number[]> {
  const allPages = Array.from({ length: pdf.numPages }, (_, i) => i + 1);
  if (!problemLabel?.trim()) return allPages;

  try {
    const pageTexts: string[] = [];
    for (let pageNumber = 1; pageNumber <= pdf.numPages; pageNumber++) {
      const page = await pdf.getPage(pageNumber);
      pageTexts.push(await extractPdfPageText(page));
    }
    const sections = findProblemSections(pageTexts);
    const matched = pagesForProblem(pdf.numPages, sections, problemLabel);
    return matched ?? allPages;
  } catch {
    return allPages;
  }
}
