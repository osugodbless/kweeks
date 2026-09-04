import { cn } from "@/lib/cn";
import { NavLink } from "react-router-dom";
import { useAuth } from "@/lib/auth";
import { naira } from "@/lib/player";

export const INSTRUCTOR_LINKS = [
  { label: "Dashboard", to: "/instructor/dashboard", activeKey: "dashboard" },
  { label: "Create quiz", to: "/instructor/quiz-builder", activeKey: "create" },
  { label: "History", to: "/instructor/history", activeKey: "history" },
];

export function InstructorNav({
  activeKey,
  right,
  liveLabel,
}: {
  activeKey?: string;
  right?: React.ReactNode;
  liveLabel?: string;
}) {
  return (
    <header className="flex h-16 w-full shrink-0 items-center justify-between bg-surface px-7">
      <div className="flex items-center gap-2">
        <span className="font-display text-[22px] font-extrabold text-paper">KWEEKS</span>
        {liveLabel && (
          <span className="ml-1 inline-flex items-center gap-[5px] rounded-full bg-surface px-2.5 py-1">
            <span className="h-2 w-2 rounded-full bg-red" />
            <span className="font-body text-[11px] font-bold tracking-widest text-red">
              {liveLabel}
            </span>
          </span>
        )}
        <nav className="ml-5 flex items-center gap-1">
          {INSTRUCTOR_LINKS.map((l) => (
            <NavLink
              key={l.to}
              to={l.to}
              className={cn(
                "rounded-full px-[13px] py-[9px] font-body text-[13px] font-bold transition-colors",
                activeKey === l.activeKey
                  ? "bg-surface-2 text-gold"
                  : "bg-surface text-text-2 hover:text-paper",
              )}
            >
              {l.label}
            </NavLink>
          ))}
        </nav>
      </div>
      <div className="flex items-center gap-3.5">{right}</div>
    </header>
  );
}

export function NavRight({ amount, center, initials }: { amount?: string; center?: React.ReactNode; initials?: string }) {
  const wallet = useAuth((s) => s.wallet);
  const instructor = useAuth((s) => s.instructor);
  const shown = amount ?? (wallet ? naira(wallet.balanceNaira) : "₦0");
  const avatar = initials ?? instructor?.avatar ?? "AP";
  return (
    <div className="flex items-center gap-3.5">
      {center}
      <div className="flex items-center gap-1 rounded-full bg-surface-2 px-3 py-2">
        <span className="font-body text-[11px] font-bold tracking-widest text-text-3">WALLET</span>
        <span className="font-display text-[15px] font-extrabold text-naira">{shown}</span>
      </div>
      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-surface-2">
        <span className="font-body text-[13px] font-extrabold text-violet">{avatar}</span>
      </div>
    </div>
  );
}

export function InstructorFooter() {
  return (
    <footer className="w-full border-t border-stroke bg-bg">
      <div className="mx-auto flex h-[44px] w-full max-w-[1180px] items-center justify-between px-6">
        <span className="font-body text-[12px] text-text-3">© 2026 Kweeks</span>
        <span className="font-body text-[12px] text-text-3">Live money quiz · NGN</span>
        <span className="font-body text-[12px] text-text-3">Support · Terms · Privacy</span>
      </div>
    </footer>
  );
}
