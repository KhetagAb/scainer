import { useEffect, useId, useRef, useState, type MouseEvent } from "react";
import { createPortal } from "react-dom";
import confetti from "canvas-confetti";
import { readConfettiColors } from "@/app/theme/confettiColors";

const AUTO_CLOSE_MS = 2400;
const AUTO_CLOSE_WITH_NEXT_MS = 3600;

type Props = {
  open: boolean;
  onClose: () => void;
  problemLabel: string;
  nextProblemId: string | null;
  onNextProblem: () => void;
};

function ArrowRightIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <path d="M5 12h14" />
      <path d="m12 5 7 7-7 7" />
    </svg>
  );
}

function fireBursts() {
  const colors = readConfettiColors();
  const common = {
    disableForReducedMotion: true,
    zIndex: 10000,
    colors: colors.length ? colors : undefined,
  } as const;

  void confetti({
    ...common,
    particleCount: 90,
    spread: 70,
    origin: { x: 0.2, y: 0.65 },
  });
  void confetti({
    ...common,
    particleCount: 90,
    spread: 70,
    origin: { x: 0.8, y: 0.65 },
  });
  window.setTimeout(() => {
    void confetti({
      ...common,
      particleCount: 60,
      spread: 100,
      startVelocity: 35,
      origin: { x: 0.5, y: 0.4 },
      colors: colors.length > 3 ? colors.slice(0, 4) : colors,
    });
  }, 90);
}

export default function ReviewCelebrateOverlay({
  open,
  onClose,
  problemLabel,
  nextProblemId,
  onNextProblem,
}: Props) {
  const titleId = useId();
  const onCloseRef = useRef(onClose);
  const nextHoverRef = useRef(false);
  const closeWhenLeaveRef = useRef(false);
  const [mounted, setMounted] = useState(open);
  const [visible, setVisible] = useState(false);
  const hasNext = Boolean(nextProblemId);

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    if (open) {
      nextHoverRef.current = false;
      closeWhenLeaveRef.current = false;
      setMounted(true);
      const id = requestAnimationFrame(() => {
        requestAnimationFrame(() => setVisible(true));
      });
      fireBursts();
      const ms = hasNext ? AUTO_CLOSE_WITH_NEXT_MS : AUTO_CLOSE_MS;
      const auto = window.setTimeout(() => {
        if (nextHoverRef.current) {
          closeWhenLeaveRef.current = true;
          return;
        }
        onCloseRef.current();
      }, ms);
      return () => {
        cancelAnimationFrame(id);
        window.clearTimeout(auto);
      };
    }
    setVisible(false);
    const t = window.setTimeout(() => setMounted(false), 110);
    return () => window.clearTimeout(t);
  }, [open, hasNext]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCloseRef.current();
    };
    window.addEventListener("keydown", onKey);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prevOverflow;
    };
  }, [open]);

  const onNextClick = (e: MouseEvent<HTMLButtonElement>) => {
    e.stopPropagation();
    closeWhenLeaveRef.current = false;
    onClose();
    onNextProblem();
  };

  const onNextEnter = () => {
    nextHoverRef.current = true;
  };

  const onNextLeave = () => {
    nextHoverRef.current = false;
    if (closeWhenLeaveRef.current) {
      closeWhenLeaveRef.current = false;
      onClose();
    }
  };

  if (!mounted) return null;

  return createPortal(
    <div
      className={
        "review-celebrate" + (visible ? " review-celebrate--visible" : "")
      }
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onClick={() => onClose()}
    >
      <div className="review-celebrate__copy">
        <p id={titleId} className="review-celebrate__title">
          Вы великолепны!
        </p>
        <p className="review-celebrate__subtitle">
          Вы проверили{" "}
          <span className="review-celebrate__problem">{problemLabel}</span>
        </p>
        {hasNext ? (
          <button
            type="button"
            className="review-celebrate__next"
            onClick={onNextClick}
            onMouseEnter={onNextEnter}
            onMouseLeave={onNextLeave}
          >
            Следующая задача
            <ArrowRightIcon />
          </button>
        ) : null}
      </div>
    </div>,
    document.body,
  );
}
