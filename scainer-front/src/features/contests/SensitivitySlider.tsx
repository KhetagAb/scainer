import Slider from "rc-slider";
import "rc-slider/assets/index.css";
import { SENSITIVITY_MARKS } from "@/features/contests/useParallelSensitivity";

type Props = {
  threshold: number;
  onChange: (threshold: number) => void;
  className?: string;
};

/** Слайдер чувствительности: 50% / 70% / 95%. */
export default function SensitivitySlider({ threshold, onChange, className }: Props) {
  return (
    <div className={className ? `header-control ${className}` : "header-control"}>
      <span className="control-group__label">Чувствительность</span>
      <div className="sensitivity-slider-wrap">
        <Slider
          className="sensitivity-slider"
          min={50}
          max={95}
          step={null}
          marks={SENSITIVITY_MARKS}
          value={Math.round(threshold * 100)}
          onChange={(v) => onChange((Array.isArray(v) ? v[0] : v) / 100)}
        />
      </div>
    </div>
  );
}
