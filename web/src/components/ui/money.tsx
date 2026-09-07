import { naira } from "@/lib/player";
import { cn } from "@/lib/cn";

interface MoneyProps {
  value: string | number;
  /** "dark" renders mint-on-ink (for dark chips/surfaces); default is mint-dark on light surfaces. */
  tone?: "light" | "dark";
  className?: string;
  ariaLabel?: string;
}

/**
 * Every money figure in the app renders through this component: naira green,
 * display typeface, grouped thousands. Money is ALWAYS green, never gold.
 */
export function Money({ value, tone = "light", className, ariaLabel }: MoneyProps) {
  return (
    <span
      className={cn(
        "font-display font-semibold tracking-tight",
        tone === "dark" ? "text-mint" : "text-mint-dark",
        className,
      )}
      aria-label={ariaLabel ?? naira(value)}
    >
      {naira(value)}
    </span>
  );
}
