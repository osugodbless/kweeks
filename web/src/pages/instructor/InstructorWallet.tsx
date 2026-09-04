import { useNavigate } from "react-router-dom";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { useAuth } from "@/lib/auth";
import { useDashboard, useProvisionWallet, useWallet } from "@/lib/hooks";
import { naira } from "@/lib/player";

function errText(e: unknown): string {
  if (e instanceof Error && e.message) return e.message;
  return "Something went wrong loading your dashboard.";
}

export function InstructorWallet() {
  const nav = useNavigate();
  const dash = useDashboard();
  const walletView = useWallet();
  const provision = useProvisionWallet();
  const instructor = useAuth((s) => s.instructor);
  const authBal = useAuth((s) => s.wallet?.balanceNaira);

  const name = instructor?.name || "Adeola";
  const s = dash.data;
  const quizzes = s?.quizzes ?? [];
  const live = quizzes.filter(
    (q) => Boolean(q.roomId && q.roomCode) && (q.state === "lobby" || q.state === "live"),
  );
  const balance = s?.availableNaira ?? walletView.data?.wallet.balanceNaira ?? authBal ?? "0";

  const stats = [
    { k: "QUIZZES HOSTED", v: s ? String(s.quizzesHosted) : "0", naira: false },
    { k: "PLAYERS HOSTED", v: s ? String(s.playersHosted) : "0", naira: false },
    { k: "WINNERS PAID", v: s ? String(s.winnersPaid) : "0", naira: false },
    { k: "AVAILABLE", v: naira(balance), naira: true },
  ];

  const retry = () => {
    void dash.refetch();
    void walletView.refetch();
  };

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="dashboard"
        right={<NavRight amount={naira(balance)} />}
      />
      <div className="flex flex-1 flex-col gap-[22px] bg-bg px-8 py-[30px]">
        {/* greeting */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-display text-[28px] font-extrabold text-paper">
              Welcome back, {name}
            </h1>
            <p className="font-body text-[14px] text-text-2">
              Create a quiz, fund a pool, run the room — all from here.
            </p>
          </div>
          <button
            onClick={() => nav("/instructor/quiz-builder")}
            className="flex h-[54px] items-center gap-2 rounded-2xl bg-gold px-6 font-body text-[15px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
          >
            <span className="font-display text-[20px] font-extrabold">+</span>
            CREATE A QUIZ
          </button>
        </div>

        {dash.isPending ? (
          /* loading skeleton */
          <>
            <div className="grid grid-cols-4 gap-3.5">
              {[0, 1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface px-[18px] py-4"
                >
                  <div className="h-[11px] w-24 animate-pulse rounded bg-surface-2" />
                  <div className="h-[26px] w-28 animate-pulse rounded bg-surface-2" />
                </div>
              ))}
            </div>

            <div className="flex flex-1 items-start gap-5">
              <div className="flex w-[540px] flex-col gap-4 rounded-3xl border border-stroke bg-surface-2 p-[26px]">
                <div className="h-[11px] w-32 animate-pulse rounded bg-surface" />
                <div className="h-[52px] w-52 animate-pulse rounded bg-surface" />
                <div className="h-[14px] w-64 animate-pulse rounded bg-surface" />
                <div className="h-px w-full bg-stroke" />
                <div className="h-[14px] w-40 animate-pulse rounded bg-surface" />
                <div className="h-[52px] w-[200px] animate-pulse rounded-2xl bg-surface" />
              </div>
              <div className="flex flex-1 flex-col justify-between gap-4 rounded-3xl border border-stroke bg-surface p-[22px]">
                <div className="flex items-center justify-between">
                  <div className="h-[20px] w-32 animate-pulse rounded bg-surface-2" />
                  <div className="h-[13px] w-10 animate-pulse rounded bg-surface-2" />
                </div>
                <div className="h-[64px] w-full animate-pulse rounded-2xl bg-surface-2" />
                <div className="h-[70px] w-full animate-pulse rounded-2xl bg-surface-2" />
                <div className="h-[14px] w-40 animate-pulse rounded bg-surface-2" />
              </div>
            </div>
          </>
        ) : dash.isError ? (
          /* error state */
          <div className="flex flex-1 flex-col items-center justify-center gap-4 rounded-3xl border border-stroke bg-surface px-8 py-16 text-center">
            <div className="font-display text-[22px] font-extrabold text-paper">
              Couldn't load your dashboard
            </div>
            <p className="max-w-[420px] font-body text-[13.5px] leading-relaxed text-text-2">
              {errText(dash.error)}
            </p>
            <button
              onClick={retry}
              className="mt-2 flex h-[50px] items-center rounded-2xl bg-gold px-8 font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
            >
              RETRY
            </button>
          </div>
        ) : (
          /* data */
          <>
            {/* stats */}
            <div className="grid grid-cols-4 gap-3.5">
              {stats.map((st) => (
                <div
                  key={st.k}
                  className="flex flex-col gap-1.5 rounded-2xl border border-stroke bg-surface px-[18px] py-4"
                >
                  <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
                    {st.k}
                  </span>
                  <span
                    className={
                      "font-display text-[26px] font-extrabold " + (st.naira ? "text-naira" : "text-paper")
                    }
                  >
                    {st.v}
                  </span>
                </div>
              ))}
            </div>

            {/* body */}
            <div className="flex flex-1 items-start gap-5">
              {/* wallet card */}
              <div className="flex w-[540px] flex-col gap-4 rounded-3xl border border-stroke bg-surface-2 p-[26px]">
                <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
                  ASSIGNED WALLET · NGN
                </div>
                <div className="font-display text-[52px] font-extrabold leading-none text-naira">
                  {naira(balance)}
                </div>
                <div className="font-body text-[14px] text-text-2">
                  Available to fund quiz pools and pay winners.
                </div>
                <div className="h-px w-full bg-stroke" />
                <div className="flex items-center justify-between">
                  <span className="font-body text-[12px] text-text-3">Wallet ID</span>
                  <span className="font-body text-[12px] font-bold text-paper">
                    {walletView.data?.wallet.id ?? "kweeks_ngn_…"}
                  </span>
                </div>
                <div className="flex items-center justify-between gap-2">
                  <span className="font-body text-[12px] text-text-3">BMONI rail</span>
                  {walletView.data?.wallet.bmoniUserId ? (
                    <span className="inline-flex items-center gap-1.5">
                      <span className="h-2 w-2 rounded-full bg-naira" />
                      <span className="font-body text-[12px] font-bold text-naira">Provisioned</span>
                    </span>
                  ) : (
                    <button
                      onClick={() => provision.mutate()}
                      disabled={provision.isPending}
                      className="rounded-full bg-surface px-3 py-1.5 font-body text-[12px] font-bold text-gold hover:opacity-80 disabled:opacity-40"
                    >
                      {provision.isPending ? "Provisioning…" : "Provision on BMONI"}
                    </button>
                  )}
                </div>
                {provision.error && (
                  <div className="rounded-lg bg-surface px-3 py-2 font-body text-[12px] font-semibold text-red">
                    {errText(provision.error)}
                  </div>
                )}
                <div className="flex items-center gap-2.5">
                  <button
                    onClick={() => nav("/instructor/fund")}
                    className="flex h-[52px] w-[200px] items-center justify-center rounded-2xl bg-gold font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
                  >
                    FUND WALLET
                  </button>
                  <div className="flex items-center gap-2">
                    {[
                      ["Card", "text-text-2"],
                      ["Transfer", "text-text-2"],
                      ["Instant top-up", "text-naira"],
                    ].map(([t, c]) => (
                      <span
                        key={t}
                        className={"rounded-full bg-surface px-3 py-2 font-body text-[12px] " + c}
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                </div>
                <div className="font-body text-[12.5px] leading-snug text-text-3">
                  Funds land instantly and settle in naira. Money is held in escrow only while a pool
                  is live.
                </div>
              </div>

              {/* your quizzes */}
              <div className="flex flex-1 flex-col justify-between gap-4 rounded-3xl border border-stroke bg-surface p-[22px]">
                <div className="flex items-center justify-between">
                  <h2 className="font-display text-[20px] font-extrabold text-paper">Your quizzes</h2>
                  {live.length > 0 && (
                    <span className="font-body text-[13px] font-bold text-text-3">
                      {live.length} live
                    </span>
                  )}
                </div>

                {quizzes.length === 0 ? (
                  <div className="flex flex-col items-center gap-2.5 rounded-2xl border border-dashed border-stroke bg-surface-2 px-6 py-10 text-center">
                    <span className="text-[30px]">🎯</span>
                    <span className="font-body text-[14px] font-bold text-paper">
                      No quizzes yet — create your first
                    </span>
                    <span className="font-body text-[12.5px] text-text-3">
                      Set a pool, write questions, open the room.
                    </span>
                  </div>
                ) : (
                  <div className="flex flex-col gap-2.5">
                    {quizzes.map((q) => {
                      const liveQuiz =
                        Boolean(q.roomId && q.roomCode) &&
                        (q.state === "lobby" || q.state === "live");
                      return (
                        <div
                          key={q.id}
                          className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface-2 p-4"
                        >
                          <div className="flex items-center justify-between">
                            <span className="font-body text-[15px] font-bold text-paper">
                              {q.title}
                            </span>
                            {liveQuiz && (
                              <span className="inline-flex items-center gap-1 rounded-full bg-surface px-2.5 py-1">
                                <span className="h-2 w-2 rounded-full bg-naira" />
                                <span className="font-body text-[11px] font-bold tracking-widest text-naira">
                                  LIVE
                                </span>
                              </span>
                            )}
                          </div>
                          <div className="flex items-center justify-between">
                            <span className="font-body text-[12.5px] text-text-3">
                              {q.questionCount} questions · pool {naira(q.poolNaira)} ·{" "}
                              {q.winnerCount} {q.winnerCount === 1 ? "winner" : "winners"}
                              {q.roomCode ? ` · room ${q.roomCode}` : ""}
                            </span>
                            {liveQuiz && q.roomId && (
                              <button
                                onClick={() => nav(`/instructor/live-room?room=${q.roomId}`)}
                                className="font-body text-[13px] font-bold text-gold"
                              >
                                Open room →
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}

                <div className="flex items-center justify-between gap-3.5 rounded-2xl border border-gold bg-surface-2 p-4">
                  <div>
                    <div className="font-body text-[14px] font-bold text-paper">
                      Start another live quiz
                    </div>
                    <div className="font-body text-[12.5px] text-text-3">
                      Author questions, set a pool, invite a room.
                    </div>
                  </div>
                  <button
                    onClick={() => nav("/instructor/quiz-builder")}
                    className="rounded-xl bg-gold px-4 py-3 font-body text-[12.5px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
                  >
                    CREATE QUIZ
                  </button>
                </div>

                <div className="flex items-center gap-2">
                  <span className="font-body text-[12.5px] text-text-3">
                    Completed quizzes live in your history.
                  </span>
                  <button
                    onClick={() => nav("/instructor/history")}
                    className="font-body text-[12.5px] font-bold text-gold"
                  >
                    View history →
                  </button>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
