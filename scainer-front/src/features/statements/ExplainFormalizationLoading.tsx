import { DotLottieReact } from "@lottiefiles/dotlottie-react";
import { Pencil } from "lucide-react";
import aiTwinkleLoading from "@/assets/ai-twinkle-loading.lottie?url";

export default function ExplainFormalizationLoading() {
  return (
    <>
      <div className="problem-statement-explain__head">
        <div className="problem-statement-explain__chips">
          <span className="chip problem-statement-explain__source-chip problem-statement-explain__source-chip--ai">
            Сформировано AI
          </span>
        </div>
        <button
          type="button"
          className="problem-statement-explain__edit-btn problem-statement-explain__edit-btn--pending"
          aria-hidden
          tabIndex={-1}
          disabled
        >
          <Pencil size={16} strokeWidth={2} aria-hidden />
        </button>
      </div>
      <div className="problem-statement-explain__body problem-statement-explain__body--loading">
        <div className="problem-statement-explain__loading">
          <div className="problem-statement-explain__loading-lottie">
            <DotLottieReact
              src={aiTwinkleLoading}
              loop
              autoplay
              className="problem-statement-explain__loading-lottie-player"
            />
          </div>
          <p className="problem-statement-explain__loading-label">Формализуем условие...</p>
        </div>
      </div>
    </>
  );
}
