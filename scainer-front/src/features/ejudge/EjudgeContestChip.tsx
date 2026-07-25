import type { ReactNode } from "react";
import { EjudgeChip } from "@/features/ejudge/EjudgeChip";
import { useEjudgeContest } from "@/features/ejudge/useEjudgeContest";

type Props = {
  children: ReactNode;
  className?: string;
  title?: string;
};

export function EjudgeContestChip({ children, className, title }: Props) {
  const { available, isLoading, openContest } = useEjudgeContest();
  return (
    <EjudgeChip
      className={className}
      title={
        isLoading
          ? "загрузка ejudge…"
          : available
            ? "Открыть контест в ejudge"
            : title
      }
      onClick={available && !isLoading ? openContest : undefined}
    >
      {children}
    </EjudgeChip>
  );
}
