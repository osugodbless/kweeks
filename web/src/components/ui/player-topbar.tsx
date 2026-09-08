import { Hash } from "lucide-react";
import { cn } from "@/lib/cn";
import { Chip, LiveDot } from "@/components/ui/chip";
import { Wordmark } from "@/components/ui/wordmark";

export type PlayerStatus = "join" | "lobby" | "live" | "standings" | "podium" | "claim";

const STATUS: Record<PlayerStatus, { label: string; tone: "gold" | "coral" | "violet" | "mint"; dot?: boolean }> = {
  join: { label: "Live money quiz", tone: "violet" },
  lobby: { label: "Waiting", tone: "gold", dot: true },
  live: { label: "Live", tone: "coral", dot: true },
  standings: { label: "Live", tone: "coral", dot: true },
  podium: { label: "Game over", tone: "mint" },
  claim: { label: "Claim payout", tone: "gold" },
};

interface PlayerTopBarProps {
  status: PlayerStatus;
  code?: string | null;
  className?: string;
}

export function PlayerTopBar({ status, code, className }: PlayerTopBarProps) {
  const meta = STATUS[status];
  return (
    <header className={cn("sticky top-0 z-50 border-b border-ink/5 bg-canvas/85 backdrop-blur-md", className)}>
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4 sm:px-6">
        <Wordmark />
        <div className="flex items-center gap-2.5">
          {code && (
            <span className="flex items-center gap-1.5 rounded-full border border-ink/10 bg-cream px-3 py-1.5 font-display text-sm font-semibold tracking-wide text-ink">
              <Hash className="size-3.5 text-soft" />
              {code}
            </span>
          )}
          <Chip tone={meta.tone}>
            {meta.dot && <LiveDot color={meta.tone === "gold" ? "gold" : "coral"} />}
            {meta.label}
          </Chip>
        </div>
      </div>
    </header>
  );
}
