import { FileText } from "lucide-react";
import { useCallback, type KeyboardEvent, type Ref } from "react";
import ProblemStatementModal from "@/features/statements/ProblemStatementModal";

type Props = {
  contestId: string;
  contestName?: string;
  problemLabel?: string | null;
  statementAvailable: boolean;
  open: boolean;
  onOpen: () => void;
  onClose: () => void;
  headRef?: Ref<HTMLDivElement>;
};

export default function ProblemStatementScrollZone({
  contestId,
  contestName,
  problemLabel,
  statementAvailable,
  open,
  onOpen,
  onClose,
  headRef,
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

  const enabledLabel = "Условие";
  const disabledLabel = "Невозможно получить условие";

  return (
    <>
      <div
        ref={headRef}
        className={
          "review-side-rail__head review-side-rail__scroll--head" +
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
        <span
          className="review-scroll-hint review-scroll-hint--statement"
          aria-hidden={!statementAvailable}
        >
          <FileText size={28} strokeWidth={2} />
          <span className="review-scroll-hint__label">
            {statementAvailable ? enabledLabel : disabledLabel}
          </span>
        </span>
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
