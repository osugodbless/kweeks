import { useEffect, useMemo, useState } from "react";
import type { CSSProperties } from "react";
import { useNavigate } from "react-router-dom";
import { Copy, Crown, Mail, PartyPopper, ShieldCheck, Trophy } from "lucide-react";
import { api, ApiError, type ClaimResult } from "@/lib/api";
import { useRoom } from "@/lib/hooks";
import { splitPodium, usePlayer } from "@/lib/player";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { Money } from "@/components/ui/money";
import { PlayerTopBar } from "@/components/ui/player-topbar";

const CONFETTI = [
  { cls: "bg-coral/70", pos: "left-[8%] top-[18%]", size: "size-3", tilt: "-6deg", delay: "0s" },
  { cls: "bg-gold/80", pos: "left-[16%] top-[64%]", size: "size-2", tilt: "12deg", delay: "0.8s" },
  { cls: "bg-mint/70", pos: "right-[10%] top-[20%]", size: "size-2.5", tilt: "20deg", delay: "1.6s" },
  { cls: "bg-violet/60", pos: "left-[42%] top-[8%]", size: "size-2", tilt: "30deg", delay: "2.2s" },
  { cls: "bg-sky/70", pos: "right-[28%] bottom-[10%]", size: "size-2", tilt: "-12deg", delay: "1.1s" },
  { cls: "bg-coral/50", pos: "right-[5%] top-[58%]", size: "size-3", tilt: "8deg", delay: "2.6s" },
];

const CLAIM_STEPS = [
  { title: "Claim locked", copy: "Your share of the pool is reserved the moment you redeem." },
  { title: "Invite sent", copy: "We reach out to your payout email to confirm your details." },
  { title: "Paid out", copy: "The money lands. You'll see it on your history too." },
];

export function PlayerPodium() {
  const navigate = useNavigate();
  const { roomId, email, participantId } = usePlayer();
  const { data: room } = useRoom(roomId ?? undefined);

  const [claim, setClaim] = useState<ClaimResult | null>(null);
  const [claiming, setClaiming] = useState(false);
  const [claimError, setClaimError] = useState("");
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!roomId) {
      navigate("/join", { replace: true });
      return;
    }
    if (room?.state === "lobby") navigate("/lobby");
    else if (room?.state === "live") navigate("/question");
  }, [room?.state, roomId, navigate]);

  const winners = room?.winners ?? [];
  const winnerCount = room?.winnerCount ?? 3;
  const shares = useMemo(() => splitPodium(room?.poolNaira ?? "0", winnerCount), [room?.poolNaira, winnerCount]);

  const myIndex = winners.findIndex((w) => w.participantId === participantId);
  const isWinner = myIndex >= 0;
  const myShare = isWinner ? shares[myIndex] ?? 0 : 0;
  const place = isWinner ? myIndex + 1 : null;

  async function redeem() {
    if (!roomId) return;
    setClaiming(true);
    setClaimError("");
    try {
      const res = await api.post<ClaimResult>(`/rooms/${roomId}/redeem`, { email: email ?? "" });
      setClaim(res);
    } catch (err) {
      setClaimError(err instanceof ApiError ? err.message : "Redeem failed — try again.");
    } finally {
      setClaiming(false);
    }
  }

  async function copyClaim() {
    if (!claim?.claimCode) return;
    try {
      await navigator.clipboard.writeText(claim.claimCode);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard unavailable */
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-32 size-72 rounded-full bg-gold/20 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-24 size-80 rounded-full bg-violet/20 blur-3xl" />

      <PlayerTopBar status="podium" code={room?.code} />

      <div className="relative mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 lg:py-16">
        <div className="grid items-start gap-12 lg:grid-cols-[1fr_1fr]">
          {/* Left: my result */}
          <div className="lg:sticky lg:top-24">
            <div className="relative overflow-hidden rounded-[2rem] bg-ink p-8 text-center text-cream sm:p-10">
              <span className="pointer-events-none absolute inset-0 bg-dots-light" />
              {CONFETTI.map((c, i) => (
                <span
                  key={i}
                  className={`pointer-events-none absolute hidden rounded-[3px] animate-float sm:block ${c.cls} ${c.pos} ${c.size}`}
                  style={{ "--tilt": c.tilt, animationDelay: c.delay } as CSSProperties}
                />
              ))}

              <div className="relative">
                <Chip tone={isWinner ? "gold" : "soft"}>
                  {isWinner ? `${place}${place === 1 ? "st" : place === 2 ? "nd" : place === 3 ? "rd" : "th"} place` : "Final"}
                </Chip>
                <div className="mx-auto mt-6 grid size-20 animate-float place-items-center rounded-full bg-gold text-ink">
                  {isWinner ? <Crown className="size-10" /> : <PartyPopper className="size-10" />}
                </div>
                <h1 className="mt-6 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
                  {isWinner ? (
                    <>
                      You banked <Money value={myShare} tone="dark" />
                    </>
                  ) : (
                    "Better luck next time"
                  )}
                </h1>
                <p className="mt-3 text-white/60">
                  {isWinner
                    ? `${room?.participantCount ?? 0} players · top ${winnerCount} split the ${"pool"}.`
                    : `${room?.title ?? "The room"} has wrapped — the winners took the pool.`}
                </p>
                <p className="mt-4 text-sm font-bold text-white/60">
                  {room?.title} · <Money value={room?.poolNaira ?? "0"} tone="dark" className="text-sm" /> pool
                </p>
              </div>
            </div>
          </div>

          {/* Right: winners + claim */}
          <section className="space-y-6">
            <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6 sm:p-8">
              <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
                <Trophy className="size-5 text-gold" /> Final standings
              </h2>
              <ul className="mt-5 space-y-3">
                {winners.map((w, i) => {
                  const isMe = w.participantId === participantId;
                  return (
                    <li
                      key={w.participantId}
                      className={`flex items-center gap-4 rounded-2xl border-2 px-4 py-3 ${
                        isMe ? "border-gold bg-gold/10" : "border-ink/5 bg-white"
                      }`}
                    >
                      <span className="font-display text-xl font-semibold">{i + 1}</span>
                      <Avatar id={w.avatar} className="size-10" />
                      <span className="min-w-0 flex-1 truncate font-bold">
                        {w.nickname}
                        {isMe && (
                          <span className="ml-2 rounded-full bg-gold px-2 py-0.5 text-[10px] font-bold uppercase text-ink">You</span>
                        )}
                      </span>
                      <span className="text-xs font-bold text-soft">{w.correctCount} correct</span>
                      <Money value={shares[i] ?? 0} className="text-base" />
                    </li>
                  );
                })}
                {winners.length === 0 && (
                  <li className="rounded-2xl border-2 border-dashed border-ink/15 bg-white px-4 py-8 text-center text-sm font-bold text-soft">
                    Winners are being declared — check back in a moment.
                  </li>
                )}
              </ul>
            </div>

            {isWinner && (
              <div className="rounded-[2rem] border-2 border-gold/40 bg-gold/10 p-6 sm:p-8">
                <p className="text-xs font-bold uppercase tracking-[0.25em] text-gold-dark">Your claim code</p>

                {claim?.claimCode ? (
                  <>
                    <div className="mt-3 flex items-center gap-3">
                      <p className="flex-1 rounded-2xl bg-white px-5 py-4 font-display text-xl font-semibold tracking-wide text-ink">
                        {claim.claimCode}
                      </p>
                      <Button variant="ghost" size="md" onClick={() => void copyClaim()}>
                        <Copy className="size-4" />
                        {copied ? "Copied" : "Copy"}
                      </Button>
                    </div>
                    <p className="mt-3 flex items-center gap-2 text-sm font-bold text-mint-dark">
                      <ShieldCheck className="size-4" /> Claim locked · <Money value={claim.amountNaira} className="text-sm" />
                    </p>
                  </>
                ) : (
                  <>
                    <p className="mt-2 text-sm leading-relaxed text-ink">
                      Lock in your <Money value={myShare} className="text-sm" /> share now. Your claim code is
                      generated right here — keep it safe.
                    </p>
                    {claimError && (
                      <p role="alert" className="mt-3 rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                        {claimError}
                      </p>
                    )}
                    <Button
                      variant="gold"
                      size="lg"
                      loading={claiming}
                      onClick={() => void redeem()}
                      className="mt-5 w-full"
                      icon={<Mail className="size-5" />}
                    >
                      {claiming ? "Redeeming…" : "Redeem"}
                    </Button>
                    <p className="mt-2 flex items-center gap-2 text-xs text-soft">
                      <Mail className="size-3.5" /> Payout details go to {email ?? "your payout email"}
                    </p>
                  </>
                )}
              </div>
            )}

            {isWinner && !claim?.claimCode && (
              <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6">
                <p className="text-xs font-bold uppercase tracking-widest text-soft">How it lands</p>
                <ol className="mt-4 space-y-4">
                  {CLAIM_STEPS.map((step, i) => (
                    <li key={step.title} className="flex items-start gap-4">
                      <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-violet/10 font-display text-base font-semibold text-violet-dark">
                        {i + 1}
                      </span>
                      <div>
                        <p className="font-bold">{step.title}</p>
                        <p className="text-sm text-soft">{step.copy}</p>
                      </div>
                    </li>
                  ))}
                </ol>
              </div>
            )}

            {!isWinner && (
              <Button variant="violet" size="lg" onClick={() => navigate("/")} className="w-full">
                Back to Kweeks
              </Button>
            )}
          </section>
        </div>
      </div>

      <Footer variant="player" />
    </main>
  );
}
