export function DesktopFrame({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div className="flex min-h-screen w-full flex-col bg-bg">
      <div
        className={
          "mx-auto flex w-full max-w-[1180px] flex-1 flex-col " + (className ?? "")
        }
      >
        {children}
      </div>
    </div>
  );
}
