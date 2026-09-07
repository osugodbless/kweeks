import type { ReactNode } from "react";
import { Coins } from "lucide-react";
import { cn } from "@/lib/cn";
import { Money } from "@/components/ui/money";
import { Wordmark } from "@/components/ui/wordmark";

interface AuthShellProps {
  children: ReactNode;
  tagline?: string;
  chipAmount?: string;
  className?: string;
}

export function AuthShell({ children, tagline, chipAmount, className }: AuthShellProps) {
  return (
    <main className="relative flex min-h-screen items-center overflow-hidden py-14">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-16 size-72 rounded-full bg-violet/20 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-16 size-80 rounded-full bg-coral/20 blur-3xl" />
      <span className="pointer-events-none absolute left-1/2 top-0 size-64 rounded-full bg-gold/10 blur-3xl" />

      <div className="relative mx-auto grid w-full max-w-5xl items-center gap-10 px-4 sm:px-6 lg:grid-cols-[0.9fr_1.1fr]">
        {/* Brand panel */}
        <aside className="relative hidden overflow-hidden rounded-[2.5rem] bg-ink p-10 text-cream lg:block">
          <span className="pointer-events-none absolute inset-0 bg-dots-light" />
          <div className="relative">
            <Wordmark to="/" dot={false} />
            <h1 className="mt-8 font-display text-4xl font-semibold leading-tight tracking-tight">
              Fund the pool. <br />
              Run the room. <br />
              <span className="text-shine">Pay the winners.</span>
            </h1>
            <p className="mt-4 max-w-sm text-white/60">
              One instructor account holds your wallet, your quizzes and your full
              payout history in one place.
            </p>
            <div className="mt-10 flex items-center gap-3 rounded-3xl border border-white/10 bg-white/5 p-5">
              <span className="grid size-12 shrink-0 place-items-center rounded-2xl bg-gold text-ink">
                <Coins className="size-6" fill="currentColor" />
              </span>
              <div>
                <p className="text-xs font-bold uppercase tracking-widest text-white/50">{tagline ?? "Wallet ready"}</p>
                <Money value={chipAmount ?? "0"} tone="dark" className="text-2xl" />
              </div>
            </div>
          </div>
        </aside>

        {/* Form */}
        <section className={cn("relative w-full max-w-md justify-self-center lg:justify-self-stretch", className)}>
          {children}
        </section>
      </div>
    </main>
  );
}
