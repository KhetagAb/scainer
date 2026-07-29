import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { postContestsMutation } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";
import {
  parseEjudgeContestTable,
  parallelLabel,
  type ParsedImportRow,
  type SkippedImportRow,
} from "@/features/contests/shared/parallels";
import { UNGROUPED_PARALLEL } from "@/features/contests/shared/contestHelpers";

type Props = {
  onSuccess: () => void;
};

type ImportOutcome = { row: ParsedImportRow; kind: "added" | "duplicate" | "error" };

type Step = "paste" | "preview" | "done";

export default function ImportContestsPanel({ onSuccess }: Props) {
  const [isOpen, setIsOpen] = useState(false);
  const [raw, setRaw] = useState("");
  const [step, setStep] = useState<Step>("paste");
  const [rows, setRows] = useState<ParsedImportRow[]>([]);
  const [skipped, setSkipped] = useState<SkippedImportRow[]>([]);
  const [isImporting, setIsImporting] = useState(false);
  const [outcomes, setOutcomes] = useState<ImportOutcome[]>([]);

  const mutation = useMutation(postContestsMutation());

  const close = () => {
    setIsOpen(false);
    setRaw("");
    setStep("paste");
    setRows([]);
    setSkipped([]);
    setOutcomes([]);
  };

  const handleParse = () => {
    const result = parseEjudgeContestTable(raw);
    setRows(result.rows);
    setSkipped(result.skipped);
    setStep("preview");
  };

  const handleConfirm = async () => {
    setIsImporting(true);
    const results: ImportOutcome[] = [];
    for (const row of rows) {
      try {
        await mutation.mutateAsync({
          body: { id: row.id, parallelId: row.parallelId },
          headers: authHeaders(),
        });
        results.push({ row, kind: "added" });
      } catch (e) {
        const isDuplicate = (e as { status?: number })?.status === 400;
        results.push({ row, kind: isDuplicate ? "duplicate" : "error" });
      }
    }
    setIsImporting(false);
    setOutcomes(results);
    setStep("done");
    onSuccess();
  };

  if (!isOpen) {
    return (
      <button type="button" className="btn import-dock__btn" onClick={() => setIsOpen(true)}>
        Импорт
      </button>
    );
  }

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && step !== "done" && close()}>
      <div className="form-card modal-card">
        {step === "paste" && (
          <>
            <p className="control-group__label" style={{ marginBottom: "0.5rem" }}>
              Вставьте табличку списка контестов из ejudge
            </p>
            <textarea
              value={raw}
              onChange={(e) => setRaw(e.target.value)}
              rows={10}
              placeholder={"2\t50050\tЛКШ.2026.Параллель R.Template\t...\n3\t50051\tЛКШ.2026.Параллель R.День 01.Разнобой\t..."}
              style={{
                width: "100%",
                boxSizing: "border-box",
                fontFamily: "var(--font-body)",
                padding: "0.5rem 0.65rem",
                border: "1px solid var(--chip-border)",
                borderRadius: "var(--radius-chip)",
                marginBottom: "1rem",
              }}
            />
            <div className="form-actions">
              <button type="button" className="btn btn--primary" disabled={!raw.trim()} onClick={handleParse}>
                Разобрать
              </button>
              <button type="button" className="btn btn--ghost" onClick={close}>
                Отмена
              </button>
            </div>
          </>
        )}

        {step === "preview" && (
          <>
            <p className="control-group__label" style={{ marginBottom: "0.5rem" }}>
              Будет импортировано: {rows.length}. Пропущено: {skipped.length}.
            </p>
            <div className="modal-card__scroll">
              <ul className="contest-list">
                {rows.map((r) => (
                  <li key={r.id} className="contest-row">
                    <div>
                      <div className="contest-row__id">{r.name}</div>
                      <div className="contest-row__meta">ID: {r.id}</div>
                    </div>
                    <span className="chip contest-chip">
                      {parallelLabel(r.parallelId, UNGROUPED_PARALLEL)}
                    </span>
                  </li>
                ))}
                {skipped.map((s) => (
                  <li key={s.id} className="contest-row contest-row--skipped">
                    <div>
                      <div className="contest-row__id">{s.name}</div>
                      <div className="contest-row__meta">ID: {s.id}</div>
                    </div>
                    <span className="chip">пропущено: {s.reason}</span>
                  </li>
                ))}
              </ul>
            </div>
            <div className="form-actions">
              <button
                type="button"
                className="btn btn--primary"
                disabled={rows.length === 0 || isImporting}
                onClick={() => void handleConfirm()}
              >
                {isImporting ? "Импортирую…" : `Импортировать ${rows.length} контестов`}
              </button>
              <button type="button" className="btn btn--ghost" onClick={() => setStep("paste")} disabled={isImporting}>
                Назад
              </button>
              <button type="button" className="btn btn--ghost" onClick={close} disabled={isImporting}>
                Отмена
              </button>
            </div>
          </>
        )}

        {step === "done" && (
          <>
            <p className="control-group__label" style={{ marginBottom: "0.5rem" }}>
              Импортировано: {outcomes.filter((o) => o.kind === "added").length}
              {outcomes.some((o) => o.kind === "duplicate")
                ? `, уже были добавлены: ${outcomes.filter((o) => o.kind === "duplicate").length}`
                : ""}
              {outcomes.some((o) => o.kind === "error")
                ? `, с ошибкой: ${outcomes.filter((o) => o.kind === "error").length}`
                : ""}
            </p>
            <div className="modal-card__scroll">
              <ul className="contest-list">
                {outcomes.map((o) => (
                  <li
                    key={o.row.id}
                    className={`contest-row${
                      o.kind === "error" ? " contest-row--error" : o.kind === "duplicate" ? " contest-row--skipped" : ""
                    }`}
                  >
                    <div>
                      <div className="contest-row__id">{o.row.name}</div>
                      <div className="contest-row__meta">
                        ID: {o.row.id} · {parallelLabel(o.row.parallelId, UNGROUPED_PARALLEL)}
                        {o.kind === "duplicate" ? " · уже был добавлен" : ""}
                        {o.kind === "error" ? " · ошибка" : ""}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
            <div className="form-actions">
              <button type="button" className="btn" onClick={close}>
                Закрыть
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
