import type { FindingView, ReportData } from "@/client/types.gen";

export function submissionPanelId(submissionId: string): string {
  return `review-sub-${submissionId.replace(/[^a-zA-Z0-9_-]/g, "-")}`;
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

function findingsForSubmission(
  report: ReportData | undefined,
  submissionId: string,
): FindingView[] {
  if (!report?.findings) return [];
  return report.findings.filter((f) => findingTouchesSubmission(f, submissionId));
}

export function findingsForSubmissionVisible(
  report: ReportData | undefined,
  submissionId: string,
  threshold: number,
): FindingView[] {
  return findingsForSubmission(report, submissionId).filter((f) => f.score >= threshold);
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
