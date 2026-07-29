import type { ReactNode } from "react";
import { useSensitivity } from "@/features/contests/shared/SensitivityContext";
import SensitivitySlider from "@/features/contests/ui/SensitivitySlider";
import ReviewParticipantFilter from "@/features/review/ReviewParticipantFilter";
import ReviewVerdictFilterChip from "@/features/review/ReviewVerdictFilterChip";
import type { ReviewVerdictFilter } from "@/features/review/reviewFilterUtils";

type Props = {
  sync: ReactNode;
  showSensitivity?: boolean;
  verdictFilter?: ReviewVerdictFilter;
  onVerdictFilterChange?: (value: ReviewVerdictFilter) => void;
  participantQuery?: string;
  onParticipantQueryChange?: (value: string) => void;
  hideParticipantFilter?: boolean;
};

export default function ContestHeadToolbar({
  sync,
  showSensitivity = true,
  verdictFilter,
  onVerdictFilterChange,
  participantQuery,
  onParticipantQueryChange,
  hideParticipantFilter = false,
}: Props) {
  const { threshold, setThreshold } = useSensitivity();
  const showReviewFilters =
    verdictFilter != null &&
    onVerdictFilterChange != null &&
    participantQuery != null &&
    onParticipantQueryChange != null;

  return (
    <div
      className={
        "contest-head__actions" +
        (showSensitivity ? "" : " contest-head__actions--compact")
      }
    >
      {showReviewFilters ? (
        <div className="contest-head__meta">
          <div className="contest-head__verdict-filter">
            <span className="contest-head__filter-label">Статус</span>
            <ReviewVerdictFilterChip
              value={verdictFilter}
              onChange={onVerdictFilterChange}
            />
          </div>
          {!hideParticipantFilter ? (
            <ReviewParticipantFilter
              value={participantQuery}
              onChange={onParticipantQueryChange}
            />
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
