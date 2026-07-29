import { useEffect, useRef, useState } from "react";
import type { SubmissionListItem } from "@/client/types.gen";
import {
  formatSubmittedAt,
  submissionPanelId,
} from "@/features/review/reviewFindings";
import {
  formatVerdictLabel,
  verdictChipTone,
} from "@/features/review/reviewVerdicts";
import ReviewSubmissionPanel from "@/features/review/ReviewSubmissionPanel";
import type { ComponentProps } from "react";

type PanelProps = ComponentProps<typeof ReviewSubmissionPanel>;

type Props = PanelProps & {
  index: number;
};

const EAGER_PANEL_COUNT = 2;
const LAZY_ROOT_MARGIN = "900px 0px";

function ReviewSubmissionPanelPlaceholder({
  submission,
}: {
  submission: SubmissionListItem;
}) {
  const tone = verdictChipTone(submission.verdict);
  return (
    <div className="review-detail__placeholder">
      <div className="review-side-head">
        <span className="review-side-head__name">{submission.participant}</span>
        <span className={`review-verdict-chip review-verdict-chip--${tone}`}>
          {formatVerdictLabel(submission.verdict)}
        </span>
      </div>
      <p className="review-detail__placeholder-hint">
        {formatSubmittedAt(submission.submitted_at)}
      </p>
    </div>
  );
}

/** Монтирует полную панель только у видимых посылок (и первых в очереди). */
export default function ReviewSubmissionPanelGate({ index, submission, ...props }: Props) {
  const panelId = submissionPanelId(submission.id);
  const ref = useRef<HTMLElement>(null);
  const [mounted, setMounted] = useState(index < EAGER_PANEL_COUNT);

  useEffect(() => {
    if (mounted) return;
    const el = ref.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting) {
          setMounted(true);
          observer.disconnect();
        }
      },
      { rootMargin: LAZY_ROOT_MARGIN, threshold: 0 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [mounted]);

  if (mounted) {
    return <ReviewSubmissionPanel submission={submission} {...props} />;
  }

  return (
    <article
      ref={ref}
      id={panelId}
      className="review-detail review-detail--panel review-detail--placeholder"
      aria-busy="true"
    >
      <ReviewSubmissionPanelPlaceholder submission={submission} />
    </article>
  );
}
