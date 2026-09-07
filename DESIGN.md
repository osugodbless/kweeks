# DESIGN — Kweeks (live money quiz)

Source of truth for the shipped frontend in `web/`. One design system, 14
routes, TypeScript + Tailwind. Light theme, warm paper stage, playful
game-show energy — with money handled like a ledger, never decoration.

## Direction contract

- **THESIS.** A live money quiz should feel like a high-energy quiz night in a
  warm room, not a fintech dashboard and not a soulless Kahoot clone. The prize
  pool is the protagonist; speed + correctness are the sport; the host is the
  showrunner.
- **OWN-WORLD.** Warm paper ground `#faf3e7` (canvas) with cream cards
  `#fffaf0` lifting off it. Ink `#24163f` (deep plum) is the text/authority
  color. Coral `#ff4d5f` is the primary action + brand dot; violet `#6c4cf1`
  is the host identity + selection. Money is **always** naira green; gold is
  reserved for wins and highlights.
- **TYPE.** Display = Fredoka (rounded, chunky, confident) for headlines, money
  figures, room codes, and big numbers. Body/UI = Karla (quirky grotesque,
  quietly readable). Display numerals carry the tension; body stays quiet.
- **SHAPING.** Radius 16–32 for cards, 8–16 for small chips; pill only for
  chips and primary CTAs. Hard "press-3d" buttons (flat shadow, raise on hover,
  press down on click). No gradients except the hero `text-shine`/`text-pop`
  wordmarks; no glass; dots texture (`bg-dots`) for atmosphere; soft blurred
  color glows behind content for depth.
- **MOTION.** One orchestrated entrance per page (`rise`/`pop` with stagger),
  `float` for decorations, `pulse-ring` for the lobby's waiting card, a gold
  countdown bar on live questions. All disabled under `prefers-reduced-motion`.

## System tokens

Colors (RGB-triplet CSS vars in `src/index.css`, mapped 1:1 in
`tailwind.config.js` so alpha utilities work):

| Token | Hex | Role |
|---|---|---|
| canvas | `#faf3e7` | page ground |
| cream | `#fffaf0` | cards, elevated surfaces |
| ink | `#24163f` | primary text / logo tile |
| soft | `#6b5f8d` | secondary text |
| coral | `#ff4d5f` | primary action / live / wrong |
| coral-dark | `#e23a4b` | press depth |
| violet | `#6c4cf1` | host identity / selection |
| violet-dark | `#5133d6` | press depth |
| mint | `#16c47f` | money / correct / secure |
| mint-dark | `#0b9a62` | money text on light surfaces |
| gold | `#f6a91b` | win / prize / highlight / countdown |
| gold-dark | `#da8c05` | press depth / text-on-gold hover |
| sky | `#38a8ff` | info / streak |
| sky-dark | `#1785de` | text-on-sky |

Type: `font-display` = Fredoka, `font-body` = Karla (Google Fonts in
`index.html`). Space scale 4/8/16/24/32/48; radius 16 (cards), 24 (hero
cards), pill (chips/CTAs).

### Semantic rules (hard, enforced everywhere)
- **Money is mint only.** Every naira figure renders through the `Money`
  component (display typeface, grouped thousands). Mint appears nowhere except
  money/correct/secure. Scores and points are gold/violet, never green.
- **Coral = primary + live + wrong.** Primary CTAs, the wordmark dot, the LIVE
  chip, incorrect answers.
- **Violet = host + selection.** Instructor identity, auth CTAs, selected
  options, focus rings.
- **Gold = win/you/prize.** Podium highlight, "You" row, prize chips, the
  countdown fill, the lobby waiting pulse.

## Signature

The **pool as a live naira figure** is the running motif: on the player join
screen as the entry stakes (`₦50,000`), on the lobby as "On the line", on the
podium as "You banked ₦X", and on the instructor wallet as the available
balance. The **4-letter room code** is the entry artifact everywhere (join
input, lobby pill, instructor live-room). The **gold countdown bar** is the
live-question motif.

## Pages

1. **Landing Page** `/` — public nav (wordmark, scroll anchors For players /
   For instructors / How it works / Security, Play + Host links, Host login) →
   hero ("Answer fast. Take the pool." + live-room mock card: `LIVE · Question
   3`, room code AB12, `₦50,000` pool chip, leaderboard with lucide avatars,
    gold countdown) → **How it works** (Create & fund / Open the room / Pay the
   podium) → **For players** (real money fast: no app, join by code, live
   questions, winner-only claims) → **For instructors** (fund a pool in
   seconds / rooms players love / you decide the winners / a clean ledger) →
   **Security** (naira only / escrow while live / winner-only claims / full
   history) on an ink band → footer.
2. **Player · Join** `/join` — player top bar (wordmark, room code pill) +
   two-column: left pool promo (`PRIZE POOL`, live `₦`, "how it plays" steps,
   payout-address note) · right join card (room code entry → room found →
   nickname + email + avatar grid, coral JOIN THE GAME). Errors are friendly
   and inline. Join persists the player session.
3. **Player · Lobby** `/lobby` — waiting room: "You're in, {name}", "On the
   line" pool card with prize-split chips (mirrors the backend weighted split),
   copy-code button; right "In the room" roster chips + gold waiting card with
   pulse-ring. Auto-advances to question on live.
4. **Player · Question** `/question` — "Question N of M" + Standings link,
   gold countdown bar (server timing), 2×2 option tiles (tap to lock), correct
   = mint / wrong = coral feedback, then auto-advance to standings.
5. **Player · Standings** `/standings` — "After question N · X players", "You
  're {rank}" gold pill, wide leaderboard (correct count violet, speed), YOUR
   row highlighted gold, auto-advances when the host opens the next question.
6. **Player · Podium** `/podium` — "Final standings · Game over" top bar +
   two-column: left ink hero (confetti, place chip, "You banked ₦X" or "Better
   luck next time") · right winners list (rank, avatar, correct count, share)
   + winner-only claim card (per-winner claim code, COPY, gold REDEEM, 3-step
   how-it-lands). Redeem → claim code locked, green confirmation.
7. **Instructor · Sign up** `/instructor/signup` — split layout: brand panel
   (wordmark, "Fund the pool. Run the room. Pay the winners.", wallet-ready
   chip) + cream card (full name / email / password, coral CREATE ACCOUNT,
   sign-in link). Wallet is issued at signup.
8. **Instructor · Log in** `/instructor/login` — same split, email + password,
   violet LOG IN, sign-up link.
9. **Instructor · Wallet (dashboard)** `/instructor/dashboard` — instructor
   nav (Dashboard / Create quiz / History + wallet chip + avatar) → "Welcome
   back, {name}" + gold CREATE A QUIZ; stats row (Quizzes hosted / Players
   hosted / Winners paid / Available); left wallet card (`ASSIGNED WALLET ·
   NGN`, balance, wallet id, FUND WALLET, method chips, live-room banner);
   right "Your quizzes" list (open-room / edit actions) + history link.
10. **Instructor · Fund wallet** `/instructor/fund` — amount field (₦), quick
    picks ₦1k/₦5k/₦50k/₦100k, funding methods (Wallet credit / Debit card /
    Bank transfer), mint FUND button, instant-credit note. 502 from card/rail
    surfaces as a recoverable error pointing back to wallet credit.
11. **Instructor · Quiz Builder** `/instructor/quiz-builder` — title, prize
    pool slider (₦1k–₦200k), winner count (1/3/5), pacing (manual/auto),
    default question time, per-question editor (prompt, 4 options with mint
    correct ring, per-question time), add/delete, SAVE & OPEN ROOM. Accepts
    `?id=…` to edit an existing quiz. Insufficient balance links to funding.
12. **Instructor · Live Room** `/instructor/live-room?room=…` — projector
    preview (question, gold countdown, option tiles, realtime-on chip),
    join-card with the big room code + copy + player roster, control card
    (START / NEXT / DECLARE WINNERS), live top-3 standings. Podium state shows
    a room-ended panel linking to history.
13. **Instructor · History** `/instructor/history` — unified ledger table
    (Activity / Type / Amount / Status) with funding, hosted quizzes and paid
    winners, newest first; money inflow green, outflow neutral.
14. **Instructor · History (empty)** `/instructor/history-empty` — same nav +
    centered empty state ("No history yet", CREATE A QUIZ) + footer. The
    History page renders this when the ledger is empty.

### Navigation & footers (system-wide)
- Instructor pages share the top nav (Dashboard / Create quiz / History, active
  pill violet, wallet chip, avatar AP, sign out) with a mobile bottom row.
- Player pages share a player top bar (wordmark + room code pill + status chip
  WAITING/LIVE/GAME OVER).
- Footers: full variant (`kweeks. · Live money quiz · NGN` + legal links) and a
  shortened player variant.

### Player flow (auto-advance)
Join → lobby (waiting) → question (lock-in, countdown) → standings (between
questions) → … → podium + redeem. The frontend reacts to the server-owned room
state; pacing is `manual` (host advances) or `auto` (scheduler). The player
session persists across refresh.

### Instructor flow
Sign up → wallet issued → fund wallet → build quiz → open room (pool escrowed
from the wallet) → run live room → declare podium → winners redeem with
per-winner claim codes → all activity lands in History.

## Contrast (computed)
Body/placeholder text ≥ 4.5:1 on canvas/cream; ink/canvas ≈ 10.9:1;
soft/canvas ≈ 4.7:1; mint-dark/canvas ≈ 3.3:1 (display figures only, always
bold ≥ 18px → passes the large-text bar); white on coral ≈ 3.3:1 (large bold
display CTA text); ink on gold ≈ 9:1; mint on ink ≈ 6.7:1 (dark chips).

## Imagery & identity
Identity is zero-image: 16 lucide avatars (id + pastel bg + fg) picked at join,
rendered from the id string the backend stores. Instructor identity = violet
initial avatar. No photography; atmosphere comes from dots texture + color
glows. Icons are lucide throughout — consistent 24px grid, no emoji-as-icon.

## Implementation notes
- REST: `src/lib/api.ts` mirrors `web/docs/API_CONTRACT.md`; all requests go to
  the Go backend (Vite dev proxy `:5173 → :8080`, ws enabled; production via
  `KWEEKS_WEB_ROOT` same-origin SPA). No Supabase anywhere.
- Server-authoritative: money math (`splitPodium` mirror), pacing, and
  standings are backend-owned; the UI only renders public state and refetches
  on socket events.
- Gate: `npm run typecheck`, `npm run lint`, `npm test`, `npm run build`.
