import { PlayerFrame } from "@/components/ui/phone";
import { cn } from "@/lib/cn";

const winners = [
  { rk: "1", e: "🦊", nm: "Ada", amt: "₦25,000", you: false },
  { rk: "2", e: "🐙", nm: "YOU", amt: "₦15,000", you: true },
  { rk: "3", e: "🐼", nm: "Tobi", amt: "₦10,000", you: false },
];

const steps = [
  { n: "1", t: "Claim locked", s: "Your win is recorded against your email." },
  { n: "2", t: "Invite sent", s: "We email you a secure redemption link." },
  { n: "3", t: "Paid", s: "Money moves from the pool wallet to you." },
];

export function PlayerPodium() {
  return (
    <PlayerFrame statusTime="9:42">
      <div className="flex flex-col">
        {/* header */}
        <div className="flex items-center justify-between">
          <h1 className="font-display text-[20px] font-extrabold text-paper">Final standings</h1>
          <span className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            GAME OVER
          </span>
        </div>

        {/* hero */}
        <div className="mt-4 overflow-hidden rounded-2xl">
          <img
            src="https://images.unsplash.com/photo-1693670984742-c008c239daf2?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&h=480&ixlib=rb-1.2.1&q=80&w=1080"
            alt="Celebration with confetti"
            className="h-[150px] w-full object-cover"
          />
        </div>
        <div className="mt-3 font-body text-[15px] text-text-3">2nd place · 8 players</div>
        <div className="font-display text-[52px] font-extrabold leading-none text-naira">
          ₦15,000
        </div>
        <div className="mt-1 font-body text-[13.5px] text-text-2">
          You banked ₦15,000 from the pool. Claim it below.
        </div>

        {/* winners list */}
        <div className="mt-4 flex flex-col gap-2">
          {winners.map((w) => (
            <div
              key={w.rk}
              className={cn(
                "flex items-center gap-3 rounded-[16px] px-4 py-2.5",
                w.you ? "border border-gold bg-gold/10" : "bg-surface",
              )}
            >
              <span
                className={cn(
                  "font-display text-[20px] font-extrabold",
                  w.you ? "text-gold-ink" : w.rk === "1" ? "text-gold" : "text-text-2",
                )}
              >
                {w.rk}
              </span>
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-surface-2 text-[20px]">
                {w.e}
              </div>
              <span
                className={cn(
                  "flex-1 font-body text-[16px] font-medium",
                  w.you ? "font-extrabold text-gold-ink" : "text-paper",
                )}
              >
                {w.nm}
              </span>
              <span
                className={cn(
                  "font-display text-[17px] font-extrabold",
                  w.you ? "text-gold-ink" : "text-naira",
                )}
              >
                {w.amt}
              </span>
            </div>
          ))}
        </div>

        {/* claim card */}
        <div className="mt-5 rounded-[18px] border border-stroke bg-surface-2 px-5 py-4">
          <div className="font-body text-[12px] font-bold tracking-[0.12em] text-text-3">
            YOUR CLAIM CODE
          </div>
          <div className="mt-2 flex items-center justify-between">
            <span className="font-display text-[20px] font-extrabold tracking-wide text-gold">
              KWEEKS-7F3A-9Z
            </span>
            <span className="rounded-lg bg-surface px-3 py-1.5 font-body text-[12px] font-bold text-paper">
              COPY
            </span>
          </div>
          <div className="mt-2 font-body text-[12.5px] text-text-3">
            Only this device and email can redeem it. Don't share it.
          </div>
        </div>

        {/* redeem */}
        <button className="mt-4 flex h-[56px] w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90">
          REDEEM ₦15,000 NOW
        </button>

        {/* steps */}
        <div className="mt-5 flex flex-col gap-3 pb-2">
          {steps.map((st) => (
            <div key={st.n} className="flex items-start gap-3">
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface-2">
                <span className="font-display text-[13px] font-extrabold text-naira">{st.n}</span>
              </div>
              <div>
                <div className="font-body text-[14px] font-semibold text-paper">{st.t}</div>
                <div className="font-body text-[12.5px] text-text-3">{st.s}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </PlayerFrame>
  );
}
