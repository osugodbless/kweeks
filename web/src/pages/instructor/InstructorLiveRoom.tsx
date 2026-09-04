import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";
import { QuizListItem } from "@/lib/api";
import {
  qk,
  useOpenRoom,
  useQuizzes,
  useQuiz,
  useRoom,
  useRoomControl,
  useRoomSocket,
  useStandings,
} from "@/lib/hooks";
import { naira } from "@/lib/player";

const LETTERS = ["A", "B", "C", "D"];

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : "Something went wrong.";
}

function isOpenRoom(q: QuizListItem): boolean {
  return Boolean(q.roomId && q.state && q.state !== "ended");
}

function Chrome({ children }: { children: React.ReactNode }) {
  return (
    <DesktopFrame>
      <InstructorNav activeKey="create" right={<NavRight />} />
      {children}
      <InstructorFooter />
    </DesktopFrame>
  );
}

function Wait({ children = "Loading…" }: { children?: React.ReactNode }) {
  return (
    <div className="flex flex-1 items-center justify-center bg-bg px-7 py-20">
      <span className="font-body text-[15px] text-text-2">{children}</span>
    </div>
  );
}

function Fail({
  message,
  onBack,
}: {
  message: string;
  onBack?: () => void;
}) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 bg-bg px-7 py-20">
      <div className="max-w-[420px] rounded-2xl border border-red/40 bg-surface px-6 py-5 text-center">
        <div className="font-body text-[13px] font-bold tracking-widest text-red">API ERROR</div>
        <div className="mt-1 font-body text-[14px] text-paper">{message}</div>
      </div>
      {onBack && (
        <button
          onClick={onBack}
          className="rounded-xl bg-gold px-5 py-2.5 font-body text-[13px] font-extrabold text-gold-ink"
        >
          PICK A QUIZ
        </button>
      )}
    </div>
  );
}

export function InstructorLiveRoom() {
  const [sp] = useSearchParams();
  const roomId = sp.get("room");
  const quizId = sp.get("quiz");

  if (roomId) return <RoomView roomId={roomId} />;
  if (quizId) return <QuizOpenView quizId={quizId} />;
  return <PickerView />;
}

function PickerView() {
  const nav = useNavigate();
  const quizzesQ = useQuizzes();
  const openRoom = useOpenRoom();
  const [busyId, setBusyId] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  const quizzes = quizzesQ.data ?? [];

  const handleOpen = async (q: QuizListItem) => {
    setBusyId(q.id);
    setErr(null);
    try {
      const room = await openRoom.mutateAsync(q.id);
      nav(`/instructor/live-room?room=${room.id}`);
    } catch (e) {
      setErr(errMsg(e));
    } finally {
      setBusyId(null);
    }
  };

  return (
    <Chrome>
      <div className="flex flex-1 flex-col bg-bg px-7 py-7">
        <h1 className="font-display text-[26px] font-extrabold text-paper">Open a live room</h1>
        <p className="mt-1 font-body text-[14px] text-text-2">
          Pick a quiz to put on the projector. Players join with the room code.
        </p>

        <div className="mt-6 flex flex-1 flex-col gap-3">
          {quizzesQ.isPending && <Wait>Loading your quizzes…</Wait>}
          {quizzesQ.isError && (
            <Fail message={errMsg(quizzesQ.error)} onBack={() => nav("/instructor/dashboard")} />
          )}
          {!quizzesQ.isPending && !quizzesQ.isError && quizzes.length === 0 && (
            <div className="flex flex-1 flex-col items-center justify-center gap-4 rounded-2xl border border-stroke bg-surface px-6 py-16 text-center">
              <span className="text-[40px]">🎯</span>
              <div className="font-display text-[20px] font-extrabold text-paper">
                No quizzes yet
              </div>
              <div className="max-w-[380px] font-body text-[13.5px] text-text-2">
                Author your first quiz in the builder, then come back here to put it live.
              </div>
              <button
                onClick={() => nav("/instructor/quiz-builder")}
                className="flex items-center gap-1.5 rounded-xl bg-gold px-5 py-3 font-body text-[14px] font-extrabold text-gold-ink"
              >
                <span className="font-display text-[18px]">+</span>CREATE A QUIZ
              </button>
            </div>
          )}
          {!quizzesQ.isPending &&
            !quizzesQ.isError &&
            quizzes.map((q) => {
              const liveRoom = isOpenRoom(q);
              const busy = busyId === q.id;
              return (
                <div
                  key={q.id}
                  className="flex items-center justify-between gap-4 rounded-2xl border border-stroke bg-surface px-5 py-4"
                >
                  <div className="min-w-0">
                    <div className="flex items-center gap-2.5">
                      <span className="truncate font-body text-[15px] font-bold text-paper">
                        {q.title}
                      </span>
                      {liveRoom && (
                        <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-surface px-2.5 py-0.5">
                          <span className="h-2 w-2 rounded-full bg-naira" />
                          <span className="font-body text-[11px] font-bold tracking-widest text-naira">
                            LIVE
                          </span>
                        </span>
                      )}
                    </div>
                    <div className="mt-0.5 font-body text-[12.5px] text-text-3">
                      {q.questionCount} questions · pool {naira(q.poolNaira)} · {q.winnerCount}{" "}
                      winners · {q.pacing}
                    </div>
                  </div>
                  {liveRoom ? (
                    <button
                      onClick={() => nav(`/instructor/live-room?room=${q.roomId}`)}
                      className="shrink-0 rounded-xl border-[1.5px] border-naira bg-transparent px-5 py-2.5 font-body text-[13px] font-extrabold text-naira"
                    >
                      Open live room →
                    </button>
                  ) : (
                    <button
                      onClick={() => handleOpen(q)}
                      disabled={busy}
                      className="shrink-0 rounded-xl bg-gold px-5 py-2.5 font-body text-[13px] font-extrabold text-gold-ink disabled:opacity-50"
                    >
                      {busy ? "OPENING…" : "OPEN ROOM"}
                    </button>
                  )}
                </div>
              );
            })}
        </div>

        {err && (
          <div className="mt-4 rounded-2xl border border-red/40 bg-surface px-5 py-4 font-body text-[13.5px] text-red">
            {err}
          </div>
        )}
      </div>
    </Chrome>
  );
}

function QuizOpenView({ quizId }: { quizId: string }) {
  const nav = useNavigate();
  const quizQ = useQuiz(quizId);
  const quizzesQ = useQuizzes();
  const openRoom = useOpenRoom();
  const [err, setErr] = useState<string | null>(null);

  const quiz = quizQ.data;
  const liveRoom = quizzesQ.data?.find(
    (q) => q.id === quizId && q.roomId && q.state && q.state !== "ended",
  );

  const handleOpen = async () => {
    setErr(null);
    try {
      const room = await openRoom.mutateAsync(quizId);
      nav(`/instructor/live-room?room=${room.id}`);
    } catch (e) {
      setErr(errMsg(e));
    }
  };

  return (
    <Chrome>
      <div className="flex flex-1 items-center justify-center bg-bg px-7 py-7">
        {quizQ.isPending && <Wait>Loading quiz…</Wait>}
        {quizQ.isError && (
          <Fail message={errMsg(quizQ.error)} onBack={() => nav("/instructor/live-room")} />
        )}
        {!quizQ.isPending && !quizQ.isError && quiz && (
          <div className="flex w-[540px] flex-col gap-6 rounded-3xl border border-stroke bg-surface p-8">
            <div>
              <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
                READY TO HOST
              </div>
              <h1 className="mt-1 font-display text-[26px] font-extrabold leading-tight text-paper">
                {quiz.title}
              </h1>
              <div className="mt-2 font-body text-[13.5px] text-text-2">
                {quiz.questions.length} questions ·{" "}
                <span className="font-bold text-naira">{naira(quiz.poolNaira)}</span> pool ·{" "}
                {quiz.winnerCount} winners · {quiz.pacing}
              </div>
            </div>

            <div className="flex flex-col gap-2.5">
              <button
                onClick={handleOpen}
                disabled={openRoom.isPending}
                className="flex h-14 w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink disabled:opacity-50"
              >
                {openRoom.isPending ? "OPENING ROOM…" : "OPEN ROOM"}
              </button>
              {liveRoom && (
                <button
                  onClick={() => nav(`/instructor/live-room?room=${liveRoom.roomId}`)}
                  className="flex h-12 w-full items-center justify-center rounded-2xl border-[1.5px] border-naira bg-transparent font-body text-[14px] font-extrabold text-naira"
                >
                  Room {liveRoom.roomCode ?? ""} is already live → join it
                </button>
              )}
            </div>

            {err && (
              <div className="rounded-2xl border border-red/40 bg-surface px-5 py-4 font-body text-[13.5px] text-red">
                {err}
              </div>
            )}

            <div className="font-body text-[12.5px] leading-relaxed text-text-3">
              Opening a room deducts the pool from your wallet and holds it until winners redeem.
              Players open the join page and type the room code shown on the live screen.
            </div>
          </div>
        )}
      </div>
    </Chrome>
  );
}

function RoomView({ roomId }: { roomId: string }) {
  const nav = useNavigate();
  const qc = useQueryClient();

  const roomQ = useRoom(roomId);
  const standQ = useStandings(roomId);
  const quizQ = useQuiz(roomQ.data?.quizId);
  const control = useRoomControl();
  const { event } = useRoomSocket(roomId);

  const room = roomQ.data;
  const q = room?.currentQuestion ?? null;
  const standings = standQ.data ?? [];
  const state = room?.state;
  const live = state === "live";
  const lobby = state === "lobby";
  const podium = state === "podium";
  // The last question is in play when the current index is the final one.
  const onLast = Boolean(live && room && room.questionCount > 0 && room.currentIndex >= room.questionCount - 1);
  const lastDone = Boolean(live && !q && room && room.currentIndex >= room.questionCount - 1);
  const correctIndex =
    live && q ? quizQ.data?.questions.find((qq) => qq.id === q.id)?.correctIndex ?? -1 : -1;

  useEffect(() => {
    if (!event) return;
    void qc.invalidateQueries({ queryKey: qk.room(roomId) });
    void qc.invalidateQueries({ queryKey: qk.standings(roomId) });
  }, [event, qc, roomId]);

  const [secs, setSecs] = useState(0);
  useEffect(() => {
    if (state !== "live" || !q) return;
    setSecs(Math.max(0, Math.ceil(q.remainingMs / 1000)));
    const t = window.setInterval(() => setSecs((s) => Math.max(0, s - 1)), 1000);
    return () => window.clearInterval(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state, q?.id]);

  const totalSecs = q ? Math.max(1, Math.ceil(q.durationMs / 1000)) : 30;
  const frac = totalSecs > 0 ? Math.min(1, Math.max(0, secs / totalSecs)) : 0;
  const CIRC = 188.5;

  const rightInfo = room
    ? live && q
      ? `Q ${q.index + 1} · ${room.participantCount} players`
      : `${room.participantCount} players · room ${room.code}`
    : undefined;

  const phase = !room
    ? ""
    : podium
      ? "Podium — winners below"
      : state === "ended"
        ? "Room ended"
        : lobby
          ? `Waiting in the lobby · ${room.code}`
          : live && q
            ? `Question ${q.index + 1} of ${room.questionCount}`
            : live && lastDone
              ? "Final question complete — declare winners"
              : "Preparing next question…";

  let content: React.ReactNode;
  if (roomQ.isPending) {
    content = <Wait>Opening room…</Wait>;
  } else if (roomQ.isError || !room) {
    content = (
      <Fail message={errMsg(roomQ.error ?? "Room not found")} onBack={() => nav("/instructor/live-room")} />
    );
  } else {
    content = (
      <div className="flex flex-1 gap-6 bg-bg px-7 py-7">
        <div className="flex flex-1 flex-col">
          <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            PROJECTOR PREVIEW
          </div>
          <div className="mt-3 flex flex-1 flex-col rounded-3xl border border-stroke bg-surface px-10 py-8">
            <div className="flex items-start justify-between gap-4">
              <div className="flex min-w-0 flex-col gap-0.5">
                <span className="truncate font-body text-[15px] font-semibold text-paper">
                  {room.title}
                </span>
                <span className="font-body text-[13px] text-text-3">{phase}</span>
              </div>
              <span className="shrink-0 font-display text-[18px] font-extrabold text-naira">
                {naira(room.poolNaira)} POOL
              </span>
            </div>

            {lobby && (
              <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
                <div className="font-display text-[28px] font-extrabold text-paper">
                  Room is in the lobby — press START
                </div>
                <div className="font-body text-[14px] text-text-2">
                  Players on the join screen will see the first question the moment you start.
                </div>
              </div>
            )}

            {live && q && (
              <>
                <div className="mt-8 flex flex-col items-center">
                  <div className="relative flex h-[72px] w-[72px] items-center justify-center">
                    <svg viewBox="0 0 72 72" className="h-[72px] w-[72px] -rotate-90">
                      <circle cx="36" cy="36" r="30" fill="none" stroke="var(--color-surface-2)" strokeWidth="6" />
                      <circle
                        cx="36"
                        cy="36"
                        r="30"
                        fill="none"
                        stroke="var(--color-gold)"
                        strokeWidth="6"
                        strokeDasharray={`${CIRC}`}
                        strokeDashoffset={CIRC * (1 - frac)}
                        strokeLinecap="round"
                      />
                    </svg>
                    <span className="absolute font-display text-[36px] font-extrabold text-paper">
                      {secs}
                    </span>
                  </div>
                </div>

                <h1 className="mt-6 text-center font-display text-[34px] font-extrabold leading-tight text-paper">
                  {q.prompt}
                </h1>

                <div className="mt-8 grid grid-cols-2 gap-4">
                  {q.options.map((opt, i) => {
                    const correct = correctIndex === i;
                    return (
                      <div
                        key={i}
                        className={cn(
                          "flex items-center gap-3 rounded-2xl border px-5 py-4",
                          correct ? "border-naira bg-surface-2" : "border-stroke bg-surface",
                        )}
                      >
                        <span
                          className={cn(
                            "flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface-2 text-[15px]",
                            correct ? "text-gold-ink" : "text-text-2",
                          )}
                        >
                          {LETTERS[i]}
                        </span>
                        <span
                          className={cn(
                            "min-w-0 flex-1 font-body text-[16px]",
                            correct ? "font-semibold text-naira" : "text-paper",
                          )}
                        >
                          {opt}
                        </span>
                        {correct && <span className="text-[15px] text-naira">✓</span>}
                      </div>
                    );
                  })}
                </div>

                {standings.length > 0 && (
                  <div className="mt-auto flex items-center gap-2 rounded-2xl bg-surface-2 px-5 py-3.5">
                    <span className="font-body text-[14px] font-bold text-naira">
                      ▲ {standings[0].correctCount}
                    </span>
                    <span className="font-body text-[14px] text-paper">{standings[0].nickname}</span>
                    <span className="font-body text-[13.5px] text-text-2">
                      leads with {standings[0].correctCount} correct — {standings.length} on the
                      board
                    </span>
                  </div>
                )}
              </>
            )}

            {live && !q && (
              <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
                <div className="font-display text-[26px] font-extrabold text-paper">
                  {lastDone
                    ? "That was the last question."
                    : "Waiting for the next question…"}
                </div>
                <div className="font-body text-[14px] text-text-2">
                  {lastDone
                    ? "Declare winners to lock the podium and pay out the pool."
                    : "Players are locked out until you move things forward."}
                </div>
              </div>
            )}

            {(podium || state === "ended") && (
              <div className="mt-6 flex flex-1 flex-col gap-3">
                {room.winners && room.winners.length > 0 ? (
                  room.winners.map((w, i) => (
                    <div
                      key={w.participantId}
                      className={cn(
                        "flex items-center gap-4 rounded-2xl border px-5 py-4",
                        i === 0 ? "border-gold bg-surface-2" : "border-stroke bg-surface",
                      )}
                    >
                      <span
                        className={cn(
                          "flex h-9 w-9 shrink-0 items-center justify-center rounded-full font-display text-[16px] font-extrabold",
                          i === 0 ? "bg-gold text-gold-ink" : "bg-surface-2 text-paper",
                        )}
                      >
                        {i === 0 ? "🥇" : i === 1 ? "🥈" : i === 2 ? "🥉" : i + 1}
                      </span>
                      <span className="text-[22px]">{w.avatar}</span>
                      <div className="flex min-w-0 flex-1 flex-col">
                        <span className="truncate font-body text-[15px] font-bold text-paper">
                          {w.nickname}
                        </span>
                        <span className="font-body text-[12.5px] text-text-3">
                          {w.correctCount} correct
                        </span>
                      </div>
                      {i === 0 && (
                        <span className="font-display text-[20px] font-extrabold text-naira">
                          {naira(room.poolNaira)}
                        </span>
                      )}
                    </div>
                  ))
                ) : (
                  <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
                    <div className="font-display text-[24px] font-extrabold text-paper">
                      No winners this round
                    </div>
                    <div className="max-w-[360px] font-body text-[13.5px] text-text-2">
                      Nobody answered a question correctly, so the pool stays in the wallet. Run the
                      quiz again with a fresh room.
                    </div>
                    <button
                      onClick={() => nav("/instructor/dashboard")}
                      className="mt-1 rounded-xl bg-surface px-5 py-2.5 font-body text-[13px] font-bold text-text-2 hover:text-paper"
                    >
                      Back to dashboard
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="flex w-[330px] shrink-0 flex-col gap-5">
          <div className="flex flex-col gap-4 rounded-2xl border border-stroke bg-surface px-5 py-5">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              PLAYERS JOIN HERE
            </div>
            <div className="text-center font-body text-[12px] text-text-3">
              Open the room code below on a player phone
            </div>
            <div className="rounded-2xl border border-stroke bg-surface-2 px-4 py-4 text-center">
              <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
                ROOM CODE
              </div>
              <div className="mt-1 font-display text-[44px] font-extrabold tracking-[0.1em] text-gold">
                {room.code}
              </div>
            </div>
            <button
              onClick={() => {
                const url = `${window.location.origin}/join?code=${room.code}`;
                void navigator.clipboard?.writeText(url).catch(() => {});
              }}
              className="flex items-center justify-center gap-2 rounded-xl bg-surface px-4 py-3 font-body text-[13px] font-bold text-text-2 hover:text-paper"
            >
              <span className="font-display text-[14px]">⧉</span> COPY JOIN LINK
            </button>
            <div className="flex flex-col items-center gap-0.5">
              <span className="font-body text-[12px] text-text-3">or have players open</span>
              <span className="font-display text-[14px] font-bold text-paper">
                {window.location.origin}/join
              </span>
              <span className="font-body text-[12px] text-text-3">and type the code</span>
            </div>
            <div className="flex min-h-[28px] items-center justify-center gap-1.5">
              {room.participants.length === 0 && (
                <span className="font-body text-[12px] text-text-3">No players yet</span>
              )}
              {room.participants.slice(0, 8).map((p) => (
                <div
                  key={p.id}
                  className="flex h-[28px] w-[28px] items-center justify-center rounded-full bg-surface-2 text-[17px]"
                >
                  {p.avatar}
                </div>
              ))}
              {room.participants.length > 8 && (
                <span className="flex h-[28px] w-[28px] items-center justify-center rounded-full bg-surface-2 font-body text-[12px] font-bold text-text-2">
                  +{room.participants.length - 8}
                </span>
              )}
            </div>
            <div className="text-center font-body text-[12px] font-semibold text-text-2">
              {room.participantCount} players joined
            </div>
          </div>

          <div className="flex flex-col gap-4 rounded-2xl border border-stroke bg-surface px-5 py-5">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              CONTROL · {room.pacing.toUpperCase()}
            </div>

            {lobby && (
              <button
                onClick={() => control.start.mutate(roomId)}
                disabled={control.start.isPending}
                className="flex h-12 w-full items-center justify-center rounded-2xl bg-gold font-body text-[15px] font-extrabold text-gold-ink disabled:opacity-50"
              >
                {control.start.isPending ? "STARTING…" : "START QUIZ →"}
              </button>
            )}

            {live && !lastDone && room.pacing === "manual" && !onLast && (
              <button
                onClick={() => control.next.mutate(roomId)}
                disabled={control.next.isPending}
                className="flex h-12 w-full items-center justify-center rounded-2xl bg-gold font-body text-[15px] font-extrabold text-gold-ink disabled:opacity-50"
              >
                {control.next.isPending ? "ADVANCING…" : "NEXT QUESTION →"}
              </button>
            )}

            {(lastDone || (live && room.pacing === "manual" && onLast)) && (
              <button
                onClick={() => control.podium.mutate(roomId)}
                disabled={control.podium.isPending}
                className="flex h-12 w-full items-center justify-center rounded-2xl border-[1.5px] border-naira bg-transparent font-body text-[14px] font-extrabold text-naira disabled:opacity-50"
              >
                {control.podium.isPending ? "DECLARING…" : "DECLARE WINNERS"}
              </button>
            )}

            {podium && (
              <div className="rounded-2xl border border-gold bg-surface-2 px-4 py-3 font-body text-[12.5px] leading-snug text-gold">
                Winners locked in — they redeem the pool from their own screens.
              </div>
            )}

            {state === "ended" && (
              <div className="rounded-2xl bg-surface-2 px-4 py-3 font-body text-[12.5px] leading-snug text-text-2">
                This room is closed. Open a fresh one from your dashboard.
              </div>
            )}

            <div className="font-body text-[12px] text-text-3">
              {room.pacing === "manual"
                ? "Manual waits for your tap. Move on when the room has answered."
                : "Auto is on a 30s timer. Rooms close after the last question."}
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="create"
        liveLabel={state === "live" ? "LIVE" : undefined}
        right={
          <NavRight
            center={
              rightInfo ? (
                <span className="font-body text-[13px] text-text-2">{rightInfo}</span>
              ) : undefined
            }
          />
        }
      />
      {content}
      <InstructorFooter />
    </DesktopFrame>
  );
}
