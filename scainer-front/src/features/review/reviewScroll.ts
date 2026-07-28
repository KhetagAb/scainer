let scrollLockUntil = 0;
let pickActiveAfterScroll: (() => void) | null = null;
let activeScrollFrame = 0;

export function isReviewScrollLocked(): boolean {
  return performance.now() < scrollLockUntil;
}

/** Called from useReviewActivePanel to refresh active id after programmatic scroll. */
export function registerReviewScrollPickActive(fn: () => void): () => void {
  pickActiveAfterScroll = fn;
  return () => {
    if (pickActiveAfterScroll === fn) pickActiveAfterScroll = null;
  };
}

function releaseScrollLock() {
  scrollLockUntil = 0;
  pickActiveAfterScroll?.();
}

function lockScrollFor(ms: number) {
  scrollLockUntil = performance.now() + ms;
}

/** Fixed duration — native smooth scales with distance and feels sluggish on long stacks. */
const REVIEW_SCROLL_MS = 450;

/** Smooth scroll to a review panel; respects scroll-margin and prefers-reduced-motion. */
export function scrollToReviewPanel(el: HTMLElement) {
  const marginTop = parseFloat(getComputedStyle(el).scrollMarginTop) || 0;
  const targetTop = Math.max(0, el.getBoundingClientRect().top + window.scrollY - marginTop);
  const startY = window.scrollY;
  const distance = targetTop - startY;

  if (Math.abs(distance) < 1) return;

  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reducedMotion) {
    window.scrollTo(0, targetTop);
    pickActiveAfterScroll?.();
    return;
  }

  if (activeScrollFrame) {
    cancelAnimationFrame(activeScrollFrame);
    activeScrollFrame = 0;
  }

  lockScrollFor(REVIEW_SCROLL_MS + 50);

  const start = performance.now();
  const tick = (now: number) => {
    const t = Math.min((now - start) / REVIEW_SCROLL_MS, 1);
    const eased = 1 - (1 - t) ** 3;
    window.scrollTo(0, startY + distance * eased);
    if (t < 1) {
      activeScrollFrame = requestAnimationFrame(tick);
    } else {
      activeScrollFrame = 0;
      releaseScrollLock();
    }
  };
  activeScrollFrame = requestAnimationFrame(tick);
}
