import { Radar } from "lucide-react";

type Props = {
  size?: number;
  className?: string;
};

export function ContestSyncIcon({ size, className }: Props) {
  return (
    <Radar
      {...(size !== undefined ? { size } : {})}
      className={className}
      strokeWidth={2}
      aria-hidden
    />
  );
}
