import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";
import { cn } from "@/lib/cn";

type HRow = { act: string; when: string; ty: string; pool: string; poolC: string; st: string; stC: string };

const rows: HRow[] = [
  { act: "Funded wallet · card", when: "Today · 9:41 AM", ty: "Wallet credit", pool: "+₦200,000", poolC: "text-naira", st: "Completed", stC: "text-naira" },
  { act: "Hosted · Naija General Knowledge", when: "Today · 10:02 AM", ty: "Quiz", pool: "₦50,000", poolC: "text-paper", st: "Live · AB12", stC: "text-gold" },
  { act: "Paid winners · Ada, Zainab, Tobi", when: "Today · 10:32 AM", ty: "Payout", pool: "−₦50,000", poolC: "text-red", st: "Paid", stC: "text-naira" },
  { act: "Hosted · Jollof Capital", when: "Aug 30", ty: "Quiz", pool: "₦20,000", poolC: "text-paper", st: "Ended", stC: "text-text-2" },
  { act: "Hosted · Naija Anthem Line", when: "Aug 28", ty: "Quiz", pool: "₦15,000", poolC: "text-paper", st: "Ended", stC: "text-text-2" },
];

const colw = ["w-[430px]", "w-[120px]", "w-[110px]", ""];

export function InstructorHistory() {
  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="history"
        right={<NavRight amount="₦150,000" />}
      />
      <div className="flex flex-col gap-5 bg-bg px-8 py-7">
        <h1 className="font-display text-[26px] font-extrabold text-paper">
          Quiz & wallet history
        </h1>
        <p className="-mt-2 font-body text-[14px] text-text-2">
          Every pool you've funded, quiz you've hosted, and winner you've paid.
        </p>

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
          {rows.map((r, i) => (
            <div
              key={r.act}
              className={cn(
                "flex items-center gap-0 px-5 py-3.5",
                i % 2 === 0 ? "bg-surface" : "bg-surface-2",
              )}
            >
              <div className="flex w-[430px] flex-col gap-0.5">
                <span className="font-body text-[14px] font-bold text-paper">{r.act}</span>
                <span className="font-body text-[12px] text-text-3">{r.when}</span>
              </div>
              <span className="w-[120px] font-body text-[13px] text-text-2">{r.ty}</span>
              <span className={cn("w-[110px] font-display text-[14px] font-extrabold", r.poolC)}>
                {r.pool}
              </span>
              <span className="flex-1">
                <span
                  className={cn(
                    "rounded-full bg-surface px-2.5 py-1 font-body text-[12px] font-bold",
                    r.stC,
                  )}
                >
                  {r.st}
                </span>
              </span>
            </div>
          ))}
        </div>
      </div>
      <InstructorFooter />
    </DesktopFrame>
  );
}
