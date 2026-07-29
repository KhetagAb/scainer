import type { ReactNode } from "react";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import SensitivitySlider from "@/features/contests/ui/SensitivitySlider";
import ReviewPrOnlyFilter from "@/features/review/ReviewPrOnlyFilter";

type Props = {
  sync: ReactNode;
  submissionCount?: number;
  showSensitivity?: boolean;
  prOnly?: boolean;
  onPrOnlyChange?: (value: boolean) => void;
};

export default function ContestHeadToolbar({
  sync,
  submissionCount,
  showSensitivity = true,
  prOnly,
  onPrOnlyChange,
}: Props) {
  const { threshold, setThreshold } = useSensitivity();
  const hasStat = submissionCount != null;
  const showPrOnly = prOnly != null && onPrOnlyChange != null;
  const hasMeta = showPrOnly || hasStat;

  return (
    <div
      className={
        "contest-head__actions" +
        (showSensitivity ? "" : " contest-head__actions--compact")
      }
    >
      {hasMeta ? (
        <div className="contest-head__meta">
          {showPrOnly ? (
            <ReviewPrOnlyFilter checked={prOnly} onChange={onPrOnlyChange} />
          ) : null}
          {hasStat ? (
            <span className="contest-head__stat">{submissionCount} посылок</span>
          ) : null}
        </div>
      ) : null}
      <div className="contest-head__tools">
        <div className="contest-head__job-actions">{sync}</div>
        {showSensitivity ? (
          <>
            <span className="contest-head__sep" aria-hidden>
              |
            </span>
            <SensitivitySlider threshold={threshold} onChange={setThreshold} />
          </>
        ) : null}
      </div>
    </div>
  );
}
