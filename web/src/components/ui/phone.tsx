import { cn } from "@/lib/cn";

export function StatusBar({ time = "9:42", className }: { time?: string; className?: string }) {
  return (
    <div
      className={cn(
        "flex h-[54px] w-full items-center justify-between px-[28px]",
        className,
      )}
    >
      <span className="font-body text-[14px] font-medium text-paper">{time}</span>
      <span className="font-body text-[12px] tracking-wider text-text-2">●●●● ▮▮ 🔋</span>
    </div>
  );
}

export function PhoneFooter({ className }: { className?: string }) {
  return (
    <footer
      className={cn(
        "mt-auto w-full border-t border-stroke bg-bg",
        className,
      )}
    >
      <div className="flex h-[44px] w-full items-center justify-between px-6">
        <span className="font-body text-[12px] text-text-3">© 2026 Kweeks</span>
        <span className="font-body text-[12px] text-text-3">live money quiz</span>
      </div>
    </footer>
  );
}

/** Wraps a mobile player frame: 390-wide phone canvas centered on a dark backdrop. */
export function PhoneShell({
  children,
  className,
  darkBackdrop = true,
}: {
  children: React.ReactNode;
  className?: string;
  darkBackdrop?: boolean;
}) {
  return (
    <div className={cn("min-h-screen w-full", darkBackdrop && "bg-black")}>
      <div
        className={cn(
          "mx-auto flex min-h-screen w-full max-w-[390px] flex-col border-x border-stroke/40 bg-bg",
          className,
        )}
      >
        {children}
      </div>
    </div>
  );
}

export function PlayerFrame({
  children,
  className,
  statusTime,
}: {
  children: React.ReactNode;
  className?: string;
  statusTime?: string;
}) {
  return (
    <PhoneShell>
      <StatusBar time={statusTime} />
      <div className={cn("flex-1 px-5", className)}>{children}</div>
      <PhoneFooter />
    </PhoneShell>
  );
}
