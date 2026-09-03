import { useNavigate } from "react-router-dom";
import { DesktopFrame } from "@/components/ui/desktop";
import { InstructorNav, NavRight, InstructorFooter } from "@/components/ui/instructor-nav";

export function InstructorHistoryEmpty() {
  const nav = useNavigate();
  return (
    <DesktopFrame>
      <InstructorNav
        activeKey="history"
        right={<NavRight amount="₦150,000" />}
      />
      <div className="flex flex-1 flex-col items-center justify-center gap-[18px] bg-bg px-8 py-10">
        <div className="flex h-24 w-24 items-center justify-center rounded-full border border-stroke bg-surface-2 text-[44px]">
          🗂️
        </div>
        <h1 className="font-display text-[28px] font-extrabold text-paper">No history yet</h1>
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
      <InstructorFooter />
    </DesktopFrame>
  );
}
