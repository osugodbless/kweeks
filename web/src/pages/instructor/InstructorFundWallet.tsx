import { useState } from "react";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";

const quicks = ["₦1,000", "₦5,000", "₦50,000", "₦100,000"];

const methods = [
  { t: "Wallet credit", s: "Instant · issued from your Kweeks balance", active: true },
  { t: "Debit card", s: "Verve / Mastercard · instant", active: false },
  { t: "Bank transfer", s: "Virtual account · 1–2 min", active: false },
];

export function InstructorFundWallet() {
  const [amount] = useState("50,000");
  const [method, setMethod] = useState(0);

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="dashboard"
        right={<NavRight amount="₦150,000" />}
      />
      <div className="flex flex-1 flex-col gap-[22px] bg-bg px-14 py-10">
        <h1 className="font-display text-[30px] font-extrabold text-paper">Fund your wallet</h1>
        <p className="-mt-4 font-body text-[14px] text-text-2">
          Money you add is ready for pools the moment it lands. All naira.
        </p>

        <div className="flex w-[560px] flex-col gap-5 rounded-3xl border border-stroke bg-surface-2 px-[26px] py-[22px]">
          <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            AMOUNT
          </div>

          <div className="flex h-[58px] items-center gap-2 rounded-[14px] border border-stroke bg-surface px-[18px]">
            <span className="font-display text-[24px] font-extrabold text-naira">₦</span>
            <span className="font-display text-[26px] font-extrabold text-paper">{amount}</span>
          </div>

          <div className="flex items-center gap-2.5">
            {quicks.map((q, i) => {
              const sel = i === 2;
              return (
                <button
                  key={q}
                  className={cn(
                    "rounded-full px-4 py-2.5 font-body text-[13px] font-bold",
                    sel ? "bg-gold text-gold-ink" : "bg-surface text-text-2",
                  )}
                >
                  {q}
                </button>
              );
            })}
          </div>

          <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            FUND WITH
          </div>

          <div className="flex flex-col gap-2">
            {methods.map((m, i) => {
              const on = method === i;
              return (
                <button
                  key={m.t}
                  onClick={() => setMethod(i)}
                  className={cn(
                    "flex items-center justify-between rounded-[14px] px-4 py-3.5 text-left",
                    on
                      ? "border-[1.5px] border-naira bg-surface"
                      : "border border-stroke bg-surface",
                  )}
                >
                  <div>
                    <div
                      className={cn(
                        "font-body text-[14px] font-bold",
                        on ? "text-naira" : "text-paper",
                      )}
                    >
                      {m.t}
                    </div>
                    <div className="font-body text-[12.5px] text-text-3">{m.s}</div>
                  </div>
                  <span className={cn("text-[16px]", on ? "text-naira" : "text-text-3")}>
                    {on ? "●" : "○"}
                  </span>
                </button>
              );
            })}
          </div>

          <button className="flex h-14 w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90">
            FUND ₦50,000
          </button>
          <div className="font-body text-[12.5px] text-text-3">
            Wallet credits post instantly to your available balance.
          </div>
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
