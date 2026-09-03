import { useState } from "react";
import { PlayerFrame } from "@/components/ui/phone";
import { cn } from "@/lib/cn";

const options = [
  { key: "A", text: "Ibadan" },
  { key: "B", text: "Lagos" },
  { key: "C", text: "Abuja" },
  { key: "D", text: "Enugu" },
];

export function PlayerQuestion() {
  const [selected, setSelected] = useState<string | null>("A");

  return (
    <PlayerFrame statusTime="9:42">
      <div className="flex flex-col">
        {/* top row */}
        <div className="flex items-center justify-between">
          <span className="font-body text-[13px] text-text-3">Question 3 of 10</span>
          <span className="font-body text-[13px] font-bold text-gold">1,250 pts</span>
        </div>

        {/* timer */}
        <div className="mt-5 flex flex-col items-center">
          <div className="relative flex h-16 w-16 items-center justify-center">
            <svg viewBox="0 0 64 64" className="h-16 w-16 -rotate-90">
              <circle cx="32" cy="32" r="27" fill="none" stroke="var(--color-surface-2)" strokeWidth="5" />
              <circle
                cx="32" cy="32" r="27" fill="none" stroke="var(--color-gold)" strokeWidth="5"
                strokeDasharray="169.6" strokeDashoffset={169.6 * (1 - 18 / 30)}
                strokeLinecap="round"
              />
            </svg>
            <span className="absolute font-display text-[26px] font-extrabold text-paper">18</span>
          </div>
          <span className="mt-2 font-body text-[12.5px] text-text-3">
            Answer before the bar runs out
          </span>
        </div>

        {/* gold progress bar */}
        <div className="mt-3 h-1 w-full overflow-hidden rounded-full bg-surface-2">
          <div className="h-full w-[60%] rounded-full bg-gold" />
        </div>

        {/* question */}
        <h1 className="mt-6 font-display text-[26px] font-extrabold leading-tight text-paper">
          Which Nigerian city is known as the Centre of Excellence?
        </h1>

        {/* options */}
        <div className="mt-6 flex flex-col gap-3">
          {options.map((o) => {
            const on = selected === o.key;
            return (
              <button
                key={o.key}
                onClick={() => setSelected(o.key)}
                className={cn(
                  "flex items-center gap-4 rounded-2xl border px-5 py-4 text-left transition",
                  on ? "border-gold bg-gold" : "border-stroke bg-surface hover:border-text-3",
                )}
              >
                <span
                  className={cn(
                    "font-display text-[15px] font-bold",
                    on ? "text-gold-ink" : "text-text-2",
                  )}
                >
                  {o.key}
                </span>
                <span
                  className={cn(
                    "font-body text-[17px] font-medium",
                    on ? "text-gold-ink" : "text-paper",
                  )}
                >
                  {o.text}
                </span>
              </button>
            );
          })}
        </div>
      </div>
    </PlayerFrame>
  );
}
