# DESIGN — Kweeks "Money Match Arena"

Written from the built world in `designs/design.pen` (pen.dev). One design
system, 7 page frames + 1 system frame.

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

1. **Design System** — wordmark, gold/dark chips, money chip.
2. **Player · Join** — room code pill + LIVE tag, pool hero, 2×6 emoji avatar
   grid (selected = gold), nickname + email, gold CTA.
3. **Player · Lobby** — room code, waiting state card, roster chips, 1st-place
   prize bar.
4. **Player · Question** — Q of N, score, 18s ring + gold bar, question,
   4 option tiles, one selected gold.
5. **Player · Standings** — LIVE leaderboard, gold "you" row + YOU tag, pts
   shown paper (not money-green), 1st gold.
6. **Player · Podium** — Unsplash confetti win hero, `You banked ₦15,000`,
   winners list (25k/15k/10k) with you highlighted, claim code chip, gold
   redeem CTA, 3-step how-it-lands.
7. **Instructor · Quiz Builder** — nav w/ wallet balance, quiz title card with
   Unsplash Lagos deck cover, pool slider, winner count segmented (3), pacing
   segmented (manual), question editor with correct answer marked green,
   question strip, open-room CTA.
8. **Instructor · Live Room** — projector preview (Q3, ring, options, live
   answer bar), join card (realistic Unsplash QR photo + kweeks.ng/r/AB12 +
   avatars), manual control card (next → declare winners).

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
