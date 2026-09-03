import { PlayerFrame } from "@/components/ui/phone";

const roster = [
  { e: "🐙", you: true },
  { e: "🦊" },
  { e: "🐼" },
  { e: "😎" },
  { e: "🤖" },
];

export function PlayerLobby() {
  return (
    <PlayerFrame statusTime="9:42">
      <div className="flex flex-col">
        {/* room header */}
        <div className="flex items-center justify-between">
          <div className="flex items-baseline gap-2">
            <span className="font-body text-[12px] text-text-3">Room</span>
            <span className="font-display text-[16px] font-extrabold text-gold">AB12</span>
          </div>
          <span className="font-body text-[12px] font-bold tracking-[0.14em] text-text-3">
            WAITING
          </span>
        </div>

        {/* hero */}
        <div className="mt-8 flex flex-col items-center text-center">
          <span className="text-[56px] leading-none">🐙</span>
          <h1 className="mt-3 font-display text-[24px] font-extrabold text-paper">
            You're in, Zainab
          </h1>
          <p className="mt-1 font-body text-[15px] font-medium text-naira">
            ₦50,000 on the line
          </p>
          <p className="mt-2 max-w-[300px] font-body text-[14px] leading-5 text-text-2">
            The host starts in a moment. Keep this tab open — questions move fast.
          </p>
        </div>

        {/* in the room */}
        <div className="mt-8">
          <div className="mb-3 font-body text-[12px] font-bold tracking-[0.12em] text-text-3">
            IN THE ROOM
          </div>
          <div className="flex items-center gap-2">
            {roster.map((r) => (
              <div
                key={r.e}
                className="flex h-9 w-9 items-center justify-center rounded-full bg-surface text-[24px]"
              >
                {r.e}
              </div>
            ))}
          </div>
          <div className="mt-2 font-body text-[13px] text-text-3">
            + 3 more · you are Zainab 🐙
          </div>
        </div>

        {/* waiting card */}
        <div className="mt-8 rounded-[18px] border border-stroke bg-surface-2 px-5 py-5">
          <div className="flex items-center gap-3">
            <span className="relative flex h-3 w-3">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gold opacity-60" />
              <span className="relative inline-flex h-3 w-3 rounded-full bg-gold" />
            </span>
            <h2 className="font-display text-[18px] font-extrabold text-paper">
              The host hasn't started yet
            </h2>
          </div>
          <p className="mt-3 font-body text-[14px] leading-5 text-text-2">
            When the first question drops, everyone in this room sees it at the same second.
            Correct + fast = climb. No answer, no points.
          </p>
        </div>

        {/* prize bar */}
        <div className="mt-6 flex items-center justify-between rounded-[16px] bg-surface px-5 py-4">
          <span className="font-body text-[14px] text-text-2">1st place takes</span>
          <span className="font-display text-[20px] font-extrabold text-naira">₦25,000</span>
        </div>
      </div>
    </PlayerFrame>
  );
}
