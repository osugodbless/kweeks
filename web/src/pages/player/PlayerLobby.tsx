import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Copy, Timer, Trophy, Users } from "lucide-react";
import { useRoom } from "@/lib/hooks";
import { splitPodium, usePlayer } from "@/lib/player";
import { Avatar } from "@/components/ui/avatar";
import { Chip } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { Money } from "@/components/ui/money";
import { PlayerTopBar } from "@/components/ui/player-topbar";

export function PlayerLobby() {
  const navigate = useNavigate();
  const { roomId, code, nickname, participantId } = usePlayer();
  const { data: room, isLoading } = useRoom(roomId ?? undefined);

  useEffect(() => {
    if (!roomId) {
      navigate("/join", { replace: true });
      return;
    }
    if (room?.state === "live") navigate("/question");
    else if (room?.state === "podium" || room?.state === "ended") navigate("/podium");
  }, [room?.state, roomId, navigate]);

  const pool = room?.poolNaira ?? "0";
  const winnerCount = room?.winnerCount ?? 3;
  const shares = splitPodium(pool, winnerCount);
  const youName = nickname ?? "Player";

  async function copyCode() {
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
    } catch {
      /* clipboard unavailable */
    }
  }

  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-24 size-72 rounded-full bg-gold/20 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-24 size-80 rounded-full bg-violet/20 blur-3xl" />

      <PlayerTopBar status="lobby" code={room?.code ?? code} />

      <div className="relative mx-auto w-full max-w-6xl px-4 py-12 sm:px-6 lg:py-16">
        <div className="grid items-start gap-12 lg:grid-cols-[1.05fr_0.95fr]">
          {/* Left: you're in */}
          <div className="lg:sticky lg:top-24">
            <h1 className="font-display text-5xl font-semibold leading-[1.02] tracking-tight sm:text-6xl">
              You&apos;re in, <span className="text-pop">{youName}</span>
            </h1>
            <p className="mt-4 max-w-md text-lg leading-relaxed text-soft">
              {room ? `"${room.title}" is almost live. Keep this tab open — the first question drops the second the host hits start.` : "Waiting for the host to open the room…"}
            </p>

            <div className="mt-8 rounded-[2rem] border-2 border-mint/30 bg-mint/10 p-6">
              <p className="text-xs font-bold uppercase tracking-[0.25em] text-mint-dark">On the line</p>
              <Money value={pool} className="mt-1 block text-5xl" />
              <p className="mt-2 text-sm leading-relaxed text-ink">
                Pool split across the top <span className="font-bold">{winnerCount}</span> finisher{winnerCount === 1 ? "" : "s"}.
              </p>
              <div className="mt-4 flex flex-wrap gap-2">
                {shares.map((s, i) => (
                  <Chip key={i} tone="gold">
                    {i + 1}
                    {i === 0 ? "st" : i === 1 ? "nd" : i === 2 ? "rd" : "th"} takes <Money value={s} className="text-xs" />
                  </Chip>
                ))}
              </div>
            </div>

            {room?.code && (
              <button
                type="button"
                onClick={copyCode}
                className="mt-6 flex cursor-pointer items-center gap-2 rounded-full border-2 border-ink/10 bg-cream px-5 py-2.5 text-sm font-bold text-ink transition-colors hover:border-violet hover:text-violet"
              >
                <Copy className="size-4" />
                Copy room code {room.code}
              </button>
            )}
          </div>

          {/* Right: in the room */}
          <section className="rounded-[2rem] border-2 border-ink/5 bg-cream p-6 card-3d sm:p-8">
            <div className="flex items-center justify-between">
              <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
                <Users className="size-5 text-violet" /> In the room
              </h2>
              <span className="rounded-full bg-ink/5 px-3 py-1 text-xs font-bold text-soft">
                {room?.participantCount ?? 1} player{room?.participantCount === 1 ? "" : "s"}
              </span>
            </div>

            <ul className="mt-5 flex flex-wrap gap-2.5">
              {room?.participants.map((p) => (
                <li
                  key={p.id}
                  className={`flex items-center gap-2 rounded-full border px-3 py-1.5 text-sm font-bold ${
                    p.id === participantId
                      ? "border-violet bg-violet/10 text-violet-dark"
                      : "border-ink/10 bg-white text-soft"
                  }`}
                >
                  <Avatar id={p.avatar} className="size-6" />
                  {p.nickname}
                  {p.id === participantId && (
                    <span className="rounded-full bg-violet px-2 py-0.5 text-[10px] font-bold uppercase text-white">You</span>
                  )}
                </li>
              ))}
              {!room && !isLoading && (
                <li className="text-sm font-bold text-soft">{nickname ?? "You"} are first in — invite friends with the code.</li>
              )}
            </ul>

            <div className="mt-6 rounded-3xl border-2 border-gold/40 bg-gold/10 p-6 text-center">
              <div className="mx-auto grid size-14 animate-pulse-ring place-items-center rounded-full bg-gold text-ink">
                <Trophy className="size-7" />
              </div>
              <p className="mt-4 font-display text-xl font-semibold">Waiting for the host…</p>
              <p className="mx-auto mt-1 max-w-xs text-sm leading-relaxed text-soft">
                {room?.participantCount && room.participantCount > 1
                  ? `${room.participantCount} players are in. The room fires on the host's signal.`
                  : "Share the code so more players can jump in before the room goes live."}
              </p>
              <p className="mt-4 flex items-center justify-center gap-2 text-sm font-bold text-soft">
                <Timer className="size-4 text-coral" />
                Rounds start automatically once the host hits start
              </p>
            </div>
          </section>
        </div>
      </div>

      <Footer variant="player" />
    </main>
  );
}
