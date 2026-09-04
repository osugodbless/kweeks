import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  PlayerDesktop,
  PlayerTopBar,
  PlayerFooter,
  KweeksBrand,
} from "@/components/ui/player-desktop";
import { cn } from "@/lib/cn";
import { useRoom, useStandings } from "@/lib/hooks";
import { fmtNaira, usePlayer } from "@/lib/player";

function ptsFor(s: { correctCount: number; totalLatencyMs: number }): number {
  if (s.correctCount <= 0) return 0;
  const speed = Math.max(0, 1 - s.totalLatencyMs / (30_000 * s.correctCount));
  return s.correctCount * 1000 + Math.round(250 * speed);
}

function ordinal(n: number): string {
  const s = ["th", "st", "nd", "rd"];
  const v = n % 100;
  return `${n}${s[(v - 20) % 10] ?? s[v] ?? s[0]}`;
}

export function PlayerStandings() {
  const nav = useNavigate();
  const player = usePlayer();
  const roomId = player.roomId ?? undefined;

  const roomQ = useRoom(roomId);
  const room = roomQ.data;
  const standingsQ = useStandings(roomId);
  const standings = standingsQ.data;

  useEffect(() => {
    if (!room) return;
    if (room.state === "podium") nav("/podium");
    else if (room.state === "live" && room.currentQuestion && room.currentQuestion.remainingMs > 0) {
      nav("/question");
    }
  }, [
    room,
    room?.state,
    room?.currentQuestion?.id,
    room?.currentQuestion?.remainingMs,
    nav,
  ]);

  if (!roomId) {
    return (
      <PlayerDesktop>
        <PlayerTopBar left={<KweeksBrand />} />
        <div className="flex flex-1 flex-col items-center justify-center gap-4">
          <h1 className="font-display text-[30px] font-extrabold text-paper">
            You're not in a room yet
          </h1>
          <button
            onClick={() => nav("/join")}
            className="flex h-[54px] items-center justify-center rounded-2xl bg-gold px-7 font-body text-[15px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
          >
            JOIN A ROOM
          </button>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  if (roomQ.isError && !room) {
    return (
      <PlayerDesktop>
        <PlayerTopBar left={<KweeksBrand />} />
        <div className="flex flex-1 flex-col items-center justify-center gap-4">
          <h1 className="font-display text-[30px] font-extrabold text-paper">
            Lost the connection to this room
          </h1>
          <button
            onClick={() => nav("/join")}
            className="flex h-[50px] items-center justify-center rounded-2xl bg-gold px-6 font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
          >
            BACK TO JOIN
          </button>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  const loading = !standings;
  const rows = standings ?? [];
  const ended = room?.state === "ended";
  const total = room?.participantCount ?? rows.length;
  const after = room ? room.currentIndex + 1 : 0;
  const questionsLeft = room ? Math.max(0, room.questionCount - room.currentIndex - 1) : 0;
  const winners = room?.winnerCount ?? 0;
  const nickname = player.nickname ?? "you";

  const myRow = rows.findIndex((s) => s.participantId === player.participantId);
  const rank = myRow >= 0 ? myRow + 1 : null;

  const pill =
    rank == null || !room ? (
      <div className="mt-1 rounded-2xl bg-surface-2 px-4 py-3.5 text-center font-body text-[14px] font-semibold text-text-2">
        Finding your spot…
      </div>
    ) : (
      <div className="mt-1 rounded-2xl bg-surface-2 px-4 py-3.5 text-center font-body text-[14px] font-semibold text-paper">
        {player.avatar ?? "🐙"} You're {ordinal(rank)} —{" "}
        {room?.state === "ended"
          ? "the game wrapped up. No winners this time."
          : room?.state === "live" && questionsLeft === 0
            ? "the last question is in. Results are being tallied."
            : `${questionsLeft} ${questionsLeft === 1 ? "question" : "questions"} left. Top ${winners} cash out.`}
      </div>
    );

  return (
    <PlayerDesktop>
      <PlayerTopBar
        left={
          <>
            <KweeksBrand />
            <span className="font-body text-[14px] text-text-3">/ Live standings</span>
          </>
        }
        right={
          ended ? (
            <span className="inline-flex items-center gap-[5px] rounded-full bg-surface-2 px-3 py-2">
              <span className="h-2 w-2 rounded-full bg-text-3" />
              <span className="font-body text-[11px] font-bold tracking-widest text-text-3">
                FINAL
              </span>
            </span>
          ) : (
            <span className="inline-flex items-center gap-[5px] rounded-full bg-surface-2 px-3 py-2">
              <span className="h-2 w-2 rounded-full bg-red" />
              <span className="font-body text-[11px] font-bold tracking-widest text-red">LIVE</span>
            </span>
          )
        }
      />

      <div className="flex flex-col gap-1.5 px-16 pt-7">
        <h1 className="font-display text-[30px] font-extrabold text-paper">Live standings</h1>
        <div className="font-body text-[14px] text-text-3">
          After question {after} · {total} players
        </div>
      </div>

      <div className="flex justify-center px-16 py-5">
        <div className="flex w-[760px] flex-col gap-2.5 rounded-[22px] border border-stroke bg-surface p-[26px]">
          <div className="flex justify-end pb-1 font-body text-[12px] text-text-3">
            pts = speed × correct · you play as {nickname}
          </div>
          {standingsQ.isError && !standings ? (
            <div className="rounded-2xl bg-surface-2 px-4 py-4 text-center font-body text-[13px] text-text-2">
              Couldn't load the board right now — retrying…
            </div>
          ) : loading
            ? [0, 1, 2, 3, 4].map((i) => (
                <div
                  key={i}
                  className="flex h-[62px] animate-pulse items-center gap-[18px] rounded-2xl bg-surface-2 px-4"
                />
              ))
            : rows.map((s, i) => {
                const you = s.participantId === player.participantId;
                const lead = i === 0;
                return (
                  <div
                    key={s.participantId}
                    className={cn(
                      "flex items-center gap-[18px] rounded-2xl px-4 py-3",
                      you ? "bg-gold" : "bg-surface-2",
                    )}
                  >
                    <span
                      className={cn(
                        "w-9 font-display text-[18px] font-extrabold",
                        you ? "text-gold-ink" : lead ? "text-gold" : "text-text-2",
                      )}
                    >
                      {i + 1}
                    </span>
                    <div className="flex h-11 w-11 items-center justify-center rounded-full bg-surface text-[22px]">
                      {s.avatar}
                    </div>
                    <div className="flex flex-1 items-center gap-2">
                      <span
                        className={cn(
                          "font-body text-[17px] font-semibold",
                          you ? "text-gold-ink" : "text-paper",
                        )}
                      >
                        {s.nickname}
                      </span>
                      {you && (
                        <span className="rounded-full bg-surface px-2 py-0.5 font-body text-[10px] font-extrabold tracking-wide text-gold-ink">
                          YOU
                        </span>
                      )}
                    </div>
                    <span
                      className={cn(
                        "font-display text-[20px] font-extrabold",
                        you ? "text-gold-ink" : lead ? "text-gold" : "text-paper",
                      )}
                    >
                      {fmtNaira(ptsFor(s))}
                    </span>
                  </div>
                );
              })}
          {!(standingsQ.isError && !standings) && pill}
        </div>
      </div>

      <PlayerFooter />
    </PlayerDesktop>
  );
}
