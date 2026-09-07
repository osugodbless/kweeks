import { Link } from "react-router-dom";
import { cn } from "@/lib/cn";

interface FooterProps {
  variant?: "full" | "player";
  className?: string;
}

export function Footer({ variant = "full", className }: FooterProps) {
  return (
    <footer className={cn("border-t border-ink/5 bg-cream/60", className)}>
      <div
        className={cn(
          "mx-auto flex w-full max-w-6xl flex-col items-center justify-between gap-4 px-4 py-8 text-sm sm:px-6 md:flex-row",
          variant === "player" && "justify-center",
        )}
      >
        <span className="flex items-center gap-2 font-display text-lg font-semibold tracking-tight">
          kweeks<span className="text-coral">.</span>
        </span>

        {variant === "full" ? (
          <>
            <p className="order-3 text-center text-soft md:order-2">
              Live money quiz · Naira only · © {new Date().getFullYear()} Kweeks
            </p>
            <nav className="order-2 flex items-center gap-6 font-bold text-soft md:order-3">
              <Link to="/" className="transition-colors hover:text-ink">Support</Link>
              <Link to="/" className="transition-colors hover:text-ink">Terms</Link>
              <Link to="/" className="transition-colors hover:text-ink">Privacy</Link>
            </nav>
          </>
        ) : (
          <p className="text-center text-sm text-soft">Live money quiz · Play free</p>
        )}
      </div>
    </footer>
  );
}
