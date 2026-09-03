import { useNavigate } from "react-router-dom";

const anchor = (id: string) => document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });

export function LandingPage() {
  const nav = useNavigate();
  return (
    <div className="min-h-screen w-full bg-bg">
      <div className="mx-auto flex w-full max-w-[1180px] flex-col">
        {/* NAV */}
        <header className="flex h-16 w-full items-center justify-between bg-surface px-7">
          <div className="flex items-center gap-7">
            <div className="flex items-center gap-2">
              <span className="font-display text-[22px] font-extrabold text-paper">KWEEKS</span>
              <span className="font-body text-[12px] font-semibold text-naira">live money quiz</span>
            </div>
            <nav className="flex items-center gap-1.5">
              {[
                ["For instructors", "instructors"],
                ["For players", "players"],
                ["How it works", "how"],
                ["Security", "security"],
              ].map(([label, id]) => (
                <button
                  key={id}
                  onClick={() => anchor(id)}
                  className="rounded-full bg-surface px-[13px] py-[9px] font-body text-[13px] font-bold text-text-2 hover:text-paper"
                >
                  {label}
                </button>
              ))}
            </nav>
          </div>
          <div className="flex items-center gap-2.5">
            <button
              onClick={() => nav("/instructor/login")}
              className="rounded-xl bg-surface px-4 py-2.5 font-body text-[13px] font-bold text-paper"
            >
              Log in
            </button>
            <button
              onClick={() => nav("/instructor/signup")}
              className="rounded-xl bg-gold px-[18px] py-2.5 font-body text-[13px] font-extrabold text-gold-ink"
            >
              Start hosting
            </button>
          </div>
        </header>

        {/* HERO */}
        <section id="top" className="flex items-center gap-10 bg-bg px-16 py-16">
          <div className="flex-1">
            <div className="font-body text-[12px] font-bold tracking-[0.18em] text-gold">
              THE LIVE MONEY QUIZ PLATFORM
            </div>
            <h1 className="mt-4 font-display text-[44px] font-extrabold leading-[1.12] text-paper">
              Run a quiz. Put real money on it. Pay winners the second it ends.
            </h1>
            <p className="mt-4 font-body text-[16px] leading-relaxed text-text-2">
              Kweeks turns any quiz into a live money event. Instructors fund a naira pool from
              their wallet, players join in seconds by room code, and the top finishers take the
              prize — paid straight from the pool, no paperwork, no delays.
            </p>
            <div className="mt-6 flex items-center gap-3">
              <button
                onClick={() => nav("/instructor/signup")}
                className="flex h-14 items-center justify-center rounded-2xl bg-gold px-7 font-body text-[15px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
              >
                START HOSTING A QUIZ
              </button>
              <button
                onClick={() => nav("/join")}
                className="flex h-14 items-center justify-center rounded-2xl border border-stroke bg-surface px-7 font-body text-[15px] font-bold tracking-wide text-paper"
              >
                JOIN A GAME AS PLAYER
              </button>
            </div>
            <p className="mt-5 font-body text-[13px] text-text-3">
              Naira only · Escrow-held pools · Instant payouts · No install for players
            </p>
          </div>

          {/* live mock card */}
          <div className="w-[380px] shrink-0">
            <div className="flex flex-col gap-3.5 rounded-3xl border border-stroke bg-surface-2 p-6">
              <div className="font-body text-[11px] font-bold tracking-[0.14em] text-text-3">
                LIVE NOW · ROOM AB12
              </div>
              <div className="font-display text-[40px] font-extrabold leading-none text-naira">
                ₦50,000
              </div>
              <div className="font-body text-[13px] text-text-3">prize pool · 3 winners</div>
              <div className="flex items-center gap-2">
                <div className="font-body text-[14px] font-semibold text-paper">
                  Which city is the Centre of Excellence?
                </div>
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full border-2 border-gold bg-surface">
                  <span className="font-display text-[15px] font-extrabold text-gold">18</span>
                </div>
              </div>
              <div className="flex items-center gap-1.5">
                {["🐙", "🦊", "🐼", "😎", "🤖"].map((e) => (
                  <div
                    key={e}
                    className="flex h-[30px] w-[30px] items-center justify-center rounded-full bg-surface text-[15px]"
                  >
                    {e}
                  </div>
                ))}
              </div>
              <div className="font-body text-[12px] text-text-3">
                8 players in the room · top 3 cash out
              </div>
            </div>
          </div>
        </section>

        {/* FOR INSTRUCTORS */}
        <section id="instructors" className="flex flex-col gap-6 bg-surface px-16 py-14">
          <div className="font-body text-[12px] font-bold tracking-[0.18em] text-gold">
            FOR INSTRUCTORS
          </div>
          <h2 className="font-display text-[30px] font-extrabold text-paper">Why hosts choose Kweeks</h2>
          <div className="grid grid-cols-4 gap-4">
            {[
              {
                t: "Fund a pool in seconds",
                s: "Add naira from your wallet the moment you're ready. Money lands instantly and is held in escrow only while a room is live.",
              },
              {
                t: "Rooms your players love",
                s: "Players join by a 4-letter room code. No account, no install, no sign-up friction — they just show up and play.",
              },
              {
                t: "You decide the winners",
                s: "Pick 1, 2, 3 or 5 winners at setup. Podium payouts are split from your pool automatically the moment the quiz ends.",
              },
              {
                t: "A clean ledger, always",
                s: "Every funding, every pool, every payout lands in your wallet history. Know exactly where each naira went.",
              },
            ].map((f) => (
              <div
                key={f.t}
                className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface-2 p-5"
              >
                <div className="font-body text-[16px] font-bold text-paper">{f.t}</div>
                <div className="font-body text-[13px] leading-relaxed text-text-2">{f.s}</div>
              </div>
            ))}
          </div>
        </section>

        {/* FOR PLAYERS */}
        <section id="players" className="flex flex-col gap-6 bg-bg px-16 py-14">
          <div className="font-body text-[12px] font-bold tracking-[0.18em] text-gold">
            FOR PLAYERS
          </div>
          <h2 className="font-display text-[30px] font-extrabold text-paper">Why players come back</h2>
          <div className="grid grid-cols-4 gap-4">
            {[
              {
                t: "Real money, real fast",
                s: "A short live quiz, a naira pool, and winners paid in minutes. Answer right and fast, climb the board, and cash out on your own screen.",
              },
              {
                t: "Join with a code",
                s: "Your host shares a 4-letter code. Tap it, pick an avatar, and you're in the room — no account, no download, no waiting.",
              },
              {
                t: "Every question is live",
                s: "Everyone in the room sees each question at the same second. Speed counts. No answer key ever leaks to the room.",
              },
              {
                t: "Your win is yours",
                s: "Place in the top spots and a claim code lands on your screen. Only you can redeem it — straight to your payout.",
              },
            ].map((f) => (
              <div
                key={f.t}
                className="flex flex-col gap-2.5 rounded-2xl border border-stroke bg-surface-2 p-5"
              >
                <div className="font-body text-[16px] font-bold text-paper">{f.t}</div>
                <div className="font-body text-[13px] leading-relaxed text-text-2">{f.s}</div>
              </div>
            ))}
          </div>
        </section>

        {/* HOW IT WORKS */}
        <section id="how" className="flex flex-col gap-6 bg-surface px-16 py-14">
          <div className="font-body text-[12px] font-bold tracking-[0.18em] text-gold">
            HOW IT WORKS
          </div>
          <h2 className="font-display text-[30px] font-extrabold text-paper">
            Three steps to a money quiz
          </h2>
          <div className="grid grid-cols-3 gap-4">
            {[
              {
                n: "1",
                t: "Create & fund",
                s: "Author your questions, set a naira pool from your wallet, and choose how many winners take it home.",
              },
              {
                n: "2",
                t: "Open the room",
                s: "Players join by your room code. Run each question live — auto-paced or on your tap — while everyone watches the board.",
              },
              {
                n: "3",
                t: "Pay the podium",
                s: "The quiz ends, the top spots lock, and each winner redeems from their own screen. Paid straight from the pool.",
              },
            ].map((st) => (
              <div
                key={st.n}
                className="flex flex-col gap-3 rounded-2xl border border-stroke bg-surface-2 p-5"
              >
                <div className="flex items-center gap-2.5">
                  <span className="font-display text-[20px] font-extrabold text-gold">{st.n}</span>
                  <span className="font-body text-[16px] font-bold text-paper">{st.t}</span>
                </div>
                <div className="font-body text-[13px] leading-relaxed text-text-2">{st.s}</div>
              </div>
            ))}
          </div>
        </section>

        {/* SECURITY */}
        <section id="security" className="grid grid-cols-4 gap-4 bg-bg px-16 py-12">
          {[
            {
              t: "Naira only, always",
              s: "Every wallet, pool and payout is NGN. No currency games, no surprises.",
            },
            {
              t: "Escrow while live",
              s: "Pool funds sit in escrow from the moment a room opens until winners are paid.",
            },
            {
              t: "Winner-only claims",
              s: "Each winner gets a private claim code on their own screen. Only they can redeem it.",
            },
            {
              t: "Full history",
              s: "Fundings, pools and payouts are all recorded in your wallet history.",
            },
          ].map((it) => (
            <div key={it.t} className="flex flex-col gap-2 rounded-2xl border border-stroke bg-surface-2 p-[18px]">
              <div className="font-body text-[14px] font-bold text-paper">{it.t}</div>
              <div className="font-body text-[12.5px] leading-snug text-text-2">{it.s}</div>
            </div>
          ))}
        </section>

        {/* CTA */}
        <section className="flex flex-col items-center gap-5 bg-surface px-16 py-14">
          <h2 className="font-display text-[34px] font-extrabold text-paper">
            Your next quiz could be a money event.
          </h2>
          <p className="font-body text-[15px] text-text-2">
            Host from your laptop or join from a phone. Either way, the money moves.
          </p>
          <div className="flex items-center gap-3">
            <button
              onClick={() => nav("/instructor/signup")}
              className="flex h-14 items-center justify-center rounded-2xl bg-gold px-[30px] font-body text-[15px] font-extrabold tracking-wide text-gold-ink hover:opacity-90"
            >
              START HOSTING
            </button>
            <button
              onClick={() => nav("/join")}
              className="flex h-14 items-center justify-center rounded-2xl border border-stroke bg-surface-2 px-[30px] font-body text-[15px] font-bold tracking-wide text-paper"
            >
              JOIN A GAME
            </button>
          </div>
          <p className="font-body text-[12.5px] text-text-3">
            Free to start · no card required · naira-only wallets
          </p>
        </section>

        {/* FOOTER */}
        <footer className="w-full border-t border-stroke bg-bg">
          <div className="flex h-[52px] w-full items-center justify-between px-6">
            <div className="flex items-center gap-2">
              <span className="font-display text-[16px] font-extrabold text-paper">KWEEKS</span>
              <span className="font-body text-[12px] text-text-3">© 2026 Kweeks · live money quiz</span>
            </div>
            <span className="font-body text-[12px] text-text-3">Support · Terms · Privacy</span>
            <span className="font-body text-[12px] text-text-3">Naira only · Bank-grade rails</span>
          </div>
        </footer>
      </div>
    </div>
  );
}
