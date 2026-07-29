import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type MouseEvent,
  type RefObject,
} from "react";

export const LINE_FLASH_MS = 1000;

export type CiteRangeState = { from: number; to: number; top: number };

export function getSequentialLineNo(_el: HTMLElement, index: number): number {
  return index + 1;
}

export function getCiteLineNoFromData(el: HTMLElement, _index?: number): number | null {
  const raw = el.dataset.citeLine;
  if (!raw) return null;
  const n = Number(raw);
  return Number.isNaN(n) ? null : n;
}

export function lineRangeFromSelection(
  root: HTMLElement,
  wrap: HTMLElement,
  getLineNo: (el: HTMLElement, index: number) => number | null,
): CiteRangeState | null {
  const sel = window.getSelection();
  if (!sel || sel.isCollapsed || sel.rangeCount === 0) return null;

  const range = sel.getRangeAt(0);
  if (!root.contains(range.commonAncestorContainer)) return null;

  const lineEls = root.querySelectorAll<HTMLElement>(".line");
  let from = -1;
  let to = -1;
  lineEls.forEach((el, i) => {
    if (range.intersectsNode(el)) {
      const n = getLineNo(el, i);
      if (n == null) return;
      if (from < 0) from = n;
      to = n;
    }
  });
  if (from < 0 || to < 0) return null;

  const rect = range.getBoundingClientRect();
  if (rect.width === 0 && rect.height === 0) return null;
  const wrapRect = wrap.getBoundingClientRect();

  return {
    from,
    to,
    top: rect.bottom - wrapRect.top + 4,
  };
}

export function onLineContentMouseDown(
  e: MouseEvent,
  skipGutterSelector: string,
  onLineNumberClick?: (lineNo: number) => void,
  skipNextCiteMouseUpRef?: { current: boolean },
): void {
  if (!onLineNumberClick) return;
  if ((e.target as HTMLElement).closest(skipGutterSelector)) return;
  if (e.detail === 2) {
    e.preventDefault();
    if (skipNextCiteMouseUpRef) skipNextCiteMouseUpRef.current = true;
  }
}

export function onLineContentDoubleClick(
  e: MouseEvent,
  skipGutterSelector: string,
  lineNo: number | null,
  lineIndex: number,
  onLineNumberClick: (lineNo: number) => void,
  clearCite: () => void,
  triggerFlash: (lineIndex: number) => void,
): void {
  if (lineNo == null) return;
  if ((e.target as HTMLElement).closest(skipGutterSelector)) return;
  e.preventDefault();
  onLineNumberClick(lineNo);
  triggerFlash(lineIndex);
  clearCite();
  window.getSelection()?.removeAllRanges();
}

type UseCodeCiteInteractionsArgs = {
  wrapRef: RefObject<HTMLDivElement | null>;
  onCiteLines?: (from: number, to: number) => void;
  onLineNumberClick?: (lineNo: number) => void;
  getLineNoFromLineEl: (el: HTMLElement, index: number) => number | null;
  resetKey: unknown;
};

export function useCodeCiteInteractions({
  wrapRef,
  onCiteLines,
  onLineNumberClick,
  getLineNoFromLineEl,
  resetKey,
}: UseCodeCiteInteractionsArgs) {
  const citeModifierSelectRef = useRef(false);
  const modifierHeldRef = useRef(false);
  const skipNextCiteMouseUpRef = useRef(false);
  const [cite, setCite] = useState<CiteRangeState | null>(null);
  const [lineFlashes, setLineFlashes] = useState<Record<number, number>>({});

  const triggerLineFlashMany = useCallback((lineIndices: number[]) => {
    if (lineIndices.length === 0) return;
    const token = Date.now();
    setLineFlashes((prev) => {
      const next = { ...prev };
      for (const idx of lineIndices) delete next[idx];
      return next;
    });
    requestAnimationFrame(() => {
      setLineFlashes((prev) => {
        const next = { ...prev };
        for (const idx of lineIndices) next[idx] = token;
        return next;
      });
      window.setTimeout(() => {
        setLineFlashes((prev) => {
          const next = { ...prev };
          for (const idx of lineIndices) {
            if (prev[idx] === token) delete next[idx];
          }
          return next;
        });
      }, LINE_FLASH_MS);
    });
  }, []);

  const triggerLineFlash = useCallback(
    (lineIndex: number) => triggerLineFlashMany([lineIndex]),
    [triggerLineFlashMany],
  );

  const triggerLineFlashRange = useCallback(
    (fromLine: number, toLine: number) => {
      const from = Math.min(fromLine, toLine);
      const to = Math.max(fromLine, toLine);
      const wrap = wrapRef.current;
      const root = wrap?.querySelector<HTMLElement>(".code-lines");
      if (!root) return;
      const indices: number[] = [];
      root.querySelectorAll<HTMLElement>(".line").forEach((el, i) => {
        const n = getLineNoFromLineEl(el, i);
        if (n != null && n >= from && n <= to) indices.push(i);
      });
      triggerLineFlashMany(indices);
    },
    [getLineNoFromLineEl, triggerLineFlashMany, wrapRef],
  );

  const handleLineNumberClick = useCallback(
    (lineNo: number, lineIndex: number) => {
      if (!onLineNumberClick) return;
      onLineNumberClick(lineNo);
      triggerLineFlash(lineIndex);
    },
    [onLineNumberClick, triggerLineFlash],
  );

  const applyCiteRange = useCallback(
    (from: number, to: number) => {
      if (!onCiteLines) return;
      onCiteLines(from, to);
      triggerLineFlashRange(from, to);
      setCite(null);
      window.getSelection()?.removeAllRanges();
    },
    [onCiteLines, triggerLineFlashRange],
  );

  const refreshCite = useCallback(() => {
    if (!onCiteLines) return;
    if (modifierHeldRef.current || citeModifierSelectRef.current) {
      setCite(null);
      return;
    }
    const wrap = wrapRef.current;
    const root = wrap?.querySelector<HTMLElement>(".code-lines");
    if (!wrap || !root) {
      setCite(null);
      return;
    }
    setCite(lineRangeFromSelection(root, wrap, getLineNoFromLineEl));
  }, [getLineNoFromLineEl, onCiteLines, wrapRef]);

  const onPreMouseDown = useCallback(
    (e: MouseEvent) => {
      if (!onCiteLines) return;
      citeModifierSelectRef.current = e.metaKey || e.ctrlKey;
    },
    [onCiteLines],
  );

  const onPreMouseUp = useCallback(
    (e: MouseEvent) => {
      if (!onCiteLines) return;

      if (skipNextCiteMouseUpRef.current) {
        skipNextCiteMouseUpRef.current = false;
        refreshCite();
        return;
      }

      const modifier = e.metaKey || e.ctrlKey || citeModifierSelectRef.current;
      citeModifierSelectRef.current = false;

      const wrap = wrapRef.current;
      const root = wrap?.querySelector<HTMLElement>(".code-lines");
      if (!wrap || !root) return;

      const range = lineRangeFromSelection(root, wrap, getLineNoFromLineEl);
      if (modifier && range) {
        applyCiteRange(range.from, range.to);
        return;
      }

      refreshCite();
    },
    [applyCiteRange, getLineNoFromLineEl, onCiteLines, refreshCite, wrapRef],
  );

  useEffect(() => {
    if (!onCiteLines) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Meta" || e.key === "Control") modifierHeldRef.current = true;
    };
    const onKeyUp = (e: KeyboardEvent) => {
      if (e.key === "Meta" || e.key === "Control") modifierHeldRef.current = false;
    };
    const onSel = () => refreshCite();
    document.addEventListener("keydown", onKeyDown);
    document.addEventListener("keyup", onKeyUp);
    document.addEventListener("selectionchange", onSel);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.removeEventListener("keyup", onKeyUp);
      document.removeEventListener("selectionchange", onSel);
    };
  }, [onCiteLines, refreshCite]);

  useEffect(() => {
    setCite(null);
    setLineFlashes({});
  }, [resetKey]);

  return {
    cite,
    setCite,
    lineFlashes,
    skipNextCiteMouseUpRef,
    handleLineNumberClick,
    triggerLineFlash,
    applyCiteRange,
    onPreMouseDown,
    onPreMouseUp,
  };
}
