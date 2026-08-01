import { Pencil } from "lucide-react";
import type { ProblemStatementExplainView } from "@/client/types.gen";
import {
  computeReadingReduction,
  readingReductionChipLabel,
  readingReductionChipTitle,
  trimStatementForExplain,
} from "@/features/statements/explainReadingReduction";
import { useExplainRawStatement } from "@/features/statements/useExplainRawStatement";

function ReadingReductionChip({
  contestId,
  problemId,
  prompt,
  formalStatement,
}: {
  contestId: string;
  problemId: string;
  prompt: string | undefined;
  formalStatement: string;
}) {
  const rawStatement = useExplainRawStatement(contestId, problemId, prompt);
  const reduction = rawStatement
    ? computeReadingReduction(trimStatementForExplain(rawStatement), formalStatement)
    : null;

  if (!reduction) return null;

  return (
    <span
      className={
        "chip problem-statement-explain__reduction-chip" +
        (reduction.percentLess > 0 ? " problem-statement-explain__reduction-chip--saved" : "") +
        (reduction.percentLess < 0 ? " problem-statement-explain__reduction-chip--longer" : "")
      }
      title={readingReductionChipTitle(reduction)}
    >
      {readingReductionChipLabel(reduction)}
    </span>
  );
}

type Props = {
  contestId: string;
  problemId: string;
  data: ProblemStatementExplainView;
  editing: boolean;
  editPending: boolean;
  onEdit: () => void;
};

export default function ExplainPanelHeader({
  contestId,
  problemId,
  data,
  editing,
  editPending,
  onEdit,
}: Props) {
  return (
    <div className="problem-statement-explain__head">
      <div className="problem-statement-explain__chips">
        {data.source === "ai" ? (
          <span className="chip problem-statement-explain__source-chip problem-statement-explain__source-chip--ai">
            Сформировано AI
          </span>
        ) : null}
        <ReadingReductionChip
          contestId={contestId}
          problemId={problemId}
          prompt={data.prompt}
          formalStatement={data.statement}
        />
      </div>
      {!editing ? (
        <button
          type="button"
          className={
            "problem-statement-explain__edit-btn" +
            (editPending ? " problem-statement-explain__edit-btn--pending" : "")
          }
          aria-label="Редактировать формализацию"
          aria-hidden={editPending}
          tabIndex={editPending ? -1 : 0}
          disabled={editPending}
          onClick={onEdit}
        >
          <Pencil size={16} strokeWidth={2} aria-hidden />
        </button>
      ) : null}
    </div>
  );
}
