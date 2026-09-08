import type { CSSProperties, ReactNode } from "react";
import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ArrowRight, Copy, FileQuestion, History, LayoutDashboard, Plus, ShieldCheck, Trophy, Users, Wallet, Zap } from "lucide-react";
import { useDashboard, useWallet, useWalletSetup } from "@/lib/hooks";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Chip, LiveDot } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { Money } from "@/components/ui/money";
import { WalletSetupWizard } from "@/pages/instructor/WalletSetupWizard";

const METHODS = ["Card", "Transfer", "Instant top-up"];

const SETUP_COPY: Record<string, { title: string; copy: string }> = {
  kyc: { title: "Finish setting up your wallet", copy: "Verify your identity to create your wallet and activate naira payouts." },
  wallet: { title: "Create your wallet", copy: "Your identity is verified — the last steps create your wallet and activate naira." },
  rail: { title: "Activate naira", copy: "One last step: activate the NGN rail so your wallet is ready to fund and pay winners." },
  unprovisioned: { title: "Set up your wallet", copy: "Provision your BMONI wallet to fund pools and pay winners." },
};

function Stat({ label, value, accent }: { label: string; value: ReactNode; accent: string }) {
  return (
    <div className="rounded-3xl border-2 border-ink/5 bg-cream p-5">
      <p className="text-[11px] font-bold uppercase tracking-widest text-soft">{label}</p>
      <p className={`mt-2 font-display text-3xl font-semibold ${accent}`}>{value}</p>
    </div>
  );
}

function shortAddress(addr: string): string {
  if (!addr) return "";
  if (addr.length <= 16) return addr;
  return `${addr.slice(0, 6)}…${addr.slice(-4)}`;
}

/** The wallet's real identity once provisioned. Nothing is shown before setup,
 *  so no fabricated internal id leaks into the UI. Once the wallet exists its
 *  on-chain address appears (truncated, copyable); when the NGN rail is ready
 *  the receiving bank account the host funds by transfer is shown first. */
function WalletIdentity({
  ready,
  address,
  accountNumber,
  bankName,
}: {
  ready: boolean;
  address?: string;
  accountNumber?: string;
  bankName?: string;
}) {
  const [copied, setCopied] = useState<string | null>(null);

  async function copy(value: string, key: string) {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(key);
      setTimeout(() => setCopied((c) => (c === key ? null : c)), 1400);
    } catch {
      /* clipboard unavailable */
    }
  }

  if (!address) return null; // not provisioned — stay empty

  return (
    <div className="mt-3 space-y-2">
      {ready && accountNumber && (
        <div>
          <p className="text-[11px] font-bold uppercase tracking-widest text-white/40">NGN receiving account</p>
          <div className="mt-1 flex items-center gap-2">
            <button
              type="button"
              onClick={() => void copy(accountNumber, "acct")}
              className="group flex cursor-pointer items-center gap-2 rounded-lg px-1.5 py-0.5 transition-colors hover:bg-white/5"
              title="Copy account number"
            >
              <span className="font-display text-lg font-semibold tracking-wide text-cream">
                {accountNumber.replace(/(\d{4})(?=\d)/g, "$1 ")}
              </span>
              <Copy className="size-3.5 text-white/40 group-hover:text-cream" />
            </button>
            {bankName && <span className="text-xs text-white/40">{bankName}</span>}
          </div>
        </div>
      )}

      <div className="flex items-center gap-1.5 text-sm text-white/50">
        <span className="text-xs font-bold uppercase tracking-widest text-white/35">Wallet address</span>
        <button
          type="button"
          onClick={() => void copy(address, "addr")}
          className="group flex cursor-pointer items-center gap-1.5 rounded-lg px-1.5 py-0.5 font-mono text-[13px] text-white/70 transition-colors hover:bg-white/5 hover:text-cream"
          title="Copy wallet address"
        >
          {shortAddress(address)}
          <Copy className="size-3 text-white/40 group-hover:text-cream" />
          {copied && <span className="font-body text-[11px] font-bold text-mint">Copied</span>}
        </button>
      </div>
    </div>
  );
}

export function InstructorWallet() {
  const navigate = useNavigate();
  const { instructor, wallet } = useAuth();
  const { data: dash } = useDashboard();
  const { data: walletView } = useWallet();
  const { data: setup } = useWalletSetup();
  const [setupOpen, setSetupOpen] = useState(false);

  const first = instructor?.firstName ?? instructor?.name?.split(" ")[0] ?? "host";
  const balance = walletView?.wallet.balanceNaira ?? wallet?.balanceNaira ?? "0";
  const quizzes = dash?.quizzes ?? [];
  const liveQuiz = quizzes.find((q) => q.roomId && (q.state === "lobby" || q.state === "live"));
  const setupPending = setup && setup.stage !== "ready";
  const setupMeta = setupPending ? SETUP_COPY[setup.stage] ?? SETUP_COPY.unprovisioned : null;

  return (
    <main className="relative min-h-screen overflow-hidden pb-24">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-violet/20 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-gold/15 blur-3xl" />

      <InstructorNav />

      <div className="relative mx-auto w-full max-w-6xl px-4 pt-12 sm:px-6 lg:pt-16">
        <div className="flex flex-col justify-between gap-6 sm:flex-row sm:items-end">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
              <LayoutDashboard className="size-4 text-gold" /> Wallet dashboard
            </span>
            <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
              Welcome back, <span className="text-pop">{first}</span>
            </h1>
          </div>
          <Button variant="coral" size="lg" icon={<Plus className="size-5" />} onClick={() => navigate("/instructor/quiz-builder")}>
            Create a quiz
          </Button>
        </div>

        {setupPending && setupMeta && (
          <button
            type="button"
            onClick={() => setSetupOpen(true)}
            className="mt-8 flex w-full cursor-pointer items-center gap-4 rounded-3xl border-2 border-violet/30 bg-violet/5 p-5 text-left transition-all hover:border-violet/50 hover:bg-violet/10"
          >
            <span className="grid size-12 shrink-0 place-items-center rounded-2xl bg-violet text-white">
              <ShieldCheck className="size-6" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block font-display text-lg font-semibold">{setupMeta.title}</span>
              <span className="block text-sm text-soft">{setupMeta.copy}</span>
            </span>
            <ArrowRight className="size-5 shrink-0 text-violet" />
          </button>
        )}

        {/* Stats */}
        <div className="mt-10 grid grid-cols-2 gap-4 lg:grid-cols-4">
          <Stat label="Quizzes hosted" value={String(dash?.quizzesHosted ?? 0)} accent="text-violet-dark" />
          <Stat label="Players hosted" value={String(dash?.playersHosted ?? 0)} accent="text-sky-dark" />
          <Stat label="Winners paid" value={String(dash?.winnersPaid ?? 0)} accent="text-coral" />
          <Stat label="Available" value={<Money value={balance} className="text-3xl" />} accent="text-mint-dark" />
        </div>

        <div className="mt-8 grid items-start gap-8 lg:grid-cols-[0.9fr_1.1fr]">
          {/* Wallet card */}
          <section className="rounded-[2rem] bg-ink p-7 text-cream sm:p-8">
            <span className="pointer-events-none absolute inset-0 rounded-[2rem] bg-dots-light opacity-60" />
            <div className="relative">
              <p className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-white/50">
                <Wallet className="size-4 text-gold" /> Your wallet · NGN
              </p>
              <div className="mt-4">
                <Money value={balance} tone="dark" className="text-5xl" />
                <WalletIdentity
                  ready={setup?.stage === "ready"}
                  address={setup?.bmoniWalletAddress || wallet?.bmoniWalletAddress}
                  accountNumber={setup?.depositAccount?.accountNumber}
                  bankName={setup?.depositAccount?.bankName}
                />
              </div>

              <Button variant="gold" size="lg" className="mt-6 w-full" onClick={() => navigate("/instructor/fund")}>
                Fund wallet
              </Button>

              <div className="mt-5 flex flex-wrap gap-2">
                {METHODS.map((m) => (
                  <Chip key={m} tone="soft" className="border-white/10 bg-white/5 text-white/70">
                    {m}
                  </Chip>
                ))}
              </div>

              {liveQuiz && (
                <div className="mt-6 rounded-3xl border border-gold/40 bg-gold/10 p-5">
                  <p className="flex items-center gap-2 text-xs font-bold uppercase tracking-widest text-gold">
                    <LiveDot color="gold" /> Room live
                  </p>
                  <p className="mt-2 font-display text-lg font-semibold">{liveQuiz.title}</p>
                  <p className="text-sm text-white/60">
                    Code <span className="font-display text-cream">{liveQuiz.roomCode}</span> ·{" "}
                    <Money value={liveQuiz.poolNaira} tone="dark" className="text-sm" /> pool
                  </p>
                  <Link
                    to={`/instructor/live-room?room=${liveQuiz.roomId}`}
                    className="mt-3 inline-flex items-center gap-2 rounded-full bg-gold px-5 py-2.5 text-sm font-bold text-ink press-3d"
                    style={{ "--btn-deep-rgb": "var(--color-gold-dark)" } as CSSProperties}
                  >
                    Open room <ArrowRight className="size-4" />
                  </Link>
                </div>
              )}
            </div>
          </section>

          {/* Your quizzes */}
          <section className="space-y-6">
            <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6 sm:p-7">
              <div className="flex items-center justify-between">
                <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
                  <FileQuestion className="size-5 text-violet" /> Your quizzes
                </h2>
                <Link to="/instructor/history" className="flex items-center gap-1.5 text-sm font-bold text-violet hover:text-violet-dark">
                  <History className="size-4" /> View history →
                </Link>
              </div>

              <ul className="mt-5 space-y-3">
                {quizzes.map((q) => (
                  <li key={q.id} className="flex items-center gap-4 rounded-2xl border-2 border-ink/5 bg-white px-4 py-3.5">
                    <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-violet/10 text-violet">
                      <Trophy className="size-5" strokeWidth={2.2} />
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-bold">{q.title}</p>
                      <p className="text-xs text-soft">
                        {q.questionCount} questions · top {q.winnerCount} ·{" "}
                        <Money value={q.poolNaira} className="text-xs" />
                      </p>
                    </div>
                    {q.roomId ? (
                      <Link
                        to={`/instructor/live-room?room=${q.roomId}`}
                        className="flex items-center gap-1.5 rounded-full bg-violet px-4 py-2 text-xs font-bold text-white press-3d"
                        style={{ "--btn-deep-rgb": "var(--color-violet-dark)" } as CSSProperties}
                      >
                        <Zap className="size-3.5" fill="currentColor" />
                        {q.state === "live" ? "Live" : "Open room"}
                      </Link>
                    ) : (
                      <Link
                        to={`/instructor/quiz-builder?id=${q.id}`}
                        className="text-xs font-bold text-soft underline-offset-2 hover:text-violet hover:underline"
                      >
                        Edit →
                      </Link>
                    )}
                  </li>
                ))}
                {quizzes.length === 0 && (
                  <li className="rounded-2xl border-2 border-dashed border-ink/15 bg-white px-6 py-10 text-center">
                    <p className="font-display text-lg font-semibold">No quizzes yet</p>
                    <p className="mt-1 text-sm text-soft">Build your first deck, fund a pool, and open a room.</p>
                    <Button variant="violet" size="md" className="mt-4" onClick={() => navigate("/instructor/quiz-builder")}>
                      <Plus className="size-4" /> Build your first quiz
                    </Button>
                  </li>
                )}
              </ul>
            </div>

            <div className="flex items-start gap-4 rounded-3xl border-2 border-gold/30 bg-gold/10 p-5">
              <span className="grid size-11 shrink-0 place-items-center rounded-2xl bg-gold text-ink">
                <Users className="size-5" strokeWidth={2.4} />
              </span>
              <p className="text-sm leading-relaxed text-ink">
                <span className="font-bold">Players join with your room code.</span> Open a room from
                any quiz above and share the code — no app, no install.
              </p>
            </div>
          </section>
        </div>
      </div>

      <Footer className="mt-16" />

      <WalletSetupWizard open={setupOpen} onClose={() => setSetupOpen(false)} />
    </main>
  );
}
