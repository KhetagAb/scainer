import type { CSSProperties } from "react";
import jplagIcon from "@/assets/jplag-icon.png";
import moonIcon from "@/assets/moon-icon.png";
import {
  detectorLabel,
  fmtScore,
  JPLAG_DETECTOR,
  NIGHT_SUBMIT_DETECTOR,
  scoreFillPercent,
  scoreLevel,
} from "@/features/findings/reportModel";

type Props = {
  detectorId: string;
  score?: number;
};

function ChipLabel({ label, score }: { label: string; score?: number }) {
  if (score === undefined) {
    return <span className="detector-chip__label detector-chip__label--static">{label}</span>;
  }
  return (
    <span className="detector-chip__label">
      <span className="detector-chip__text">{label}</span>
      <span className="detector-chip__score-hover">{fmtScore(score)}</span>
    </span>
  );
}

export function DetectorChip({ detectorId, score }: Props) {
  const label = detectorLabel(detectorId);

  if (detectorId === NIGHT_SUBMIT_DETECTOR) {
    return (
      <span
        className="chip detector-chip detector-chip--night"
        title="Посылка в интервале 22:30–06:30 МСК"
      >
        <img src={moonIcon} alt="" className="detector-chip__moon" aria-hidden />
        <ChipLabel label={label} score={score} />
      </span>
    );
  }

  if (detectorId === JPLAG_DETECTOR) {
    if (score !== undefined) {
      const level = scoreLevel(score);
      return (
        <span
          className="chip detector-chip detector-chip--jplag"
          title="Сходство кода между посылками участников (JPlag)"
          style={{ "--jplag-fill": `${scoreFillPercent(score)}%` } as CSSProperties}
        >
          <span className={`detector-chip__jplag-fill ${level}`} aria-hidden />
          <span className="detector-chip__jplag-content">
            <img src={jplagIcon} alt="" className="detector-chip__detective" aria-hidden />
            <ChipLabel label={label} score={score} />
          </span>
        </span>
      );
    }

    return (
      <span
        className="chip detector-chip detector-chip--jplag"
        title="Сходство кода между посылками участников (JPlag)"
      >
        <span className="detector-chip__jplag-content">
          <img src={jplagIcon} alt="" className="detector-chip__detective" aria-hidden />
          <span className="detector-chip__label detector-chip__label--static">{label}</span>
        </span>
      </span>
    );
  }

  return (
    <span className="chip detector-chip">
      <ChipLabel label={label} score={score} />
    </span>
  );
}
