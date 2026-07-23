import moonIcon from "@/assets/moon-icon.png";
import { detectorLabel, NIGHT_SUBMIT_DETECTOR } from "@/features/findings/reportModel";

type Props = {
  detectorId: string;
};

export function DetectorChip({ detectorId }: Props) {
  if (detectorId === NIGHT_SUBMIT_DETECTOR) {
    return (
      <span
        className="chip detector-chip detector-chip--night"
        title="Посылка в интервале 22:30–06:30 МСК"
      >
        <img src={moonIcon} alt="" className="detector-chip__moon" aria-hidden />
        <span>{detectorLabel(detectorId)}</span>
      </span>
    );
  }

  return <span className="chip detector-chip">{detectorLabel(detectorId)}</span>;
}
