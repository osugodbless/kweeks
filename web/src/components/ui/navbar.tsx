import { Link, NavLink, useLocation } from "react-router-dom";
import { LayoutDashboard, LogIn, LogOut } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { Wordmark } from "@/components/ui/wordmark";
import { cn } from "@/lib/cn";
import type { CSSProperties } from "react";

const PUBLIC_LINKS = [
  { label: "Play", to: "/join" },
  { label: "Host", to: "/instructor/signup" },
];

const ANCHORS = [
  { id: "for-players", label: "For players" },
  { id: "for-instructors", label: "For instructors" },
  { id: "how-it-works", label: "How it works" },
  { id: "security", label: "Security" },
];

function scrollTo(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: "smooth", block: "start" });
}

export function Navbar() {
  const { instructor, logout } = useAuth();
  const { pathname } = useLocation();
  const onLanding = pathname === "/";

  return (
    <header className="sticky top-0 z-50 border-b border-ink/5 bg-canvas/85 backdrop-blur-md">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-3 px-4 sm:px-6">
        <Wordmark />

        <nav className="flex items-center gap-1 sm:gap-2">
          {onLanding && (
            <ul className="hidden items-center gap-1 lg:flex">
              {ANCHORS.map((a) => (
                <li key={a.id}>
                  <button
                    type="button"
                    onClick={() => scrollTo(a.id)}
                    className="cursor-pointer rounded-full px-3 py-2 text-sm font-bold text-soft transition-colors hover:bg-ink/5 hover:text-ink"
                  >
                    {a.label}
                  </button>
                </li>
              ))}
            </ul>
          )}

          <ul className="hidden items-center gap-1 md:flex">
            {PUBLIC_LINKS.map((link) => (
              <li key={link.to}>
                <NavLink
                  to={link.to}
                  className={({ isActive }) =>
                    cn(
                      "rounded-full px-4 py-2 text-sm font-bold transition-colors hover:bg-ink/5 hover:text-ink",
                      isActive ? "bg-ink/5 text-ink" : "text-soft",
                    )
                  }
                >
                  {link.label}
                </NavLink>
              </li>
            ))}
          </ul>

          {instructor ? (
            <div className="flex items-center gap-2">
              <Link
                to="/instructor/dashboard"
                className="group flex items-center gap-2 rounded-full bg-violet px-4 py-2.5 text-sm font-bold text-white press-3d sm:px-5"
                style={{ "--btn-deep-rgb": "var(--color-violet-dark)" } as CSSProperties}
              >
                <LayoutDashboard className="size-4" strokeWidth={2.5} />
                <span>Studio</span>
              </Link>
              <button
                type="button"
                onClick={logout}
                aria-label="Sign out"
                title="Sign out"
                className="grid size-10 cursor-pointer place-items-center rounded-full border-2 border-ink/10 bg-cream text-soft transition-colors hover:border-coral hover:text-coral"
              >
                <LogOut className="size-4.5" strokeWidth={2.5} />
              </button>
            </div>
          ) : (
            <Link
              to="/instructor/login"
              className="group flex items-center gap-2 rounded-full bg-violet px-4 py-2.5 text-sm font-bold text-white press-3d sm:px-5"
              style={{ "--btn-deep-rgb": "var(--color-violet-dark)" } as CSSProperties}
            >
              <LogIn className="size-4" strokeWidth={2.5} />
              <span>Host login</span>
            </Link>
          )}
        </nav>
      </div>
    </header>
  );
}
