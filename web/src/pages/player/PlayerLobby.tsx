import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  PlayerDesktop,
  PlayerTopBar,
  PlayerFooter,
  RoomPill,
  KweeksBrand,
} from "@/components/ui/player-desktop";
import { useRoom } from "@/lib/hooks";
import { naira, usePlayer } from "@/lib/player";

const MAX_BADGES = 5;

function splitPoolNaira(pool: number, n: number): number[] {
  if (n <= 0 || pool <= 0) return [];
  if (n === 1) return [pool];
  const sum = (n * (n + 1)) / 2;
  const shares = [] as number[];
  for (let i = 0; i < n; i++) {
    shares.push(Math.floor((pool * (n - i)) / sum));
  }
  shares[0] += pool - shares.reduce((a, b) => a + b, 0);
  return shares;
}

export function PlayerLobby() {
  const nav = useNavigate();
  const player = usePlayer();
  const roomId = player.roomId ?? undefined;
  const roomQ = useRoom(roomId);
  const room = roomQ.data;

  const state = room?.state;

  useEffect(() => {
    if (!state) return;
    if (state === "live") nav("/question");
    else if (state === "podium") nav("/podium");
    else if (state === "ended") nav("/standings");
  }, [state, nav]);

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

  if (roomQ.isPending && !room) {
    return (
      <PlayerDesktop>
        <PlayerTopBar left={<KweeksBrand />} />
        <div className="flex flex-1 items-center justify-center">
          <div className="flex flex-col items-center gap-3">
            <span className="relative flex h-3 w-3">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
              <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
            </span>
            <span className="font-body text-[13px] text-text-2">Connecting to room…</span>
          </div>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  if (roomQ.isError && !room) {
    return (
      <PlayerDesktop>
        <PlayerTopBar left={<KweeksBrand />} />
        <div className="flex flex-1 flex-col items-center justify-center gap-4 px-20 text-center">
          <h1 className="font-display text-[30px] font-extrabold text-paper">
            Lost the connection to this room
          </h1>
          <p className="font-body text-[14px] text-text-2">
            The room may have closed, or the network hiccuped.
          </p>
          <div className="flex items-center gap-3">
            <button
              onClick={() => void roomQ.refetch()}
              className="flex h-[50px] items-center justify-center rounded-2xl bg-gold px-6 font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
            >
              RETRY
            </button>
            <button
              onClick={() => nav("/join")}
              className="flex h-[50px] items-center justify-center rounded-2xl border border-stroke bg-surface px-6 font-body text-[14px] font-extrabold tracking-wide text-paper"
            >
              BACK TO JOIN
            </button>
          </div>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  const roster = (room?.participants ?? []).slice(0, MAX_BADGES);
  const extra = Math.max(0, (room?.participants ?? []).length - roster.length);
  const poolN = parseInt(room?.poolNaira ?? "0", 10) || 0;
  const shares = room && room.winnerCount >= 1 ? splitPoolNaira(poolN, room.winnerCount) : [];
  const heroAvatar = player.avatar ?? "🐙";
  const who = player.nickname
    ? `you are ${player.nickname}${player.avatar ? ` ${player.avatar}` : ""}`
    : "you are in the room";

  return (
    <PlayerDesktop>
      <PlayerTopBar
        left={
          <>
            <KweeksBrand />
            {room && <RoomPill code={room.code} />}
          </>
        }
        right={
          <span className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
            WAITING
          </span>
        }
      />

      <div className="flex flex-1 items-center justify-center gap-14 px-20">
        <div className="flex w-[520px] flex-col gap-[18px]">
          <span className="text-[64px] leading-none">{heroAvatar}</span>
          <h1 className="font-display text-[38px] font-extrabold text-paper">
            You're in{player.nickname ? `, ${player.nickname}` : ""}
          </h1>
          <div className="font-body text-[18px] font-semibold text-naira">
            {room ? naira(room.poolNaira) : "₦0"} on the line
          </div>
          <p className="font-body text-[15px] leading-relaxed text-text-2">
            The host starts in a moment. Keep this tab open — questions move fast.
          </p>
          {room && shares.length > 0 ? (
            <div className="flex items-center justify-between rounded-2xl border border-stroke bg-surface px-5 py-[18px]">
              <span className="font-body text-[15px] text-text-2">1st place takes</span>
              <span className="font-display text-[24px] font-extrabold text-naira">
                {naira(shares[0])}
              </span>
            </div>
          ) : null}
        </div>

        <div className="flex w-[440px] flex-col gap-[18px]">
          <div className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
            IN THE ROOM
          </div>
          <div className="flex items-center gap-2.5">
            {roster.map((p) => (
              <div
                key={p.id}
                className="flex h-11 w-11 items-center justify-center rounded-full bg-surface-2 text-[24px]"
              >
                {p.avatar}
              </div>
            ))}
            {roster.length === 0 && (
              <div className="flex h-11 w-11 items-center justify-center rounded-full bg-surface-2 text-[16px] text-text-3">
                …
              </div>
            )}
          </div>
          <div className="font-body text-[13px] text-text-3">
            {extra > 0 ? `+ ${extra} more · ` : ""}
            {who}
          </div>

          <div className="flex flex-col gap-3 rounded-3xl border border-stroke bg-surface p-6">
            <div className="flex items-center gap-2.5">
              <span className="relative flex h-3 w-3">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
                <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
              </span>
              <h2 className="font-display text-[18px] font-extrabold text-paper">
                The host hasn't started yet
              </h2>
            </div>
            <p className="font-body text-[13.5px] leading-relaxed text-text-2">
              When the first question drops, everyone in this room sees it at the same second.
              Correct + fast = climb. No answer, no points.
            </p>
          </div>
        </div>
      </div>

      <PlayerFooter />
    </PlayerDesktop>
  );
}
