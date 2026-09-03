import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { PlayerFrame } from "@/components/ui/phone";
import { cn } from "@/lib/cn";

const FACES = ["😎", "🤠", "🦁", "🐸", "👾", "🥷"];
const ANIMALS = ["🦊", "🐼", "🐙", "🦄", "🐯", "🤖"];
const SELECTED = "🐙";

export function PlayerJoin() {
  const nav = useNavigate();
  const [avatar, setAvatar] = useState(SELECTED);
  const [nick, setNick] = useState("");
  const [email, setEmail] = useState("");

  return (
    <PlayerFrame statusTime="9:42">
      <div className="flex flex-col gap-4">
        {/* room + live */}
        <div className="flex items-center justify-between">
          <div className="flex items-baseline gap-2">
            <span className="font-body text-[12px] text-text-3">Room</span>
            <span className="font-display text-[16px] font-extrabold text-gold">AB12</span>
          </div>
          <span className="inline-flex items-center gap-[5px]">
            <span className="h-[8px] w-[8px] rounded-full bg-red" />
            <span className="font-body text-[12px] font-bold tracking-widest text-red">LIVE</span>
          </span>
        </div>

        {/* pool hero */}
        <div className="mt-3">
          <div className="font-body text-[12px] font-bold tracking-[0.14em] text-text-3">
            PRIZE POOL
          </div>
          <div className="font-display text-[42px] font-extrabold leading-none text-naira">
            ₦50,000
          </div>
          <div className="mt-2 font-body text-[15px] text-text-2">
            You're in the arena. Pick your face and grab a seat.
          </div>
          <div className="mt-1 font-body text-[13px] text-text-3">10 questions · fastest wins</div>
        </div>

        {/* avatar picker */}
        <div className="mt-4">
          <div className="mb-3 font-body text-[13px] font-medium text-text-2">Pick your avatar</div>
          <div className="flex flex-col gap-2">
            {[FACES, ANIMALS].map((row, ri) => (
              <div key={ri} className="flex gap-2">
                {row.map((e) => (
                  <button
                    key={e}
                    onClick={() => setAvatar(e)}
                    className={cn(
                      "flex h-[50px] w-[50px] items-center justify-center rounded-full text-[28px] transition",
                      avatar === e
                        ? "bg-gold ring-2 ring-gold"
                        : "bg-surface ring-1 ring-stroke",
                    )}
                  >
                    <span className={cn(avatar === e && "grayscale-0")}>{e}</span>
                  </button>
                ))}
              </div>
            ))}
          </div>
          <div className="mt-2 font-body text-[12px] text-text-3">
            One tap. It's how the room will know you.
          </div>
        </div>

        {/* fields */}
        <div className="mt-2 flex flex-col gap-3">
          <label className="block">
            <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
              NICKNAME
            </span>
            <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface px-4">
              <input
                value={nick}
                onChange={(e) => setNick(e.target.value)}
                placeholder="e.g. FastestZebra"
                className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
              />
            </div>
          </label>
          <label className="block">
            <span className="mb-[6px] block font-body text-[11px] font-bold tracking-[0.12em] text-text-3">
              EMAIL (FOR YOUR PRIZE)
            </span>
            <div className="flex h-[50px] items-center rounded-[14px] border border-stroke bg-surface px-4">
              <input
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                className="w-full bg-transparent font-body text-[15px] text-paper outline-none placeholder:text-text-3"
              />
            </div>
          </label>
        </div>

        {/* CTA */}
        <div className="flex flex-col gap-3 pt-[22px]">
          <button
            onClick={() => nav("/lobby")}
            disabled={!nick.trim() || !email.trim()}
            className="flex h-[56px] w-full items-center justify-center rounded-2xl bg-gold font-body text-[16px] font-extrabold tracking-wide text-gold-ink transition hover:opacity-90 disabled:opacity-40"
          >
            JOIN THE GAME
          </button>
          <span className="pb-1 text-center font-body text-[12px] leading-4 text-text-3">
            Winnings land in your email · protected by claim code
          </span>
        </div>
      </div>
    </PlayerFrame>
  );
}
