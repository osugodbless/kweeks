import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  PlayerDesktop,
  PlayerTopBar,
  PlayerFooter,
  KweeksBrand,
} from "@/components/ui/player-desktop";
import { ApiError } from "@/lib/api";
import { cn } from "@/lib/cn";
import { useRoom, useStandings, useSubmitAnswer } from "@/lib/hooks";
import { fmtNaira, usePlayer } from "@/lib/player";

const LETTERS = ["A", "B", "C", "D"];
const TICK_MS = 250;
const BAR_CLASSES = [
  "w-0", "w-[28px]", "w-[56px]", "w-[84px]", "w-[112px]", "w-[140px]",
  "w-[168px]", "w-[196px]", "w-[224px]", "w-[252px]", "w-[280px]", "w-[308px]",
  "w-[336px]", "w-[364px]", "w-[392px]", "w-[420px]", "w-[448px]", "w-[476px]",
  "w-[504px]", "w-[532px]", "w-[560px]",
];

function ptsFor(s: { correctCount: number; totalLatencyMs: number }): number {
  if (s.correctCount <= 0) return 0;
  const speed = Math.max(0, 1 - s.totalLatencyMs / (30_000 * s.correctCount));
  return s.correctCount * 1000 + Math.round(250 * speed);
}

type AnsState = "idle" | "sending" | "accepted" | "missed" | "retryable";

export function PlayerQuestion() {
  const nav = useNavigate();
  const player = usePlayer();
  const roomId = player.roomId ?? undefined;

  const roomQ = useRoom(roomId);
  const room = roomQ.data;
  const standingsQ = useStandings(roomId);
  const standings = standingsQ.data;
  const submit = useSubmitAnswer();

  const q = room?.state === "live" ? room.currentQuestion : null;
  const qid = q?.id;

  const [remain, setRemain] = useState(0);
  const [pick, setPick] = useState<number | null>(null);
  const [status, setStatus] = useState<AnsState>("idle");
  // remain starts at 0, which is also the "countdown finished" value. Until the
  // question's remaining time has actually been seeded, the countdown must not
  // be allowed to latch status to "missed" (otherwise every question would
  // disable its answers the moment it renders).
  const [armed, setArmed] = useState(false);

  useEffect(() => {
    setRemain(0);
    if (qid) setRemain(q?.remainingMs ?? 0);
  }, [qid]);

  useEffect(() => {
    if (!q) return;
    setRemain((r) => Math.max(0, Math.min(r, q.remainingMs)));
  }, [qid, q?.remainingMs]);

  useEffect(() => {
    if (!qid) return;
    const t = window.setInterval(() => {
      setRemain((r) => Math.max(0, r - TICK_MS));
    }, TICK_MS);
    return () => window.clearInterval(t);
  }, [qid]);

  useEffect(() => {
    setPick(null);
    setStatus("idle");
    setArmed(false);
  }, [qid]);

  // Arm the missed check only after remain has been seeded for this question.
  useEffect(() => {
    if (qid) setArmed(true);
  }, [qid, q?.remainingMs]);

  useEffect(() => {
    if (room?.state === "podium") nav("/podium");
    else if (room?.state === "ended") nav("/standings");
  }, [room?.state, nav]);

  useEffect(() => {
    if (!room || room.state !== "live" || !room.currentQuestion) return;
    if (room.pacing === "manual" && room.currentQuestion.remainingMs <= 0) {
      nav("/standings");
    }
  }, [
    room,
    room?.state,
    room?.currentQuestion?.id,
    room?.currentQuestion?.remainingMs,
    nav,
  ]);

  useEffect(() => {
    if (q && armed && remain <= 0 && status === "idle") setStatus("missed");
  }, [q, armed, remain, status]);

  const my = standings?.find((s) => s.participantId === player.participantId);
  const pts = my ? ptsFor(my) : 0;
  const seconds = q ? Math.max(0, Math.ceil(remain / 1000)) : 0;
  const durationMs = q?.durationMs ?? 0;
  const frac = durationMs > 0 ? Math.min(1, Math.max(0, remain / durationMs)) : 0;

  const canTap =
    Boolean(q) && remain > 0 && (status === "idle" || status === "retryable") && !submit.isPending;

  const ringR = 31;
  const ringC = 2 * Math.PI * ringR;
  const ringOff = ringC * (1 - frac);
  const barPct = Math.round(frac * 20);

  async function choose(optionIndex: number) {
    if (!canTap || !q || !player.roomId || !player.participantId) return;
    setPick(optionIndex);
    setStatus("sending");
    try {
      await submit.mutateAsync({
        roomId: player.roomId,
        participantId: player.participantId,
        questionId: q.id,
        optionIndex,
      });
      setStatus("accepted");
    } catch (e) {
      if (e instanceof ApiError && e.status === 0) {
        setStatus("retryable");
      } else if (e instanceof Error && /already answered/i.test(e.message)) {
        setStatus("accepted");
      } else {
        setStatus("missed");
      }
    }
  }

  let banner: string | null = null;
  if (status === "sending") banner = "Sending your answer…";
  else if (status === "accepted") banner = "✓ Answer locked in — watching the room…";
  else if (status === "retryable") banner = "Couldn't send — tap your answer again.";
  else if (status === "missed") banner = "No answer in time — this one's closed.";

  const headerLabel =
    room && room.state === "live" && q
      ? `Question ${q.index + 1} of ${room.questionCount}`
      : room && room.state === "live"
        ? "Live question"
        : room && room.state === "lobby"
          ? "Waiting to start"
          : "Standing by";

  const waiting =
    room &&
    ((room.state === "lobby" && !q) ||
      (room.state === "live" && !q) ||
      (room.state === "podium"));

  const title =
    room && room.state === "live"
      ? "Waiting for the next question…"
      : "Waiting for the host to start…";

  return (
    <PlayerDesktop>
      <PlayerTopBar
        left={
          <>
            <KweeksBrand />
            <span className="font-body text-[13px] text-text-3">{headerLabel}</span>
          </>
        }
        right={
          <span className="font-body text-[14px] font-bold text-gold">{fmtNaira(pts)} pts</span>
        }
      />

      {room && room.state === "live" && q ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-[26px] px-[72px] py-6">
          <div className="flex items-center gap-3">
            <div className="relative flex h-[76px] w-[76px] items-center justify-center">
              <svg viewBox="0 0 76 76" className="absolute inset-0 h-full w-full -rotate-90">
                <circle
                  cx="38"
                  cy="38"
                  r={ringR}
                  fill="none"
                  stroke="var(--color-surface-2)"
                  strokeWidth="6"
                />
                <circle
                  cx="38"
                  cy="38"
                  r={ringR}
                  fill="none"
                  stroke="var(--color-gold)"
                  strokeWidth="6"
                  strokeLinecap="round"
                  strokeDasharray={`${ringC}`}
                  strokeDashoffset={ringOff}
                />
              </svg>
              <span className="font-display text-[30px] font-extrabold text-paper">{seconds}</span>
            </div>
            <span className="font-body text-[13px] text-text-3">
              Answer before the bar runs out
            </span>
          </div>

          <div className="h-2 w-[560px] overflow-hidden rounded-full bg-surface-2">
            <div
              className={cn(
                "h-full rounded-full bg-gold transition-[width] duration-200",
                BAR_CLASSES[barPct],
              )}
            />
          </div>

          <h1 className="max-w-[760px] text-center font-display text-[38px] font-extrabold leading-[1.15] text-paper">
            {q.prompt}
          </h1>

          <div className="grid w-[760px] grid-cols-2 gap-4">
            {q.options.map((text, i) => {
              const on = pick === i;
              const disabled = !canTap;
              return (
                <button
                  key={i}
                  disabled={disabled}
                  onClick={() => void choose(i)}
                  className={cn(
                    "flex h-16 items-center gap-4 rounded-2xl border px-[22px] transition",
                    on ? "border-gold bg-gold" : "border-stroke bg-surface hover:border-text-3",
                    disabled && !on && "opacity-60",
                    disabled && "cursor-default",
                  )}
                >
                  <span
                    className={cn(
                      "font-display text-[18px] font-extrabold",
                      on ? "text-gold-ink" : "text-gold",
                    )}
                  >
                    {LETTERS[i] ?? i + 1}
                  </span>
                  <span
                    className={cn(
                      "font-body text-[18px] font-semibold",
                      on ? "text-gold-ink" : "text-paper",
                    )}
                  >
                    {text}
                  </span>
                </button>
              );
            })}
          </div>

          <div className="flex h-[22px] items-center justify-center">
            {banner && (
              <span
                className={cn(
                  "font-body text-[13px] font-semibold",
                  status === "accepted"
                    ? "text-naira"
                    : status === "missed" || status === "retryable"
                      ? "text-text-2"
                      : "text-gold",
                )}
              >
                {banner}
              </span>
            )}
          </div>
        </div>
      ) : waiting ? (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 px-20">
          <span className="relative flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
          </span>
          <h1 className="font-display text-[26px] font-extrabold text-paper">{title}</h1>
          <p className="font-body text-[13.5px] text-text-2">
            Keep this tab open — questions move fast.
          </p>
        </div>
      ) : (
        <div className="flex flex-1 items-center justify-center">
          <span className="font-body text-[13px] text-text-2">Loading…</span>
        </div>
      )}

      <PlayerFooter />
    </PlayerDesktop>
  );
}
