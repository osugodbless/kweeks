import { useNavigate } from "react-router-dom";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";
import { useAuth } from "@/lib/auth";
import { useHistory, useWallet } from "@/lib/hooks";
import { naira } from "@/lib/player";
import type { HistoryItem } from "@/lib/api";

const TYPE_LABEL: Record<HistoryItem["type"], string> = {
  fund: "Wallet credit",
  quiz: "Quiz",
  payout: "Payout",
  room: "Room",
};

const colw = ["w-[430px]", "w-[120px]", "w-[110px]", ""];

function errText(e: unknown): string {
  if (e instanceof Error && e.message) return e.message;
  return "Something went wrong loading your history.";
}

function fmtWhen(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  if (sameDay) {
    const t = d.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
    return `Today · ${t}`;
  }
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}

function fmtPool(raw?: string): { text: string; cls: string } {
  if (!raw) return { text: "—", cls: "text-text-2" };
  const t = raw.trim();
  const neg = t.startsWith("-");
  const digits = t.replace(/[^0-9]/g, "");
  const text = (neg ? "−" : "") + naira(digits === "" ? "0" : digits);
  return { text, cls: neg ? "text-red" : "text-naira" };
}

function chipFor(h: HistoryItem): { text: string | null; cls: string } {
  const st = h.state?.trim();
  const meta = h.meta?.trim();
  if (!st && !meta) return { text: null, cls: "text-text-2" };
  const cap = (v: string) => v.charAt(0).toUpperCase() + v.slice(1);
  const text = st && meta ? `${cap(st)} · ${meta}` : cap(st ?? meta ?? "");
  const k = (st ?? text).toLowerCase();
  let cls = "text-text-2";
  if (k.includes("live") || k.includes("lobby")) cls = "text-gold";
  else if (/paid|paid-out|completed|settled|success|credited|redeemed/.test(k)) cls = "text-naira";
  return { text, cls };
}

export function InstructorHistory() {
  const nav = useNavigate();
  const his = useHistory();
  const walletView = useWallet();
  const authBal = useAuth((s) => s.wallet?.balanceNaira);
  const balance = walletView.data?.wallet.balanceNaira ?? authBal ?? "0";

  const items = his.data ?? [];

  const emptyState = (
    <div className="flex flex-col items-center justify-center gap-[18px] rounded-3xl border border-stroke bg-surface px-8 py-16">
      <div className="flex h-24 w-24 items-center justify-center rounded-full border border-stroke bg-surface-2 text-[44px]">
        🗂️
      </div>
      <h2 className="font-display text-[28px] font-extrabold text-paper">No history yet</h2>
      <p className="max-w-[420px] text-center font-body text-[14px] leading-relaxed text-text-2">
        When you fund your wallet or host a quiz, every move lands here — funding, pools, and
        payouts in one place.
      </p>
      <button
        onClick={() => nav("/instructor/quiz-builder")}
        className="mt-2 flex h-[54px] items-center gap-2 rounded-2xl bg-gold px-[26px] font-body text-[15px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
      >
        <span className="font-display text-[20px] font-extrabold">+</span>
        CREATE A QUIZ
      </button>
    </div>
  );

  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="history"
        right={<NavRight amount={naira(balance)} />}
      />
      <div className="flex flex-col gap-5 bg-bg px-8 py-7">
        <h1 className="font-display text-[26px] font-extrabold text-paper">
          Quiz & wallet history
        </h1>
        <p className="-mt-2 font-body text-[14px] text-text-2">
          Every pool you've funded, quiz you've hosted, and winner you've paid.
        </p>

        {his.isPending ? (
          /* loading skeleton */
          <div className="overflow-hidden rounded-3xl border border-stroke bg-surface">
            <div className="flex items-center gap-0 px-5 py-3">
              {["Activity", "Type", "Pool", "Status"].map((h, i) => (
                <span
                  key={h}
                  className={cn(
                    "font-body text-[11px] font-bold tracking-[0.1em] text-text-3",
                    i < 3 ? colw[i] : "flex-1",
                  )}
                >
                  {h}
                </span>
              ))}
            </div>
            {[0, 1, 2, 3, 4].map((r) => (
              <div
                key={r}
                className={cn(
                  "flex items-center gap-0 px-5 py-3.5",
                  r % 2 === 0 ? "bg-surface" : "bg-surface-2",
                )}
              >
                <div className="flex w-[430px] flex-col gap-1.5">
                  <div className="h-[14px] w-3/4 animate-pulse rounded bg-surface-2" />
                  <div className="h-[12px] w-1/3 animate-pulse rounded bg-surface-2" />
                </div>
                <div className="h-[13px] w-20 animate-pulse rounded bg-surface-2" />
                <div className="h-[14px] w-16 animate-pulse rounded bg-surface-2" />
              </div>
            ))}
          </div>
        ) : his.isError ? (
          /* error state */
          <div className="flex flex-col items-center justify-center gap-4 rounded-3xl border border-stroke bg-surface px-8 py-16 text-center">
            <div className="font-display text-[22px] font-extrabold text-paper">
              Couldn't load your history
            </div>
            <p className="max-w-[420px] font-body text-[13.5px] leading-relaxed text-text-2">
              {errText(his.error)}
            </p>
            <button
              onClick={() => void his.refetch()}
              className="mt-2 flex h-[50px] items-center rounded-2xl bg-gold px-8 font-body text-[14px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
            >
              RETRY
            </button>
          </div>
        ) : items.length === 0 ? (
          emptyState
        ) : (
          <div className="overflow-hidden rounded-3xl border border-stroke bg-surface">
            {/* header */}
            <div className="flex items-center gap-0 px-5 py-3">
              {["Activity", "Type", "Pool", "Status"].map((h, i) => (
                <span
                  key={h}
                  className={cn(
                    "font-body text-[11px] font-bold tracking-[0.1em] text-text-3",
                    i < 3 ? colw[i] : "flex-1",
                  )}
                >
                  {h}
                </span>
              ))}
            </div>
            {items.map((r, i) => {
              const amt = fmtPool(r.amountNaira);
              const chip = chipFor(r);
              return (
                <div
                  key={r.id}
                  className={cn(
                    "flex items-center gap-0 px-5 py-3.5",
                    i % 2 === 0 ? "bg-surface" : "bg-surface-2",
                  )}
                >
                  <div className="flex w-[430px] flex-col gap-0.5">
                    <span className="font-body text-[14px] font-bold text-paper">{r.title}</span>
                    <span className="font-body text-[12px] text-text-3">{fmtWhen(r.at)}</span>
                  </div>
                  <span className="w-[120px] font-body text-[13px] text-text-2">
                    {TYPE_LABEL[r.type]}
                  </span>
                  <span className={cn("w-[110px] font-display text-[14px] font-extrabold", amt.cls)}>
                    {amt.text}
                  </span>
                  <span className="flex-1">
                    {chip.text && (
                      <span
                        className={cn(
                          "inline-block rounded-full bg-surface px-2.5 py-1 font-body text-[12px] font-bold",
                          chip.cls,
                        )}
                      >
                        {chip.text}
                      </span>
                    )}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
