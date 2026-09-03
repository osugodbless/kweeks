export function AuthShell({
  lead,
  sub,
  chipLabel,
  foot,
  children,
}: {
  lead: string;
  sub: string;
  chipLabel: string;
  foot: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen w-full flex-col bg-bg">
      <div className="mx-auto flex w-full max-w-[1180px] flex-1">
        {/* brand panel */}
        <aside className="flex w-[430px] flex-col justify-between bg-surface px-9 py-11">
          <div>
            <div className="font-display text-[30px] font-extrabold tracking-wide text-paper">
              KWEEKS
            </div>
            <div className="mt-1 font-body text-[13px] font-semibold text-naira">
              host a live money quiz
            </div>
          </div>
          <div className="flex flex-col gap-5">
            <h1 className="font-display text-[30px] font-extrabold leading-[1.15] text-paper">
              {lead}
            </h1>
            <p className="font-body text-[14px] leading-relaxed text-text-2">{sub}</p>
            <div className="flex w-fit items-center gap-1 rounded-full bg-surface-2 px-3.5 py-2.5">
              <span className="font-display text-[16px] font-extrabold text-naira">₦</span>
              <span className="font-body text-[13px] font-semibold text-text-2">{chipLabel}</span>
            </div>
            <p className="font-body text-[12px] text-text-3">{foot}</p>
          </div>
        </aside>

        {/* form column */}
        <div className="flex flex-1 items-center justify-center bg-bg px-16 py-10">
          <div className="w-[420px] rounded-3xl border border-stroke bg-surface-2 px-8 py-9">
            {children}
          </div>
        </div>
      </div>
      {/* footer */}
      <footer className="w-full">
        <div className="mx-auto flex h-[44px] w-full max-w-[1180px] items-center justify-between border-t border-stroke px-6">
          <span className="font-body text-[12px] text-text-3">© 2026 Kweeks</span>
          <span className="font-body text-[12px] text-text-3">Live money quiz · NGN</span>
          <span className="font-body text-[12px] text-text-3">Support · Terms · Privacy</span>
        </div>
      </footer>
    </div>
  );
}
