import { cn } from "@/lib/cn";

export function KweeksWordmark({
  className,
  size = "sm",
}: {
  className?: string;
  size?: "sm" | "md" | "lg";
}) {
  return (
    <div className={cn("flex items-center", className)}>
      <span
        className={cn(
          "font-display font-extrabold tracking-wide text-paper",
          size === "lg" && "text-[30px]",
          size === "md" && "text-[22px]",
          size === "sm" && "text-[16px]",
        )}
      >
        KWEEKS
      </span>
    </div>
  );
}
