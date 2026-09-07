import { useEffect } from "react";
import type { CSSProperties } from "react";
import { useNavigate } from "react-router-dom";
import { Crown, Medal, Trophy, Users } from "lucide-react";
import { useRoom, useStandings } from "@/lib/hooks";
import { usePlayer } from "@/lib/player";
import { Avatar } from "@/components/ui/avatar";
import { Chip, LiveDot } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { PlayerTopBar } from "@/components/ui/player-topbar";
import { cn } from "@/lib/cn";

export function PlayerStandings() {
  const navigate = useNavigate();
  const { roomId, participantId, lastAnsweredIndex } = usePlayer();
  const { data: room } = useRoom(roomId ?? undefined);
  const { data: standings = [] } = useStandings(roomId ?? undefined);

  useEffect(() => {
    if (!roomId) {
      navigate("/join", { replace: true });
      return;
    }
    if (room?.state === "lobby") navigate("/lobby");
    else if (room?.state === "podium" || room?.state === "ended") navigate("/podium");
    else if (
      room?.state === "live" &&
      room.currentQuestion &&
      (lastAnsweredIndex == null || room.currentQuestion.index > lastAnsweredIndex)
    ) {
      navigate("/question");
    }
  }, [room?.state, room?.currentQuestion, lastAnsweredIndex, roomId, navigate]);

  const afterQuestion = room && room.currentIndex >= 0 ? room.currentIndex + 1 : 1;
  const myRow = standings.find((s) => s.participantId === participantId);
  const myRank = myRow ? standings.indexOf(myRow) + 1 : null;

  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -left-24 top-32 size-72 rounded-full bg-sky/20 blur-3xl" />
      <span className="pointer-events-none absolute -right-24 bottom-24 size-80 rounded-full bg-violet/20 blur-3xl" />

      <PlayerTopBar status="standings" code={room?.code} />

      <div className="relative mx-auto w-full max-w-4xl px-4 py-12 sm:px-6 lg:py-16">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
              <Trophy className="size-4 text-gold" /> Live standings
            </span>
            <h1 className="mt-4 font-display text-4xl font-semibold tracking-tight sm:text-5xl">
              After question {afterQuestion}
              {room && <span className="text-soft"> · {room.participantCount} players</span>}
            </h1>
          </div>
          {myRank && (
            <span className="flex items-center gap-2 rounded-full bg-gold px-5 py-2.5 font-display text-lg font-semibold text-ink">
              <Crown className="size-5" /> You&apos;re {myRank}{myRank === 1 ? "st" : myRank === 2 ? "nd" : myRank === 3 ? "rd" : "th"}
            </span>
          )}
        </div>

        <div className="mt-10 overflow-hidden rounded-[2rem] border-2 border-ink/5 bg-cream card-3d">
          <div className="grid grid-cols-[3rem_1fr_5rem_6rem] items-center gap-3 border-b border-ink/5 px-6 py-3 text-[11px] font-bold uppercase tracking-widest text-soft sm:grid-cols-[4rem_1fr_5rem_7rem]">
            <span>#</span>
            <span>Player</span>
            <span className="text-right">Correct</span>
            <span className="text-right">Speed</span>
          </div>
          <ul>
            {standings.map((row, i) => {
              const isMe = row.participantId === participantId;
              return (
                <li
                  key={row.participantId}
                  className={cn(
                    "grid grid-cols-[3rem_1fr_5rem_6rem] items-center gap-3 px-6 py-4 sm:grid-cols-[4rem_1fr_5rem_7rem]",
                    i > 0 && "border-t border-ink/5",
                    isMe && "bg-gold",
                  )}
                >
                  <span className="grid size-8 place-items-center font-display text-lg font-semibold">
                    {i === 0 ? (
                      <Trophy className="size-5 text-gold-dark" fill="currentColor" />
                    ) : i === 1 ? (
                      <Medal className="size-5 text-soft" />
                    ) : i === 2 ? (
                      <Medal className="size-5 text-gold/70" />
                    ) : (
                      <span className={isMe ? "text-ink" : "text-soft"}>{i + 1}</span>
                    )}
                  </span>
                  <span className="flex min-w-0 items-center gap-3">
                    <Avatar id={row.avatar} className="size-10" />
                    <span className="truncate font-bold">{row.nickname}</span>
                    {isMe && (
                      <span className="rounded-full bg-ink px-2.5 py-0.5 text-[10px] font-bold uppercase text-cream">You</span>
                    )}
                  </span>
                  <span
                    className={cn(
                      "text-right font-display text-lg font-semibold",
                      isMe ? "text-ink" : "text-violet-dark",
                    )}
                  >
                    {row.correctCount}
                  </span>
                  <span className={cn("text-right text-sm font-bold", isMe ? "text-ink/80" : "text-soft")}>
                    {(row.totalLatencyMs / 1000).toFixed(1)}s
                  </span>
                </li>
              );
            })}
            {standings.length === 0 && (
              <li className="px-6 py-10 text-center text-sm font-bold text-soft">
                No scores yet — be the first to answer.
              </li>
            )}
          </ul>
        </div>

        <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
          {room?.state === "live" && (
            <button
              type="button"
              onClick={() => navigate("/question")}
              className="flex cursor-pointer items-center gap-2 rounded-full bg-violet px-6 py-3 text-sm font-bold text-white press-3d"
              style={{ "--btn-deep-rgb": "var(--color-violet-dark)" } as CSSProperties}
            >
              Back to question
            </button>
          )}
          <span className="flex items-center gap-2 text-sm font-bold text-soft">
            <LiveDot color="coral" />
            Next question starts on the host&apos;s signal
          </span>
          <Chip tone="soft">
            <Users className="size-3.5" /> {room?.participantCount ?? 0} playing
          </Chip>
        </div>
      </div>

      <Footer variant="player" />
    </main>
  );
}
