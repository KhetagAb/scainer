/** §4 — pure layout: measured inputs → mode, sizes, positions. No DOM. */

export const SIDE_RAIL_HYST_PX = 8;

export type SideRailMode = "pinned" | "glue" | "static" | "inactive";
export type SideRailGlue = "none" | "top" | "bottom";

export type SideRailLayoutInput = {
  chromeTop: number;
  chromeBottom: number;
  padTop: number;
  padBottom: number;
  codeTop: number;
  codeBottom: number;
  codeH: number;
  headH: number;
  footH: number;
  bodyNat: number;
  bodyMin: number;
  gapMin: number;
  wasStatic: boolean;
};

export type SideRailLayoutOutput = {
  mode: SideRailMode;
  glue: SideRailGlue;
  bodyH: number;
  bodyTop: number;
  gapEff: number;
  gapA: number;
  gapB: number;
  scrollComments: boolean;
  bandTop: number;
  bandBottom: number;
  bandH: number;
  cap: number;
  anchor: number;
  workTop: number;
  workBottom: number;
};

export function clamp(v: number, lo: number, hi: number): number {
  if (lo > hi) return lo;
  return Math.min(Math.max(v, lo), hi);
}

export function applyStaticHysteresis(
  roomBelowThreshold: boolean,
  wasStatic: boolean,
  threshold: number,
  roomStat: number,
): boolean {
  if (wasStatic) return roomStat < threshold + SIDE_RAIL_HYST_PX;
  return roomBelowThreshold;
}

/** Fit body height into available cap (PINNED). */
export function fitBodyHeight(
  bodyNat: number,
  bodyMin: number,
  cap: number,
  gapMin: number,
): number {
  const inner = Math.max(cap - 2 * gapMin, Math.min(bodyMin, cap));
  return Math.max(0, Math.round(Math.min(bodyNat, inner)));
}

export function computeSideRailLayout(input: SideRailLayoutInput): SideRailLayoutOutput {
  const {
    chromeTop,
    chromeBottom,
    padTop,
    padBottom,
    codeTop,
    codeBottom,
    codeH,
    headH,
    footH,
    bodyNat,
    bodyMin,
    gapMin,
    wasStatic,
  } = input;

  const workTop = chromeTop + padTop;
  const workBottom = chromeBottom - padBottom;
  const anchor = (workTop + workBottom) / 2;

  const bandTop = Math.max(workTop, codeTop);
  const bandBottom = Math.min(workBottom, codeBottom);
  const bandH = bandBottom - bandTop;
  const cap = Math.max(0, bandH - headH - footH);

  const base: SideRailLayoutOutput = {
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
    anchor,
    workTop,
    workBottom,
  };

  if (bandH <= 0) return base;

  const roomStat = Math.min(workBottom - workTop, codeH);
  const threshold = headH + footH + bodyMin;
  const isStatic = applyStaticHysteresis(
    roomStat < threshold,
    wasStatic,
    threshold,
    roomStat,
  );

  if (isStatic) {
    return {
      ...base,
      mode: "static",
      bodyH: bodyNat,
      gapEff: gapMin,
      scrollComments: false,
    };
  }

  if (cap >= bodyMin) {
    const bodyH = fitBodyHeight(bodyNat, bodyMin, cap, gapMin);
    const gapEff = Math.max(0, Math.floor(Math.min(gapMin, (cap - bodyH) / 2)));
    const lo = bandTop + headH + gapEff;
    const hi = bandBottom - footH - gapEff - bodyH;
    const bodyTop = clamp(anchor - bodyH / 2, lo, hi);
    const gapA = bodyTop - (bandTop + headH);
    const gapB = bandBottom - footH - (bodyTop + bodyH);

    return {
      ...base,
      mode: "pinned",
      bodyH,
      bodyTop,
      gapEff,
      gapA,
      gapB,
      scrollComments: bodyH < bodyNat,
    };
  }

  const bodyH = Math.max(0, Math.min(bodyMin, codeH - headH - footH));
  const glue: SideRailGlue = codeTop > workTop ? "top" : "bottom";

  return {
    ...base,
    mode: "glue",
    glue,
    bodyH,
    gapEff: 0,
    scrollComments: bodyH < bodyNat,
  };
}
