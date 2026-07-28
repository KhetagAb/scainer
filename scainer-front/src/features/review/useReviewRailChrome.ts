import { useEffect } from "react";

const TOPBAR_SELECTOR = "[data-app-topbar]";
const CONTEST_HEAD_SELECTOR = "[data-contest-head]";

function measureReviewChromeTop() {
  const contestHead = document.querySelector(CONTEST_HEAD_SELECTOR);
  const topbar = document.querySelector(TOPBAR_SELECTOR);
  const chromeBottom = contestHead
    ? contestHead.getBoundingClientRect().bottom
    : topbar
      ? topbar.getBoundingClientRect().bottom
      : 0;
  document.documentElement.style.setProperty(
    "--review-chrome-top",
    `${Math.max(0, chromeBottom)}px`,
  );
}

/** Syncs --review-chrome-top from measured topbar + contest head (review page). */
export function useReviewRailChrome(enabled: boolean) {
  useEffect(() => {
    if (!enabled) return;

    measureReviewChromeTop();

    const observed = new Set<Element>();
    const ro = new ResizeObserver(() => measureReviewChromeTop());

    const observe = (selector: string) => {
      const el = document.querySelector(selector);
      if (el && !observed.has(el)) {
        observed.add(el);
        ro.observe(el);
      }
    };

    observe(TOPBAR_SELECTOR);
    observe(CONTEST_HEAD_SELECTOR);

    window.addEventListener("resize", measureReviewChromeTop);
    return () => {
      ro.disconnect();
      window.removeEventListener("resize", measureReviewChromeTop);
    };
  }, [enabled]);
}
