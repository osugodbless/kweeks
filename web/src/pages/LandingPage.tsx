import { Link } from "react-router-dom";
import type { CSSProperties } from "react";
import {
  ArrowRight,
  Banknote,
  Coins,
  Crown,
  Gift,
  Hash,
  Medal,
  PartyPopper,
  Play,
  ShieldCheck,
  Sparkles,
  Timer,
  Trophy,
  Users,
  Wallet,
  Zap,
} from "lucide-react";
import { Avatar } from "@/components/ui/avatar";
import { Money } from "@/components/ui/money";
import { LiveDot } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { Navbar } from "@/components/ui/navbar";

const LEADERBOARD = [
  { name: "Ada", points: "1,240", avatarId: "fish" },
  { name: "Tobi", points: "980", avatarId: "cat" },
  { name: "Zainab", points: "815", avatarId: "rocket" },
];

const STEPS = [
  {
    Icon: Banknote,
    title: "Create & fund",
    copy: "A host picks a quiz and puts real naira in the prize pool. No tokens, no fine print — one wallet, one ledger.",
  },
  {
    Icon: Hash,
    title: "Open the room",
    copy: "Players join on their phones with a 4-letter code. Everyone answers the same question at the same second.",
  },
  {
    Icon: Crown,
    title: "Pay the podium",
    copy: "Speed and accuracy decide the top finishers. Winners redeem their share right from their own screen.",
  },
];

const INSTRUCTOR_FEATURES = [
  { Icon: Wallet, title: "Fund a pool in seconds", copy: "Credit your wallet instantly and put real money on every question." },
  { Icon: Users, title: "Rooms players love", copy: "A join code, an avatar, zero installs. The crowd is live in seconds." },
  { Icon: Trophy, title: "You decide the winners", copy: "Winner count is yours. The podium splits the pool exactly how you set it." },
  { Icon: Banknote, title: "A clean ledger", copy: "Every funding, room and payout shows up in your history. Nothing sneaks." },
];

const PLAYER_POINTS = [
  { Icon: Coins, text: "Real money, real fast — no signup walls" },
  { Icon: Hash, text: "Join with a code your host shows on screen" },
  { Icon: Zap, text: "Every question is live; speed is the sport" },
  { Icon: ShieldCheck, text: "Your win is yours — escrow until the podium" },
];

const SECURITY = [
  { Icon: Coins, title: "Naira only", copy: "One currency, no conversion surprises." },
  { Icon: ShieldCheck, title: "Escrow while live", copy: "The pool is locked in a real wallet until winners are declared." },
  { Icon: Gift, title: "Winner-only claims", copy: "Only the podium can claim, only with their own session." },
  { Icon: Banknote, title: "Full history", copy: "Every activity is recorded on the instructor ledger." },
];

export function LandingPage() {
  return (
    <main>
      <Navbar />

      {/* HERO */}
      <section className="relative overflow-hidden">
        <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
        <span className="pointer-events-none absolute -left-20 top-20 size-72 rounded-full bg-violet/20 blur-3xl" />
        <span className="pointer-events-none absolute -right-20 bottom-10 size-80 rounded-full bg-coral/20 blur-3xl" />

        <div className="relative mx-auto grid w-full max-w-6xl items-center gap-14 px-4 pb-20 pt-14 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:pb-28 lg:pt-20">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full bg-cream px-4 py-2 text-sm font-bold text-ink shadow-[0_2px_10px_rgba(36,22,63,0.08)]">
              <PartyPopper className="size-4 text-coral" />
              Live money quiz — real naira on every question
            </span>

            <h1 className="mt-6 font-display text-5xl font-semibold leading-[0.98] tracking-tight sm:text-6xl lg:text-7xl">
              Answer fast.
              <br />
              <span className="text-shine">Take the pool.</span>
            </h1>

            <p className="mt-6 max-w-xl text-lg leading-relaxed text-soft">
              Instructors fund a naira prize pool, players join on their phones with a
              code, and the fastest minds split the pot. Same second. Same question.
              Real money out.
            </p>

            <div className="mt-9 flex flex-wrap items-center gap-4">
              <Link
                to="/join"
                className="inline-flex items-center gap-2.5 rounded-full bg-coral px-8 py-4 font-display text-lg font-semibold text-white press-3d"
                style={{ "--btn-deep-rgb": "var(--color-coral-dark)" } as CSSProperties}
              >
                <Play className="size-5 fill-current" />
                Join a quiz
              </Link>
              <Link
                to="/instructor/signup"
                className="inline-flex items-center gap-2.5 rounded-full bg-violet px-8 py-4 font-display text-lg font-semibold text-white press-3d"
                style={{ "--btn-deep-rgb": "var(--color-violet-dark)" } as CSSProperties}
              >
                Start hosting
                <ArrowRight className="size-5 transition-transform group-hover:translate-x-1" />
              </Link>
            </div>

            <ul className="mt-9 flex flex-wrap gap-x-7 gap-y-3 text-sm font-bold text-soft">
              <li className="flex items-center gap-2">
                <Hash className="size-4 text-violet" /> Join with a 4-letter code
              </li>
              <li className="flex items-center gap-2">
                <Timer className="size-4 text-coral" /> Live rounds, seconds long
              </li>
              <li className="flex items-center gap-2">
                <ShieldCheck className="size-4 text-mint-dark" /> Escrow-held prize pools
              </li>
            </ul>
          </div>

          {/* Live room mock */}
          <div className="relative animate-rise" style={{ animationDelay: "0.18s" }}>
            <span className="pointer-events-none absolute -left-6 -top-8 size-28 rounded-full bg-sky/25 blur-2xl" />
            <span className="pointer-events-none absolute -bottom-10 -right-6 size-36 rounded-full bg-coral/25 blur-2xl" />

            <div className="relative mx-auto max-w-md rotate-2 rounded-[2rem] bg-ink p-6 text-cream card-3d transition-transform duration-300 hover:rotate-0 sm:p-7">
              <span className="pointer-events-none absolute inset-0 rounded-[2rem] bg-dots-light opacity-60" />

              <div className="relative flex items-center justify-between">
                <span className="inline-flex items-center gap-2 rounded-full bg-white/10 px-3 py-1.5 text-xs font-bold uppercase tracking-widest">
                  <LiveDot color="coral" />
                  Live · Question 3
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-full bg-cream/10 px-3 py-1.5 text-xs font-bold uppercase tracking-widest text-mint">
                  <Coins className="size-3.5" fill="currentColor" />
                  Pool
                </span>
              </div>

              <div className="relative mt-6">
                <p className="text-xs font-bold uppercase tracking-[0.25em] text-white/50">Join with room code</p>
                <div className="mt-3 flex gap-2.5" aria-label="Room code AB12">
                  {["A", "B", "1", "2"].map((d, i) => (
                    <span
                      key={i}
                      className="grid size-12 place-items-center rounded-xl bg-cream font-display text-2xl font-semibold text-ink shadow-[inset_0_-4px_0_rgba(36,22,63,0.15)] sm:size-13"
                    >
                      {d}
                    </span>
                  ))}
                </div>
              </div>

              <div className="relative mt-5 flex items-center justify-between rounded-2xl bg-white/5 px-4 py-3">
                <span className="text-xs font-bold uppercase tracking-widest text-white/60">Prize pool</span>
                <Money value="50000" tone="dark" className="text-xl" />
              </div>

              <div className="relative mt-4 rounded-2xl bg-white/5 p-4">
                <div className="mb-3 flex items-center justify-between">
                  <span className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-white/60">
                    <Trophy className="size-4 text-gold" /> Live standings
                  </span>
                  <span className="text-xs font-bold text-gold">+500 pts / round</span>
                </div>
                <ul className="space-y-2">
                  {LEADERBOARD.map((row, i) => (
                    <li key={row.name} className="flex items-center gap-3 rounded-xl bg-white/5 px-3 py-2">
                      <span className="grid size-6 place-items-center">
                        {i === 0 ? (
                          <Trophy className="size-4 text-gold" fill="currentColor" />
                        ) : (
                          <Medal className={`size-4 ${i === 1 ? "text-white/80" : "text-gold/60"}`} />
                        )}
                      </span>
                      <Avatar id={row.avatarId} className="size-8" />
                      <span className="flex-1 truncate text-sm font-bold">{row.name}</span>
                      <span className="font-display text-sm font-semibold text-gold">{row.points}</span>
                    </li>
                  ))}
                </ul>
              </div>

              <div className="relative mt-4 flex items-center justify-between text-xs font-bold text-white/60">
                <span className="inline-flex items-center gap-1.5">
                  <Timer className="size-4 text-coral" /> 12s left
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <Sparkles className="size-4 text-sky" /> Streak ×4
                </span>
              </div>
            </div>

            <span className="absolute -right-3 -top-7 grid size-14 animate-float place-items-center rounded-full bg-gold text-ink shadow-xl lg:-right-6">
              <Coins className="size-7" fill="currentColor" />
            </span>
            <span className="absolute -bottom-6 -left-4 flex animate-float-slow items-center gap-2 rounded-2xl bg-cream px-4 py-2.5 font-bold card-3d lg:-left-8">
              <Trophy className="size-5 text-coral" />
              <span className="text-sm">
                Top <span className="font-display text-coral">3</span> split the{" "}
                <Money value="50000" className="text-sm" />
              </span>
            </span>
          </div>
        </div>
      </section>

      {/* HOW IT WORKS */}
      <section id="how-it-works" className="relative scroll-mt-20 border-y border-ink/5 bg-cream/60 py-20 lg:py-24">
        <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
          <div className="mx-auto max-w-2xl text-center">
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-wider text-soft">
              <Sparkles className="size-4 text-gold" /> How it works
            </span>
            <h2 className="mt-5 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
              One code. Any room. <span className="text-pop">Real prizes.</span>
            </h2>
            <p className="mt-4 text-lg text-soft">
              Hosts fund the pool and open a room. Players jump in by code, answer live, and
              the fastest correct minds take it home.
            </p>
          </div>

          <div className="mt-14 grid gap-6 md:grid-cols-3">
            {STEPS.map((step, i) => (
              <article
                key={step.title}
                className="group relative flex flex-col overflow-hidden rounded-3xl border-2 border-ink/5 bg-cream p-8 transition-all duration-300 hover:-translate-y-2 card-3d"
                style={{ animation: "rise 0.6s cubic-bezier(0.16,1,0.3,1) both", animationDelay: `${i * 0.12 + 0.1}s` }}
              >
                <span className="pointer-events-none absolute right-5 top-4 font-display text-5xl font-bold text-ink/5 transition-colors group-hover:text-ink/10">
                  {String(i + 1).padStart(2, "0")}
                </span>
                <span className="mb-5 grid size-14 place-items-center rounded-2xl bg-coral text-white shadow-lg transition-transform duration-300 group-hover:-rotate-6 group-hover:scale-110">
                  <step.Icon className="size-7" strokeWidth={2.2} />
                </span>
                <h3 className="font-display text-2xl font-semibold tracking-tight">{step.title}</h3>
                <p className="mt-2 text-[15px] leading-relaxed text-soft">{step.copy}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      {/* FOR PLAYERS */}
      <section id="for-players" className="scroll-mt-20 py-20 lg:py-24">
        <div className="mx-auto grid w-full max-w-6xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-wider text-soft">
              <Users className="size-4 text-violet" /> For players
            </span>
            <h2 className="mt-5 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
              Real money, <span className="text-pop">real fast</span>
            </h2>
            <p className="mt-4 text-lg leading-relaxed text-soft">
              No app to download, no account to make. Grab the code on the screen, pick an
              avatar, and race the clock.
            </p>
            <ul className="mt-8 space-y-3.5">
              {PLAYER_POINTS.map((p) => (
                <li key={p.text} className="flex items-center gap-3 text-[15px] font-semibold">
                  <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-violet/10 text-violet">
                    <p.Icon className="size-4.5" strokeWidth={2.4} />
                  </span>
                  {p.text}
                </li>
              ))}
            </ul>
          </div>
        </div>
      </section>

      {/* FOR INSTRUCTORS */}
      <section id="for-instructors" className="scroll-mt-20 pb-20 lg:pb-24">
        <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
          <div className="max-w-2xl">
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-wider text-soft">
              <Wallet className="size-4 text-gold-dark" /> For instructors
            </span>
            <h2 className="mt-5 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
              Run a room. <span className="text-shine">Fund a pool.</span>
            </h2>
            <p className="mt-4 text-lg leading-relaxed text-soft">
              One wallet, one ledger. Fund the prize pool in seconds, open a room players
              love, and pay the podium from your history.
            </p>
          </div>

          <div className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {INSTRUCTOR_FEATURES.map((f, i) => (
              <div
                key={f.title}
                className="group flex flex-col rounded-3xl border-2 border-ink/5 bg-cream p-6 transition-all duration-300 hover:-translate-y-1.5 hover:border-gold/50 card-3d"
                style={{ animation: "rise 0.6s cubic-bezier(0.16,1,0.3,1) both", animationDelay: `${i * 0.08}s` }}
              >
                <span className="grid size-12 place-items-center rounded-2xl bg-gold/15 text-gold-dark transition-transform duration-300 group-hover:-rotate-6 group-hover:scale-110">
                  <f.Icon className="size-6" strokeWidth={2.2} />
                </span>
                <h3 className="mt-4 font-display text-lg font-semibold">{f.title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-soft">{f.copy}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* SECURITY */}
      <section id="security" className="scroll-mt-20 px-4 sm:px-6">
        <div className="relative mx-auto w-full max-w-6xl overflow-hidden rounded-[2.5rem] bg-ink px-6 py-16 text-center text-cream sm:px-12 lg:py-20">
          <span className="pointer-events-none absolute inset-0 bg-dots-light" />
          <span className="pointer-events-none absolute -left-20 -top-24 size-72 rounded-full bg-coral/30 blur-3xl" />
          <span className="pointer-events-none absolute -bottom-28 -right-16 size-80 rounded-full bg-violet/40 blur-3xl" />

          <div className="relative">
            <span className="inline-flex items-center gap-2 rounded-full bg-white/10 px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-gold">
              <ShieldCheck className="size-4" fill="currentColor" />
              The money rules
            </span>
            <h2 className="mx-auto mt-6 max-w-2xl font-display text-4xl font-semibold leading-tight tracking-tight sm:text-6xl">
              Escrow-kept pools. <span className="text-shine">Winner-only claims.</span>
            </h2>

            <div className="mx-auto mt-12 grid max-w-4xl gap-5 text-left sm:grid-cols-2 lg:grid-cols-4">
              {SECURITY.map((s) => (
                <div key={s.title} className="rounded-3xl border border-white/10 bg-white/5 p-5">
                  <s.Icon className="size-6 text-gold" strokeWidth={2.2} />
                  <h3 className="mt-3 font-display text-lg font-semibold">{s.title}</h3>
                  <p className="mt-1 text-sm leading-relaxed text-white/60">{s.copy}</p>
                </div>
              ))}
            </div>

            <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
              <Link
                to="/join"
                className="inline-flex items-center gap-2.5 rounded-full bg-coral px-9 py-4 font-display text-lg font-semibold text-white press-3d"
                style={{ "--btn-deep-rgb": "var(--color-coral-dark)" } as CSSProperties}
              >
                <Play className="size-5 fill-current" />
                Join a quiz
              </Link>
              <Link
                to="/instructor/signup"
                className="inline-flex items-center gap-2.5 rounded-full border-2 border-white/25 bg-white/5 px-9 py-4 font-display text-lg font-semibold text-white backdrop-blur transition-all duration-200 hover:border-white/50 hover:bg-white/10"
              >
                Start hosting
                <ArrowRight className="size-5" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      <Footer className="mt-16" />
    </main>
  );
}
