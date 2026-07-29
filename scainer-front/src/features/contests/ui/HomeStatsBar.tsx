import {
  contestsCountLabel,
  contestsCountWord,
} from "@/features/contests/shared/contestHelpers";
import {
  formatSubmissionCount,
  submissionCountWord,
} from "@/features/findings/reportModel";

const numberFormat = new Intl.NumberFormat("ru-RU");

type Props = {
  contestCount: number;
  submissionCount: number;
};

type MetricProps = {
  value: number;
  label: string;
  ariaLabel: string;
};

function HomeSummaryMetric({ value, label, ariaLabel }: MetricProps) {
  return (
    <div className="home-summary__metric" aria-label={ariaLabel}>
      <span className="home-summary__sigma" aria-hidden>
        Σ
      </span>
      <div className="home-summary__metric-body">
        <span className="home-summary__value">{numberFormat.format(value)}</span>
        <span className="home-summary__label">{label}</span>
      </div>
    </div>
  );
}

export default function HomeStatsBar({ contestCount, submissionCount }: Props) {
  return (
    <section className="home-summary" aria-label="Сводка">
      <div className="home-summary__metrics">
        <HomeSummaryMetric
          value={contestCount}
          label={contestsCountWord(contestCount)}
          ariaLabel={contestsCountLabel(contestCount)}
        />
        <HomeSummaryMetric
          value={submissionCount}
          label={submissionCountWord(submissionCount)}
          ariaLabel={formatSubmissionCount(submissionCount)}
        />
      </div>
    </section>
  );
}
