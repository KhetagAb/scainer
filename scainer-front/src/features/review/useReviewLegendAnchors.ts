import { useEffect, type RefObject } from "react";

const ANCHOR_NAMES = [
  "picker",
  "cite",
  "statement",
  "pr-only",
  "verdict",
  "nav",
] as const;

type AnchorName = (typeof ANCHOR_NAMES)[number];

function centerYRelativeToContainer(
  el: Element,
  containerRect: DOMRect,
): number {
  const rect = el.getBoundingClientRect();
  return rect.top + rect.height / 2 - containerRect.top;
}

function measureCiteY(card: HTMLElement, containerRect: DOMRect): number | null {
  const lines = card.querySelectorAll<HTMLElement>(".line-match");
  if (lines.length === 0) return null;

  let top = Infinity;
  let bottom = -Infinity;
  for (const line of lines) {
    const rect = line.getBoundingClientRect();
    top = Math.min(top, rect.top);
    bottom = Math.max(bottom, rect.bottom);
  }
  if (!Number.isFinite(top) || !Number.isFinite(bottom)) return null;
  return (top + bottom) / 2 - containerRect.top;
}

function clampY(y: number, calloutHeight: number, containerHeight: number): number {
  const half = calloutHeight / 2;
  if (containerHeight <= 0) return y;
  if (calloutHeight >= containerHeight) return containerHeight / 2;
  return Math.max(half, Math.min(containerHeight - half, y));
}

function measureRawAnchors(
  cardWrap: HTMLElement,
  containerRect: DOMRect,
): Partial<Record<AnchorName, number>> {
  const raw: Partial<Record<AnchorName, number>> = {};

  const picker = cardWrap.querySelector('[data-legend-anchor="picker"]');
  if (picker) raw.picker = centerYRelativeToContainer(picker, containerRect);

  const statement = cardWrap.querySelector('[data-legend-anchor="statement"]');
  if (statement) raw.statement = centerYRelativeToContainer(statement, containerRect);

  const prOnly = cardWrap.querySelector('[data-legend-anchor="pr-only"]');
  if (prOnly) raw["pr-only"] = centerYRelativeToContainer(prOnly, containerRect);

  const verdict = cardWrap.querySelector('[data-legend-anchor="verdict"]');
  if (verdict) raw.verdict = centerYRelativeToContainer(verdict, containerRect);

  const nav = cardWrap.querySelector('[data-legend-anchor="nav"]');
  if (nav) raw.nav = centerYRelativeToContainer(nav, containerRect);

  const cite = measureCiteY(cardWrap, containerRect);
  if (cite != null) raw.cite = cite;

  return raw;
}

function applyAnchors(cardWrap: HTMLElement, diagram: HTMLElement): void {
  const containerRect = cardWrap.getBoundingClientRect();
  if (containerRect.height <= 0) return;

  const raw = measureRawAnchors(cardWrap, containerRect);
  const sections = diagram.querySelectorAll<HTMLElement>(
    ".legend-diagram__notes .legend-callout__sections",
  );

  for (const sectionsEl of sections) {
    const sectionsRect = sectionsEl.getBoundingClientRect();
    const topOffset = containerRect.top - sectionsRect.top;
    const boundsHeight = sectionsRect.height || containerRect.height;

    for (const callout of sectionsEl.querySelectorAll<HTMLElement>(
      ".legend-callout__section--align-anchor",
    )) {
      const name = callout.dataset.legendAnchor as AnchorName | undefined;
      if (!name || raw[name] == null) continue;

      const calloutHeight = callout.getBoundingClientRect().height;
      const yInSections = raw[name]! + topOffset;
      const y = clampY(yInSections, calloutHeight, boundsHeight);
      diagram.style.setProperty(`--legend-y-${name}`, `${y}px`);
    }
  }
}

function clearAnchors(diagram: HTMLElement): void {
  for (const name of ANCHOR_NAMES) {
    diagram.style.removeProperty(`--legend-y-${name}`);
  }
}

export function useReviewLegendAnchors(
  cardWrapRef: RefObject<HTMLElement | null>,
  diagramRef: RefObject<HTMLElement | null>,
  enabled: boolean,
): void {
  useEffect(() => {
    if (!enabled) return;

    const cardWrap = cardWrapRef.current;
    const diagram = diagramRef.current;
    if (!cardWrap || !diagram) return;

    let rafId = 0;

    const measure = () => {
      cancelAnimationFrame(rafId);
      rafId = requestAnimationFrame(() => {
        if (cardWrapRef.current && diagramRef.current) {
          applyAnchors(cardWrapRef.current, diagramRef.current);
        }
      });
    };

    measure();

    const observer = new ResizeObserver(measure);
    observer.observe(cardWrap);
    observer.observe(diagram);

    window.addEventListener("resize", measure);

    const diagramEl = diagram;

    return () => {
      cancelAnimationFrame(rafId);
      observer.disconnect();
      window.removeEventListener("resize", measure);
      clearAnchors(diagramEl);
    };
  }, [cardWrapRef, diagramRef, enabled]);
}
