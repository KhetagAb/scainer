import type { FindingView, ReportData, SubmissionListItem } from "@/client/types.gen";
import { problemDisplay } from "@/features/findings/reportModel";

export function prCountByProblem(items: SubmissionListItem[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const item of items) {
    if (item.verdict !== "PR") continue;
    map.set(item.problem, (map.get(item.problem) ?? 0) + 1);
  }
  return map;
}

/** Первая задача (в порядке чипов) с числом PR > 0. */
export function firstProblemWithPr(
  problems: { id: string; name?: string | null }[],
  items: SubmissionListItem[],
): string | null {
  const counts = prCountByProblem(items);
  const sorted = problems.slice().sort((a, b) => {
    const la = problemDisplay(a.id, a.name);
    const lb = problemDisplay(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });
  for (const p of sorted) {
    if ((counts.get(p.id) ?? 0) > 0) return p.id;
  }
  for (const [id, n] of counts) {
    if (n > 0) return id;
  }
  return null;
}

/** Следующая задача с PR после current (порядок как у чипов). */
export function nextProblemWithPr(
  problems: { id: string; name?: string | null }[],
  items: SubmissionListItem[],
  currentProblemId: string,
): string | null {
  const counts = prCountByProblem(items);
  const sorted = problems.slice().sort((a, b) => {
    const la = problemDisplay(a.id, a.name);
    const lb = problemDisplay(b.id, b.name);
    return la < lb ? -1 : la > lb ? 1 : 0;
  });
  const idx = sorted.findIndex((p) => p.id === currentProblemId);
  const start = idx >= 0 ? idx + 1 : 0;
  for (let i = start; i < sorted.length; i++) {
    if ((counts.get(sorted[i].id) ?? 0) > 0) return sorted[i].id;
  }
  return null;
}

export function prQueueForProblem(
  items: SubmissionListItem[],
  problemId: string,
): SubmissionListItem[] {
  return items
    .filter((s) => s.problem === problemId && s.verdict === "PR")
    .slice()
    .sort((a, b) => a.submitted_at.localeCompare(b.submitted_at));
}

/** Следующая PR после currentId в очереди (current исключается). */
export function nextPrInQueue(
  queue: SubmissionListItem[],
  currentId: string,
): SubmissionListItem | undefined {
  const idx = queue.findIndex((s) => s.id === currentId);
  if (idx >= 0) {
    const after = queue.slice(idx + 1).find((s) => s.id !== currentId);
    if (after) return after;
  }
  return queue.find((s) => s.id !== currentId);
}

export function submissionPanelId(submissionId: string): string {
  return `review-sub-${submissionId.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
}

export function submissionIdsInFindings(report: ReportData | undefined): Set<string> {
  const ids = new Set<string>();
  if (!report) return ids;
  for (const id of Object.keys(report.submissions ?? {})) {
    ids.add(id);
  }
  for (const finding of report.findings ?? []) {
    if (finding.subject?.submission) ids.add(finding.subject.submission);
    for (const signal of finding.signals ?? []) {
      for (const evidence of signal.evidence ?? []) {
        for (const span of evidence.spans ?? []) {
          if (span.submission) ids.add(span.submission);
        }
      }
    }
  }
  return ids;
}

function findingTouchesSubmission(finding: FindingView, submissionId: string): boolean {
  if (finding.subject?.submission === submissionId) return true;
  for (const signal of finding.signals ?? []) {
    for (const evidence of signal.evidence ?? []) {
      for (const span of evidence.spans ?? []) {
        if (span.submission === submissionId) return true;
      }
    }
  }
  return false;
}

export function findingsForSubmission(
  report: ReportData | undefined,
  submissionId: string,
): FindingView[] {
  if (!report?.findings) return [];
  return report.findings.filter((f) => findingTouchesSubmission(f, submissionId));
}

export function findingsForProblem(
  report: ReportData | undefined,
  problemId: string,
  submissionItems: SubmissionListItem[] = [],
): FindingView[] {
  if (!report?.findings) return [];
  const problemSubIds = new Set<string>();
  for (const s of submissionItems) {
    if (s.problem === problemId) problemSubIds.add(s.id);
  }
  for (const [id, sub] of Object.entries(report.submissions ?? {})) {
    if (sub.problem === problemId) problemSubIds.add(id);
  }
  return report.findings.filter((f) => {
    if (f.subject?.problem === problemId) return true;
    for (const id of problemSubIds) {
      if (findingTouchesSubmission(f, id)) return true;
    }
    return false;
  });
}

export function topFindingKey(findings: FindingView[]): string | undefined {
  if (!findings.length) return undefined;
  let best = findings[0];
  for (const f of findings) {
    if (f.score > best.score) best = f;
  }
  return best.key;
}

export function formatSubmittedAt(value: string): string {
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
