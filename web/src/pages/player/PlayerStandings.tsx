import { PlayerFrame } from "@/components/ui/phone";
import { cn } from "@/lib/cn";

type Row = { rk: string; e: string; nm: string; pts: string; you?: boolean; lead?: boolean };

const rows: Row[] = [
  { rk: "1", e: "🦊", nm: "Ada", pts: "3,400", lead: true },
  { rk: "2", e: "🐼", nm: "Tobi", pts: "3,050" },
  { rk: "3", e: "🐙", nm: "Zainab", pts: "2,900", you: true },
  { rk: "4", e: "😎", nm: "Chidi", pts: "2,600" },
  { rk: "5", e: "🤖", nm: "Uche", pts: "2,400" },
];

export function PlayerStandings() {
  return (
    <PlayerFrame statusTime="9:42">
      <div className="flex flex-col">
        {/* header */}
        <div className="flex items-center justify-between">
          <h1 className="font-display text-[22px] font-extrabold text-paper">Live standings</h1>
          <span className="flex items-center gap-[5px]">
            <span className="h-2 w-2 rounded-full bg-red" />
            <span className="font-body text-[11px] font-bold tracking-widest text-red">LIVE</span>
          </span>
        </div>
        <div className="mt-1 flex items-center gap-1 font-body text-[13px] text-text-3">
          <span>After question 4</span>
          <span>·</span>
          <span>8 players</span>
        </div>

        {/* col hint */}
        <div className="mt-3 font-body text-[11.5px] text-text-3">
          pts = speed × correct · you play as Zainab
        </div>

        {/* rows */}
        <div className="mt-3 flex flex-col gap-2">
          {rows.map((r) => {
            const isYou = r.you;
            return (
              <div
                key={r.rk}
                className={cn(
                  "flex items-center gap-3 rounded-[16px] border px-4 py-2.5",
                  isYou ? "border-gold/70 bg-gold/10" : "border-transparent bg-surface",
                )}
              >
                <span
                  className={cn(
                    "w-5 font-display text-[18px] font-extrabold",
                    r.lead || isYou ? "text-gold" : "text-text-2",
                  )}
                >
                  {r.rk}
                </span>
                <div className="flex h-[38px] w-[38px] items-center justify-center rounded-full bg-surface-2 text-[21px]">
                  {r.e}
                </div>
                <div className="flex flex-1 items-center gap-2">
                  <span className="font-body text-[16px] font-medium text-paper">{r.nm}</span>
                  {isYou && (
                    <span className="rounded-full bg-gold px-2 py-0.5 font-body text-[10px] font-extrabold text-gold-ink">
                      YOU
                    </span>
                  )}
                </div>
                <span
                  className={cn(
                    "font-display text-[17px] font-extrabold",
                    r.lead || isYou ? "text-gold" : "text-paper",
                  )}
                >
                  {r.pts}
                </span>
              </div>
            );
          })}
        </div>

        {/* you pill */}
        <div className="mt-5 rounded-[16px] bg-surface px-4 py-3 text-center font-body text-[13.5px] text-paper">
          🐙 You're 3rd — 4 questions left. Top 3 cash out.
        </div>
      </div>
    </PlayerFrame>
  );
}
