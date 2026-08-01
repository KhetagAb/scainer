import { Check, Trash2, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import type { ProblemStatementExplainView } from "@/client/types.gen";
import ExplainFormalizationLoading from "@/features/statements/ExplainFormalizationLoading";
import ExplainLlmDebug from "@/features/statements/ExplainLlmDebug";
import ExplainMarkdown from "@/features/statements/ExplainMarkdown";
import ExplainPanelHeader from "@/features/statements/ExplainPanelHeader";
import { deleteProblemStatementExplain } from "@/features/statements/deleteProblemStatementExplain";
import type { FormalizationView } from "@/features/statements/formalizationView";
import { updateProblemStatementExplain } from "@/features/statements/updateProblemStatementExplain";

type Props = {
  contestId: string;
  problemId: string;
  view: FormalizationView;
  onUpdated: (data: ProblemStatementExplainView) => void;
  onDeleted: () => void;
};

function viewData(view: FormalizationView): ProblemStatementExplainView | null {
  if (view.phase === "typing" || view.phase === "ready") return view.data;
  return null;
}

export default function ProblemStatementExplainPanel({
  contestId,
  problemId,
  view,
  onUpdated,
  onDeleted,
}: Props) {
  const [editing, setEditing] = useState(false);
  const [draftStatement, setDraftStatement] = useState("");
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const data = view.phase === "ready" ? view.data : null;
  const meta = viewData(view);
  const isTyping = view.phase === "typing";
  const displayText =
    view.phase === "typing" ? view.text : view.phase === "ready" ? view.data.statement : "";

  useEffect(() => {
    if (!data || editing) return;
    setDraftStatement(data.statement);
  }, [data, editing]);

  const startEdit = useCallback(() => {
    if (!data) return;
    setDraftStatement(data.statement);
    setSaveError(null);
    setEditing(true);
  }, [data]);

  const cancelEdit = useCallback(() => {
    setSaveError(null);
    setEditing(false);
  }, []);

  const saveEdit = useCallback(() => {
    if (!data) return;
    setSaving(true);
    setSaveError(null);
    void updateProblemStatementExplain(contestId, problemId, {
      title: data.title,
      statement: draftStatement.trim(),
    })
      .then((updated) => {
        onUpdated(updated);
        setEditing(false);
        setSaving(false);
      })
      .catch((err: unknown) => {
        setSaveError(err instanceof Error ? err.message : "Не удалось сохранить");
        setSaving(false);
      });
  }, [contestId, problemId, data, draftStatement, onUpdated]);

  const deleteEdit = useCallback(() => {
    if (!data) return;
    setSaving(true);
    setSaveError(null);
    void deleteProblemStatementExplain(contestId, problemId)
      .then(() => {
        onDeleted();
        setEditing(false);
        setSaving(false);
      })
      .catch((err: unknown) => {
        setSaveError(err instanceof Error ? err.message : "Не удалось удалить");
        setSaving(false);
      });
  }, [contestId, problemId, data, onDeleted]);

  const isDeleteMode = editing && !draftStatement.trim();
  const isFetching = view.phase === "fetching";

  return (
    <aside className="problem-statement-explain" aria-live="polite" aria-label="Формализация условия">
      {isFetching ? <ExplainFormalizationLoading /> : null}
      {view.phase === "error" ? (
        <p className="problem-statement-explain__error">{view.message}</p>
      ) : null}
      {meta && view.phase !== "error" && !isFetching ? (
        <>
          <ExplainPanelHeader
            contestId={contestId}
            problemId={problemId}
            data={meta}
            editing={editing}
            editPending={isTyping}
            onEdit={startEdit}
          />

          <ExplainLlmDebug data={meta} />

          {isTyping || (view.phase === "ready" && !editing) ? (
            <div className="problem-statement-explain__body">
              <ExplainMarkdown>{displayText}</ExplainMarkdown>
            </div>
          ) : editing ? (
            <div className="problem-statement-explain__editor">
              <label className="problem-statement-explain__field">
                <span className="problem-statement-explain__field-label">Формализация (Markdown)</span>
                <textarea
                  className="problem-statement-explain__textarea"
                  value={draftStatement}
                  onChange={(e) => setDraftStatement(e.target.value)}
                  rows={14}
                />
              </label>
              {saveError ? <p className="problem-statement-explain__error">{saveError}</p> : null}
              <div className="problem-statement-explain__editor-actions">
                <button
                  type="button"
                  className="problem-statement-explain__action-btn"
                  disabled={saving}
                  onClick={cancelEdit}
                >
                  <X size={16} strokeWidth={2} aria-hidden />
                  Отмена
                </button>
                <button
                  type="button"
                  className={
                    "problem-statement-explain__action-btn problem-statement-explain__action-btn--primary" +
                    (isDeleteMode ? " problem-statement-explain__action-btn--danger" : "")
                  }
                  disabled={saving}
                  onClick={isDeleteMode ? deleteEdit : saveEdit}
                >
                  {isDeleteMode ? (
                    <Trash2 size={16} strokeWidth={2} aria-hidden />
                  ) : (
                    <Check size={16} strokeWidth={2} aria-hidden />
                  )}
                  {saving
                    ? isDeleteMode
                      ? "Удаляем…"
                      : "Сохраняем…"
                    : isDeleteMode
                      ? "Удалить"
                      : "Сохранить"}
                </button>
              </div>
            </div>
          ) : null}
        </>
      ) : null}
    </aside>
  );
}
