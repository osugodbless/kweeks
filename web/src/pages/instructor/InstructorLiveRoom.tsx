import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Copy, Crown, Play, Presentation, Trophy, Users, Video } from "lucide-react";
import { useRoom, useRoomControl, useRoomSocket, useStandings } from "@/lib/hooks";
import { ApiError } from "@/lib/api";
import { naira } from "@/lib/player";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Chip, LiveDot } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { Money } from "@/components/ui/money";
import { cn } from "@/lib/cn";

export function InstructorLiveRoom() {
  const [params] = useSearchParams();
  const roomId = params.get("room") ?? undefined;

  const { data: room } = useRoom(roomId);
  const { data: standings = [] } = useStandings(roomId);
  const control = useRoomControl();
  const { connected } = useRoomSocket(roomId);

  const [copied, setCopied] = useState(false);
  const [copiedLink, setCopiedLink] = useState(false);
  const [error, setError] = useState("");
  const [now, setNow] = useState(() => Date.now());

  const question = room?.currentQuestion ?? null;
  const isManual = room?.pacing === "manual";
  const state = room?.state ?? "lobby";

  // Tick for the projector countdown.
  useEffect(() => {
    if (!question) return;
    const t = setInterval(() => setNow(Date.now()), 250);
    return () => clearInterval(t);
  }, [question]);

  const remaining = useMemo(() => {
    if (!question) return 0;
    const deadline = new Date(question.startedAt).getTime() + question.durationMs;
    return Math.max(0, deadline - now);
  }, [question, now]);

  function joinUrl(code: string) {
    return `${window.location.origin}/join?code=${code}`;
  }

  async function copyCode() {
    if (!room?.code) return;
    try {
      await navigator.clipboard.writeText(room.code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard unavailable */
    }
  }

  async function copyJoinLink(code: string) {
    try {
      await navigator.clipboard.writeText(joinUrl(code));
      setCopiedLink(true);
      setTimeout(() => setCopiedLink(false), 1500);
    } catch {
      /* clipboard unavailable */
    }
  }

  function run(fn: () => Promise<unknown>, label: string) {
    setError("");
    fn().catch((err) => {
      setError(err instanceof ApiError ? `${label}: ${err.message}` : `${label} failed.`);
    });
  }

  return (
    <main className="relative min-h-screen overflow-hidden pb-24">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-coral/20 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-violet/15 blur-3xl" />

      <InstructorNav />

      <div className="relative mx-auto w-full max-w-6xl px-4 pt-12 sm:px-6 lg:pt-16">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
              <Video className="size-4 text-coral" /> Live room
            </span>
            <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
              {room?.title ?? "Room"}
            </h1>
            {room && (
              <p className="mt-2 flex items-center gap-2 text-sm font-bold text-soft">
                <LiveDot color={state === "live" ? "coral" : state === "podium" ? "mint" : "gold"} />
                {state === "live" ? "Live" : state === "podium" ? "Podium" : "Lobby"} · {room.participantCount} player
                {room.participantCount === 1 ? "" : "s"} ·{" "}
                <Money value={room.poolNaira} className="text-sm" /> pool
              </p>
            )}
          </div>

          {room?.code && (
            <div className="flex flex-col items-end gap-2">
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold uppercase tracking-widest text-soft">Room code</span>
                <button
                  type="button"
                  onClick={() => void copyCode()}
                  className="flex cursor-pointer items-center gap-3 rounded-2xl border-2 border-violet/30 bg-cream px-5 py-3 font-display text-2xl font-semibold tracking-wide text-ink transition-colors hover:border-violet"
                >
                  {room.code}
                  <Copy className="size-4 text-soft" />
                  {copied && <span className="text-xs font-bold text-violet-dark">Copied</span>}
                </button>
              </div>
              <button
                type="button"
                onClick={() => void copyJoinLink(room.code)}
                className="flex max-w-full cursor-pointer items-center gap-2 rounded-full border border-ink/10 bg-cream px-3 py-1.5 text-sm font-bold text-soft transition-colors hover:border-violet hover:text-violet"
                title="Copy join link for players"
              >
                <Presentation className="size-3.5" />
                <span className="truncate font-mono text-xs">{joinUrl(room.code)}</span>
                <Copy className="size-3.5" />
                {copiedLink && <span className="text-xs font-bold text-violet-dark">Copied</span>}
              </button>
            </div>
          )}
        </div>

        {error && (
          <p role="alert" className="mt-6 rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
            {error}
          </p>
        )}

        <div className="mt-10 grid items-start gap-8 lg:grid-cols-[1.15fr_0.85fr]">
          {/* Projector preview */}
          <section className="rounded-[2rem] bg-ink p-7 text-cream sm:p-9">
            <span className="pointer-events-none absolute inset-0 rounded-[2rem] bg-dots-light opacity-50" />
            <div className="relative">
              <div className="flex items-center justify-between">
                <Chip tone="soft" className="border-white/10 bg-white/5 text-white/60">
                  <Presentation className="size-3.5" /> Projector preview
                </Chip>
                {connected && <Chip tone="mint"><LiveDot color="mint" /> Realtime on</Chip>}
              </div>

              {question ? (
                <>
                  <p className="mt-6 font-display text-lg text-white/50">
                    Question {question.index + 1} of {room?.questionCount} · {naira(room?.poolNaira ?? "0")} pool
                  </p>
                  <h2 className="mt-3 font-display text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
                    {question.prompt}
                  </h2>

                  <div className="mt-4">
                    <div className="flex items-center justify-between text-xs font-bold uppercase tracking-widest text-white/50">
                      <span>{(remaining / 1000).toFixed(0)}s left</span>
                      <span>{isManual ? "Manual pacing" : "Auto pacing"}</span>
                    </div>
                    <div className="mt-2 h-3 w-full overflow-hidden rounded-full bg-white/10">
                      <div className="h-full rounded-full bg-gold transition-[width] duration-200" style={{ width: `${(remaining / question.durationMs) * 100}%` }} />
                    </div>
                  </div>

                  <div className="mt-6 grid gap-3 sm:grid-cols-2">
                    {question.options.map((opt, i) => (
                      <div
                        key={i}
                        className={cn(
                          "flex items-center gap-3 rounded-2xl border-2 px-4 py-3.5 font-bold",
                          "border-white/10 bg-white/5",
                        )}
                      >
                        <span className="grid size-8 shrink-0 place-items-center rounded-xl bg-white/10 font-display text-base font-semibold text-white/70">
                          {String.fromCharCode(65 + i)}
                        </span>
                        <span className="flex-1">{opt}</span>
                      </div>
                    ))}
                  </div>
                </>
              ) : state === "lobby" ? (
                <div className="py-14 text-center">
                  <div className="mx-auto grid size-20 animate-pulse-ring place-items-center rounded-full bg-gold text-ink">
                    <Users className="size-10" />
                  </div>
                  <h2 className="mt-6 font-display text-3xl font-semibold">Waiting to start</h2>
                  <p className="mx-auto mt-2 max-w-sm text-white/60">
                    {room?.participantCount ?? 0} player{room?.participantCount === 1 ? " is" : "s are"} in the
                    room. Hit start when the crowd is ready.
                  </p>
                  <p className="mt-4 text-sm font-bold text-gold">
                    Share the code: <span className="font-display text-cream">{room?.code ?? "—"}</span>
                  </p>
                </div>
              ) : (
                <div className="py-14 text-center">
                  <div className="mx-auto grid size-20 place-items-center rounded-full bg-gold text-ink">
                    <Trophy className="size-10" />
                  </div>
                  <h2 className="mt-6 font-display text-3xl font-semibold">
                    {state === "podium" ? "Winners declared" : "All questions played"}
                  </h2>
                  {state === "podium" ? (
                    <p className="mt-2 text-white/60">The podium has been set — share the good news.</p>
                  ) : (
                    <p className="mt-2 text-white/60">Declare the podium to lock in the winners.</p>
                  )}
                </div>
              )}
            </div>
          </section>

          {/* Side column */}
          <section className="space-y-6">
            {/* Players */}
            <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6">
              <div className="flex items-center justify-between">
                <h3 className="flex items-center gap-2 font-display text-lg font-semibold">
                  <Users className="size-5 text-violet" /> In the room
                </h3>
                <span className="rounded-full bg-ink/5 px-3 py-1 text-xs font-bold text-soft">
                  {room?.participantCount ?? 0}
                </span>
              </div>
              <ul className="mt-4 flex flex-wrap gap-2">
                {room?.participants.map((p) => (
                  <li key={p.id} className="flex items-center gap-1.5 rounded-full border border-ink/10 bg-white px-2.5 py-1 text-xs font-bold text-soft">
                    <Avatar id={p.avatar} className="size-5" />
                    {p.nickname}
                  </li>
                ))}
                {!room?.participants.length && (
                  <li className="text-sm font-bold text-soft">No players yet — share the code.</li>
                )}
              </ul>
            </div>

            {/* Controls */}
            <div className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6">
              <h3 className="flex items-center gap-2 font-display text-lg font-semibold">
                <Crown className="size-5 text-gold" /> Controls
              </h3>
              <div className="mt-4 space-y-3">
                {state === "lobby" && (
                  <Button variant="coral" size="lg" loading={control.start.isPending} className="w-full" onClick={() => run(() => control.start.mutateAsync(roomId!), "Start")}>
                    <Play className="size-5 fill-current" /> Start the quiz
                  </Button>
                )}
                {state === "live" && isManual && (
                  <Button variant="violet" size="lg" loading={control.next.isPending} className="w-full" onClick={() => run(() => control.next.mutateAsync(roomId!), "Next question")}>
                    Next question
                  </Button>
                )}
                {state === "live" && (
                  <Button variant="gold" size="lg" loading={control.podium.isPending} className="w-full" onClick={() => run(() => control.podium.mutateAsync(roomId!), "Declare winners")}>
                    <Trophy className="size-5" /> Declare winners
                  </Button>
                )}
                {state === "podium" && (
                  <div className="rounded-2xl border-2 border-mint/30 bg-mint/10 px-4 py-4 text-center">
                    <p className="text-xs font-bold uppercase tracking-widest text-violet-dark">Room ended</p>
                    <Link to="/instructor/history" className="mt-2 inline-block text-sm font-bold text-violet hover:text-violet-dark">
                      View history →
                    </Link>
                  </div>
                )}
              </div>

              {state === "live" && standings.length > 0 && (
                <div className="mt-5 border-t border-ink/5 pt-4">
                  <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Live standings</p>
                  <ul className="mt-3 space-y-2">
                    {standings.slice(0, 3).map((s, i) => (
                      <li key={s.participantId} className="flex items-center gap-2 text-sm font-bold">
                        <span className="w-4 text-soft">{i + 1}</span>
                        <Avatar id={s.avatar} className="size-6" />
                        <span className="min-w-0 flex-1 truncate">{s.nickname}</span>
                        <span className="font-display text-violet-dark">{s.correctCount}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </section>
        </div>
      </div>

      <Footer className="mt-16" />
    </main>
  );
}
