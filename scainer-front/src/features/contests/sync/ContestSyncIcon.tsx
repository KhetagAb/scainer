import { Radar } from "lucide-react";

type Props = {
  size?: number;
  className?: string;
};

export function ContestSyncIcon({ size = 20, className }: Props) {
  return (
    <Radar
      size={size}
      className={className}
      strokeWidth={2}
      aria-hidden
    />
  );
}
