import { useEffect, useRef, useState } from "react";

const TICK_MS = 20;
const TARGET_DURATION_MS = 2500;

type Options = {
  onComplete?: () => void;
};

function prefersReducedMotion(): boolean {
  if (typeof window === "undefined") return false;
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/** Постепенно раскрывает текст — анимация печати при первом показе формализации. */
export function useTypewriterText(text: string | null, options: Options = {}): { displayed: string } {
  const { onComplete } = options;
  const onCompleteRef = useRef(onComplete);
  onCompleteRef.current = onComplete;

  const [displayed, setDisplayed] = useState("");

  useEffect(() => {
    if (!text) {
      setDisplayed("");
      return;
    }

    if (prefersReducedMotion()) {
      setDisplayed(text);
      onCompleteRef.current?.();
      return;
    }

    setDisplayed("");
    let index = 0;
    const charsPerTick = Math.max(1, Math.ceil(text.length / (TARGET_DURATION_MS / TICK_MS)));
    const id = window.setInterval(() => {
      index = Math.min(text.length, index + charsPerTick);
      setDisplayed(text.slice(0, index));
      if (index >= text.length) {
        window.clearInterval(id);
        onCompleteRef.current?.();
      }
    }, TICK_MS);

    return () => window.clearInterval(id);
  }, [text]);

  return { displayed: text ? displayed : "" };
}
