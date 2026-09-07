import { useNavigate } from "react-router-dom";
import { History, Trophy } from "lucide-react";
import { useHistory } from "@/lib/hooks";
import { naira } from "@/lib/player";
import { Button } from "@/components/ui/button";
import { Chip } from "@/components/ui/chip";
import { Footer } from "@/components/ui/footer";
import { InstructorNav } from "@/components/ui/instructor-nav";
import { HistoryEmpty } from "@/pages/instructor/InstructorHistoryEmpty";
import { cn } from "@/lib/cn";

const TYPE_META: Record<string, { label: string; tone: "mint" | "violet" | "gold" | "coral" }> = {
  fund: { label: "Funding", tone: "mint" },
  credit: { label: "Credit", tone: "mint" },
  pool: { label: "Pool", tone: "gold" },
  quiz: { label: "Quiz", tone: "violet" },
  payout: { label: "Payout", tone: "gold" },
  room: { label: "Room", tone: "coral" },
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString("en-NG", {
      day: "numeric",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function InstructorHistory() {
  const navigate = useNavigate();
  const { data: items = [], isLoading } = useHistory();

  return (
    <main className="relative flex min-h-screen flex-col overflow-hidden">
      <div className="bg-dots pointer-events-none absolute inset-0 opacity-40" />
      <span className="pointer-events-none absolute -right-24 -top-10 size-80 rounded-full bg-violet/20 blur-3xl" />
      <span className="pointer-events-none absolute -left-24 bottom-0 size-72 rounded-full bg-gold/15 blur-3xl" />

      <InstructorNav active="/instructor/history" />

      <div className="relative mx-auto flex w-full max-w-6xl flex-1 flex-col px-4 pt-12 sm:px-6 lg:pt-16">
        <span className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-cream px-4 py-1.5 text-sm font-bold uppercase tracking-widest text-soft">
          <History className="size-4 text-violet" /> History
        </span>
        <h1 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-tight sm:text-5xl">
          Every funding, room &amp; <span className="text-pop">payout</span>
        </h1>

        {isLoading ? (
          <div className="mt-10 flex justify-center py-16 text-sm font-bold text-soft">Loading history…</div>
        ) : items.length === 0 ? (
          <div className="mt-10 flex-1">
            <HistoryEmpty onCreate={() => navigate("/instructor/quiz-builder")} />
          </div>
        ) : (
          <div className="mt-10 overflow-hidden rounded-[2rem] border-2 border-ink/5 bg-cream card-3d">
            <div className="hidden grid-cols-[1fr_7rem_8rem_8rem] gap-3 border-b border-ink/5 px-6 py-3 text-[11px] font-bold uppercase tracking-widest text-soft sm:grid">
              <span>Activity</span>
              <span>Type</span>
              <span className="text-right">Amount</span>
              <span className="text-right">Status</span>
            </div>
            <ul>
              {items.map((item, i) => {
                const meta = TYPE_META[item.type] ?? { label: item.type, tone: "violet" as const };
                const hasAmount = item.amountNaira !== undefined && item.amountNaira !== "";
                const isInflow = item.type === "fund" || item.type === "credit";
                return (
                  <li
                    key={item.id}
                    className={cn(
                      "grid grid-cols-1 gap-2 px-6 py-4 sm:grid-cols-[1fr_7rem_8rem_8rem] sm:items-center sm:gap-3",
                      i > 0 && "border-t border-ink/5",
                    )}
                  >
                    <div className="min-w-0">
                      <p className="truncate font-bold">{item.title}</p>
                      <p className="text-xs text-soft">{formatDate(item.at)}</p>
                      {item.meta && <p className="mt-0.5 text-xs text-soft">{item.meta}</p>}
                    </div>
                    <div>
                      <Chip tone={meta.tone}>{meta.label}</Chip>
                    </div>
                    <div className="text-right">
                      {hasAmount ? (
                        <span
                          className={cn(
                            "font-display text-base font-semibold",
                            isInflow ? "text-mint-dark" : "text-soft",
                          )}
                        >
                          {isInflow ? "+" : "−"}
                          {naira(item.amountNaira!)}
                        </span>
                      ) : (
                        <span className="text-sm font-bold text-soft">—</span>
                      )}
                    </div>
                    <div className="text-right">
                      {item.state ? (
                        <Chip tone={item.state.toLowerCase() === "paid" ? "mint" : "soft"}>{item.state}</Chip>
                      ) : (
                        <span className="text-sm font-bold text-soft">Done</span>
                      )}
                    </div>
                  </li>
                );
              })}
            </ul>
          </div>
        )}

        <div className="mt-8 flex justify-center">
          <Button variant="coral" size="lg" onClick={() => navigate("/instructor/quiz-builder")}>
            <Trophy className="size-5" /> Create a quiz
          </Button>
        </div>
      </div>

      <Footer className="mt-16" />
    </main>
  );
}

export default InstructorHistory;
