import { Check } from "lucide-react";
import { AVATARS, avatarById, type AvatarDef } from "@/lib/avatars";
import { cn } from "@/lib/cn";

interface AvatarProps {
  id?: string | null;
  def?: AvatarDef;
  className?: string;
  iconClassName?: string;
  /** Applies an inline ring around the circle (e.g. selection, podium). */
  ring?: boolean;
}

export function Avatar({ id, def, className, iconClassName, ring }: AvatarProps) {
  const a = def ?? avatarById(id);
  const Icon = a.Icon;
  return (
    <span
      className={cn("grid shrink-0 place-items-center rounded-full", ring && "ring-4 ring-violet/20", className)}
      style={{ background: a.bg, color: a.fg }}
    >
      <Icon className={cn("size-1/2", iconClassName)} strokeWidth={2.3} />
    </span>
  );
}

interface AvatarPickerProps {
  value: string | null;
  onChange: (id: string) => void;
  error?: string;
}

export function AvatarPicker({ value, onChange, error }: AvatarPickerProps) {
  return (
    <div>
      <div
        role="radiogroup"
        aria-label="Choose an avatar"
        className="grid grid-cols-4 gap-3 sm:grid-cols-8 xl:grid-cols-8"
      >
        {AVATARS.map((a) => {
          const selected = a.id === value;
          const Icon = a.Icon;
          return (
            <button
              key={a.id}
              type="button"
              role="radio"
              aria-checked={selected}
              title={`Choose ${a.label}`}
              onClick={() => onChange(a.id)}
              className={cn(
                "group relative flex cursor-pointer flex-col items-center gap-1.5 rounded-2xl border-2 px-1 pb-2 pt-3 transition-all duration-150",
                selected
                  ? "border-violet bg-white shadow-[0_0_0_4px_rgba(108,76,241,0.15)]"
                  : "border-ink/5 bg-white/50 hover:-translate-y-0.5 hover:border-ink/20 hover:bg-white",
              )}
            >
              {selected && (
                <span className="absolute right-1.5 top-1.5 grid size-4.5 place-items-center rounded-full bg-violet text-white">
                  <Check className="size-3" strokeWidth={3.5} />
                </span>
              )}
              <span
                className={cn(
                  "grid size-10 place-items-center rounded-full transition-transform duration-200 sm:size-11",
                  selected ? "scale-110" : "group-hover:scale-105",
                )}
                style={{ background: a.bg, color: a.fg }}
              >
                <Icon className="size-5.5" strokeWidth={2.3} />
              </span>
              <span
                className={cn(
                  "text-[10px] font-bold leading-none",
                  selected ? "text-violet" : "text-soft/80 group-hover:text-ink",
                )}
              >
                {a.label}
              </span>
            </button>
          );
        })}
      </div>
      {error && <div className="mt-1.5 text-sm font-bold text-coral">{error}</div>}
    </div>
  );
}

export function InstructorAvatar({ name, className }: { name: string; className?: string }) {
  const initial = (name.trim().charAt(0) || "?").toUpperCase();
  return (
    <span
      className={cn(
        "grid size-10 shrink-0 place-items-center rounded-full bg-violet font-display font-semibold text-white",
        className,
      )}
    >
      {initial}
    </span>
  );
}
