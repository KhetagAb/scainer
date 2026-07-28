import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type RefObject,
} from "react";
import {
  computeSideRailLayout,
  fitBodyHeight,
  clamp,
  type SideRailGlue,
  type SideRailLayoutOutput,
  type SideRailMode,
} from "@/features/review/sideRail/reviewSideRailLayout";
import {
  isSideRailDebugEnabled,
  measureSideRailChrome,
  measureSideRailSubmission,
  readZoneHeights,
  setRailVar,
  type SideRailChrome,
  type SideRailSubmissionMeasure,
} from "@/features/review/sideRail/reviewSideRailMeasure";

export type SideRailDebugSnapshot = SideRailLayoutOutput & {
  headH: number;
  footH: number;
  bodyNat: number;
  bodyMin: number;
  recalculations: number;
  styleWritesLastFrame: number;
};

type Env = SideRailChrome & {
  workTop: number;
  workBottom: number;
  anchor: number;
};

type RailState = {
  measure: SideRailSubmissionMeasure;
  isStatic: boolean;
  mode: SideRailMode;
  glue: SideRailGlue;
  live: SideRailLayoutOutput | null;
};

type Options = {
  codeRef: RefObject<HTMLElement | null>;
  headRef: RefObject<HTMLElement | null>;
  footRef: RefObject<HTMLElement | null>;
  bodyRef: RefObject<HTMLElement | null>;
  metaRef: RefObject<HTMLElement | null>;
  formRef: RefObject<HTMLElement | null>;
  listContentRef: RefObject<HTMLElement | null>;
  listScrollRef: RefObject<HTMLElement | null>;
  /** Re-run layout when content size may change. */
  contentKey: string;
};

function setModeAttr(
  rail: HTMLElement,
  mode: SideRailMode,
  prev: SideRailMode | null,
): boolean {
  if (prev === mode) return false;
  if (mode === "inactive") {
    rail.dataset.mode = "inactive";
    rail.setAttribute("aria-hidden", "true");
  } else if (mode === "static") {
    rail.dataset.mode = "static";
    rail.removeAttribute("aria-hidden");
  } else {
    rail.dataset.mode = "dynamic";
    rail.removeAttribute("aria-hidden");
  }
  return true;
}

function clearBodyTop(rail: HTMLElement, cache: Record<string, string>) {
  delete cache["--body-top"];
  rail.style.removeProperty("--body-top");
}

function setGlueAttr(
  rail: HTMLElement,
  glue: SideRailGlue,
  prev: SideRailGlue | null,
): boolean {
  if (prev === glue) return false;
  if (glue === "none") {
    delete rail.dataset.glue;
  } else {
    rail.dataset.glue = glue;
  }
  return true;
}

function isFormFocused(body: HTMLElement | null): boolean {
  const active = document.activeElement;
  if (!active || !body) return false;
  return body.contains(active) && active !== body;
}

export function useReviewSideRail({
  codeRef,
  headRef,
  footRef,
  bodyRef,
  metaRef,
  formRef,
  listContentRef,
  listScrollRef,
  contentKey,
}: Options) {
  const railRef = useRef<HTMLDivElement>(null);
  const stateRef = useRef<RailState | null>(null);
  const envRef = useRef<Env | null>(null);
  const rootCacheRef = useRef<Record<string, string>>({});
  const railCacheRef = useRef<Record<string, string>>({});
  const frozenBodyTopRef = useRef<number | null>(null);

  const [ready, setReady] = useState(false);
  const [mode, setMode] = useState<SideRailMode>("inactive");
  const [glue, setGlue] = useState<SideRailGlue>("none");
  const [scrollComments, setScrollComments] = useState(false);
  const [gapEff, setGapEff] = useState(0);
  const [debugSnapshot, setDebugSnapshot] = useState<SideRailDebugSnapshot | null>(null);

  const debugRef = useRef(isSideRailDebugEnabled());
  const recalcsRef = useRef(0);
  const writesLastRef = useRef(0);

  const applyLayout = useCallback(
    (layout: SideRailLayoutOutput, measure: SideRailSubmissionMeasure, writes: { n: number }) => {
      const rail = railRef.current;
      if (!rail) return;

      const prev = stateRef.current;
      if (setModeAttr(rail, layout.mode, prev?.mode ?? null)) writes.n++;
      if (setGlueAttr(rail, layout.glue, prev?.glue ?? null)) writes.n++;

      if (layout.mode === "static") {
        if (setRailVar(rail, "--gap-eff", measureSideRailChrome().gapMin, railCacheRef.current)) {
          writes.n++;
        }
        setScrollComments(false);
        setGapEff(measureSideRailChrome().gapMin);
        return;
      }

      if (layout.mode === "inactive") {
        setGlueAttr(rail, "none", prev?.glue ?? null);
        clearBodyTop(rail, railCacheRef.current);
        setScrollComments(false);
        setGapEff(0);
        return;
      }

      if (layout.mode === "glue") {
        clearBodyTop(rail, railCacheRef.current);
      }

      if (setRailVar(rail, "--gap-eff", layout.gapEff, railCacheRef.current)) writes.n++;
      if (setRailVar(rail, "--body-h", layout.bodyH, railCacheRef.current)) writes.n++;
      setGapEff(layout.gapEff);
      setScrollComments(layout.scrollComments);

      if (layout.mode === "pinned") {
        let bodyTop = layout.bodyTop;
        if (isFormFocused(bodyRef.current)) {
          if (frozenBodyTopRef.current == null) {
            const cached = railCacheRef.current["--body-top"];
            frozenBodyTopRef.current = cached ? parseFloat(cached) : bodyTop;
          }
          bodyTop = frozenBodyTopRef.current;
        } else {
          frozenBodyTopRef.current = null;
        }
        if (setRailVar(rail, "--body-top", bodyTop, railCacheRef.current)) writes.n++;
      }

      if (debugRef.current) {
        setDebugSnapshot({
          ...layout,
          headH: measure.headH,
          footH: measure.footH,
          bodyNat: measure.bodyNat,
          bodyMin: measure.bodyMin,
          recalculations: recalcsRef.current,
          styleWritesLastFrame: writes.n,
        });
      }
    },
    [bodyRef],
  );

  const syncZoneHeights = useCallback(
    (writes: { n: number }) => {
      const rail = railRef.current;
      if (!rail) return { headH: 0, footH: 0 };
      const { headH, footH } = readZoneHeights(headRef.current, footRef.current);
      if (setRailVar(rail, "--head-h", headH, railCacheRef.current)) writes.n++;
      if (setRailVar(rail, "--foot-h", footH, railCacheRef.current)) writes.n++;
      if (stateRef.current) {
        stateRef.current.measure = { ...stateRef.current.measure, headH, footH };
      }
      return { headH, footH };
    },
    [footRef, headRef],
  );

  /** Scroll-time layout — mirrors example.html tick(): cap vs bodyMin → pinned | glue. */
  const tick = useCallback(() => {
    const rail = railRef.current;
    const code = codeRef.current;
    const body = bodyRef.current;
    const env = envRef.current;
    const st = stateRef.current;
    if (!rail || !code || !body || !env || !st || st.isStatic) return;

    const writes = { n: 0 };
    const scrollY = window.scrollY;
    const codeTop = st.measure.codeTopDoc - scrollY;
    const codeBottom = st.measure.codeBottomDoc - scrollY;
    const { headH, footH } = syncZoneHeights(writes);
    const { bodyMin, bodyNat } = st.measure;

    const bandTop = Math.max(env.workTop, codeTop);
    const bandBottom = Math.min(env.workBottom, codeBottom);
    const bandH = bandBottom - bandTop;
    const cap = Math.max(0, bandH - headH - footH);

    if (bandH <= 0) {
      const inactive: SideRailLayoutOutput = {
        mode: "inactive",
        glue: "none",
        bodyH: 0,
        bodyTop: 0,
        gapEff: 0,
        gapA: 0,
        gapB: 0,
        scrollComments: false,
        bandTop,
        bandBottom,
        bandH,
        cap,
        anchor: env.anchor,
        workTop: env.workTop,
        workBottom: env.workBottom,
      };
      applyLayout(inactive, st.measure, writes);
      setMode("inactive");
      setGlue("none");
      st.mode = "inactive";
      st.glue = "none";
      st.live = inactive;
      writesLastRef.current = writes.n;
      return;
    }

    let layout: SideRailLayoutOutput;

    if (cap >= bodyMin) {
      const bodyH = fitBodyHeight(bodyNat, bodyMin, cap, env.gapMin);
      const gapEff = Math.max(0, Math.floor(Math.min(env.gapMin, (cap - bodyH) / 2)));
      const bodyTop = clamp(
        env.anchor - bodyH / 2,
        env.workTop + headH + gapEff,
        env.workBottom - footH - gapEff - bodyH,
      );
      layout = {
        mode: "pinned",
        glue: "none",
        bodyH,
        bodyTop,
        gapEff,
        gapA: bodyTop - (bandTop + headH),
        gapB: bandBottom - footH - (bodyTop + bodyH),
        scrollComments: bodyH < bodyNat,
        bandTop,
        bandBottom,
        bandH,
        cap,
        anchor: env.anchor,
        workTop: env.workTop,
        workBottom: env.workBottom,
      };
    } else {
      const bodyH = Math.max(0, Math.min(bodyMin, st.measure.codeH - headH - footH));
      const glue: SideRailGlue = codeTop > env.workTop ? "top" : "bottom";
      layout = {
        mode: "glue",
        glue,
        bodyH,
        bodyTop: 0,
        gapEff: 0,
        gapA: 0,
        gapB: 0,
        scrollComments: bodyH < bodyNat,
        bandTop,
        bandBottom,
        bandH,
        cap,
        anchor: env.anchor,
        workTop: env.workTop,
        workBottom: env.workBottom,
      };
    }

    applyLayout(layout, st.measure, writes);
    setMode(layout.mode);
    setGlue(layout.glue);
    st.mode = layout.mode;
    st.glue = layout.glue;
    st.live = layout;
    writesLastRef.current = writes.n;
  }, [applyLayout, bodyRef, codeRef, syncZoneHeights]);

  const roRef = useRef<ResizeObserver | null>(null);
  const observedRef = useRef<Set<Element>>(new Set());

  const observeEl = useCallback((el: Element | null | undefined) => {
    if (!el || observedRef.current.has(el)) return;
    observedRef.current.add(el);
    roRef.current?.observe(el);
  }, []);

  const recompute = useCallback(() => {
    const rail = railRef.current;
    const code = codeRef.current;
    const body = bodyRef.current;
    if (!rail || !code || !body) return;

    const scrollY = window.scrollY;
    const chrome = measureSideRailChrome();
    const workTop = chrome.chromeTop + chrome.padTop;
    const workBottom = chrome.chromeBottom - chrome.padBottom;
    const anchor = (workTop + workBottom) / 2;
    envRef.current = { ...chrome, workTop, workBottom, anchor };

    const writes = { n: 0 };
    const root = document.documentElement;
    if (setRailVar(root, "--review-rail-work-top", workTop, rootCacheRef.current)) writes.n++;
    if (setRailVar(root, "--review-rail-anchor", anchor, rootCacheRef.current)) writes.n++;
    if (setRailVar(root, "--review-rail-pad-bottom", chrome.padBottom, rootCacheRef.current)) {
      writes.n++;
    }

    const measure = measureSideRailSubmission(
      code,
      headRef.current,
      footRef.current,
      body,
      metaRef.current,
      formRef.current,
      listScrollRef.current,
      scrollY,
    );

    observeEl(code);
    observeEl(headRef.current);
    observeEl(footRef.current);
    observeEl(body);
    observeEl(metaRef.current);
    observeEl(formRef.current);
    observeEl(listContentRef.current);
    observeEl(document.querySelector("[data-app-topbar]"));
    observeEl(document.querySelector(".contest-head--review"));

    const roomStat = Math.min(workBottom - workTop, measure.codeH);
    const threshold = measure.headH + measure.footH + measure.bodyMin;
    const wasStatic = stateRef.current?.isStatic ?? false;
    const isStatic =
      wasStatic ? roomStat < threshold + 8 : roomStat < threshold;

    if (setRailVar(rail, "--head-h", measure.headH, railCacheRef.current)) writes.n++;
    if (setRailVar(rail, "--foot-h", measure.footH, railCacheRef.current)) writes.n++;

    const layout = computeSideRailLayout({
      chromeTop: chrome.chromeTop,
      chromeBottom: chrome.chromeBottom,
      padTop: chrome.padTop,
      padBottom: chrome.padBottom,
      codeTop: measure.codeTopDoc - scrollY,
      codeBottom: measure.codeBottomDoc - scrollY,
      codeH: measure.codeH,
      headH: measure.headH,
      footH: measure.footH,
      bodyNat: measure.bodyNat,
      bodyMin: measure.bodyMin,
      gapMin: chrome.gapMin,
      wasStatic,
    });

    const effectiveMode: SideRailMode = isStatic ? "static" : layout.mode === "inactive" ? "inactive" : layout.mode;

    stateRef.current = {
      measure,
      isStatic,
      mode: effectiveMode,
      glue: isStatic ? "none" : layout.glue,
      live: isStatic ? null : layout,
    };

    if (isStatic) {
      const staticLayout: SideRailLayoutOutput = {
        ...layout,
        mode: "static",
        glue: "none",
        bodyH: measure.bodyNat,
        gapEff: chrome.gapMin,
        scrollComments: false,
      };
      applyLayout(staticLayout, measure, writes);
      setMode("static");
      setGlue("none");
    } else {
      applyLayout({ ...layout, mode: layout.mode }, measure, writes);
      setMode(layout.mode);
      setGlue(layout.glue);
    }

    recalcsRef.current++;
    writesLastRef.current = writes.n;
    setReady(true);
    tick();
  }, [
    applyLayout,
    codeRef,
    headRef,
    footRef,
    bodyRef,
    metaRef,
    formRef,
    listContentRef,
    listScrollRef,
    tick,
    observeEl,
  ]);

  useLayoutEffect(() => {
    let recomputeScheduled = false;
    let tickScheduled = false;

    const scheduleRecompute = () => {
      if (recomputeScheduled) return;
      recomputeScheduled = true;
      requestAnimationFrame(() => {
        recomputeScheduled = false;
        recompute();
      });
    };

    const scheduleTick = () => {
      if (tickScheduled) return;
      tickScheduled = true;
      requestAnimationFrame(() => {
        tickScheduled = false;
        tick();
      });
    };

    scheduleRecompute();

    window.addEventListener("scroll", scheduleTick, { passive: true });
    window.addEventListener("resize", scheduleRecompute);
    document.addEventListener("visibilitychange", () => {
      if (!document.hidden) scheduleRecompute();
    });

    const ro = new ResizeObserver(() => scheduleRecompute());
    roRef.current = ro;

    document.fonts?.ready.then(() => scheduleRecompute());

    return () => {
      window.removeEventListener("scroll", scheduleTick);
      window.removeEventListener("resize", scheduleRecompute);
      ro.disconnect();
      roRef.current = null;
      observedRef.current.clear();
    };
  }, [recompute, tick, contentKey]);

  return {
    railRef,
    ready,
    mode,
    glue,
    scrollComments,
    gapEff,
    debugSnapshot,
    debugEnabled: debugRef.current,
  };
}
