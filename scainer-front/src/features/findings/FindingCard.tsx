import { useMemo, useState, type ReactNode } from "react";
import type { EvidenceView, FindingView, ReportData, SubmissionView } from "@/client/types.gen";
import {
  PROBLEM_SUSPICION_TOOLTIP,
  problemSuspicionLevel,
} from "@/features/contests/problemSignalStats";
import {
  fmtScore,
  formatSignalCount,
  formatSubmissionCount,
  problemDisplay,
  scoreLevel,
  subjectTitle,
  uniqueDetectors,
  type GroupBy,
} from "@/features/findings/reportModel";
import SourceCode from "@/features/code/SourceCode";

type Props = {
  finding: FindingView;
  groupBy: GroupBy;
  groupKey: string;
  submissions: Record<string, SubmissionView>;
  hidden: boolean;
};

export function FindingCard({ finding, groupBy, groupKey, submissions, hidden }: Props) {
  const [open, setOpen] = useState(false);
  const [rendered, setRendered] = useState(false);
  const multiSignal = (finding.signals ?? []).length > 1;

  const onToggle = (next: boolean) => {
    setOpen(next);
    if (next && !rendered) setRendered(true);
  };

  return (
    <details
      className={`finding-card${hidden ? " hidden" : ""}`}
      open={open}
      onToggle={(e) => {
        if (e.target !== e.currentTarget) return;
        onToggle((e.currentTarget as HTMLDetailsElement).open);
      }}
      data-key={finding.key}
      data-finding-key={finding.key}
      data-score={String(finding.score)}
    >
      <summary className="finding-head">
        <span className={`score-badge ${scoreLevel(finding.score)}`}>{fmtScore(finding.score)}</span>
        <span className="finding-title">{subjectTitle(finding.subject, groupBy, groupKey)}</span>
        <span className="ai-slot">{finding.ai ? <AiBadge /> : null}</span>
        <span className="finding-meta">
          {groupBy !== "problem" ? <ProblemChip subject={finding.subject} /> : null}
        </span>
        <span className="detectors">
          {uniqueDetectors(finding.signals).map((d) => (
            <span key={d} className="chip detector-chip">
              {d}
            </span>
          ))}
        </span>
      </summary>
      <div className="finding-body">
        {rendered
          ? (finding.signals ?? []).map((sig, i) => (
              <SignalBlock key={`${sig.detector}-${i}`} signal={sig} submissions={submissions} showDetector={multiSignal} />
            ))
          : null}
      </div>
    </details>
  );
}

function AiBadge() {
  return (
    <span
      className="ai-badge"
      tabIndex={0}
      aria-label="Сигнал от AI-детектора"
      onClick={(e) => e.preventDefault()}
    >
      <span className="ai-badge-tip" role="tooltip">
        <span className="ai-badge-tip-label">AI-детектор</span>
        <span className="ai-badge-tip-text">
          В оценку вошёл сигнал LLM: модель смотрела историю посылок и отметила подозрение на использование ИИ.
        </span>
      </span>
    </span>
  );
}

function ProblemChip({ subject }: { subject: FindingView["subject"] }) {
  if (!subject.problem) return null;
  return (
    <Chip className="problem-chip" title={subject.problem}>
      {problemDisplay(subject.problem, subject.problem_name)}
    </Chip>
  );
}

function Chip({
  children,
  className,
  title,
}: {
  children: ReactNode;
  className?: string;
  title?: string;
}) {
  return (
    <span className={className ? `chip ${className}` : "chip"} title={title}>
      {children}
    </span>
  );
}

function SignalBlock({
  signal,
  submissions,
  showDetector,
}: {
  signal: FindingView["signals"][number];
  submissions: Record<string, SubmissionView>;
  showDetector: boolean;
}) {
  return (
    <div className="signal-block">
      {showDetector ? (
        <div className="signal-meta">
          <Chip>{signal.detector}</Chip>
          {signal.ai ? <AiBadge /> : null}
        </div>
      ) : null}
      {(signal.evidence ?? []).map((ev, i) => (
        <div key={i}>
          {ev.description ? <p className="evidence-text">{ev.description}</p> : null}
          <SideBySide evidence={ev} submissions={submissions} />
        </div>
      ))}
    </div>
  );
}

function SideBySide({
  evidence,
  submissions,
}: {
  evidence: EvidenceView;
  submissions: Record<string, SubmissionView>;
}) {
  const spans = evidence.spans ?? [];
  if (spans.length < 2) return null;
  const aID = spans[0].submission;
  const bID = spans[1].submission;
  const rangesA: Array<{ start: number; end: number }> = [];
  const rangesB: Array<{ start: number; end: number }> = [];
  for (const sp of spans) {
    const r = { start: sp.start_line, end: sp.end_line };
    if (sp.submission === aID) rangesA.push(r);
    else if (sp.submission === bID) rangesB.push(r);
  }
  return (
    <div className="side-by-side">
      <CodePane subID={aID} sub={submissions[aID]} ranges={rangesA} />
      <CodePane subID={bID} sub={submissions[bID]} ranges={rangesB} />
    </div>
  );
}

function CodePane({
  subID,
  sub,
  ranges,
}: {
  subID: string;
  sub?: SubmissionView;
  ranges: Array<{ start: number; end: number }>;
}) {
  const [moreOpen, setMoreOpen] = useState(false);
  const lines = sub?.source ?? ["(нет исходника)"];
  const inRange = (lineNo: number) => ranges.some((r) => lineNo >= r.start && lineNo <= r.end);

  return (
    <div className="code-pane">
      <div className="code-pane-head">
        <details
          className="code-pane-more"
          open={moreOpen}
          onToggle={(e) => {
            e.stopPropagation();
            setMoreOpen((e.currentTarget as HTMLDetailsElement).open);
          }}
        >
          <summary className="code-pane-summary">
            <span className={`code-pane-participant${!sub?.participant ? " code-pane-participant-missing" : ""}`}>
              {sub?.participant || "исходник недоступен"}
            </span>
            <span className="code-pane-toggle">{moreOpen ? "свернуть" : "подробнее"}</span>
          </summary>
          <div className="code-pane-meta">
            <Chip title={subID}>{subID}</Chip>
            {sub?.problem ? (
              <Chip className="problem-chip" title={sub.problem}>
                {problemDisplay(sub.problem, sub.problem_name)}
              </Chip>
            ) : null}
            {sub?.lang ? <Chip>{sub.lang}</Chip> : null}
          </div>
        </details>
      </div>
      <SourceCode
        lines={lines}
        lang={sub?.lang}
        matchLine={inRange}
      />
    </div>
  );
}

export function GroupTitle({
  groupBy,
  group,
  signalCount,
  submissionCount,
}: {
  groupBy: GroupBy;
  group: {
    key: string;
    contest: string;
    contestName: string;
    problem: string;
    problemName: string;
    visibleCount: number;
  };
  /** Findings задачи со score ≥ порога (для groupBy=problem). */
  signalCount?: number;
  submissionCount?: number;
}) {
  const sharePercent =
    groupBy === "problem" && submissionCount != null && submissionCount > 0
      ? (100 * (signalCount ?? 0)) / submissionCount
      : null;
  const heat = sharePercent != null ? problemSuspicionLevel(sharePercent) : null;

  return (
    <h2 className="group-title">
      {groupBy === "problem" ? (
        <>
          <Chip className="problem-chip" title={group.problem || group.key}>
            {problemDisplay(group.problem || group.key, group.problemName)}
          </Chip>
        </>
      ) : (
        <code>{group.key}</code>
      )}
      {groupBy === "problem" && submissionCount != null ? (
        <span
          className={
            "group-stats" +
            (heat === "level-high"
              ? " group-stats--high"
              : heat === "level-low"
                ? " group-stats--low"
                : "")
          }
          title={PROBLEM_SUSPICION_TOOLTIP}
        >
          {formatSignalCount(signalCount ?? 0)} / {formatSubmissionCount(submissionCount)}
        </span>
      ) : null}
    </h2>
  );
}

export function ReportMeta({ note }: { note?: string }) {
  return <p className="meta">{note || ""}</p>;
}

export function Logo() {
  return (
    <h1 className="logo" aria-label="scAIner">
      sc<span className="logo-ai">AI</span>ner
    </h1>
  );
}

export function useFilteredVisibility(
  findings: FindingView[],
  threshold: number
) {
  return useMemo(() => {
    const visible = new Set<string>();
    for (const f of findings) {
      if (f.score >= threshold) visible.add(f.key);
    }
    return visible;
  }, [findings, threshold]);
}

export type { ReportData };
