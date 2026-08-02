import { useId } from "react";

export function Logo() {
  const gradientId = useId().replace(/:/g, "");

  return (
    <svg
      className="app-logo"
      role="img"
      aria-label="scAIner"
      viewBox="0 0 108 26"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <linearGradient id={gradientId} x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="var(--logo-ai-gradient-from)" />
          <stop offset="48%" stopColor="var(--accent-violet)" />
          <stop offset="100%" stopColor="var(--logo-ai-gradient-to)" />
        </linearGradient>
      </defs>
      <text
        x="0"
        y="20.5"
        fontFamily="'Montserrat Alternates', 'Segoe UI', sans-serif"
        fontSize="26"
        fontWeight="100"
        letterSpacing="-0.52"
        fill="currentColor"
      >
        <tspan>sc</tspan>
        <tspan fill={`url(#${gradientId})`}>AI</tspan>
        <tspan fill="currentColor">ner</tspan>
      </text>
    </svg>
  );
}
