import { Sparkles } from "lucide-react";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState, type ReactNode } from "react";
import type { EvidenceView, FindingView, ReportData, SubmissionView } from "@/client/types.gen";
import {
  PROBLEM_SUSPICION_TOOLTIP,
  problemSuspicionLevel,
} from "@/features/contests/shared/problemSignalStats";
import {
  formatSignalCount,
  formatSubmissionCount,
  problemDisplay,
  subjectTitle,
  uniqueDetectors,
  type GroupBy,
} from "@/features/findings/reportModel";
import { DetectorChip } from "@/features/findings/DetectorChip";
import SourceCode from "@/features/code/SourceCode";
import { CodePaneMoreBar } from "@/features/code/CodePaneMoreBar";
import { EjudgeContestChip } from "@/features/ejudge/EjudgeContestChip";

type CardSharedProps = {
  finding: FindingView;
  groupBy: GroupBy;
  groupKey: string;
};

type HeadProps = CardSharedProps & {
  open: boolean;
  bodyId: string;
  onToggle: (open: boolean) => void;
  onMeasure?: (height: number) => void;
};

export function FindingCardHead({
  finding,
  groupBy,
  groupKey,
  open,
  bodyId,
  onToggle,
  onMeasure,
}: HeadProps) {
  const cardRef = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    if (!cardRef.current || !onMeasure) return;
    onMeasure(cardRef.current.getBoundingClientRect().height);
  });

  return (
    <div ref={cardRef} className="finding-card" data-score={String(finding.score)}>
      <button
        type="button"
        className="finding-head"
        aria-expanded={open}
        aria-controls={bodyId}
        data-key={finding.key}
        data-finding-key={finding.key}
        onClick={() => onToggle(!open)}
      >
        <span className="detectors">
          {uniqueDetectors(finding.signals).map((d) => (
            <DetectorChip key={d} detectorId={d} score={finding.score} />
          ))}
        </span>
        <span className="finding-head__main">
          <span className="finding-title">{subjectTitle(finding.subject, groupBy, groupKey)}</span>
          <span className="ai-slot">{finding.ai ? <AiBadge /> : null}</span>
          <span className="finding-meta">
            {groupBy !== "problem" ? <ProblemChip subject={finding.subject} /> : null}
          </span>
        </span>
      </button>
    </div>
  );
}

type GhostProps = CardSharedProps & {
  onCollapse: () => void;
  minHeight?: number;
};

export function FindingCardGhost({
  finding,
  groupBy,
  groupKey,
  onCollapse,
  minHeight,
}: GhostProps) {
  return (
    <div
      className="finding-card finding-card--ghost-slot"
      style={minHeight ? { minHeight: `${minHeight}px` } : undefined}
    >
      <button
        type="button"
        className="finding-card-ghost"
        aria-label="Свернуть находку"
        data-key={finding.key}
        data-finding-key={finding.key}
        onClick={onCollapse}
      >
        <span className="finding-card-ghost__title">
          {subjectTitle(finding.subject, groupBy, groupKey)}
        </span>
        <span className="finding-card-ghost__action">свернуть</span>
      </button>
    </div>
  );
}

type BodyProps = CardSharedProps & {
  submissions: Record<string, SubmissionView>;
  rendered: boolean;
};

export function FindingCardBody({ finding, submissions, rendered }: BodyProps) {
  const multiSignal = (finding.signals ?? []).length > 1;

  if (!rendered) return null;

  return (
    <div className="finding-body">
      {(finding.signals ?? []).map((sig, i) => (
        <SignalBlock
          key={`${sig.detector}-${i}`}
          signal={sig}
          submissions={submissions}
          showDetector={multiSignal}
          subjectSubmission={finding.subject.submission ?? undefined}
          findingScore={finding.score}
        />
      ))}
    </div>
  );
}

type RowProps = {
  pair: FindingView[];
  groupBy: GroupBy;
  groupKey: string;
  submissions: Record<string, SubmissionView>;
  deepLinkKey?: string | null;
};

export function FindingRow({
  pair,
  groupBy,
  groupKey,
  submissions,
  deepLinkKey,
}: RowProps) {
  const deepLinkInRow = deepLinkKey != null && pair.some((f) => f.key === deepLinkKey);
  const [openKey, setOpenKey] = useState<string | null>(deepLinkInRow ? deepLinkKey : null);
  const [renderedKeys, setRenderedKeys] = useState<Set<string>>(() =>
    deepLinkInRow && deepLinkKey ? new Set([deepLinkKey]) : new Set(),
  );
  const [slotHeights, setSlotHeights] = useState<Record<string, number>>({});

  const onMeasureHead = useCallback((key: string, height: number) => {
    setSlotHeights((prev) => (prev[key] === height ? prev : { ...prev, [key]: height }));
  }, []);

  useEffect(() => {
    if (deepLinkKey && pair.some((f) => f.key === deepLinkKey)) {
      setOpenKey(deepLinkKey);
      setRenderedKeys((prev) => new Set(prev).add(deepLinkKey));
    }
  }, [deepLinkKey, pair]);

  const openFinding = openKey ? pair.find((f) => f.key === openKey) : undefined;
  const openSlot = openFinding ? pair.indexOf(openFinding) : null;

  const onToggle = (key: string, next: boolean) => {
    setOpenKey(next ? key : null);
    if (next) {
      setRenderedKeys((prev) => new Set(prev).add(key));
    }
  };

  const slot0 = pair[0];
  const slot1 = pair[1];
  const bodyId = openFinding ? `finding-body-${openFinding.key}` : undefined;

  const renderSlot = (finding: FindingView) => {
    const isOpen = openKey === finding.key;
    const id = bodyId ?? `finding-body-${finding.key}`;
    if (isOpen) {
      return (
        <FindingCardGhost
          finding={finding}
          groupBy={groupBy}
          groupKey={groupKey}
          minHeight={slotHeights[finding.key]}
          onCollapse={() => onToggle(finding.key, false)}
        />
      );
    }
    return (
      <FindingCardHead
        finding={finding}
        groupBy={groupBy}
        groupKey={groupKey}
        open={false}
        bodyId={id}
        onMeasure={(height) => onMeasureHead(finding.key, height)}
        onToggle={(next) => onToggle(finding.key, next)}
      />
    );
  };

  return (
    <div
      className={
        "finding-row" + (openSlot !== null ? ` finding-row--open-slot-${openSlot}` : "")
      }
    >
      <div className="finding-row__heads">
        {slot0 ? renderSlot(slot0) : null}
        {slot1 ? (
          renderSlot(slot1)
        ) : (
          <div className="finding-row__spacer" aria-hidden />
        )}
      </div>
      {openFinding && bodyId ? (
        <div className="finding-row__expand" id={bodyId}>
          <FindingCardBody
            finding={openFinding}
            groupBy={groupBy}
            groupKey={groupKey}
            submissions={submissions}
            rendered={renderedKeys.has(openFinding.key)}
          />
        </div>
      ) : null}
    </div>
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
      <Sparkles className="ai-badge__icon" size={16} strokeWidth={2} aria-hidden />
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
    <EjudgeContestChip className="problem-chip" title={subject.problem}>
      {problemDisplay(subject.problem, subject.problem_name)}
    </EjudgeContestChip>
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
  subjectSubmission,
  findingScore,
}: {
  signal: FindingView["signals"][number];
  submissions: Record<string, SubmissionView>;
  showDetector: boolean;
  subjectSubmission?: string;
  findingScore?: number;
}) {
  return (
    <div className="signal-block">
      {showDetector ? (
        <div className="signal-meta">
          <DetectorChip detectorId={signal.detector} score={findingScore} />
          {signal.ai ? <AiBadge /> : null}
        </div>
      ) : null}
      {(signal.evidence ?? []).map((ev, i) => (
        <div key={i}>
          {ev.description ? <p className="evidence-text">{ev.description}</p> : null}
          <SideBySide
            evidence={ev}
            submissions={submissions}
            fallbackSubmissionId={subjectSubmission}
          />
        </div>
      ))}
    </div>
  );
}

function SideBySide({
  evidence,
  submissions,
  fallbackSubmissionId,
}: {
  evidence: EvidenceView;
  submissions: Record<string, SubmissionView>;
  fallbackSubmissionId?: string;
}) {
  const spans = evidence.spans ?? [];
  if (spans.length === 0) {
    if (!fallbackSubmissionId) return null;
    return (
      <div className="side-by-side side-by-side--single">
        <CodePane
          subID={fallbackSubmissionId}
          sub={submissions[fallbackSubmissionId]}
          ranges={[]}
        />
      </div>
    );
  }

  if (spans.length === 1) {
    const id = spans[0].submission;
    return (
      <div className="side-by-side side-by-side--single">
        <CodePane subID={id} sub={submissions[id]} ranges={[]} />
      </div>
    );
  }

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
  const lines = sub?.source ?? ["(нет исходника)"];
  const inRange = (lineNo: number) => ranges.some((r) => lineNo >= r.start && lineNo <= r.end);

  return (
    <div className="code-pane">
      <div className="code-pane-head">
        <CodePaneMoreBar
          key={subID}
          participant={sub?.participant || "исходник недоступен"}
          participantClassName={`code-pane-participant${!sub?.participant ? " code-pane-participant-missing" : ""}`}
        >
          <EjudgeContestChip title={subID}>{subID}</EjudgeContestChip>
          {sub?.problem ? (
            <Chip className="problem-chip" title={sub.problem}>
              {problemDisplay(sub.problem, sub.problem_name)}
            </Chip>
          ) : null}
          {sub?.lang ? <Chip>{sub.lang}</Chip> : null}
        </CodePaneMoreBar>
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
          <EjudgeContestChip
            className="problem-chip"
            title={group.problem || group.key}
          >
            {problemDisplay(group.problem || group.key, group.problemName)}
          </EjudgeContestChip>
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
