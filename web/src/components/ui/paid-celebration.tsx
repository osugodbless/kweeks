import { useEffect, useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { PartyPopper, ShieldCheck, Sparkles, Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Money } from "@/components/ui/money";

interface PaidCelebrationProps {
  open: boolean;
  onClose: () => void;
  amountNaira: string | number;
  accountLast4?: string;
  bankName?: string;
  payoutRef?: string;
}

const CONFETTI_COLORS = ["#ff4d5f", "#f6a91b", "#16c47f", "#6c4cf1", "#38a8ff"];

// Deterministic 0..1 pseudo-random from an index — pure, so render stays pure.
function prand(n: number): number {
  const x = Math.sin(n * 127.1 + 311.7) * 43758.5453;
  return x - Math.floor(x);
}

/**
 * The winner's reward moment: a full-screen celebration shown the instant a
 * prize is credited — confetti, a floating trophy, the amount counting up, and
 * the settlement receipt. Portalled to <body> so it truly covers the screen.
 */
export function PaidCelebration({ open, onClose, amountNaira, accountLast4, bankName, payoutRef }: PaidCelebrationProps) {
  const amount = typeof amountNaira === "number" ? amountNaira : parseInt(amountNaira || "0", 10) || 0;
  const [shown, setShown] = useState(0);

  // Lock scroll + escape to dismiss while open.
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
    };
  }, [open, onClose]);

  // Amount counts up with a satisfying ease-out. The component is mounted only
  // while open, so `shown` always starts at 0.
  useEffect(() => {
    if (!open) return;
    let raf = 0;
    const start = performance.now();
    const dur = 1000;
    const tick = (t: number) => {
      const p = Math.min(1, (t - start) / dur);
      setShown(Math.round(amount * (1 - Math.pow(1 - p, 3))));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [open, amount]);

  // A fresh burst each time it opens.
  const confetti = useMemo(() => {
    if (!open) return [];
    return Array.from({ length: 64 }, (_, i) => ({
      key: i,
      left: prand(i * 3 + 1) * 100,
      delay: prand(i * 5 + 2) * 1.4,
      duration: 2.6 + prand(i * 7 + 3) * 2,
      size: 7 + prand(i * 11 + 4) * 10,
      round: prand(i * 13 + 5) > 0.6,
      color: CONFETTI_COLORS[i % CONFETTI_COLORS.length],
    }));
  }, [open]);

  if (!open) return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[120] grid place-items-center overflow-hidden p-4"
      role="dialog"
      aria-modal="true"
      aria-label="Payment credited"
    >
      <button
        type="button"
        aria-label="Close"
        onClick={onClose}
        className="absolute inset-0 cursor-default bg-ink/70 backdrop-blur-md"
      />

      {/* Warm glow behind the card */}
      <span className="pointer-events-none absolute left-1/2 top-1/2 size-[44rem] -translate-x-1/2 -translate-y-1/2 rounded-full bg-gold/25 blur-3xl" />
      <span className="pointer-events-none absolute left-[22%] top-[30%] size-80 rounded-full bg-mint/20 blur-3xl" />

      {/* Confetti */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        {confetti.map((c) => (
          <span
            key={c.key}
            className="animate-confetti absolute top-0 block will-change-transform"
            style={{
              left: `${c.left}%`,
              width: c.size,
              height: c.round ? c.size : c.size * 0.6,
              background: c.color,
              borderRadius: c.round ? "50%" : 3,
              animationDuration: `${c.duration}s`,
              animationDelay: `${c.delay}s`,
            }}
          />
        ))}
      </div>

      {/* Card */}
      <div className="relative w-full max-w-xl animate-pop rounded-[2.5rem] border-2 border-gold/40 bg-cream p-7 text-center card-3d sm:p-11">
        <div className="pointer-events-none absolute -right-4 -top-5 grid size-16 animate-float place-items-center rounded-full bg-coral text-white shadow-xl">
          <Sparkles className="size-7" />
        </div>

        <div className="mx-auto grid size-20 animate-float place-items-center rounded-full bg-gold text-ink shadow-[0_14px_34px_-10px_rgba(246,169,27,0.9)]">
          <Trophy className="size-10" />
        </div>

        <p className="mt-6 inline-flex items-center gap-2 rounded-full bg-mint/10 px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-mint-dark">
          <span className="grid size-4 place-items-center rounded-full bg-mint text-white">&#10003;</span>
          Payment credited
        </p>

        <h2 className="mt-4 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
          You&rsquo;ve been <span className="text-shine">paid</span>
        </h2>
        <p className="mt-2 text-lg text-soft">Your prize just landed. Go celebrate.</p>

        <div className="mt-6 rounded-[2rem] border-2 border-mint/30 bg-mint/10 px-6 py-6">
          <p className="text-xs font-bold uppercase tracking-[0.25em] text-mint-dark">Amount credited</p>
          <Money value={shown} className="mt-2 block text-6xl sm:text-7xl" />
        </div>

        <div className="mt-5 grid gap-3 text-left sm:grid-cols-2">
          <div className="rounded-2xl border-2 border-ink/5 bg-white px-5 py-4">
            <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Paid to</p>
            <p className="mt-1 font-bold">
              &#8226;&#8226;&#8226;&#8226; {accountLast4 || "••••"}
              {bankName ? ` · ${bankName}` : ""}
            </p>
          </div>
          <div className="rounded-2xl border-2 border-ink/5 bg-white px-5 py-4">
            <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Reference</p>
            <p className="mt-1 truncate font-mono text-sm font-bold">{payoutRef || "—"}</p>
          </div>
        </div>

        <div className="mt-7 flex flex-col items-center gap-3">
          <Button
            variant="gold"
            size="lg"
            onClick={onClose}
            icon={<PartyPopper className="size-5" />}
            className="w-full"
            autoFocus
          >
            Done — that&rsquo;s my money!
          </Button>
          <p className="flex items-center gap-2 text-xs font-bold text-soft">
            <ShieldCheck className="size-3.5 text-mint-dark" /> Settled and recorded on the host ledger.
          </p>
        </div>
      </div>
    </div>,
    document.body,
  );
}
