import { useRef, useState, type FormEvent, type KeyboardEvent, type MouseEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { postContestsMutation } from "@/client/@tanstack/react-query.gen";
import { authHeaders } from "@/features/auth/authStorage";

type Props = {
  parallelId: string;
  parallelName: string;
  onDone: () => void;
};

// Карточка в сетке контестов: курсив «Добавить контест» + поле ID.
export default function AddContestForm({ parallelId, parallelName, onDone }: Props) {
  const [contestId, setContestId] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);
  const mutation = useMutation(postContestsMutation());
  const pending = mutation.isPending;

  const focusInput = (e: MouseEvent) => {
    if ((e.target as HTMLElement).closest("input")) return;
    e.preventDefault();
    inputRef.current?.focus();
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const id = contestId.trim();
    if (!id || pending) return;
    try {
      await mutation.mutateAsync({
        body: { id, parallelId, parallelName },
        headers: authHeaders(),
      });
      setContestId("");
      onDone();
    } catch (error) {
      if ((error as { status?: number })?.status === 400) {
        setContestId("");
        onDone();
        return;
      }
      console.error("Failed to add contest:", error);
    }
  };

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Escape" && !pending) {
      setContestId("");
      (e.target as HTMLInputElement).blur();
    }
  };

  return (
    <form
      className={`contest-card contest-card--add${pending ? " is-busy" : ""}`}
      onSubmit={(e) => void handleSubmit(e)}
      onMouseDown={focusInput}
      aria-busy={pending}
    >
      <span className="contest-card__name contest-card__name--add">
        {pending ? "Добавление…" : "Добавить контест"}
      </span>
      <input
        ref={inputRef}
        className="contest-card__id-input"
        type="text"
        value={contestId}
        onChange={(e) => setContestId(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder="ID, напр. 50601"
        required
        readOnly={pending}
        aria-label={`ID контеста для ${parallelName}`}
      />
      {mutation.isError && !pending && (
        <span className="contest-card__error">Не удалось добавить</span>
      )}
    </form>
  );
}
