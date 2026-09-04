import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";
import { api, QuizDetail, QuizQuestion } from "@/lib/api";
import { useCreateQuiz, useOpenRoom, useQuiz, qk } from "@/lib/hooks";
import { fmtNaira, naira } from "@/lib/player";

const LETTERS = ["A", "B", "C", "D"];

const DEFAULT_TITLE = "Naija General Knowledge";
const DEFAULT_DURATION_MS = 30000;
const POOL_MIN = 1000;
const POOL_MAX = 150000;
const POOL_STEP = 1000;

interface DraftQuestion {
  id: string;
  prompt: string;
  options: string[];
  correctIndex: number;
  durationMs: number;
}

function blankQuestion(overrides?: Partial<DraftQuestion>): DraftQuestion {
  return {
    id: crypto.randomUUID(),
    prompt: "",
    options: ["", "", "", ""],
    correctIndex: 0,
    durationMs: DEFAULT_DURATION_MS,
    ...overrides,
  };
}

const SEED_QUESTION = blankQuestion({
  prompt: "Which Nigerian city is called the Centre of Excellence?",
  options: ["Ibadan", "Lagos", "Abuja", "Enugu"],
  correctIndex: 1,
});

function errMsg(e: unknown): string {
  return e instanceof Error ? e.message : "Something went wrong.";
}

function shortLabel(prompt: string): string {
  const t = prompt.trim();
  if (!t) return "New question";
  return t.length > 18 ? `${t.slice(0, 18)}…` : t;
}

export function InstructorQuizBuilder() {
  const nav = useNavigate();
  const [sp] = useSearchParams();
  const quizId = sp.get("quiz") ?? undefined;
  const editing = Boolean(quizId);

  const quizQ = useQuiz(quizId);
  const createQuiz = useCreateQuiz();
  const openRoom = useOpenRoom();
  const qc = useQueryClient();
  const editingReady = editing && !quizQ.isError;
  const busy = createQuiz.isPending || openRoom.isPending || (editing && quizQ.isPending);

  const [title, setTitle] = useState(DEFAULT_TITLE);
  const [pool, setPool] = useState(50000);
  const [winnerCount, setWinnerCount] = useState(3);
  const [pacing, setPacing] = useState<"auto" | "manual">("manual");
  const [questions, setQuestions] = useState<DraftQuestion[]>([SEED_QUESTION]);
  const [active, setActive] = useState(0);
  const [appliedId, setAppliedId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const d = quizQ.data;
    if (!d || !d.id || d.id === appliedId) return;
    setAppliedId(d.id);
    setTitle(d.title);
    const p = parseInt(d.poolNaira || "", 10);
    setPool(Number.isFinite(p) ? p : 50000);
    setWinnerCount(d.winnerCount);
    setPacing(d.pacing);
    setQuestions(
      d.questions.length > 0
        ? d.questions.map((q: QuizQuestion) => ({
            id: q.id,
            prompt: q.prompt,
            options:
              q.options.length === 4 ? [...q.options] : [...q.options, ...Array(4 - q.options.length).fill("")].slice(0, 4),
            correctIndex: q.correctIndex,
            durationMs: q.durationMs,
          }))
        : [blankQuestion()],
    );
    setActive(0);
  }, [quizQ.data, appliedId]);

  const toQuizQuestion = (q: DraftQuestion) => ({
    id: q.id,
    prompt: q.prompt,
    options: q.options,
    correctIndex: q.correctIndex,
    durationMs: q.durationMs,
  });

  const body = () => ({
    title: title.trim() || DEFAULT_TITLE,
    poolNaira: String(pool),
    winnerCount,
    pacing,
    defaultDurationMs: DEFAULT_DURATION_MS,
    questions: questions.map(toQuizQuestion),
  });

  const ensureSaved = async (): Promise<string> => {
    if (editingReady && quizId) {
      await api.put<{ id: string }>(`/quizzes/${quizId}`, body());
      await qc.invalidateQueries({ queryKey: qk.quiz(quizId) });
      await qc.invalidateQueries({ queryKey: qk.quizzes });
      return quizId;
    }
    const payload: QuizDetail = { id: crypto.randomUUID(), ...body() };
    const res = await createQuiz.mutateAsync(payload);
    return res.id;
  };

  const handleSave = async () => {
    setError(null);
    try {
      const id = await ensureSaved();
      nav(`/instructor/live-room?quiz=${id}`);
    } catch (e) {
      setError(errMsg(e));
    }
  };

  const handleOpen = async () => {
    setError(null);
    try {
      const id = await ensureSaved();
      const room = await openRoom.mutateAsync(id);
      nav(`/instructor/live-room?room=${room.id}`);
    } catch (e) {
      setError(errMsg(e));
    }
  };

  const setPrompt = (qi: number, v: string) =>
    setQuestions((qs) => qs.map((q, i) => (i === qi ? { ...q, prompt: v } : q)));

  const setOption = (qi: number, oi: number, v: string) =>
    setQuestions((qs) =>
      qs.map((q, i) => (i === qi ? { ...q, options: q.options.map((o, j) => (j === oi ? v : o)) } : q)),
    );

  const setCorrect = (qi: number, oi: number) =>
    setQuestions((qs) => qs.map((q, i) => (i === qi ? { ...q, correctIndex: oi } : q)));

  const addQuestion = () => {
    const next = questions.length;
    setQuestions((qs) => [...qs, blankQuestion()]);
    setActive(next);
  };

  const removeQuestion = (qi: number) => {
    if (questions.length <= 1) return;
    const next = questions.filter((_, i) => i !== qi);
    setQuestions(next);
    let na = active;
    if (active === qi) na = Math.min(qi, next.length - 1);
    else if (active > qi) na = active - 1;
    setActive(Math.max(0, Math.min(na, next.length - 1)));
  };

  const pct = Math.min(100, Math.max(0, ((pool - POOL_MIN) / (POOL_MAX - POOL_MIN)) * 100));

  return (
    <DesktopFrame>
      <InstructorNav activeKey="create" right={<NavRight />} />
      <div className="flex flex-1 gap-6 bg-bg px-7 py-7">
        <div className="flex w-[360px] shrink-0 flex-col gap-4">
          <div className="flex flex-col overflow-hidden rounded-2xl border border-stroke bg-surface">
            <img
              src="https://images.unsplash.com/photo-1710075769969-4b12fe6e223c?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=360&ixlib=rb-1.2.1&q=80&w=1080"
              alt="Lagos"
              className="h-24 w-full object-cover"
            />
            <div className="flex flex-col gap-2 px-5 py-4">
              <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
                QUIZ TITLE
              </span>
              <input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder={DEFAULT_TITLE}
                className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
              />
              {editing && quizQ.isPending && (
                <span className="font-body text-[11px] text-text-3">Loading quiz…</span>
              )}
            </div>
          </div>

          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              PRIZE POOL (FROM YOUR WALLET)
            </span>
            <div className="flex items-baseline gap-1">
              <span className="font-display text-[24px] font-extrabold text-naira">₦</span>
              <span className="font-display text-[32px] font-extrabold leading-none text-naira">
                {fmtNaira(pool)}
              </span>
            </div>
            <div className="relative h-[6px] w-full rounded-full bg-surface-2">
              <div
                className="absolute left-0 top-0 h-full rounded-full bg-naira"
                style={{ width: `${pct}%` }}
              />
              <input
                type="range"
                min={POOL_MIN}
                max={POOL_MAX}
                step={POOL_STEP}
                value={pool}
                onChange={(e) => setPool(parseInt(e.target.value, 10))}
                aria-label="Prize pool in naira"
                className="absolute inset-0 h-full w-full cursor-pointer appearance-none bg-transparent opacity-0"
              />
              <div
                className="pointer-events-none absolute top-1/2 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-naira bg-gold"
                style={{ left: `${pct}%` }}
              />
            </div>
            <div className="flex justify-between font-body text-[11px] text-text-3">
              <span>₦1k</span>
              <span>₦50k</span>
              <span>₦150k</span>
            </div>
            <span className="font-body text-[12px] leading-snug text-text-3">
              Pool is deducted when the room opens and held until winners redeem.
            </span>
          </div>

          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              WINNERS
            </span>
            <div className="flex gap-1.5">
              {[1, 2, 3, 5].map((w) => (
                <button
                  key={w}
                  onClick={() => setWinnerCount(w)}
                  className={cn(
                    "flex h-9 flex-1 items-center justify-center rounded-full font-display text-[16px] font-bold",
                    winnerCount === w ? "bg-gold text-gold-ink" : "bg-surface-2 text-text-2",
                  )}
                >
                  {w}
                </button>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              PACING
            </span>
            <div className="flex gap-1.5">
              {(["auto", "manual"] as const).map((p) => (
                <button
                  key={p}
                  onClick={() => setPacing(p)}
                  className={cn(
                    "flex h-9 flex-1 items-center justify-center rounded-full font-body text-[13px] font-bold",
                    pacing === p ? "bg-gold text-gold-ink" : "bg-surface-2 text-text-2",
                  )}
                >
                  {p.toUpperCase()}
                </button>
              ))}
            </div>
            <span className="font-body text-[12.5px] leading-snug text-text-2">
              Auto advances on a timer. Manual = you tap next. Venue mode: Manual.
            </span>
          </div>
        </div>

        <div className="flex flex-1 flex-col gap-4">
          <div className="flex items-center justify-between">
            <h2 className="font-display text-[20px] font-extrabold text-paper">Questions</h2>
            <button
              onClick={addQuestion}
              className="flex items-center gap-1.5 rounded-xl bg-gold px-4 py-2.5 font-body text-[14px] font-extrabold text-gold-ink"
            >
              <span className="font-display text-[18px]">+</span>Add question
            </button>
          </div>

          {editing && quizQ.isError && (
            <div className="rounded-2xl border border-red/40 bg-surface px-5 py-4 font-body text-[13.5px] text-red">
              Couldn't load that quiz ({errMsg(quizQ.error)}). You can keep editing below — saving
              will create a fresh quiz.
            </div>
          )}

          <div className="flex flex-col gap-3">
            {questions.map((q, qi) => {
              const on = active === qi;
              return (
                <div
                  key={q.id}
                  onClick={() => setActive(qi)}
                  className={cn(
                    "flex cursor-pointer flex-col gap-4 rounded-2xl border bg-surface px-6 py-5",
                    on ? "border-gold" : "border-stroke",
                  )}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
                      QUESTION {qi + 1} · {Math.round(q.durationMs / 1000)}s
                    </span>
                    {questions.length > 1 && (
                      <button
                        type="button"
                        onClick={(e) => {
                          e.stopPropagation();
                          removeQuestion(qi);
                        }}
                        className="font-body text-[12px] font-bold text-text-3 transition-colors hover:text-red"
                      >
                        REMOVE
                      </button>
                    )}
                  </div>
                  <textarea
                    rows={2}
                    value={q.prompt}
                    onChange={(e) => setPrompt(qi, e.target.value)}
                    placeholder="Type the question…"
                    className="w-full resize-none bg-transparent font-display text-[20px] font-extrabold leading-snug text-paper outline-none placeholder:text-text-3"
                  />
                  <div className="grid grid-cols-2 gap-3">
                    {[0, 1, 2, 3].map((oi) => {
                      const correct = q.correctIndex === oi;
                      return (
                        <div
                          key={oi}
                          onClick={() => setCorrect(qi, oi)}
                          className={cn(
                            "flex cursor-pointer items-center gap-3 rounded-2xl border px-4 py-3",
                            correct ? "border-naira bg-surface-2" : "border-stroke bg-surface",
                          )}
                        >
                          <span
                            className={cn(
                              "flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-surface-2 text-[12px]",
                              correct ? "text-gold-ink" : "text-paper",
                            )}
                          >
                            {LETTERS[oi]}
                          </span>
                          <input
                            value={q.options[oi]}
                            onClick={(e) => e.stopPropagation()}
                            onChange={(e) => setOption(qi, oi, e.target.value)}
                            placeholder={`Option ${LETTERS[oi]}`}
                            className={cn(
                              "w-full min-w-0 bg-transparent font-body text-[15px] outline-none placeholder:text-text-3",
                              correct ? "font-semibold text-naira" : "text-paper",
                            )}
                          />
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              setCorrect(qi, oi);
                            }}
                            className={cn(
                              "ml-auto flex shrink-0 items-center gap-1 rounded-full px-2 py-1 text-[11px] font-semibold",
                              correct ? "bg-surface text-naira" : "bg-surface-2 text-text-3",
                            )}
                          >
                            {correct ? "✓ correct" : "mark ✓"}
                          </button>
                        </div>
                      );
                    })}
                  </div>
                </div>
              );
            })}
          </div>

          <div className="flex flex-wrap items-center gap-1.5">
            {questions.map((q, qi) => (
              <button
                key={q.id}
                onClick={() => setActive(qi)}
                className={cn(
                  "flex flex-col items-center rounded-xl border px-4 py-2",
                  active === qi ? "border-gold bg-surface-2" : "border-stroke bg-surface",
                )}
              >
                <span
                  className={cn(
                    "font-display text-[15px] font-extrabold",
                    active === qi ? "text-gold" : "text-text-2",
                  )}
                >
                  {qi + 1}
                </span>
                <span
                  className={cn(
                    "font-body text-[10px]",
                    active === qi ? "text-text-2" : "text-text-3",
                  )}
                >
                  {shortLabel(q.prompt)}
                </span>
              </button>
            ))}
            <button
              onClick={addQuestion}
              className="flex flex-col items-center rounded-xl border border-stroke bg-surface px-4 py-2"
            >
              <span className="font-display text-[15px] font-extrabold text-text-2">+</span>
              <span className="font-body text-[10px] text-text-2">add</span>
            </button>
          </div>

          {error && (
            <div className="rounded-2xl border border-red/40 bg-surface px-5 py-4 font-body text-[13.5px] text-red">
              {error}
            </div>
          )}

          <div className="mt-auto flex items-center justify-between gap-4">
            <span className="font-body text-[13px] text-text-3">
              {questions.length} questions · pool {naira(pool)} · {winnerCount} winners · {pacing}
            </span>
            <div className="flex items-center gap-2.5">
              <button
                onClick={handleSave}
                disabled={busy}
                className="rounded-2xl border border-naira bg-transparent px-6 py-3 font-body text-[15px] font-extrabold text-naira transition-colors hover:bg-naira/10 disabled:opacity-50"
              >
                {busy ? "WORKING…" : editingReady ? "SAVE QUIZ" : "CREATE QUIZ"}
              </button>
              <button
                onClick={handleOpen}
                disabled={busy}
                className="rounded-2xl bg-gold px-6 py-3 font-body text-[15px] font-extrabold text-gold-ink transition-opacity hover:opacity-90 disabled:opacity-50"
              >
                {busy ? "WORKING…" : "Open room →"}
              </button>
            </div>
          </div>
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
