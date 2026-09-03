import { type ButtonHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/cn";

type Variant = "gold" | "dark" | "outline" | "naira";
type Size = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
}

const variantCls: Record<Variant, string> = {
  gold: "bg-gold text-gold-ink",
  dark: "bg-surface text-paper",
  outline: "bg-surface-2 text-paper border border-stroke",
  naira: "bg-naira text-gold-ink",
};

const sizeCls: Record<Size, string> = {
  sm: "h-9 px-4 text-[12.5px] font-extrabold tracking-wide rounded-xl",
  md: "h-[52px] px-6 text-[14px] font-extrabold tracking-wide rounded-2xl",
  lg: "h-[56px] px-7 text-[15px] font-extrabold tracking-[0.02em] rounded-2xl",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "gold", size = "md", ...props }, ref) => (
    <button
      ref={ref}
      className={cn(
        "inline-flex items-center justify-center gap-2 font-body transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-40",
        variantCls[variant],
        sizeCls[size],
        className,
      )}
      {...props}
    />
  ),
);
Button.displayName = "Button";

export function NavLinkPill({
  label,
  active,
  onClick,
  className,
}: {
  label: string;
  active?: boolean;
  onClick?: () => void;
  className?: string;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "inline-flex items-center rounded-full px-[13px] py-[9px] font-body text-[13px] font-bold transition-colors",
        active ? "bg-surface-2 text-gold" : "bg-surface text-text-2 hover:text-paper",
        className,
      )}
    >
      {label}
    </button>
  );
}

export function Chip({
  label,
  className,
  children,
}: {
  label?: string;
  className?: string;
  children?: React.ReactNode;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-[12px] py-[8px] font-body text-[12px] font-semibold",
        className,
      )}
    >
      {children ?? label}
    </span>
  );
}

export function LiveDot({ className }: { className?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-[5px]", className)}>
      <span className="h-[8px] w-[8px] rounded-full bg-red" />
      <span className="font-body text-[11px] font-bold tracking-widest text-red">LIVE</span>
    </span>
  );
}
