# Hyperframes Composition Brief: Kweeks

## Objective
Create a short launch-style brag video for Kweeks, a live money quiz where an
instructor funds a real naira prize pool and players win real money from their
phones.

## Output
- Composition directory: `brag-output/composition/`
- Rendered video: `brag-output/brag.mp4`
- Format: landscape — 1920x1080
- Duration: 20.2 seconds

## Source Material
- Project root: `/home/griittyy/GOSSU/kweeks`
- Primary files read: `web/src/pages/LandingPage.tsx`,
  `web/src/pages/player/PlayerJoin.tsx`, `PlayerLobby.tsx`, `PlayerQuestion.tsx`,
  `PlayerStandings.tsx`, `PlayerPodium.tsx`,
  `web/src/pages/instructor/InstructorLiveRoom.tsx`,
  `web/src/index.css`, `DESIGN.md`, `PRODUCT.md`, `web/README.md`
- Product name: Kweeks
- Tagline / strongest claim: "Answer fast. Take the pool." (landing hero);
  money rule: "the fastest correct minds take it home."
- Key UI or visual moment to recreate: (1) the deep-plum live-room card with the
  `AB12` join code and `₦50,000` pool chip; (2) the player question screen with
  the gold countdown bar and 2×2 option tiles, mint "Correct!" receipt; (3) the
  podium "You banked ₦25,000" + per-winner claim code + gold REDEEM.
- Copy that must appear verbatim:
  - "Answer fast." / "Take the pool."
  - "₦50,000" (pool), "₦25,000" (winner share)
  - "AB12" (room code), "Question 3 of 10"
  - "Correct! +500 pts in 412ms"
  - "Real money out."
  - "Kweeks"

## Creative Direction
- Tone preset: `default`
- Creative direction: warm, high-energy Lagos quiz night — a game show with real
  money on the table.
- Interpretation: playful and confident, never corporate. Springy, fast motion
  with short hard cuts, but every line holds long enough to read. The money
  figure is the anchor and always mint-green. Jokes stay understated — the fun is
  that this cheerful cartoon quiz actually moves naira. No memes, no lens flares.
- Angle: "Quiz night, but the prize is real money." It looks like a warm quiz
  show and behaves like a ledger: the pool is escrowed naira, the winners redeem
  it from their own screens.
- Hook: `₦50,000` slams onto the warm paper stage under "PRIZE POOL · ON THE LINE",
  then "Answer fast. / Take the pool."
- Outro / punchline: Kweeks wordmark + "Real money out."
- Avoid:
  - Generic SaaS language ("streamline", "empower", "workflow")
  - Abstract filler visuals, waveform/equalizer graphics, generic particle fields
  - Redesigning Kweeks away from its warm-paper / plum / coral identity

## Visual Identity
- Background: `#faf3e7` canvas (warm paper)
- Cards: `#fffaf0` cream
- Text: `#24163f` ink (deep plum); secondary `#6b5f8d` soft
- Accent: coral `#ff4d5f` (primary/live), violet `#6c4cf1` (host/selection)
- Money / correct: mint `#16c47f`, `mint-dark #0b9a62` on light
- Win / prize: gold `#f6a91b`
- Display font: Fredoka (local `assets/fonts/fredoka-latin*.woff2`, variable
  400–700) — headlines, money, codes, big numbers
- Body font: Karla (local `assets/fonts/karla-latin*.woff2`) — labels, captions
- Visual references from the project: dots texture (`radial-gradient`, 18px),
  soft blurred violet/coral glows, `card-3d` hard drop shadow, `press-3d` button
  shadow, `text-pop` (coral→violet→sky) and `text-shine` (gold) wordmarks,
  pastel lucide-avatar chips (zero image assets), the `bg-dots-light` texture on
  the ink surfaces.

## Storyboard
Use the storyboard in `brag-output/brag-plan.md` as the creative contract.

Scene summary:
1. Hook: the pool is real — 3.70s — `₦50,000` slams in under "PRIZE POOL · ON THE
   LINE", then "Answer fast." / "Take the pool." (text-pop gradient).
2. Join by code — 2.64s — deep-plum live-room card, `A B 1 2` tiles pop in one by
   one, "One code. Any room."
3. The live question — 5.26s — cream question card, "Question 3 of 10", gold
   countdown draining, 2×2 options; a simulated cursor taps the correct answer →
   mint check + "Correct! +500 pts in 412ms"; a coral wrong tile for contrast;
   "Lock in. First correct tap counts."
4. Speed decides the podium — 3.16s — standings board with pastel avatar chips,
   "You" row gold; rows arrive one by one; "Speed decides the podium."
5. Podium: real money out — 4.54s — ink podium hero, "1st place", "You banked
   ₦25,000", claim code + COPY + gold REDEEM, then Kweeks wordmark + "Real money
   out."

## Audio
- Audio role: warm upbeat bed with motion-matched accents
- Audio arc: pool slam opens with weight → code tiles tick in → the correct-tap
  lands on a strong cue with a bright payoff → rows accelerate → podium payout is
  the loudest beat → short music fade under the wordmark hit.
- Music: `assets/music/happy-beats-business-moves-vol-9-by-ende-dot-app.mp3`
  (114.84 BPM; 113.64s source, use first 19.3s; volume 0.34, fade out over the
  final ~0.5s)
- Music cue guidance: bundled preset —
  `assets/music/cues/happy-beats-business-moves-vol-9-by-ende-dot-app.music-cues.json`
  (copy of the skill preset; see `brag-plan.md`). Strong cues in window: 6.34,
  10.54, 12.65, 3.70, 8.44. Beat grid ~0.52s. Planned locks: 6.34 (question
  reveal), 10.54 (correct-tap payoff), 12.65 (standings). Sequential code tiles
  snap to beats 4.23/4.75/5.28/5.80; standings rows to 12.12/12.65/13.18/13.70.
- Audio-reactive treatment: subtle; use music RMS/bass to let the pool card's
  mint glow and the ink live-room card's depth breathe. No waveform/equalizer
  visuals. If extraction is unavailable (no helper), document and skip — do not
  block the render.
- Audio-coupled moments:
  - Scene 1 — pool figure slam (impact), headline pop (soft drop)
  - Scene 2 — 4 code tiles, one card sound each (accent first + last)
  - Scene 3 — simulated cursor move + tap (UI click), mint "Correct!" payoff
    (announcement bell), coral wrong tile for contrast
  - Scene 4 — leaderboard rows one by one; gold "You" row gets the strongest
    small accent
  - Scene 5 — payout chips/coin on the amount, soft confirmation on the claim
    code, logo hit over the music fade
- SFX selection guidance: motion-matched; card sounds for card-like reveals,
  short announcement cue for the correct payoff, UI click for the simulated tap,
  chips/coin for money, bell for the logo. Lower high-frequency-risk files for
  repeated/polished moments. Reference: `sfx-analysis.md` (skill library).
- Exact SFX choice: Hyperframes chooses filenames, timestamps, density, and
  volume based on the implemented animation (files already staged under
  `assets/sfx/`).
- Audio files: music copied to `assets/music/`; SFX staged in `assets/sfx/`.

## Hyperframes Instructions
Load the composition-building Hyperframes domain skills — `hyperframes-core`
(composition contract + `data-*` timing), `hyperframes-animation` (motion),
`hyperframes-creative` (design spec, beats, audio-reactive), `hyperframes-keyframes`
(seek-safe keyframes), and `hyperframes-cli` (lint/check/render). /brag is its own
workflow: do not enter the `hyperframes` entry-point intent interview and do not
route into its generic promo / launch-video workflow. Prefer native Hyperframes
conventions over anything in `/brag`.

Requirements:
- Show at least one real UI, copy, or visual element from the source project
  (this recreates the code entry, question card, standings, and podium).
- Keep all text readable in the final render.
- Keep the video within 15–25 seconds.
- Include the planned music/SFX layer.
- Treat `/brag` audio notes as guidance, not a fixed cue sheet. Choose SFX after
  the visual animation exists.
- Treat music cue metadata as optional timing hints.
- Use local assets for audio and fonts; no render-time network fetches.
- Run `npx hyperframes check` before render — brag's single gate.
