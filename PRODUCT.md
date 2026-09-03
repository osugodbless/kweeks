# PRODUCT — Kweeks (frontend design brief)

Derived from `godbless-master-design-20260902-160919.md` (APPROVED) for the
pen.dev UI mockups in `designs/`.

## One line
Kweeks is a live money quiz: an instructor funds a naira prize pool, players
join a room on their phones by code, everyone answers the same question at the
same second, and the top winners take home real money redeemed from their own
screen.

## Audience & scene
Venue demo (BMONI Embedded API hackathon, Friday 4 Sept 2026, Lagos). Players
are 5–8 phones in a room, plus a projected instructor laptop. Crowd energy and
one real money transfer landing are the show. Zero installs.

## Surfaces (both in this design system)
- Instructor (desktop 1180): author quiz + fund pool → run a live room with
  projector preview + join code + control.
- Player (mobile 390): join → lobby → live question → standings → podium +
  redeem.

## Money rules that bind the UI
- Pool is a real wallet balance (BMONI sandbox). Shown as ₦, naira-green.
- Podium: top N winners (instructor-chosen count), pool split across them.
- Redemption is winner-driven: per-winner claim_code shown only on the winner's
  screen, redeem → invited → paid. Green is reserved for money/correct;
  never used for decoration.

## Instructor account & wallet
- Sign up → the account is assigned a NGN wallet immediately (one wallet per
  instructor, created at signup). Log in returns to that same wallet.
- Funding the wallet = BMONI sandbox test-credit (and/or card/transfer rails);
  funds land instantly and are spendable on pools. Naira only, always.
- Demo ledger reconciles: wallet created → funded +₦200,000 → pool escrow
  −₦50,000 → ₦150,000 available (as shown on the Wallet + nav chips).

## Persona used across the mockups
Zainab, avatar 🐙, joins room AB12, pool ₦50,000, 3 winners (25k/15k/10k),
manual pacing (venue default), answers the Lagos "Centre of Excellence" deck.

## Content
- 24 emoji avatars (faces + animals) rendered as large text — zero image
  assets for identity.
- Deck: Lagos/Naija general knowledge, 4 options each.
