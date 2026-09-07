import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

type Tone = "gold" | "mint" | "coral" | "violet" | "sky" | "ink" | "soft";

const TONES: Record<Tone, string> = {
  gold: "border-gold/30 bg-gold/10 text-gold-dark",
  mint: "border-mint/30 bg-mint/10 text-mint-dark",
  coral: "border-coral/30 bg-coral/10 text-coral",
  violet: "border-violet/30 bg-violet/10 text-violet-dark",
  sky: "border-sky/30 bg-sky/10 text-sky-dark",
  ink: "border-ink/10 bg-ink/5 text-ink",
  soft: "border-ink/10 bg-cream text-soft",
};

interface ChipProps {
  children: ReactNode;
  tone?: Tone;
  className?: string;
  pulse?: boolean;
}

export function Chip({ children, tone = "soft", className, pulse }: ChipProps) {
  return (
    <span
      className={cn(
        "inline-flex w-fit items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-bold uppercase tracking-widest",
        TONES[tone],
        pulse && "animate-pulse-soft",
        className,
      )}
    >
      {children}
    </span>
  );
}

export function LiveDot({ color = "coral", className }: { color?: "coral" | "gold" | "mint"; className?: string }) {
  const dot = { coral: "bg-coral", gold: "bg-gold", mint: "bg-mint" }[color];
  return (
    <span className={cn("relative flex size-2", className)}>
      <span className={cn("absolute inline-flex size-full animate-ping rounded-full opacity-75", dot)} />
      <span className={cn("relative inline-flex size-2 rounded-full", dot)} />
    </span>
  );
}
