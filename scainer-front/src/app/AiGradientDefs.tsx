/** Shared SVG gradient for AI icons (stroke via url(#scainer-ai-gradient)). */
export default function AiGradientDefs() {
  return (
    <svg aria-hidden="true" width={0} height={0} className="ai-gradient-defs">
      <defs>
        <linearGradient id="scainer-ai-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="var(--logo-ai-gradient-from)" />
          <stop offset="48%" stopColor="var(--accent-violet)" />
          <stop offset="100%" stopColor="var(--logo-ai-gradient-to)" />
        </linearGradient>
      </defs>
    </svg>
  );
}
