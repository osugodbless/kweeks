import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";

const OPTIONS = [
  { ch: "A", tx: "Ibadan", correct: false },
  { ch: "B", tx: "Lagos", correct: true },
  { ch: "C", tx: "Abuja", correct: false },
  { ch: "D", tx: "Enugu", correct: false },
];

export function InstructorLiveRoom() {
  return (
    <DesktopFrame>
      <header className="flex h-16 w-full shrink-0 items-center justify-between bg-surface px-7">
        <div className="flex items-center gap-2">
          <span className="font-display text-[22px] font-extrabold text-paper">KWEEKS</span>
          <span className="ml-1 inline-flex items-center gap-[5px] rounded-full bg-surface px-2.5 py-1">
            <span className="h-2 w-2 rounded-full bg-red" />
            <span className="font-body text-[11px] font-bold tracking-widest text-red">LIVE</span>
          </span>
        </div>
        <div className="flex items-center gap-3.5">
          <span className="font-body text-[13px] text-text-2">Q 3 · 8 players</span>
          <div className="flex items-center gap-1 rounded-full bg-surface-2 px-3 py-2">
            <span className="font-body text-[11px] font-bold tracking-widest text-text-3">
              WALLET
            </span>
            <span className="font-display text-[15px] font-extrabold text-naira">₦150,000</span>
          </div>
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-surface-2">
            <span className="font-body text-[13px] font-extrabold text-violet">AP</span>
          </div>
        </div>
      </header>

      <div className="flex flex-1 gap-6 bg-bg px-7 py-7">
        {/* MAIN projector preview */}
        <div className="flex flex-1 flex-col">
          <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            PROJECTOR PREVIEW
          </div>
          <div className="mt-3 flex flex-1 flex-col rounded-3xl border border-stroke bg-surface px-10 py-8">
            <div className="flex items-center justify-between">
              <span className="font-body text-[15px] text-text-3">Question 3 of 10</span>
              <span className="font-display text-[18px] font-extrabold text-naira">
                ₦50,000 POOL
              </span>
            </div>

            <div className="mt-8 flex flex-col items-center">
              <div className="relative flex h-[72px] w-[72px] items-center justify-center">
                <svg viewBox="0 0 72 72" className="h-[72px] w-[72px] -rotate-90">
                  <circle cx="36" cy="36" r="30" fill="none" stroke="var(--color-surface-2)" strokeWidth="6" />
                  <circle
                    cx="36" cy="36" r="30" fill="none" stroke="var(--color-gold)" strokeWidth="6"
                    strokeDasharray="188.5" strokeDashoffset={188.5 * (1 - 18 / 30)}
                    strokeLinecap="round"
                  />
                </svg>
                <span className="absolute font-display text-[36px] font-extrabold text-paper">
                  18
                </span>
              </div>
            </div>

            <h1 className="mt-6 text-center font-display text-[34px] font-extrabold leading-tight text-paper">
              Which Nigerian city is known as the Centre of Excellence?
            </h1>

            <div className="mt-8 grid grid-cols-2 gap-4">
              {OPTIONS.map((o) => (
                <div
                  key={o.ch}
                  className={cn(
                    "flex items-center gap-3 rounded-2xl border px-5 py-4",
                    o.correct ? "border-naira bg-surface-2" : "border-stroke bg-surface",
                  )}
                >
                  <span
                    className={cn(
                      "flex h-7 w-7 items-center justify-center rounded-full bg-surface-2 text-[15px]",
                      o.correct ? "text-gold-ink" : "text-text-2",
                    )}
                  >
                    {o.ch}
                  </span>
                  <span
                    className={cn(
                      "font-body text-[16px]",
                      o.correct ? "font-semibold text-naira" : "text-paper",
                    )}
                  >
                    {o.tx}
                  </span>
                  {o.correct && (
                    <span className="ml-auto text-[15px] text-gold-ink">
                      ✓
                    </span>
                  )}
                </div>
              ))}
            </div>

            {/* live answer bar */}
            <div className="mt-auto flex items-center gap-2 rounded-2xl bg-surface-2 px-5 py-3.5">
              <span className="font-body text-[14px] font-bold text-naira">▲ 3</span>
              <span className="font-body text-[14px] text-paper">Zainab</span>
              <span className="font-body text-[13.5px] text-text-2">just answered in 2.1s — correct</span>
            </div>
          </div>
        </div>

        {/* SIDE */}
        <div className="flex w-[330px] shrink-0 flex-col gap-5">
          {/* join card */}
          <div className="flex flex-col gap-4 rounded-2xl border border-stroke bg-surface px-5 py-5">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              PLAYERS JOIN HERE
            </div>
            <div className="text-center font-body text-[12px] text-text-3">Scan or type</div>
            <div className="mx-auto flex h-[120px] w-[120px] items-center justify-center overflow-hidden rounded-xl bg-white">
              <img
                src="https://images.unsplash.com/photo-1550482768-88b710a445fd?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=480&ixlib=rb-1.2.1&q=80&w=480"
                alt="Join QR"
                className="h-full w-full object-cover"
              />
            </div>
            <div className="flex items-baseline justify-center gap-1">
              <span className="font-body text-[15px] text-text-3">kweeks.ng/r/</span>
              <span className="font-display text-[20px] font-extrabold text-gold">AB12</span>
            </div>
            <div className="flex items-center justify-center gap-1.5">
              {["🐙", "🦊", "🐼", "😎", "🤖"].map((e) => (
                <div
                  key={e}
                  className="flex h-[28px] w-[28px] items-center justify-center rounded-full bg-surface-2 text-[17px]"
                >
                  {e}
                </div>
              ))}
            </div>
          </div>

          {/* control card */}
          <div className="flex flex-col gap-4 rounded-2xl border border-stroke bg-surface px-5 py-5">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              CONTROL · MANUAL
            </div>
            <button className="flex h-12 w-full items-center justify-center rounded-2xl bg-gold font-body text-[15px] font-extrabold text-gold-ink">
              NEXT QUESTION →
            </button>
            <button className="flex h-12 w-full items-center justify-center rounded-2xl border-[1.5px] border-naira bg-transparent font-body text-[14px] font-extrabold text-naira">
              DECLARE WINNERS
            </button>
            <div className="font-body text-[12px] text-text-3">
              Auto is on a 30s timer. Manual waits for your tap.
            </div>
          </div>
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
