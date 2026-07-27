import { Scroll } from "lucide-react";
import { useCallback, type KeyboardEvent } from "react";
import ProblemStatementModal from "@/features/statements/ProblemStatementModal";

type Props = {
  contestId: string;
  contestName?: string;
  problemLabel?: string | null;
  statementAvailable: boolean;
  open: boolean;
  onOpen: () => void;
  onClose: () => void;
};

export default function ProblemStatementScrollZone({
  contestId,
  contestName,
  problemLabel,
  statementAvailable,
  open,
  onOpen,
  onClose,
}: Props) {
  const onKeyDown = useCallback(
    (e: KeyboardEvent<HTMLDivElement>) => {
      if (!statementAvailable) return;
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onOpen();
      }
    },
    [onOpen, statementAvailable],
  );

  const enabledLabel = "Условие задачи";
  const disabledLabel = "Невозможно получить условие";

  return (
    <>
      <div
        className={
          "review-side-rail__scroll review-side-rail__scroll--head" +
          (statementAvailable ? "" : " review-side-rail__scroll--head-disabled")
        }
        role="button"
        tabIndex={statementAvailable ? 0 : -1}
        aria-disabled={!statementAvailable}
        aria-label={statementAvailable ? enabledLabel : disabledLabel}
        onClick={() => {
          if (statementAvailable) onOpen();
        }}
        onKeyDown={onKeyDown}
      >
        <div className="review-scroll-head__bar">
          <span
            className="review-scroll-hint review-scroll-hint--statement"
            aria-hidden={!statementAvailable}
          >
            <Scroll size={28} strokeWidth={2} />
            <span className="review-scroll-hint__label">
              {statementAvailable ? enabledLabel : disabledLabel}
            </span>
          </span>
        </div>
        {statementAvailable ? (
          <div className="review-scroll-head__hover-fill" aria-hidden />
        ) : null}
      </div>
      {statementAvailable ? (
        <ProblemStatementModal
          open={open}
          contestId={contestId}
          title={contestName}
          problemLabel={problemLabel}
          onClose={onClose}
        />
      ) : null}
    </>
  );
}
