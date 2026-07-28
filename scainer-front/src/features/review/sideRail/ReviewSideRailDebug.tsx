import { useEffect } from "react";
import { isSideRailDebugEnabled } from "@/features/review/sideRail/reviewSideRailMeasure";
import type { SideRailDebugSnapshot } from "@/features/review/sideRail/useReviewSideRail";

export function ReviewSideRailDebugGuides() {
  useEffect(() => {
    const on = isSideRailDebugEnabled();
    document.documentElement.toggleAttribute("data-review-rail-debug", on);
    return () => {
      document.documentElement.removeAttribute("data-review-rail-debug");
    };
  }, []);

  if (!isSideRailDebugEnabled()) return null;

  return (
    <div className="review-side-rail-guides" aria-hidden>
      <div className="review-side-rail-guides__line review-side-rail-guides__line--work-top">
        <span>workTop</span>
      </div>
      <div className="review-side-rail-guides__line review-side-rail-guides__line--anchor">
        <span>anchor</span>
      </div>
      <div className="review-side-rail-guides__line review-side-rail-guides__line--work-bottom">
        <span>workBottom</span>
      </div>
    </div>
  );
}

export function ReviewSideRailDebugHud({
  panelLabel,
  snapshot,
}: {
  panelLabel: string;
  snapshot: SideRailDebugSnapshot | null;
}) {
  if (!isSideRailDebugEnabled() || !snapshot) return null;

  const lines = [
    `${panelLabel}`,
    `mode=${snapshot.mode} glue=${snapshot.glue}`,
    `band ${snapshot.bandTop | 0}…${snapshot.bandBottom | 0} (${snapshot.bandH | 0}) cap=${snapshot.cap | 0}`,
    `head=${snapshot.headH | 0} foot=${snapshot.footH | 0} body=${snapshot.bodyH | 0} gapEff=${snapshot.gapEff | 0}`,
    `gapA=${snapshot.gapA.toFixed(0)} gapB=${snapshot.gapB.toFixed(0)} scrollComments=${snapshot.scrollComments}`,
    `recalcs=${snapshot.recalculations} writes=${snapshot.styleWritesLastFrame}`,
  ];

  return <pre className="review-side-rail-hud">{lines.join("\n")}</pre>;
}
