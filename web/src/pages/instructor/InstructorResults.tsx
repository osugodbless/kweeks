import { useEffect } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { ArrowLeft, Crown, Medal, Trophy, Users } from "lucide-react";
import { useQuizResults } from "@/lib/hooks";
import { splitPodium } from "@/lib/player";
import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { Money } from "@/components/ui/money";
import { cn } from "@/lib/cn";

const STATE_LABEL: Record<string, string> = {
  lobby: "Lobby",
  live: "Live",
  podium: "Podium",
  ended: "Ended",
};

function fmtDate(iso: string | undefined): string {
  if (!iso) return "";
  try {
    return new Date(iso).toLocaleString("en-NG", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

/**
 * The host's persistent, post-quiz leaderboard. Standings are recomputed from
 * stored participants + answers, so this page keeps working after the room has
 * closed — including after logging out and back in.
 */
export function InstructorResults() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const quizId = params.get("quiz") ?? undefined;
  const { data, isLoading, isError } = useQuizResults(quizId);

  useEffect(() => {
    if (!quizId) navigate("/instructor/dashboard", { replace: true });
  }, [quizId, navigate]);

  const standings = data?.standings ?? [];
  const winners = data?.winners ?? [];
  const winnerCount = data?.quiz.winnerCount ?? 0;
  const shares = data ? splitPodium(data.quiz.poolNaira, winnerCount) : [];
  const winnerIds = new Set(winners.map((w) => w.participantId));

  return (
    <main className="relative min-h-screen overflow-hidden pb-24">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-gold/15 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-violet/15 blur-3xl" />

      <InstructorNav active="/instructor/dashboard" />

      <div className="relative mx-auto w-full max-w-4xl px-4 pt-12 sm:px-6 lg:pt-16">
        <Link
          to="/instructor/dashboard"
          className="inline-flex items-center gap-2 text-sm font-bold text-soft transition-colors hover:text-violet"
        >
          <ArrowLeft className="size-4" /> Back to dashboard
        </Link>

        {isLoading && (
          <div className="mt-10 flex justify-center py-16 text-sm font-bold text-soft">Loading results…</div>
        )}

        {isError && (
          <div className="mt-10 rounded-[2rem] border-2 border-coral/30 bg-coral/10 px-6 py-10 text-center">
            <p className="font-display text-xl font-semibold text-coral">We couldn't load these results.</p>
            <p className="mt-1 text-sm text-soft">The quiz may have been removed. Head back to your dashboard.</p>
          </div>
        )}

        {data && (
          <>
            <div className="mt-8 flex flex-wrap items-end justify-between gap-4">
              <div>
                <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
                  <Trophy className="size-4 text-gold" /> Quiz results
                </span>
                <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
                  {data.quiz.title}
                </h1>
                <p className="mt-3 text-lg text-soft">
                  {data.room ? (
                    <>
                      Room <span className="font-display text-ink">{data.room.code}</span> ·{" "}
                      <span
                        className={cn(
                          "font-bold",
                          data.room.state === "live"
                            ? "text-coral"
                            : data.room.state === "podium"
                              ? "text-mint-dark"
                              : "text-soft",
                        )}
                      >
                        {STATE_LABEL[data.room.state] ?? data.room.state}
                      </span>{" "}
                      · {fmtDate(data.room.startedAt ?? data.quiz.createdAt)}
                    </>
                  ) : (
                    "This quiz hasn't been opened in a room yet."
                  )}
                </p>
              </div>
              {winners.length > 0 && (
                <Chip tone="gold">
                  <Crown className="size-3.5" /> {winners.length} winner{winners.length === 1 ? "" : "s"}
                </Chip>
              )}
            </div>

            {/* Summary */}
            <div className="mt-8 grid grid-cols-2 gap-4 sm:grid-cols-4">
              <div className="rounded-3xl border-2 border-ink/5 bg-cream p-5">
                <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Prize pool</p>
                <Money value={data.quiz.poolNaira} className="mt-2 block text-3xl" />
              </div>
              <div className="rounded-3xl border-2 border-ink/5 bg-cream p-5">
                <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Players</p>
                <p className="mt-2 font-display text-3xl font-semibold text-violet-dark">{standings.length}</p>
              </div>
              <div className="rounded-3xl border-2 border-ink/5 bg-cream p-5">
                <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Questions</p>
                <p className="mt-2 font-display text-3xl font-semibold text-sky-dark">{data.quiz.questionCount}</p>
              </div>
              <div className="rounded-3xl border-2 border-ink/5 bg-cream p-5">
                <p className="text-[11px] font-bold uppercase tracking-widest text-soft">Winners</p>
                <p className="mt-2 font-display text-3xl font-semibold text-gold-dark">{winnerCount}</p>
              </div>
            </div>

            {/* Winners */}
            {winners.length > 0 && (
              <section className="mt-8">
                <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
                  <Crown className="size-5 text-gold" /> Podium
                </h2>
                <ul className="mt-4 grid gap-3 sm:grid-cols-3">
                  {winners.map((w, i) => (
                    <li
                      key={w.participantId}
                      className={cn(
                        "rounded-3xl border-2 bg-cream p-5 text-center",
                        i === 0 ? "border-gold" : "border-ink/5",
                      )}
                    >
                      <div className="mx-auto grid size-10 place-items-center rounded-full bg-gold/15 font-display text-lg font-semibold text-gold-dark">
                        {i + 1}
                      </div>
                      <Avatar id={w.avatar} className="mx-auto mt-3 size-14" />
                      <p className="mt-3 truncate font-bold">{w.nickname}</p>
                      <p className="text-xs font-bold text-soft">
                        {w.correctCount} correct · {(w.totalLatencyMs / 1000).toFixed(1)}s
                      </p>
                      <Money value={shares[i] ?? 0} className="mt-3 block text-2xl" />
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {/* Full leaderboard */}
            <section className="mt-8">
              <h2 className="flex items-center gap-2 font-display text-xl font-semibold">
                <Users className="size-5 text-violet" /> Full leaderboard
              </h2>

              {standings.length === 0 ? (
                <div className="mt-4 rounded-[2rem] border-2 border-dashed border-ink/15 bg-cream px-6 py-12 text-center">
                  <p className="font-display text-lg font-semibold">No scores recorded</p>
                  <p className="mt-1 text-sm text-soft">
                    This quiz hasn't had a live run with answers yet — the leaderboard appears once players answer.
                  </p>
                  <Button variant="violet" size="md" className="mt-5" onClick={() => navigate("/instructor/dashboard")}>
                    Back to dashboard
                  </Button>
                </div>
              ) : (
                <div className="mt-4 overflow-hidden rounded-[2rem] border-2 border-ink/5 bg-cream card-3d">
                  <div className="grid grid-cols-[3rem_1fr_5rem_5rem] items-center gap-3 border-b border-ink/5 px-6 py-3 text-[11px] font-bold uppercase tracking-widest text-soft sm:grid-cols-[4rem_1fr_6rem_7rem]">
                    <span>#</span>
                    <span>Player</span>
                    <span className="text-right">Correct</span>
                    <span className="text-right">Speed</span>
                  </div>
                  <ul>
                    {standings.map((row, i) => {
                      const isWinner = winnerIds.has(row.participantId);
                      return (
                        <li
                          key={row.participantId}
                          className={cn(
                            "grid grid-cols-[3rem_1fr_5rem_5rem] items-center gap-3 px-6 py-4 sm:grid-cols-[4rem_1fr_6rem_7rem]",
                            i > 0 && "border-t border-ink/5",
                            isWinner && "bg-gold/10",
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
                              <span className="text-soft">{i + 1}</span>
                            )}
                          </span>
                          <span className="flex min-w-0 items-center gap-3">
                            <Avatar id={row.avatar} className="size-10" />
                            <span className="truncate font-bold">{row.nickname}</span>
                            {isWinner && (
                              <span className="rounded-full bg-gold px-2.5 py-0.5 text-[10px] font-bold uppercase text-ink">
                                Paid
                              </span>
                            )}
                          </span>
                          <span className="text-right font-display text-lg font-semibold text-violet-dark">
                            {row.correctCount}
                          </span>
                          <span className="text-right text-sm font-bold text-soft">
                            {(row.totalLatencyMs / 1000).toFixed(1)}s
                          </span>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              )}
            </section>

            <p className="mt-6 flex items-center justify-center gap-2 text-sm font-bold text-soft">
              <Trophy className="size-4 text-gold" />
              Results are saved — you can reopen this page any time.
            </p>
          </>
        )}
      </div>

      <Footer className="mt-16" />
    </main>
  );
}
