import { DotLottieReact } from "@lottiefiles/dotlottie-react";
import sonarRadarUrl from "@/assets/sonar-radar.lottie?url";

export function RadarAttentionIcon() {
  return (
    <DotLottieReact
      src={sonarRadarUrl}
      loop
      autoplay
      className="review-findings-link__icon"
      aria-hidden
    />
  );
}
