import { useState } from "react";
import { NavLink } from "react-router-dom";
import { FileQuestion, History, LayoutDashboard, LogOut, Wallet } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/cn";
import { Money } from "@/components/ui/money";
import { InstructorAvatar } from "@/components/ui/avatar";
import { Chip } from "@/components/ui/chip";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

const NAV = [
  { label: "Dashboard", to: "/instructor/dashboard", Icon: LayoutDashboard },
  { label: "Create quiz", to: "/instructor/quiz-builder", Icon: FileQuestion },
  { label: "History", to: "/instructor/history", Icon: History },
];

export function InstructorNav({ active }: { active?: string }) {
  const { instructor, wallet, logout } = useAuth();
  const [confirmLogout, setConfirmLogout] = useState(false);

  return (
    <header className="sticky top-0 z-50 border-b border-ink/5 bg-canvas/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-3 px-4 sm:px-6">
        <div className="flex items-center gap-5">
          <NavLink to="/instructor/dashboard" className="flex items-center gap-2.5">
            <span className="grid size-9 place-items-center rounded-xl bg-ink text-cream">
              <Wallet className="size-5" strokeWidth={2.5} />
            </span>
            <span className="hidden font-display text-2xl font-semibold tracking-tight sm:block">
              kweeks<span className="text-coral">.</span>
            </span>
          </NavLink>

          <nav className="hidden items-center gap-1 md:flex" aria-label="Instructor navigation">
            {NAV.map(({ label, to, Icon }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-2 rounded-full px-4 py-2 text-sm font-bold transition-colors",
                    isActive || active === to
                      ? "bg-violet text-white"
                      : "text-soft hover:bg-ink/5 hover:text-ink",
                  )
                }
              >
                <Icon className="size-4" strokeWidth={2.5} />
                {label}
              </NavLink>
            ))}
          </nav>
        </div>

        <div className="flex items-center gap-2.5">
          <div className="hidden items-center gap-2 rounded-full border border-mint/30 bg-mint/10 py-1.5 pl-3 pr-4 sm:flex">
            <span className="text-[10px] font-bold uppercase tracking-widest text-mint-dark">Wallet</span>
            <Money value={wallet?.balanceNaira ?? "0"} className="text-sm" />
          </div>
          {instructor && (
            <div className="flex items-center gap-2.5">
              <InstructorAvatar name={instructor.name} className="size-9" />
              <button
                type="button"
                onClick={() => setConfirmLogout(true)}
                aria-label="Sign out"
                title="Sign out"
                className="grid size-9 cursor-pointer place-items-center rounded-full border-2 border-ink/10 bg-cream text-soft transition-colors hover:border-coral hover:text-coral"
              >
                <LogOut className="size-4" strokeWidth={2.5} />
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="border-t border-ink/5 md:hidden">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-2 px-4 py-2 sm:px-6">
          <nav className="flex items-center gap-1">
            {NAV.map(({ label, to }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    "rounded-full px-3 py-1.5 text-xs font-bold",
                    isActive || active === to ? "bg-violet text-white" : "text-soft",
                  )
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
          <Chip tone="mint">
            <Money value={wallet?.balanceNaira ?? "0"} className="text-xs" />
          </Chip>
        </div>
      </div>

      <ConfirmDialog
        open={confirmLogout}
        signOut
        title="Sign out of kweeks?"
        body="You'll need to sign in again to host quizzes or manage your wallet."
        confirmLabel="Sign out"
        cancelLabel="Stay signed in"
        onConfirm={() => {
          setConfirmLogout(false);
          logout();
        }}
        onCancel={() => setConfirmLogout(false)}
      />
    </header>
  );
}
