# Hyperframes Composition Brief: Kweeks full product tour

## Objective
A 63s polished launch/demo film touring every Kweeks screen and the full money
lifecycle, for a hackathon panel and a projected venue demo.

## Output
- Composition: `demo-output/composition/`
- Render: `demo-output/demo.mp4`
- Format: landscape 1920x1080, 30fps
- Duration: 63.0s, 14 screens

## Source material
- Project: `/home/griittyy/GOSSU/kweeks` (React SPA in `web/src`)
- Screens recreated from: LandingPage, InstructorSignup, InstructorWallet,
  InstructorFundWallet, InstructorQuizBuilder, InstructorLiveRoom,
  InstructorHistory, PlayerJoin, PlayerLobby, PlayerQuestion, PlayerStandings,
  PlayerPodium, ClaimPage; tokens from `web/src/index.css` + `DESIGN.md`.
- Copy must appear verbatim where used (pool ₦50,000, winner share ₦25,000, code
  AB12, claim code KWA-8F2D, "Answer fast. Take the pool.", "Real money out.",
  "Correct! +500 pts · 412ms", "Wallet credited with ₦50,000 — it is spendable
  right away.", "You banked ₦25,000").

## Creative direction
- Tone: `polished` — elegant restraint, mixed case, generous spacing, slow
  dissolves, one accent idea per screen, no jokes, no hype.
- Shared stage: warm paper + dots + slow blurred glows; screens are the product's
  own UI, framed by the app's top bar (instructor nav or player room bar).
- Hook: "Quiz night. The prize is real money." + the pool counting to ₦50,000.
- Outro: money rules row + wordmark + "Real money out."

## Screens / storyboard
1 Hook 0–5.5 · 2 Host signup 5.5–10 · 3 Wallet dashboard 10–14.5 · 4 Fund wallet
14.5–18.5 · 5 Quiz builder 18.5–23.5 · 6 Live room 23.5–29 · 7 History 29–33 ·
8 Player join 33–37.5 · 9 Lobby 37.5–41 · 10 Question 41–46.5 · 11 Standings
46.5–50 · 12 Podium 50–54.5 · 13 Claim & payout 54.5–59 · 14 Close 59–63.
Transitions: 0.5s crossfades over the shared background.

## Audio
- Music: `assets/music/kweeks-demo-bed.mp3` (vol-12, 63.6s, vol ~0.30, fade out).
- Audio-reactive: `assets/audio-data.js` — subtle bass/RMS breathing on money
  glows and ink cards.
- SFX: sparse motion-matched accents from `assets/sfx/` (cards on reveals,
  announcement hit on the podium payout, one logo hit on the close).
- Beat sync: use the bundled vol-12 preset (first 25s) and `hyperframes beats`
  for the remainder; lock 3–5 major reveals within ±0.15s of strong beats.

## Hyperframes instructions
Standalone composition. Local fonts (@font-face), local GSAP, local audio — no
network at render. One paused timeline registered on
`window.__timelines["kweeks-demo"]`. Shared background outside clips; per-scene
content wrappers crossfaded. Run `hyperframes check` before render.
