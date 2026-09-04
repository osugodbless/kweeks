import { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import {
  PlayerDesktop,
  PlayerTopBar,
  PlayerFooter,
  RoomPill,
  KweeksBrand,
} from "@/components/ui/player-desktop";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/cn";
import { useJoinRoom, useRoomByCode } from "@/lib/hooks";
import { AVATARS, naira, usePlayer } from "@/lib/player";

const ROWS = [AVATARS.slice(0, 6), AVATARS.slice(6, 12)];
const EMAIL_RE = /^\S+@\S+\.\S+$/;

const HOW = [
  "Same question, same second, everyone in the room.",
  "Correct + fast = points. No answer, no points.",
  "Top 3 take the pool — paid from their own screen.",
];

function stateLabel(state: string): { text: string; dot: string; on: boolean } {
  if (state === "live") return { text: "LIVE", dot: "bg-red", on: true };
  if (state === "lobby") return { text: "WAITING", dot: "bg-gold", on: false };
  if (state === "podium") return { text: "PODIUM", dot: "bg-naira", on: false };
  return { text: "ENDED", dot: "bg-surface-2", on: false };
}

export function PlayerJoin() {
  const nav = useNavigate();
  const player = usePlayer();
  const [sp] = useSearchParams();
  const urlCode = (sp.get("code") ?? "").toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 4);

  const [code, setCode] = useState(urlCode || player.code || "");
  const [avatar, setAvatar] = useState(player.avatar ?? "🐙");
  const [nick, setNick] = useState(player.nickname ?? "");
  const [email, setEmail] = useState(player.email ?? "");

  const lookup = useRoomByCode(code.length === 4 ? code : undefined);
  const room = lookup.data;

  const join = useJoinRoom();
  const joining = join.isPending;

  const codeOk = code.length === 4;
  const roomOk = Boolean(room);
  const nickOk = nick.trim().length > 0;
  const emailOk = EMAIL_RE.test(email.trim());
  const canJoin = codeOk && roomOk && nickOk && emailOk && !joining;

  const notFound =
    codeOk && !room && lookup.isError && lookup.error instanceof ApiError && lookup.error.status === 404;
  const roomError =
    codeOk && !room && lookup.isError && !(lookup.error instanceof ApiError && lookup.error.status === 404);

  function doJoin() {
    if (!canJoin || !room) return;
    join.mutate(
      {
        roomId: room.id,
        email: email.trim(),
        nickname: nick.trim(),
        avatar,
      },
      {
        onSuccess: (p) => {
          player.setRoom({ roomId: p.roomId, code: room.code });
          player.setParticipant({
            id: p.id,
            email: p.email,
            nickname: p.nickname,
            avatar: p.avatar,
          });
          nav("/lobby");
        },
      },
    );
  }

  const st = room ? stateLabel(room.state) : null;

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
          room ? (
            <span className="font-body text-[12.5px] text-text-3">
              {room.questionCount} questions · fastest wins
            </span>
          ) : null
        }
      />

      <div className="flex flex-1 items-center justify-center gap-14 px-20">
        <div className="flex w-[430px] flex-col gap-4">
          <div className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
            PRIZE POOL
          </div>
          <div className="font-display text-[76px] font-extrabold leading-none text-naira">
            {room ? naira(room.poolNaira) : "—"}
          </div>
          {room && st ? (
            <>
              <div className="flex items-center gap-2">
                <span className="inline-flex items-center gap-1.5 rounded-full bg-surface-2 px-3 py-1.5">
                  <span className={cn("h-2 w-2 rounded-full", st.dot)} />
                  <span className={cn("font-body text-[11px] font-bold tracking-widest", st.on ? "text-red" : "text-text-2")}>
                    {st.text}
                  </span>
                </span>
                <span className="font-body text-[15px] font-semibold text-paper">{room.title}</span>
              </div>
              <p className="font-body text-[13.5px] text-text-2">
                {room.questionCount} questions · top {room.winnerCount} share the pool · join below
              </p>
            </>
          ) : (
            <p className="font-body text-[18px] leading-normal text-text-2">
              You're in the arena. Type your room code, pick a face and grab a seat.
            </p>
          )}
          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface p-[22px]">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              HOW IT PLAYS
            </div>
            {HOW.map((t) => (
              <p key={t} className="font-body text-[13.5px] leading-relaxed text-text-2">
                • {t}
              </p>
            ))}
          </div>
        </div>
        <div className="flex w-[560px] flex-col gap-[18px] rounded-3xl border border-stroke bg-surface p-[30px]">
          <h1 className="font-display text-[24px] font-extrabold text-paper">Join the room</h1>
          <label className="block">
            <span className="mb-1.5 block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
              ROOM CODE
            </span>
            <div
              className={cn(
                "flex h-[54px] items-center rounded-[14px] border bg-surface-2 px-4",
                codeOk ? "border-gold" : "border-stroke",
              )}
            >
              <input
                value={code}
                onChange={(e) =>
                  setCode(e.target.value.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 4))
                }
                onKeyDown={(e) => {
                  if (e.key === "Enter") (e.target as HTMLInputElement).blur();
                }}
                placeholder="AB12"
                autoComplete="off"
                spellCheck={false}
                className="w-full bg-transparent pl-[0.4em] text-center font-display text-[26px] font-extrabold tracking-[0.4em] text-gold outline-none placeholder:font-body placeholder:text-[16px] placeholder:font-semibold placeholder:text-text-3"
              />
            </div>
          </label>
          {notFound ? (
            <div className="-mt-2 font-body text-[12.5px] font-semibold text-red">
              Room not found — check the code and try again.
            </div>
          ) : roomError ? (
            <div className="-mt-2 font-body text-[12.5px] font-semibold text-red">
              Couldn't reach the server — is the host running?
            </div>
          ) : !codeOk ? (
            <div className="-mt-2 font-body text-[12.5px] text-text-3">
              Type the 4-letter code your host shared.
            </div>
          ) : lookup.isFetching && !room ? (
            <div className="-mt-2 font-body text-[12.5px] text-text-3">Looking up room…</div>
          ) : room ? (
            <div className="-mt-2 font-body text-[12.5px] text-text-2">
              {room.poolNaira ? `${naira(room.poolNaira)} pool · ` : ""}
              {room.questionCount} questions · top {room.winnerCount} paid
            </div>
          ) : null}

          <div className="font-body text-[13px] font-semibold text-text-2">Pick your avatar</div>
          <div className="flex flex-col gap-2">
            {ROWS.map((row, ri) => (
              <div key={ri} className="flex gap-2">
                {row.map((e) => (
                  <button
                    key={e}
                    type="button"
                    onClick={() => setAvatar(e)}
                    className={cn(
                      "flex h-16 w-16 items-center justify-center rounded-full text-[30px] transition",
                      avatar === e
                        ? "bg-gold ring-1 ring-gold"
                        : "bg-surface-2 ring-1 ring-stroke",
                    )}
                  >
                    {e}
                  </button>
                ))}
              </div>
            ))}
          </div>
          <div className="font-body text-[12.5px] text-text-3">
            One tap. It's how the room will know you.
          </div>
          <div className="flex items-center gap-3">
            <label className="block flex-1">
              <span className="mb-1.5 block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
                NICKNAME
              </span>
              <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface-2 px-4">
                <input
                  value={nick}
                  onChange={(e) => setNick(e.target.value)}
                  placeholder="e.g. FastestZebra"
                  className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
                />
              </div>
            </label>
            <label className="block flex-1">
              <span className="mb-1.5 block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
                EMAIL (FOR YOUR PRIZE)
              </span>
              <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface-2 px-4">
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") void doJoin();
                  }}
                  placeholder="you@example.com"
                  className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
                />
              </div>
            </label>
          </div>

          {join.isError ? (
            <div className="font-body text-[12.5px] font-semibold text-red">
              {join.error instanceof Error
                ? join.error.message
                : "Couldn't join — check your details."}
            </div>
          ) : null}

          <button
            onClick={() => void doJoin()}
            disabled={!canJoin}
            className="flex h-[58px] w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90 disabled:opacity-40"
          >
            {joining ? "JOINING…" : "JOIN THE GAME"}
          </button>
          <div className="text-center font-body text-[12.5px] text-text-3">
            Winnings land in your email · protected by claim code
          </div>
        </div>
      </div>

      <PlayerFooter />
    </PlayerDesktop>
  );
}
