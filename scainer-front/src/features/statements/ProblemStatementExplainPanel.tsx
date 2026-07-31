import type { ProblemStatementExplainView } from "@/client/types.gen";

type Props = {
  data: ProblemStatementExplainView | null;
  loading: boolean;
  error: string | null;
};

export default function ProblemStatementExplainPanel({ data, loading, error }: Props) {
  return (
    <aside
      className="problem-statement-explain"
      aria-live="polite"
      aria-label="Разбор условия"
    >
      {loading ? <p className="problem-statement-explain__status">Разбираем условие…</p> : null}
      {error ? <p className="problem-statement-explain__error">{error}</p> : null}
      {data && !error ? (
        <>
          <p className="problem-statement-explain__source">Распознано из PDF</p>
          <h2 className="problem-statement-explain__title">{data.title}</h2>
          <div className="problem-statement-explain__body">{data.statement}</div>
        </>
      ) : null}
    </aside>
  );
}
