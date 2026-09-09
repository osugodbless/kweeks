import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Check, FileQuestion, Plus, Presentation, Trash2, Trophy, Wallet } from "lucide-react";
import { ApiError, type QuizDetail, type QuizQuestion } from "@/lib/api";
import { useCreateQuiz, useOpenRoom, useQuiz, useUpdateQuiz, useWallet } from "@/lib/hooks";
import { naira } from "@/lib/player";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Field } from "@/components/ui/field";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { cn } from "@/lib/cn";

const POOL_MIN = 1000;
const POOL_MAX = 200000;
const POOL_STEP = 500;

const WINNER_COUNTS = [1, 3, 5];
const PACING = [
  { id: "manual", label: "Manual" },
  { id: "auto", label: "Auto" },
];
const DURATIONS = [
  { id: 10000, label: "10s" },
  { id: 15000, label: "15s" },
  { id: 20000, label: "20s" },
  { id: 30000, label: "30s" },
];

function newQuestion(): QuizQuestion {
  return {
    id: crypto.randomUUID(),
    prompt: "",
    options: ["", "", "", ""],
    correctIndex: 0,
    durationMs: 15000,
  };
}

export function InstructorQuizBuilder() {
  const [params] = useSearchParams();
  const existingId = params.get("id") ?? undefined;
  const { data: existing, isLoading } = useQuiz(existingId);

  if (existingId && isLoading) {
    return <BuilderLoading label="Loading your quiz…" />;
  }

  return (
    <BuilderForm
      key={existing?.id ?? "new"}
      existing={existing}
      existingId={existingId}
    />
  );
}

function BuilderLoading({ label }: { label: string }) {
  return (
    <main className="relative min-h-screen overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <InstructorNav />
      <div className="relative mx-auto flex w-full max-w-6xl justify-center px-4 pt-24 text-sm font-bold text-soft sm:px-6">
        {label}
      </div>
    </main>
  );
}

interface BuilderFormProps {
  existing?: QuizDetail;
  existingId?: string;
}

function BuilderForm({ existing, existingId }: BuilderFormProps) {
  const navigate = useNavigate();
  const createQuiz = useCreateQuiz();
  const updateQuiz = useUpdateQuiz();
  const openRoom = useOpenRoom();
  const { data: walletView } = useWallet();

  const [title, setTitle] = useState(existing?.title ?? "");
  const [poolNaira, setPoolNaira] = useState(existing?.poolNaira ?? "50000");
  const [winnerCount, setWinnerCount] = useState(existing?.winnerCount ?? 3);
  const [pacing, setPacing] = useState<"manual" | "auto">(existing?.pacing ?? "manual");
  const [defaultDurationMs, setDefaultDurationMs] = useState(existing?.defaultDurationMs ?? 15000);
  const [questions, setQuestions] = useState<QuizQuestion[]>(
    existing?.questions.length ? existing.questions.map((q) => ({ ...q })) : [newQuestion()],
  );
  const [error, setError] = useState("");
  const [confirmDelete, setConfirmDelete] = useState<QuizQuestion | null>(null);
  const [addedToast, setAddedToast] = useState<{ id: string; num: number } | null>(null);

  const poolNum = useMemo(() => parseInt(poolNaira || "0", 10), [poolNaira]);

  // Live affordability so the host sees the gap at the point of action, before
  // ever clicking "Save & open room".
  const available = useMemo(() => parseInt(walletView?.wallet.balanceNaira ?? "0", 10) || 0, [walletView]);
  const shortfall = Math.max(0, poolNum - available);
  const insufficient = shortfall > 0;

  const ctaRef = useRef<HTMLDivElement | null>(null);
  // When a submit fails, keep the message on screen right where the button is.
  useEffect(() => {
    if (error) ctaRef.current?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [error]);

  function updateQuestion(id: string, patch: Partial<QuizQuestion>) {
    setQuestions((qs) => qs.map((q) => (q.id === id ? { ...q, ...patch } : q)));
  }

  function removeQuestion(id: string) {
    setQuestions((qs) => qs.filter((q) => q.id !== id));
    setConfirmDelete(null);
  }

  function addQuestion() {
    const q = newQuestion();
    setQuestions((qs) => [...qs, q]);
    setAddedToast({ id: q.id, num: questions.length + 1 });
  }

  // Scroll the freshly added question into view and auto-dismiss the toast.
  useEffect(() => {
    if (!addedToast) return;
    const timer = setTimeout(() => setAddedToast(null), 2600);
    setTimeout(() => {
      document.getElementById(`question-${addedToast.id}`)?.scrollIntoView({ behavior: "smooth", block: "center" });
    }, 60);
    return () => clearTimeout(timer);
  }, [addedToast]);

  function validate(): string | null {
    if (!title.trim()) return "Give the quiz a title.";
    if (poolNum < POOL_MIN) return `Pool must be at least ${naira(POOL_MIN)}.`;
    if (questions.length === 0) return "Add at least one question.";
    for (const q of questions) {
      if (!q.prompt.trim()) return "Every question needs a prompt.";
      const filled = q.options.filter((o) => o.trim());
      if (filled.length < 2) return `"${q.prompt.trim() || "Untitled"}" needs at least 2 options.`;
      if (q.options[q.correctIndex]?.trim() === "") return `Mark a valid correct answer for "${q.prompt.trim()}".`;
    }
    return null;
  }

  async function handleOpenRoom() {
    setError("");
    const problem = validate();
    if (problem) {
      setError(problem);
      return;
    }
    try {
      const quiz = {
        id: existingId ?? "",
        title: title.trim(),
        poolNaira: String(poolNum),
        winnerCount,
        pacing,
        defaultDurationMs,
        questions: questions.map((q) => ({
          ...q,
          options: q.options.map((o) => o.trim()),
        })),
      };
      const created = existingId
        ? await updateQuiz.mutateAsync({ ...quiz, id: existingId })
        : await createQuiz.mutateAsync(quiz);
      const room = await openRoom.mutateAsync(created.id);
      navigate(`/instructor/live-room?room=${room.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not open the room — try again.");
    }
  }

  const saving = createQuiz.isPending || updateQuiz.isPending || openRoom.isPending;

  return (
    <main className="relative min-h-screen overflow-hidden pb-24">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-violet/20 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-gold/15 blur-3xl" />

      <InstructorNav />

      <div className="relative mx-auto w-full max-w-4xl px-4 pt-12 sm:px-6 lg:pt-16">
        <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
          <FileQuestion className="size-4 text-violet" /> Quiz builder
        </span>
        <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
          Build your <span className="text-pop">quiz</span>
        </h1>
        <p className="mt-3 max-w-xl text-lg text-soft">
          {existingId ? "Editing an existing deck — save and open a room when ready." : "Name the deck, set the prize, write the questions. Open the room when it's ready."}
        </p>

        {/* Deck settings */}
        <section className="mt-10 rounded-[2rem] border-2 border-ink/5 bg-cream p-7 card-3d">
          <div className="grid gap-6 lg:grid-cols-2">
            <div className="lg:col-span-2">
              <Field
                name="title"
                label="Quiz title"
                value={title}
                onChange={(e) => setTitle(e.target.value.slice(0, 60))}
                placeholder="e.g. Naija General Knowledge"
                autoComplete="off"
                spellCheck="false"
              />
            </div>

            <div>
              <div className="mb-2 flex items-center justify-between">
                <label htmlFor="pool" className="text-sm font-bold">Prize pool</label>
                <Chip tone="mint"><Trophy className="size-3" /> {naira(poolNum)}</Chip>
              </div>
              <input
                id="pool"
                type="range"
                min={POOL_MIN}
                max={POOL_MAX}
                step={POOL_STEP}
                value={Math.min(Math.max(poolNum, POOL_MIN), POOL_MAX)}
                onChange={(e) => setPoolNaira(e.target.value)}
                className="w-full accent-mint-dark"
              />
              <div className="mt-1 flex justify-between text-xs font-bold text-soft">
                <span>{naira(POOL_MIN)}</span>
                <span>{naira(POOL_MAX)}</span>
              </div>
            </div>

            <div>
              <p className="text-sm font-bold">Winner count</p>
              <div className="mt-2 flex gap-2">
                {WINNER_COUNTS.map((n) => (
                  <button
                    key={n}
                    type="button"
                    onClick={() => setWinnerCount(n)}
                    className={cn(
                      "cursor-pointer rounded-full border-2 px-4 py-2 text-sm font-bold transition-all",
                      winnerCount === n
                        ? "border-gold bg-gold text-ink"
                        : "border-ink/10 bg-white text-soft hover:border-gold hover:text-gold-dark",
                    )}
                  >
                    {n} {n === 1 ? "winner" : "winners"}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <p className="text-sm font-bold">Pacing</p>
              <div className="mt-2 flex gap-2">
                {PACING.map((p) => (
                  <button
                    key={p.id}
                    type="button"
                    onClick={() => setPacing(p.id as "manual" | "auto")}
                    className={cn(
                      "cursor-pointer rounded-full border-2 px-4 py-2 text-sm font-bold transition-all",
                      pacing === p.id
                        ? "border-violet bg-violet text-white"
                        : "border-ink/10 bg-white text-soft hover:border-violet hover:text-violet",
                    )}
                  >
                    {p.label}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <p className="text-sm font-bold">Default question time</p>
              <div className="mt-2 flex gap-2">
                {DURATIONS.map((d) => (
                  <button
                    key={d.id}
                    type="button"
                    onClick={() => setDefaultDurationMs(d.id)}
                    className={cn(
                      "cursor-pointer rounded-full border-2 px-4 py-2 text-sm font-bold transition-all",
                      defaultDurationMs === d.id
                        ? "border-sky bg-sky text-white"
                        : "border-ink/10 bg-white text-soft hover:border-sky hover:text-sky-dark",
                    )}
                  >
                    {d.label}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Questions */}
        <div className="mt-10 flex items-center justify-between">
          <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
            <Presentation className="size-5 text-violet" /> Questions ({questions.length})
          </h2>
          <Button variant="outline" size="sm" icon={<Plus className="size-4" />} onClick={addQuestion}>
            Add question
          </Button>
        </div>

        <div className="mt-5 space-y-6">
          {questions.map((q, qi) => (
            <section key={q.id} id={`question-${q.id}`} className="scroll-mt-24 rounded-[2rem] border-2 border-ink/5 bg-cream p-6 sm:p-7">
              <div className="flex items-start justify-between gap-4">
                <span className="flex items-center gap-2 font-display text-lg font-semibold">
                  <span className="grid size-8 place-items-center rounded-xl bg-violet/10 text-violet-dark">{qi + 1}</span>
                  Question {qi + 1}
                </span>
                <button
                  type="button"
                  onClick={() => setConfirmDelete(q)}
                  disabled={questions.length === 1}
                  aria-label={`Delete question ${qi + 1}`}
                  className="grid size-9 cursor-pointer place-items-center rounded-xl text-soft transition-colors hover:bg-coral/10 hover:text-coral disabled:cursor-not-allowed disabled:opacity-40"
                >
                  <Trash2 className="size-4.5" />
                </button>
              </div>

              <Field
                name={`prompt-${q.id}`}
                label="Prompt"
                value={q.prompt}
                onChange={(e) => updateQuestion(q.id, { prompt: e.target.value })}
                placeholder="Which city is Nigeria's federal capital?"
                autoComplete="off"
                spellCheck="false"
                className="mt-4"
              />

              <div className="mt-4">
                <p className="mb-2 text-sm font-bold">Options — tap the ring to mark the correct answer</p>
                <div className="grid gap-3 sm:grid-cols-2">
                  {q.options.map((opt, oi) => {
                    const isCorrect = q.correctIndex === oi;
                    return (
                      <div
                        key={oi}
                        className={cn(
                          "flex items-center gap-3 rounded-2xl border-2 bg-white px-3 py-1.5 transition-colors",
                          isCorrect ? "border-mint/60 bg-mint/5" : "border-ink/10",
                        )}
                      >
                        <button
                          type="button"
                          role="radio"
                          aria-checked={isCorrect}
                          aria-label={`Mark option ${String.fromCharCode(65 + oi)} as correct`}
                          onClick={() => updateQuestion(q.id, { correctIndex: oi })}
                          className={cn(
                            "grid size-6 shrink-0 cursor-pointer place-items-center rounded-full border-2 transition-colors",
                            isCorrect ? "border-mint bg-mint text-white" : "border-ink/20 hover:border-mint",
                          )}
                        >
                          {isCorrect && <Check className="size-3.5" strokeWidth={3.5} />}
                        </button>
                        <input
                          value={opt}
                          onChange={(e) => {
                            const options = [...q.options];
                            options[oi] = e.target.value;
                            updateQuestion(q.id, { options });
                          }}
                          placeholder={`Option ${String.fromCharCode(65 + oi)}`}
                          className="w-full bg-transparent py-2.5 font-bold text-ink outline-none placeholder:font-normal placeholder:text-soft/50"
                          maxLength={80}
                          autoComplete="off"
                        />
                      </div>
                    );
                  })}
                </div>
              </div>

              <div className="mt-4 flex items-center gap-2">
                <p className="text-sm font-bold text-soft">Time limit</p>
                <div className="flex gap-1.5">
                  {DURATIONS.map((d) => (
                    <button
                      key={d.id}
                      type="button"
                      onClick={() => updateQuestion(q.id, { durationMs: d.id })}
                      className={cn(
                        "cursor-pointer rounded-full border px-3 py-1 text-xs font-bold transition-colors",
                        q.durationMs === d.id
                          ? "border-sky bg-sky text-white"
                          : "border-ink/10 bg-white text-soft hover:border-sky hover:text-sky-dark",
                      )}
                    >
                      {d.label}
                    </button>
                  ))}
                </div>
              </div>
            </section>
          ))}
        </div>

        <div ref={ctaRef} className="mt-8 scroll-mt-24 rounded-[2rem] border-2 border-gold/30 bg-gold/10 p-6 text-center">
          <p className="font-display text-xl font-semibold text-ink">
            Ready? Open the room and put <Chip tone="mint"><Trophy className="size-3" /> {naira(poolNum)}</Chip> on the line.
          </p>
          <p className="mt-1 text-sm text-soft">
            The pool is funded from your wallet when the room opens. Players join with a code, no app needed.
          </p>

          {(insufficient || error) && (
            <div
              role="alert"
              className={
                "mt-5 rounded-2xl border-2 border-coral/40 bg-coral/10 px-4 py-3 text-left " +
                (insufficient ? "" : "")
              }
            >
              <p className="flex items-start gap-2 text-sm font-bold text-coral">
                <Wallet className="mt-0.5 size-4 shrink-0" />
                <span>
                  {insufficient ? (
                    <>
                      Not enough to fund this pool. Available {naira(available)} — you're short{" "}
                      <span className="whitespace-nowrap">{naira(shortfall)}</span>.
                    </>
                  ) : (
                    error
                  )}
                </span>
              </p>
              {(insufficient || (error && error.toLowerCase().includes("insufficient"))) && (
                <button
                  type="button"
                  onClick={() => navigate("/instructor/fund")}
                  className="mt-3 cursor-pointer rounded-full bg-coral px-5 py-2.5 text-sm font-bold text-white press-3d"
                >
                  Fund wallet →
                </button>
              )}
            </div>
          )}

          <Button
            variant="coral"
            size="lg"
            loading={saving}
            disabled={insufficient}
            onClick={() => void handleOpenRoom()}
            className="mt-5"
          >
            {saving ? "Opening room…" : "Save & open room"}
            {!saving && <Presentation className="size-5" />}
          </Button>
        </div>
      </div>

      {/* Confirm before deleting a question */}
      <ConfirmDialog
        open={Boolean(confirmDelete)}
        danger
        title={confirmDelete ? `Delete question ${questions.findIndex((q) => q.id === confirmDelete.id) + 1}?` : "Delete question?"}
        body="This question and its options will be removed from the quiz. This can't be undone."
        confirmLabel="Delete question"
        cancelLabel="Keep it"
        onConfirm={() => {
          if (confirmDelete) removeQuestion(confirmDelete.id);
        }}
        onCancel={() => setConfirmDelete(null)}
      />

      {/* Transient "question added" toast */}
      {addedToast && (
        <div className="pointer-events-none fixed inset-x-0 bottom-6 z-[90] flex justify-center px-4">
          <div className="flex items-center gap-2 rounded-full border-2 border-mint/40 bg-cream px-5 py-3 font-bold text-mint-dark card-3d">
            <span className="grid size-6 shrink-0 place-items-center rounded-full bg-mint text-white">
              <Check className="size-3.5" strokeWidth={3.5} />
            </span>
            Question {addedToast.num} added — scrolling you down to enter it.
          </div>
        </div>
      )}

      <Footer className="mt-16" />
    </main>
  );
}
