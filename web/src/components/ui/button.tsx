import { forwardRef, type ButtonHTMLAttributes, type CSSProperties, type ReactNode } from "react";
import { cn } from "@/lib/cn";
import { LoaderCircle } from "lucide-react";

type Variant = "coral" | "violet" | "gold" | "mint" | "ink" | "ghost" | "outline";
type Size = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  loading?: boolean;
  icon?: ReactNode;
}

const DEEP: Record<Variant, string> = {
  coral: "var(--color-coral-dark)",
  violet: "var(--color-violet-dark)",
  gold: "var(--color-gold-dark)",
  mint: "var(--color-mint-dark)",
  ink: "var(--color-ink)",
  ghost: "var(--color-coral-dark)",
  outline: "var(--color-coral-dark)",
};

const SOLID: Record<Variant, string> = {
  coral: "bg-coral text-white",
  violet: "bg-violet text-white",
  gold: "bg-gold text-ink",
  mint: "bg-mint text-white",
  ink: "bg-ink text-cream",
  ghost: "bg-cream text-ink",
  outline: "bg-cream text-ink",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { className, variant = "coral", size = "md", loading, icon, children, disabled, ...props },
  ref,
) {
  const sizes: Record<Size, string> = {
    sm: "px-4 py-2 text-sm gap-2",
    md: "px-5 py-3 text-sm gap-2",
    lg: "px-8 py-4 font-display text-lg gap-2.5",
  };

  const isSolid = variant !== "ghost" && variant !== "outline";

  return (
    <button
      ref={ref}
      disabled={disabled || loading}
      className={cn(
        "inline-flex cursor-pointer items-center justify-center rounded-full font-bold transition-all duration-150",
        "focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-violet/30 focus-visible:ring-offset-2 focus-visible:ring-offset-canvas",
        "disabled:cursor-not-allowed disabled:opacity-75",
        SOLID[variant],
        isSolid && "press-3d",
        variant === "ghost" && "hover:bg-ink/5",
        variant === "outline" && "border-2 border-ink/10 hover:border-violet hover:text-violet",
        sizes[size],
        className,
      )}
      style={{ "--btn-deep-rgb": DEEP[variant] } as CSSProperties}
      {...props}
    >
      {loading ? (
        <>
          <LoaderCircle className="size-4.5 animate-spin" aria-hidden />
          <span>{typeof children === "string" ? children : "Please wait…"}</span>
        </>
      ) : (
        <>
          {icon}
          {children}
        </>
      )}
    </button>
  );
});
