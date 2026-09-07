import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Check, Timer, Trophy, Users, X } from "lucide-react";
import { api, ApiError, type AnswerReceipt, type PublicQuestion } from "@/lib/api";
import { useRoom } from "@/lib/hooks";
import { usePlayer } from "@/lib/player";
import { Button } from "@/components/ui/button";
import { Footer } from "@/components/ui/footer";
import { PlayerTopBar } from "@/components/ui/player-topbar";
import { cn } from "@/lib/cn";

type Phase = "live" | "answered" | "timeout";

export function PlayerQuestion() {
  const navigate = useNavigate();
  const { roomId, participantId, markAnswered } = usePlayer();
  const { data: room } = useRoom(roomId ?? undefined);

  const question: PublicQuestion | null = room?.currentQuestion ?? null;
  const [phase, setPhase] = useState<Phase>("live");
  const [picked, setPicked] = useState<number | null>(null);
  const [receipt, setReceipt] = useState<AnswerReceipt | null>(null);
  const [error, setError] = useState("");
  const [now, setNow] = useState(() => Date.now());
  const submittedFor = useRef<string | null>(null);

  // Reset the local answer state the instant a brand-new question appears
  // (advance from the host, or a fresh question id on mount).
  useEffect(() => {
    if (!question) return;
    if (submittedFor.current === question.id) return;
    submittedFor.current = question.id;
    setPhase("live");
    setPicked(null);
    setReceipt(null);
    setError("");
  }, [question]);

  // Local countdown tick while a question is on screen. The timeout transition
  // happens inside the tick callback (not the effect body).
  useEffect(() => {
    if (!question || phase !== "live") return;
    const tick = () => {
      setNow(Date.now());
      const deadline = new Date(question.startedAt).getTime() + question.durationMs;
      if (deadline - Date.now() <= 0) {
        setPhase("timeout");
        markAnswered(question.index);
      }
    };
    const t = setInterval(tick, 250);
    return () => clearInterval(t);
  }, [question, phase, markAnswered]);

  const remaining = useMemo(() => {
    if (!question) return 0;
    const deadline = new Date(question.startedAt).getTime() + question.durationMs;
    return Math.max(0, deadline - now);
  }, [question, now]);

  async function answer(index: number) {
    if (!question || !roomId || !participantId || phase !== "live") return;
    if (submittedFor.current !== question.id) return;
    setPicked(index);
    setError("");
    try {
      const res = await api.post<AnswerReceipt>(`/rooms/${roomId}/answer`, {
        participantId,
        questionId: question.id,
        optionIndex: index,
      });
      setReceipt(res);
      markAnswered(question.index);
      setPhase("answered");
    } catch (err) {
      setPicked(null);
      const message = err instanceof ApiError ? err.message : "Answer did not go through — try again.";
      if (err instanceof ApiError && (err.status === 400 || err.status === 409)) {
        // Late / duplicate answer → treat as answered and move on.
        markAnswered(question.index);
        setPhase("timeout");
      } else {
        setError(message);
      }
    }
  }

  // Head to standings shortly after answering.
  useEffect(() => {
    if (phase !== "answered") return;
    const t = setTimeout(() => navigate("/standings"), 1100);
    return () => clearTimeout(t);
  }, [phase, navigate]);

  // Guard + cross-state navigation.
  useEffect(() => {
    if (!roomId) {
      navigate("/join", { replace: true });
      return;
    }
    if (room?.state === "lobby") navigate("/lobby");
    else if (room?.state === "podium" || room?.state === "ended") navigate("/podium");
  }, [room?.state, roomId, navigate]);

  const isCorrect = receipt?.correct;
  const questionCount = room?.questionCount ?? 0;

  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute left-1/2 top-0 size-96 -translate-x-1/2 rounded-full bg-violet/15 blur-3xl" />

      <PlayerTopBar status="live" code={room?.code} />

      <div className="relative mx-auto w-full max-w-4xl px-4 py-10 sm:px-6 lg:py-14">
        {question ? (
          <>
            <div className="flex items-center justify-between">
              <p className="font-display text-xl font-semibold text-soft">
                Question <span className="text-ink">{question.index + 1}</span>
                {questionCount > 0 && (
                  <span className="text-soft"> of {questionCount}</span>
                )}
              </p>
              <Button variant="outline" size="sm" onClick={() => navigate("/standings")}>
                <Users className="size-4" /> Standings
              </Button>
            </div>

            {/* Timer */}
            <div className="mt-6">
              <div className="flex items-center justify-between text-xs font-bold uppercase tracking-widest text-soft">
                <span>Time left</span>
                <span className="font-display text-lg text-ink">{(remaining / 1000).toFixed(1)}s</span>
              </div>
              <div className="mt-2 h-3.5 w-full overflow-hidden rounded-full bg-ink/5">
                <div
                  className={cn(
                    "h-full rounded-full transition-[width] duration-200 ease-linear",
                    phase === "live" ? "bg-gold" : phase === "answered" && isCorrect ? "bg-mint" : "bg-coral",
                  )}
                  style={{ width: `${Math.max(0, (remaining / question.durationMs) * 100)}%` }}
                />
              </div>
            </div>

            {/* Question card */}
            <div
              className="mt-8 rounded-[2rem] border-2 border-ink/5 bg-cream p-8 card-3d sm:p-10"
              style={{ animation: "rise 0.5s cubic-bezier(0.16,1,0.3,1) both" }}
            >
              <h1 className="font-display text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">
                {question.prompt}
              </h1>

              <div className="mt-8 grid gap-4 sm:grid-cols-2">
                {question.options.map((opt, i) => {
                  const isPicked = picked === i;
                  const showResult = phase !== "live";
                  const isCorrectOption = showResult && isPicked && isCorrect;
                  const isWrongPicked = showResult && isPicked && !isCorrect;
                  return (
                    <button
                      key={i}
                      type="button"
                      disabled={phase !== "live"}
                      onClick={() => void answer(i)}
                      className={cn(
                        "group relative flex min-h-20 items-center gap-4 rounded-2xl border-2 px-5 py-4 text-left font-bold transition-all duration-150",
                        phase === "live" &&
                          (isPicked
                            ? "border-violet bg-violet/10 shadow-[0_0_0_4px_rgba(108,76,241,0.15)]"
                            : "border-ink/10 bg-white hover:-translate-y-0.5 hover:border-violet/60 hover:shadow-[0_10px_25px_-12px_rgba(36,22,63,0.4)]"),
                        isCorrectOption && "border-mint bg-mint/10",
                        isWrongPicked && "border-coral bg-coral/10",
                        phase === "live" && "cursor-pointer",
                      )}
                    >
                      <span
                        className={cn(
                          "grid size-9 shrink-0 place-items-center rounded-xl font-display text-base font-semibold",
                          phase === "live" && (isPicked ? "bg-violet text-white" : "bg-ink/5 text-soft"),
                          isCorrectOption && "bg-mint text-white",
                          isWrongPicked && "bg-coral text-white",
                        )}
                      >
                        {isCorrectOption ? (
                          <Check className="size-4" strokeWidth={3} />
                        ) : isWrongPicked ? (
                          <X className="size-4" strokeWidth={3} />
                        ) : (
                          String.fromCharCode(65 + i)
                        )}
                      </span>
                      <span className="text-base sm:text-lg">{opt}</span>
                    </button>
                  );
                })}
              </div>

              {phase === "live" && (
                <p className="mt-6 text-center text-sm font-bold text-soft">
                  Tap an option to lock your answer. First correct tap counts.
                </p>
              )}
            </div>

            {error && (
              <p role="alert" className="mt-4 rounded-2xl border-2 border-coral/30 bg-coral/10 px-4 py-3 text-sm font-bold text-coral">
                {error}
              </p>
            )}

            {phase === "answered" && receipt && (
              <div
                className={cn(
                  "mt-6 flex items-center gap-4 rounded-[2rem] border-2 p-6",
                  isCorrect ? "border-mint/40 bg-mint/10" : "border-coral/40 bg-coral/10",
                )}
                style={{ animation: "pop 0.4s cubic-bezier(0.16,1,0.3,1) both" }}
              >
                <span
                  className={cn(
                    "grid size-14 shrink-0 place-items-center rounded-full text-white",
                    isCorrect ? "bg-mint" : "bg-coral",
                  )}
                >
                  {isCorrect ? <Check className="size-7" strokeWidth={3} /> : <X className="size-7" strokeWidth={3} />}
                </span>
                <div>
                  <p className="font-display text-2xl font-semibold">{isCorrect ? "Correct!" : "Not quite."}</p>
                  <p className="text-sm text-soft">
                    {isCorrect
                      ? `+${receipt.score} pts in ${receipt.latencyMs}ms — heading to the standings.`
                      : "The podium is decided by the whole board — standings incoming."}
                  </p>
                </div>
              </div>
            )}

            {phase === "timeout" && (
              <div className="mt-6 flex items-center gap-4 rounded-[2rem] border-2 border-coral/40 bg-coral/10 p-6">
                <span className="grid size-14 shrink-0 place-items-center rounded-full bg-coral text-white">
                  <Timer className="size-7" />
                </span>
                <div>
                  <p className="font-display text-2xl font-semibold">Time&apos;s up.</p>
                  <Button variant="ghost" size="sm" onClick={() => navigate("/standings")} className="mt-2">
                    <Trophy className="size-4" /> See standings
                  </Button>
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="mx-auto mt-20 max-w-md rounded-[2rem] border-2 border-ink/5 bg-cream p-8 text-center card-3d">
            <p className="font-display text-2xl font-semibold">Waiting for the next question…</p>
            <p className="mt-2 text-sm text-soft">The host controls the pace. This screen refreshes automatically.</p>
            <Button variant="ghost" className="mt-5" onClick={() => navigate("/standings")}>
              <Trophy className="size-4" /> View standings
            </Button>
          </div>
        )}
      </div>

      <Footer variant="player" />
    </main>
  );
}
