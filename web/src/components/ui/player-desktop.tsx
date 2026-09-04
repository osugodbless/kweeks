export function PlayerTopBar({
  left,
  right,
}: {
  left: React.ReactNode;
  right?: React.ReactNode;
}) {
  return (
    <div className="flex h-14 w-full shrink-0 items-center justify-between bg-bg px-12">
      <div className="flex items-center gap-2">{left}</div>
      {right}
    </div>
  );
}

export function PlayerDesktop({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className="flex min-h-screen w-full flex-col bg-bg">
      <div className={"mx-auto flex w-full max-w-[1180px] flex-1 flex-col " + (className ?? "")}>
        {children}
      </div>
    </div>
  );
}

export function PlayerFooter() {
  return (
    <footer className="w-full border-t border-stroke bg-bg">
      <div className="flex h-[44px] w-full items-center justify-between px-12">
        <span className="font-body text-[12px] text-text-3">© 2026 Kweeks</span>
        <span className="font-body text-[12px] text-text-3">live money quiz</span>
      </div>
    </footer>
  );
}

export function RoomPill({ code = "AB12" }: { code?: string }) {
  return (
    <div className="flex items-center gap-1.5 rounded-full bg-surface px-3.5 py-2">
      <span className="font-body text-[12px] text-text-3">Room</span>
      <span className="font-display text-[15px] font-extrabold text-gold">{code}</span>
    </div>
  );
}

export function KweeksBrand({ size = 20 }: { size?: number }) {
  const cls =
    size === 22 ? "text-[22px]" : size === 20 ? "text-[20px]" : "text-[20px]";
  return <span className={"font-display font-extrabold text-paper " + cls}>KWEEKS</span>;
}
