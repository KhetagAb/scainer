import type { ProblemStatementExplainView } from "@/client/types.gen";
import { useSession } from "@/features/auth/SessionProvider";

const EXPLAIN_LLM_DEBUG_USERNAME = "khetag_dz";

type Props = {
  data: ProblemStatementExplainView;
};

export default function ExplainLlmDebug({ data }: Props) {
  const { username } = useSession();
  if (username !== EXPLAIN_LLM_DEBUG_USERNAME) return null;
  if (!data.model && !data.prompt) return null;

  return (
    <details className="problem-statement-explain__debug">
      <summary className="problem-statement-explain__debug-summary">
        <span>Debug: LLM</span>
      </summary>
      {data.model ? (
        <p className="problem-statement-explain__debug-line">
          <span className="problem-statement-explain__debug-label">model</span>
          <code className="problem-statement-explain__debug-value">{data.model}</code>
        </p>
      ) : null}
      {data.prompt ? (
        <pre className="problem-statement-explain__debug-prompt">{data.prompt}</pre>
      ) : null}
    </details>
  );
}
