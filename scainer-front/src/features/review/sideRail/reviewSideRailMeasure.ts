/** Runtime measurement helpers for side-rail layout (§4, §10). */

export type SideRailChrome = {
  chromeTop: number;
  chromeBottom: number;
  padTop: number;
  padBottom: number;
  gapMin: number;
};

export type SideRailSubmissionMeasure = {
  codeTopDoc: number;
  codeBottomDoc: number;
  codeH: number;
  headH: number;
  footH: number;
  bodyNat: number;
  bodyMin: number;
};

const CHROME_TOP_SELECTORS = ["[data-app-topbar]", ".contest-head--review"];

function readRootPx(name: string, fallback: number): number {
  const raw = getComputedStyle(document.documentElement).getPropertyValue(name);
  const parsed = parseFloat(raw);
  return Number.isFinite(parsed) ? parsed : fallback;
}

/** Bottom edge of overlapping top chrome; top of bottom chrome or viewport height. */
export function measureSideRailChrome(viewportH = window.innerHeight): SideRailChrome {
  let chromeTop = 0;
  for (const sel of CHROME_TOP_SELECTORS) {
    const el = document.querySelector(sel);
    if (!el) continue;
    const rect = el.getBoundingClientRect();
    if (rect.bottom > 0 && rect.top < viewportH) {
      chromeTop = Math.max(chromeTop, rect.bottom);
    }
  }

  let chromeBottom = viewportH;
  const bottomChrome = document.querySelector("[data-review-bottom-chrome]");
  if (bottomChrome) {
    const rect = bottomChrome.getBoundingClientRect();
    if (rect.top < viewportH && rect.bottom > 0) {
      chromeBottom = Math.min(chromeBottom, rect.top);
    }
  }

  return {
    chromeTop,
    chromeBottom,
    padTop: readRootPx("--review-rail-pad-top", 12),
    padBottom: readRootPx("--review-rail-pad-bottom", 12),
    gapMin: readRootPx("--review-rail-gap-min", 24),
  };
}

/** Live head/foot heights — offsetHeight works while ancestors use opacity, not visibility. */
export function readZoneHeights(
  headEl: HTMLElement | null,
  footEl: HTMLElement | null,
): { headH: number; footH: number } {
  return {
    headH: headEl ? Math.ceil(headEl.offsetHeight) : 0,
    footH: footEl ? Math.ceil(footEl.offsetHeight) : 0,
  };
}

function measureCommentsBlock(listScrollEl: HTMLElement | null): number {
  if (!listScrollEl) return 0;
  const section = listScrollEl.closest(".review-comments");
  if (!section) return 0;
  const title = section.querySelector(".review-comments__title");
  const titleH = title ? (title as HTMLElement).offsetHeight : 0;
  const titleMargin = title ? parseFloat(getComputedStyle(title).marginBottom) || 0 : 0;
  const sectionMargin = parseFloat(getComputedStyle(section as Element).marginTop) || 0;
  const viewportH = readRootPx("--review-rail-comments-max-h", 144);
  return Math.ceil(sectionMargin + titleH + titleMargin + viewportH);
}

export function measureSideRailSubmission(
  codeEl: HTMLElement,
  headEl: HTMLElement | null,
  footEl: HTMLElement | null,
  bodyEl: HTMLElement,
  metaEl: HTMLElement | null,
  formEl: HTMLElement | null,
  listScrollEl: HTMLElement | null,
  scrollY: number,
): SideRailSubmissionMeasure {
  const codeRect = codeEl.getBoundingClientRect();
  const bs = getComputedStyle(bodyEl);
  const shell =
    parseFloat(bs.paddingTop) +
    parseFloat(bs.paddingBottom) +
    2 * parseFloat(bs.borderTopWidth) +
    2 * (parseFloat(bs.rowGap) || 0);

  const fixed =
    shell +
    (metaEl?.offsetHeight ?? 0) +
    (formEl?.offsetHeight ?? 0);

  const commentsH = measureCommentsBlock(listScrollEl);
  const zones = readZoneHeights(headEl, footEl);

  return {
    codeTopDoc: codeRect.top + scrollY,
    codeBottomDoc: codeRect.bottom + scrollY,
    codeH: codeRect.height,
    headH: zones.headH,
    footH: zones.footH,
    bodyNat: Math.ceil(fixed + commentsH),
    bodyMin: Math.ceil(fixed + commentsH),
  };
}

/** Write a CSS custom property only when the rounded px value changed. */
export function setRailVar(
  el: HTMLElement,
  name: string,
  px: number,
  cache: Record<string, string>,
): boolean {
  const v = `${Math.round(px)}px`;
  if (cache[name] === v) return false;
  cache[name] = v;
  el.style.setProperty(name, v);
  return true;
}

export function isSideRailDebugEnabled(): boolean {
  try {
    return localStorage.getItem("review-side-rail-debug") === "1";
  } catch {
    return false;
  }
}
