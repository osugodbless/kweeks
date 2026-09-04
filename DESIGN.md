# DESIGN — Kweeks "Money Match Arena"

Written from the built world in `designs/design.pen` (pen.dev). One design
system, 15 top-level frames: 1 marketing landing page + 13 app pages + the
Design System token row folded into the landing nav.

## Direction contract

- **THESIS.** A live money quiz should feel like a broadcast scoreboard in a
  dim venue, not a Kahoot clone and not a generic fintech app. The prize is the
  protagonist; speed + correctness are the sport.
- **OWN-WORLD.** Deep arena-night plum ink (`#171220`) ground; surfaces lift
  one step (`#211B2C`) to two (`#2B2438`); warm paper text (`#FBF7EE`).
  Naira green `#2ED08F` is locked to money and correct answers only. Gold
  `#FFC53D` is action/win/you. Red `#FF5A67` = wrong/live. Violet `#8B7CFF`
  only for the instructor identity initial.
- **TYPE.** Display = Bricolage Grotesque (heavy 700/800) for numbers, money,
  question text. Body/UI = Work Sans. Display numerals carry the tension; body
  stays quiet.
- **SHAPING.** Radius 12–18 for cards/buttons, 8–14 for small chips; pill only
  for live/status chips. No gradients, no glass, no drop shadows except
  elevation borders. Money figures always render as `<font-display 800>` in
  naira green.

## System tokens (SetVariables)

Colors: bg, surface, surface2, stroke, paper, text2, text3, gold, gold-deep,
gold-ink (text on gold), naira, naira-deep, red, violet. Type: font-display,
font-body. Space scale 4/8/16/24/32. Radius 16/24.

## Signature

The **pool as a live naira figure** is the running motif: on the player join
screen as the entry stakes, on the instructor projector as the stage header
(`₦50,000 POOL`), and at the podium as the personal `You banked ₦X`. The
countdown ring/bar is the secondary motif (18s, gold fill) on question + stage
frames.

## Pages

1. **Landing Page** (replaces the old token-frame slot) — the marketing frame
   that tells instructors and players what Kweeks is and why to choose it.
   Structure: top nav (KWEEKS + `live money quiz`, anchor pills For
   instructors / For players / How it works / Security, gold START HOSTING +
   LOG IN) → hero ("Run a quiz. Put real money on it." + live-room mock card:
   `LIVE NOW · ROOM AB12`, ₦50,000 pool, 18s ring, avatar stack) → **For
   instructors** (fund a pool in seconds / rooms players love / you decide the
   winners / a clean ledger) → **For players** (real money real fast / join
   with a code / every question is live / your win is yours) → **How it works**
   (3 steps: create & fund → open the room → pay the podium) → **Security**
   (naira only / escrow while live / winner-only claims / full history) →
   closing CTA band + footer. Design-system tokens (wordmark, gold/dark chips,
   money chip) were the prior occupant and are now described in the tokens
   section above rather than as a standalone frame.
2. **Player · Join** (desktop 1180) — top bar (wordmark, Room AB12 pill, LIVE)
   + two-column body: left pool promo (`PRIZE POOL`, ₦50,000, "HOW IT PLAYS"
   card) · right join card (avatar grid 2×6 with 🐙 gold-selected, nickname +
   email fields, gold JOIN THE GAME). Same content as the mobile version,
   re-composed for a wide screen.
3. **Player · Lobby** (desktop 1180) — top bar (wordmark, Room AB12, WAITING) +
   two-column: left hero (🐙, "You're in, Zainab", ₦50,000 on the line, prize
   bar "1st place takes ₦25,000") · right IN THE ROOM roster chips + waiting
   card with gold pulse.
4. **Player · Question** (desktop 1180) — top bar (wordmark, "Question 3 of
   10", 1,250 pts) + centered stage: 18s ring + gold bar, large question, 2×2
   option tiles (one gold-selected). Footer pinned.
5. **Player · Standings** (desktop 1180) — top bar (wordmark, Live standings,
   LIVE) + heading ("After question 4 · 8 players") + wide leaderboard card:
   pts colhint, 5 rows (Ada/Tobi/Zainab/Chidi/Uche), gold YOU row (bg gold,
   gold-ink), you-pill at the base.
6. **Player · Podium** (desktop 1180) — top bar (wordmark, Final standings,
   GAME OVER) + two-column: left confetti hero (Unsplash, "2nd place · 8
   players", ₦15,000, "You banked ₦15,000") · right Winners list
   (Ada🦊/YOU🐙/Tobi🐼 with amounts), claim-code card (KWEEKS-7F3A-9Z, COPY),
   gold REDEEM ₦15,000 NOW, and 3-step how-it-lands (Claim locked / Invite
   sent / Paid).
7. **Instructor · Quiz Builder** — nav w/ wallet balance, quiz title card with
   Unsplash Lagos deck cover, pool slider, winner count segmented (3), pacing
   segmented (manual), question editor with correct answer marked green,
   question strip, open-room CTA.
8. **Instructor · Live Room** — projector preview (Q3, ring, options, live
   answer bar), join card (realistic Unsplash QR photo + kweeks.ng/r/AB12 +
   avatars), manual control card (next → declare winners).
9. **Instructor · Sign up** — split brand panel (KWEEKS wordmark, "Fund the
   pool. Run the room. Pay the winners." lead, money chip) + 420-wide auth card
   (full name / email / password, gold CREATE ACCOUNT, log-in link).
10. **Instructor · Log in** — same split, email + password, gold LOG IN, sign-up
    link; brand panel money chip reads `₦150,000 ready to host`.
11. **Instructor · Wallet** (login landing / dashboard) — nav (KWEEKS / Wallet,
    WALLET ₦150,000, avatar AP); greeting `Welcome back, Adeola` + gold
    **CREATE A QUIZ** button; stats row (QUIZZES HOSTED 3, PLAYERS HOSTED 48,
    WINNERS PAID 9, AVAILABLE ₦150,000); left balance card (`ASSIGNED WALLET ·
    NGN`, wallet ID `kweeks_ngn_8f2c1a`, FUND WALLET, method chips Card /
    Transfer / Instant top-up); right "Your quizzes" list (live quiz row with
    OPEN ROOM →, start-another-quiz CTA, history link). This is where an
    instructor lands after log in.
12. **Instructor · Fund wallet** — nav + amount input (₦50,000), quick picks
    ₦1k/₦5k/₦50k/₦100k (₦50k selected), funding methods (Wallet credit
    selected, Debit card, Bank transfer), gold `FUND ₦50,000`, production note
    ("Wallet credits post instantly to your available balance.").
13. **Instructor · History** — the target of the wallet dashboard's
    `View history →`. Nav (History active) + history table (Activity / Type /
    Pool / Status columns) with 5 reconciled rows: wallet credit +₦200,000,
    hosted quiz (live AB12), payout −₦50,000, two ended quizzes. Plus footer.
14. **Instructor · History (empty)** — the no-history alternative: same nav
    (History active) + centered empty state (🗂️ icon, "No history yet", CTA
    CREATE A QUIZ) + footer.

### Navigation & footers (system-wide)
- All six instructor desktop frames share the top nav: KWEEKS logo + link pills
  (Dashboard / Create quiz / History) with the active page highlighted (gold
  text on `$surface2` pill), plus right-side WALLET chip and avatar AP. Live
  Room keeps its ● LIVE status chip beside the logo.
- Every frame (all 14) carries a minimal footer: 1px stroke rule + `© 2026
  Kweeks` · `Live money quiz · NGN` · `Support · Terms · Privacy`. Player
  frames show a shortened footer (`live money quiz`, no legal links). Auth
  frames (Sign up / Log in) keep the footer; the Design System token frame gets
  one too.

### Auth & wallet flow (canonical demo story)
Sign up → log in → the Wallet dashboard is the landing screen → a wallet was
issued at signup (`kweeks_ngn_8f2c1a`) → fund it (instant wallet credit) →
funds land in the wallet → the wallet funds the quiz pool when the room opens
and pays winners at the podium. One instructor persona throughout: avatar AP,
wallet ₦150,000, hosted quiz "Naija General Knowledge" (pool ₦50,000, 3
winners). All auth/wallet copy reads production-ready — no sandbox or demo
wording on any screen (infra is BMONI sandbox under the hood, but the UI never
says so).

## Contrast (WCAG, computed)

All body/placeholder ≥ 4.5:1; paper/bg 17.2, text3/bg 5.2, gold-ink/gold 10.9,
naira/surface 8.4. Green only money/correct; gold only action/you/win.

## Imagery & identity (fixed this pass)
All three photos are real Unsplash (verified HTTP + orientation, no placeholders):
- Podium hero = crowd confetti celebration (landscape 1.5 → `fill` crop into
  the 350×150 banner).
- Quiz Builder deck cover = Lagos Victoria Island (landscape).
- Live Room join card = standalone Qr-style photo (not a hollow white box).

Persona avatars stay emojis (🐙 Zainab) everywhere — the committed zero-image
identity; the single photographed win/cover moments are where photography earns
its place. Standings avatar mapping fixed so Zainab=🐙 and the podium winner
avatars (Ada🦊, Zainab🐙, Tobi🐼) match the leaderboard.

## Known swaps before dev
- QR is a photoreal Unsplash stand-in — replace with a generated, scannable QR
  encoding `kweeks.ng/r/AB12` before go-live.
- Status-bar clock is a static `9:42` mock — real app shows the live time.
- Podium/cover Unsplash photos are stock — swap for real venue/win imagery at
  the exhibition if available.
