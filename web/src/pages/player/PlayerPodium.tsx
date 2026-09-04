import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  PlayerDesktop,
  PlayerTopBar,
  PlayerFooter,
  KweeksBrand,
} from "@/components/ui/player-desktop";
import { ClaimResult } from "@/lib/api";
import { cn } from "@/lib/cn";
import { useRedeem, useRoom } from "@/lib/hooks";
import { naira, usePlayer } from "@/lib/player";

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

function ordinal(n: number): string {
  const s = ["th", "st", "nd", "rd"];
  const v = n % 100;
  return `${n}${s[(v - 20) % 10] ?? s[v] ?? s[0]}`;
}

function stepReached(state: string | undefined, i: number): boolean {
  if (!state) return false;
  if (state === "created") return i === 0;
  if (state === "invited" || state === "onboarded") return i <= 1;
  if (state === "paid") return i <= 2;
  return false;
}

const STEPS = [
  { n: "1", t: "Claim locked", s: "Your win is recorded against your email." },
  { n: "2", t: "Invite sent", s: "We email you a secure redemption link." },
  { n: "3", t: "Paid", s: "Money moves from the pool wallet to you." },
];

export function PlayerPodium() {
  const nav = useNavigate();
  const player = usePlayer();
  const roomId = player.roomId ?? undefined;

  const roomQ = useRoom(roomId);
  const room = roomQ.data;
  const redeem = useRedeem();

  const [claim, setClaim] = useState<ClaimResult | null>(null);
  const [copied, setCopied] = useState(false);

  const wins = room?.winners ?? [];
  const poolN = parseInt(room?.poolNaira ?? "0", 10) || 0;
  const shares = splitPoolNaira(poolN, wins.length);
  const youIdx = wins.findIndex((w) => w.participantId === player.participantId);
  const youWin = youIdx >= 0;
  const myShare = youWin ? shares[youIdx] : 0;
  const code = claim?.claimCode;

  useEffect(() => {
    if (room?.state === "ended") nav("/standings");
  }, [room?.state, nav]);

  function doRedeem() {
    if (!roomId || !player.email) return;
    redeem.mutate(
      { roomId, email: player.email },
      { onSuccess: (c) => setClaim(c) },
    );
  }

  async function copyCode() {
    if (!code) return;
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  }

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
          <span className="font-body text-[13px] text-text-2">Connecting to room…</span>
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

  if (room && room.state !== "podium") {
    const liveQ = room.state === "live" && room.currentQuestion;
    return (
      <PlayerDesktop>
        <PlayerTopBar
          left={
            <>
              <KweeksBrand />
              <span className="font-body text-[14px] text-text-3">/ Final standings</span>
            </>
          }
          right={
            <span className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
              {room.state === "live" ? "LIVE" : room.state === "lobby" ? "WAITING" : room.state.toUpperCase()}
            </span>
          }
        />
        <div className="flex flex-1 flex-col items-center justify-center gap-3 px-20">
          <span className="relative flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
          </span>
          <h1 className="font-display text-[26px] font-extrabold text-paper">
            {room.state === "live"
              ? liveQ
                ? "The game is still running…"
                : "Waiting for the next question…"
              : "Waiting for the host to start…"}
          </h1>
          {liveQ && (
            <button
              onClick={() => nav("/question")}
              className="mt-1 flex h-[50px] items-center justify-center rounded-2xl border border-stroke bg-surface px-6 font-body text-[14px] font-extrabold tracking-wide text-paper"
            >
              BACK TO QUESTION
            </button>
          )}
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  if (room && room.state === "podium" && !Array.isArray(room.winners)) {
    return (
      <PlayerDesktop>
        <PlayerTopBar left={<KweeksBrand />} />
        <div className="flex flex-1 flex-col items-center justify-center gap-3 px-20">
          <span className="relative flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
          </span>
          <h1 className="font-display text-[26px] font-extrabold text-paper">
            Loading the results…
          </h1>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  if (room && wins.length === 0) {
    return (
      <PlayerDesktop>
        <PlayerTopBar
          left={
            <>
              <KweeksBrand />
              <span className="font-body text-[14px] text-text-3">/ Final standings</span>
            </>
          }
          right={
            <span className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
              GAME OVER
            </span>
          }
        />
        <div className="flex flex-1 flex-col items-center justify-center gap-3 px-20 text-center">
          <h1 className="font-display text-[30px] font-extrabold text-paper">
            Nobody took the pool this time
          </h1>
          <p className="font-body text-[14px] text-text-2">
            No winner qualified — the pool rolls back to the host.
          </p>
          <button
            onClick={() => nav("/standings")}
            className="mt-2 flex h-[50px] items-center justify-center rounded-2xl bg-gold px-6 font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
          >
            VIEW FINAL STANDINGS
          </button>
        </div>
        <PlayerFooter />
      </PlayerDesktop>
    );
  }

  const placeText = `${ordinal(youIdx + 1)} place · ${room?.participantCount ?? 0} players`;

  return (
    <PlayerDesktop>
      <PlayerTopBar
        left={
          <>
            <KweeksBrand />
            <span className="font-body text-[14px] text-text-3">/ Final standings</span>
          </>
        }
        right={
          <span className="font-body text-[12px] font-bold tracking-[0.18em] text-text-3">
            GAME OVER
          </span>
        }
      />

      <div className="flex flex-1 items-center justify-center gap-12 px-16 py-5">
        <div className="flex w-[520px] flex-col gap-3.5">
          <div className="h-[200px] w-full overflow-hidden rounded-[20px]">
            <img
              src="https://images.unsplash.com/photo-1693670984742-c008c239daf2?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=480&ixlib=rb-1.2.1&q=80&w=1080"
              alt="Celebration with confetti"
              className="h-full w-full object-cover"
            />
          </div>
          {youWin ? (
            <>
              <div className="font-body text-[14px] text-text-3">{placeText}</div>
              <div className="font-display text-[72px] font-extrabold leading-none text-naira">
                {naira(myShare)}
              </div>
              <div className="font-body text-[16px] text-text-2">
                You banked {naira(myShare)} from the pool. Claim it below.
              </div>
            </>
          ) : (
            <>
              <div className="font-body text-[14px] text-text-3">
                Thanks for playing · {room?.participantCount ?? 0} players
              </div>
              <div className="font-display text-[44px] font-extrabold leading-none text-paper">
                See you next time
              </div>
              <div className="font-body text-[16px] text-text-2">
                You didn't crack the top {room?.winnerCount ?? 0} this round. The next pool could
                be yours.
              </div>
            </>
          )}
        </div>
        <div className="flex w-[480px] flex-col gap-4">
          <h2 className="font-display text-[20px] font-extrabold text-paper">Winners</h2>
          {wins.map((w, i) => {
            const you = w.participantId === player.participantId;
            const amt = shares[i] ?? 0;
            return (
              <div
                key={w.participantId}
                className={cn(
                  "flex items-center gap-4 rounded-2xl px-4 py-3",
                  you ? "bg-gold" : "bg-surface-2",
                )}
              >
                <span
                  className={cn(
                    "w-7 font-display text-[20px] font-extrabold",
                    you ? "text-gold-ink" : i === 0 ? "text-gold" : "text-text-2",
                  )}
                >
                  {i + 1}
                </span>
                <div className="flex h-11 w-11 items-center justify-center rounded-full bg-surface text-[22px]">
                  {w.avatar}
                </div>
                <span
                  className={cn(
                    "flex-1 font-body text-[17px] font-bold",
                    you ? "text-gold-ink" : "text-paper",
                  )}
                >
                  {you ? "YOU" : w.nickname}
                </span>
                <span
                  className={cn(
                    "font-display text-[18px] font-extrabold",
                    you ? "text-gold-ink" : "text-naira",
                  )}
                >
                  {naira(amt)}
                </span>
              </div>
            );
          })}

          {youWin ? (
            <>
              <div className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface p-5">
                <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
                  YOUR CLAIM CODE
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-display text-[22px] font-extrabold tracking-wide text-gold">
                    {code ? code : claim ? "•••• ••••" : "— — — —"}
                  </span>
                  <button
                    onClick={() => void copyCode()}
                    disabled={!code}
                    className={cn(
                      "rounded-xl bg-surface-2 px-3.5 py-2 font-body text-[12px] font-bold text-paper",
                      !code && "opacity-40",
                    )}
                  >
                    {copied ? "COPIED" : "COPY"}
                  </button>
                </div>
                <div className="font-body text-[12.5px] text-text-3">
                  {!claim
                    ? "Redeem to unlock your code. Don't share it."
                    : code
                      ? "Only this device and email can redeem it. Don't share it."
                      : `Redemption registered — your secure code was sent to ${player.email}.`}
                </div>
              </div>

              {redeem.isError && (
                <div className="font-body text-[12.5px] font-semibold text-red">
                  {redeem.error instanceof Error ? redeem.error.message : "Couldn't redeem yet."}
                </div>
              )}

              <button
                onClick={() => void doRedeem()}
                disabled={Boolean(claim) || redeem.isPending || !player.email}
                className="flex h-[58px] w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90 disabled:opacity-40"
              >
                {redeem.isPending
                  ? "REDEEMING…"
                  : claim
                    ? "CLAIMED ✓"
                    : `REDEEM ${naira(myShare)} NOW`}
              </button>

              {STEPS.map((st, i) => {
                const done = claim ? stepReached(claim.state, i) : false;
                return (
                  <div key={st.n} className="flex items-center gap-3">
                    <div
                      className={cn(
                        "flex h-7 w-7 items-center justify-center rounded-full",
                        done ? "bg-gold" : "bg-surface-2",
                      )}
                    >
                      <span
                        className={cn(
                          "font-display text-[13px] font-extrabold",
                          done ? "text-gold-ink" : "text-naira",
                        )}
                      >
                        {done ? "✓" : st.n}
                      </span>
                    </div>
                    <div>
                      <div
                        className={cn(
                          "font-body text-[14px] font-bold",
                          done ? "text-paper" : "text-text-3",
                        )}
                      >
                        {st.t}
                      </div>
                      <div className="font-body text-[12.5px] text-text-3">{st.s}</div>
                    </div>
                  </div>
                );
              })}
            </>
          ) : (
            <div className="flex flex-col items-center gap-2.5 rounded-2xl border border-stroke bg-surface p-6 text-center">
              <div className="font-display text-[18px] font-extrabold text-paper">
                Thanks for playing
              </div>
              <p className="font-body text-[13px] leading-relaxed text-text-2">
                You kept up with the room — the top spots just weren't yours today.
              </p>
            </div>
          )}
        </div>
      </div>

      <PlayerFooter />
    </PlayerDesktop>
  );
}
