import { useState } from "react";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { useAuth } from "@/lib/auth";
import { useFundWallet, useWallet } from "@/lib/hooks";
import { fmtNaira, naira } from "@/lib/player";
import { cn } from "@/lib/cn";

const QUICKS = [
  { label: "₦1,000", value: "1000" },
  { label: "₦5,000", value: "5000" },
  { label: "₦50,000", value: "50000" },
  { label: "₦100,000", value: "100000" },
];

const METHODS = [
  { method: "credit", t: "Wallet credit", s: "Instant · issued from your Kweeks balance" },
  { method: "card", t: "Debit card", s: "Verve / Mastercard · instant" },
  { method: "transfer", t: "Bank transfer", s: "Virtual account · 1–2 min" },
];

export function InstructorFundWallet() {
  const walletView = useWallet();
  const fund = useFundWallet();
  const authBal = useAuth((s) => s.wallet?.balanceNaira);
  const [amount, setAmount] = useState("50000");
  const [method, setMethod] = useState("credit");
  const [note, setNote] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  const balance = walletView.data?.wallet.balanceNaira ?? authBal ?? "0";
  const pending = fund.isPending;

  const onFund = async () => {
    if (pending) return;
    setNote(null);
    setErr(null);
    try {
      const res = await fund.mutateAsync({ amountNaira: amount, method });
      const bal = res.wallet?.balanceNaira;
      setNote(
        bal ? `Wallet credited — available balance ${naira(bal)}.` : "Wallet credited — it's ready for pools.",
      );
    } catch (x) {
      const m = x instanceof Error ? x.message : "Funding failed — please try again.";
      setErr(method === "credit" ? m : `${m} If card or transfer fails, try instant wallet credit.`);
    }
  };

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="dashboard"
        right={<NavRight amount={naira(balance)} />}
      />
      <div className="flex flex-1 flex-col gap-[22px] bg-bg px-14 py-10">
        <h1 className="font-display text-[30px] font-extrabold text-paper">Fund your wallet</h1>
        <p className="-mt-4 font-body text-[14px] text-text-2">
          Money you add is ready for pools the moment it lands. All naira.
        </p>

        <div className="flex w-[560px] flex-col gap-5 rounded-3xl border border-stroke bg-surface-2 px-[26px] py-[22px]">
          <div className="flex items-center justify-between">
            <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
              AMOUNT
            </div>
            <div className="font-body text-[12px] text-text-3">
              Available balance{" "}
              <span className="font-bold text-naira">{naira(balance)}</span>
            </div>
          </div>

          <div className="flex h-[58px] items-center gap-2 rounded-[14px] border border-stroke bg-surface px-[18px]">
            <span className="font-display text-[24px] font-extrabold text-naira">₦</span>
            <span className="font-display text-[26px] font-extrabold text-paper">{fmtNaira(amount)}</span>
          </div>

          <div className="flex items-center gap-2.5">
            {QUICKS.map((q) => {
              const sel = amount === q.value;
              return (
                <button
                  key={q.value}
                  onClick={() => setAmount(q.value)}
                  disabled={pending}
                  className={cn(
                    "rounded-full px-4 py-2.5 font-body text-[13px] font-bold disabled:opacity-40",
                    sel ? "bg-gold text-gold-ink" : "bg-surface text-text-2",
                  )}
                >
                  {q.label}
                </button>
              );
            })}
          </div>

          <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
            FUND WITH
          </div>

          <div className="flex flex-col gap-2">
            {METHODS.map((m) => {
              const on = method === m.method;
              return (
                <button
                  key={m.method}
                  onClick={() => setMethod(m.method)}
                  disabled={pending}
                  className={cn(
                    "flex items-center justify-between rounded-[14px] px-4 py-3.5 text-left disabled:opacity-60",
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

          <button
            onClick={() => void onFund()}
            disabled={pending}
            className="flex h-14 w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink hover:opacity-90 disabled:opacity-40"
          >
            {pending ? "FUNDING…" : `FUND ${naira(amount)}`}
          </button>

          {err && (
            <p className="rounded-2xl border border-stroke bg-surface px-4 py-3 font-body text-[13px] font-semibold text-red">
              {err}
            </p>
          )}

          {note && (
            <p className="rounded-2xl border border-stroke bg-surface px-4 py-3 font-body text-[13px] font-semibold text-naira">
              {note}
            </p>
          )}

          <div className="font-body text-[12.5px] text-text-3">
            Wallet credits post instantly to your available balance.
          </div>
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
