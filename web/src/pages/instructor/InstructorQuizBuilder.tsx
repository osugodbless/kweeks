import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";

const Q_OPTIONS = [
  { ch: "A", tx: "Ibadan", correct: false },
  { ch: "B", tx: "Lagos", correct: true },
  { ch: "C", tx: "Abuja", correct: false },
  { ch: "D", tx: "Enugu", correct: false },
];

const STRIP = [
  { n: "1", t: "Centre of Excellence", on: true },
  { n: "2", t: "Largest economy", on: false },
  { n: "3", t: "Independence year", on: false },
  { n: "4", t: "Naija anthem line", on: false },
  { n: "5", t: "Jollof capital", on: false },
];

export function InstructorQuizBuilder() {
  const nav = useNavigate();
  const [winners, setWinners] = useState(3);
  const [pacing] = useState("MANUAL");

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="create"
        right={<NavRight amount="₦150,000" />}
      />
      <div className="flex flex-1 gap-6 bg-bg px-7 py-7">
        {/* LEFT */}
        <div className="flex w-[360px] shrink-0 flex-col gap-4">
          {/* title card */}
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
              <span className="font-body text-[15px] text-paper">Naija General Knowledge</span>
            </div>
          </div>

          {/* pool card */}
          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              PRIZE POOL (FROM YOUR WALLET)
            </span>
            <div className="flex items-baseline gap-1">
              <span className="font-display text-[24px] font-extrabold text-naira">₦</span>
              <span className="font-display text-[32px] font-extrabold leading-none text-naira">
                50,000
              </span>
            </div>
            {/* slider track */}
            <div className="relative h-[6px] w-full rounded-full bg-surface-2">
              <div className="absolute left-0 top-0 h-full w-[38%] rounded-full bg-naira" />
              <div className="absolute left-[38%] top-1/2 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-naira bg-bg" />
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

          {/* winners */}
          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              WINNERS
            </span>
            <div className="flex gap-1.5">
              {[1, 2, 3, 5].map((w) => (
                <button
                  key={w}
                  onClick={() => setWinners(w)}
                  className={cn(
                    "flex h-9 flex-1 items-center justify-center rounded-full font-display text-[16px] font-bold",
                    winners === w ? "bg-gold text-gold-ink" : "bg-surface-2 text-text-2",
                  )}
                >
                  {w}
                </button>
              ))}
            </div>
          </div>

          {/* pacing */}
          <div className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface px-5 py-4">
            <span className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              PACING
            </span>
            <div className="flex gap-1.5">
              {["AUTO", "MANUAL"].map((p) => (
                <button
                  key={p}
                  className={cn(
                    "flex h-9 flex-1 items-center justify-center rounded-full font-body text-[13px] font-bold",
                    pacing === p ? "bg-gold text-gold-ink" : "bg-surface-2 text-text-2",
                  )}
                >
                  {p}
                </button>
              ))}
            </div>
            <span className="font-body text-[12.5px] leading-snug text-text-2">
              Auto advances on a timer. Manual = you tap next. Venue mode: Manual.
            </span>
          </div>
        </div>

        {/* RIGHT */}
        <div className="flex flex-1 flex-col gap-4">
          <div className="flex items-center justify-between">
            <h2 className="font-display text-[20px] font-extrabold text-paper">Questions</h2>
            <button className="flex items-center gap-1.5 rounded-xl bg-gold px-4 py-2.5 font-body text-[14px] font-extrabold text-gold-ink">
              <span className="font-display text-[18px]">+</span>Add question
            </button>
          </div>

          {/* question card */}
          <div className="flex flex-col gap-4 rounded-2xl border border-stroke bg-surface px-6 py-5">
            <div className="font-body text-[11px] font-bold tracking-[0.1em] text-text-3">
              QUESTION 1 · 30s
            </div>
            <div className="font-display text-[20px] font-extrabold leading-snug text-paper">
              Which Nigerian city is called the Centre of Excellence?
            </div>
            <div className="grid grid-cols-2 gap-3">
              {Q_OPTIONS.map((o) => (
                <div
                  key={o.ch}
                  className={cn(
                    "flex items-center gap-3 rounded-2xl border px-4 py-3",
                    o.correct ? "border-naira bg-surface-2" : "border-stroke bg-surface",
                  )}
                >
                  <span
                    className={cn(
                      "flex h-6 w-6 items-center justify-center rounded-full bg-surface-2 text-[12px]",
                      o.correct ? "text-gold-ink" : "text-paper",
                    )}
                  >
                    {o.ch}
                  </span>
                  <span
                    className={cn(
                      "font-body text-[15px]",
                      o.correct ? "font-semibold text-naira" : "text-paper",
                    )}
                  >
                    {o.tx}
                  </span>
                  {o.correct && (
                    <span className="ml-auto text-[14px] text-naira">
                      ✓ <span className="font-body text-[12px] font-semibold">correct</span>
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* strip */}
          <div className="flex items-center gap-1.5">
            {STRIP.map((s) => (
              <div
                key={s.n}
                className={cn(
                  "flex flex-col items-center rounded-xl border px-4 py-2",
                  s.on ? "border-gold bg-surface-2" : "border-stroke bg-surface",
                )}
              >
                <span
                  className={cn(
                    "font-display text-[15px] font-extrabold",
                    s.on ? "text-gold" : "text-text-2",
                  )}
                >
                  {s.n}
                </span>
                <span
                  className={cn(
                    "font-body text-[10px]",
                    s.on ? "text-text-2" : "text-text-3",
                  )}
                >
                  {s.t}
                </span>
              </div>
            ))}
            <div className="flex flex-col items-center rounded-xl border border-stroke bg-surface px-4 py-2">
              <span className="font-display text-[15px] font-extrabold text-text-2">+</span>
              <span className="font-body text-[10px] text-text-2">5 more</span>
            </div>
          </div>

          {/* footer */}
          <div className="mt-auto flex items-center justify-between">
            <span className="font-body text-[13px] text-text-3">
              10 questions · pool ₦50,000 · 3 winners · manual
            </span>
            <button
              onClick={() => nav("/instructor/live-room")}
              className="rounded-2xl bg-gold px-6 py-3 font-body text-[15px] font-extrabold text-gold-ink"
            >
              Open room →
            </button>
          </div>
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
