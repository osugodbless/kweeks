import { Link } from "react-router-dom";
import { Zap } from "lucide-react";
import { cn } from "@/lib/cn";

interface WordmarkProps {
  to?: string;
  className?: string;
  dot?: boolean;
}

export function Wordmark({ to = "/", className, dot = true }: WordmarkProps) {
  const inner = (
    <>
      <span className="grid size-9 place-items-center rounded-xl bg-ink text-cream shadow-[0_4px_0_rgba(0,0,0,0.15)] transition-transform duration-200 group-hover:-rotate-6 group-hover:scale-105">
        <Zap className="size-5" strokeWidth={2.5} fill="currentColor" />
      </span>
      <span className="font-display text-2xl font-semibold tracking-tight">
        kweeks{dot && <span className="text-coral">.</span>}
      </span>
    </>
  );

  if (to) {
    return (
      <Link to={to} className={cn("group flex w-fit items-center gap-2.5", className)} aria-label="Kweeks home">
        {inner}
      </Link>
    );
  }
  return <div className={cn("flex items-center gap-2.5", className)}>{inner}</div>;
}
