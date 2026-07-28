import { useEffect, useState } from "react";
import { isReviewScrollLocked, registerReviewScrollPickActive } from "./reviewScroll";

function readChromeTopPx(): number {
  const raw = getComputedStyle(document.documentElement).getPropertyValue("--review-chrome-top");
  const parsed = parseFloat(raw);
  return Number.isFinite(parsed) ? parsed : 0;
}

/**
 * Active submission = topmost panel whose section intersects the rail band below chrome.
 */
export function useReviewActivePanel(panelIds: string[]): string | null {
  const [activeId, setActiveId] = useState<string | null>(panelIds[0] ?? null);
  const panelKey = panelIds.join("\0");

  useEffect(() => {
    if (!panelIds.length) {
      setActiveId(null);
      return;
    }

    const panels = panelIds
      .map((id) => document.getElementById(id))
      .filter((el): el is HTMLElement => el != null);

    if (!panels.length) {
      setActiveId(panelIds[0] ?? null);
      return;
    }

    const chromeTop = readChromeTopPx();
    const rootMargin = `-${chromeTop}px 0px 0px 0px`;

    const pickActive = () => {
      if (isReviewScrollLocked()) return;
      const visible = panels
        .filter((panel) => {
          const rect = panel.getBoundingClientRect();
          return rect.bottom > chromeTop && rect.top < window.innerHeight;
        })
        .sort((a, b) => a.getBoundingClientRect().top - b.getBoundingClientRect().top);
      if (visible.length) setActiveId(visible[0].id);
    };

    pickActive();

    const unregisterScrollPick = registerReviewScrollPickActive(pickActive);

    const observer = new IntersectionObserver(() => pickActive(), {
      root: null,
      rootMargin,
      threshold: [0, 0.01, 0.05, 0.1],
    });

    panels.forEach((panel) => observer.observe(panel));
    window.addEventListener("resize", pickActive);

    return () => {
      unregisterScrollPick();
      observer.disconnect();
      window.removeEventListener("resize", pickActive);
    };
  }, [panelKey, panelIds]);

  return activeId;
}
