import { useNavigate } from "react-router-dom";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";

const stats = [
  { k: "QUIZZES HOSTED", v: "3" },
  { k: "PLAYERS HOSTED", v: "48" },
  { k: "WINNERS PAID", v: "9" },
  { k: "AVAILABLE", v: "₦150,000", naira: true },
];

export function InstructorWallet() {
  const nav = useNavigate();
  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="dashboard"
        right={<NavRight amount="₦150,000" />}
      />
      <div className="flex flex-1 flex-col gap-[22px] bg-bg px-8 py-[30px]">
        {/* greeting */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-display text-[28px] font-extrabold text-paper">
              Welcome back, Adeola
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

        {/* stats */}
        <div className="grid grid-cols-4 gap-3.5">
          {stats.map((s) => (
            <div
              key={s.k}
              className="flex flex-col gap-1.5 rounded-2xl border border-stroke bg-surface px-[18px] py-4"
            >
              <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
                {s.k}
              </span>
              <span
                className={
                  "font-display text-[26px] font-extrabold " + (s.naira ? "text-naira" : "text-paper")
                }
              >
                {s.v}
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
              ₦150,000
            </div>
            <div className="font-body text-[14px] text-text-2">
              Available to fund quiz pools and pay winners.
            </div>
            <div className="h-px w-full bg-stroke" />
            <div className="flex items-center justify-between">
              <span className="font-body text-[12px] text-text-3">Wallet ID</span>
              <span className="font-body text-[12px] font-bold text-paper">
                kweeks_ngn_8f2c1a
              </span>
            </div>
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
              <span className="font-body text-[13px] font-bold text-text-3">1 live</span>
            </div>

            <div className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface-2 p-4">
              <div className="flex items-center justify-between">
                <span className="font-body text-[15px] font-bold text-paper">
                  Naija General Knowledge
                </span>
                <span className="inline-flex items-center gap-1 rounded-full bg-surface px-2.5 py-1">
                  <span className="h-2 w-2 rounded-full bg-naira" />
                  <span className="font-body text-[11px] font-bold tracking-widest text-naira">
                    LIVE
                  </span>
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="font-body text-[12.5px] text-text-3">
                  10 questions · pool ₦50,000 · 3 winners · room AB12
                </span>
                <button
                  onClick={() => nav("/instructor/live-room")}
                  className="font-body text-[13px] font-bold text-gold"
                >
                  Open room →
                </button>
              </div>
            </div>

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
                className="rounded-xl bg-gold px-4 py-3 font-body text-[12.5px] font-extrabold tracking-wide text-gold-ink"
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
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
