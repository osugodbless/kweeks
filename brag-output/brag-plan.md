# Brag Plan: Kweeks

## What is this app?
Kweeks is a live money quiz: an instructor funds a real naira prize pool, players
join a room on their phones with a 4-letter code, everyone answers the same
question at the same second, and the fastest correct players split the pot — paid
out of their own screen.

## The angle
"Quiz night, but the prize is real money." It looks like a warm, playful game
show; it behaves like a ledger. The bit that earns the reaction is that the
playful cartoon quiz actually moves naira — the pool is not points, it is
`₦50,000` held in escrow and redeemed by the winners. The video plays the
game-show energy straight, and the money figure is always mint-green, never
decoration. Specific to Kweeks: the 4-letter room code, the gold countdown bar,
the "You banked ₦X" podium, and per-winner claim codes.

## Hook (first 2-3 seconds)
A big `₦50,000` slams onto the warm paper stage under "ON THE LINE" — the pool is
real money — then the product's own hero line lands: **"Answer fast. Take the
pool."** No preamble. The number is the hook.

## Key moments (the middle)
- **The 4-letter room code** — `AB12` tiles pop in one by one on the live-room
  card while players join by code. The entry artifact of the whole product.
- **The live question** — recreate the player question card: gold countdown bar
  draining from full, 2×2 option tiles, a cursor taps the correct option, it
  flips mint with a check and a latency readout ("+500 pts in 412ms").
- **The podium payout** — "You banked ₦25,000" with the winner's claim code and
  the gold REDEEM button. The money leaves the screen as real naira.

## Outro / punchline
Kweeks wordmark on the ink band: **"Real money out."** — the counterweight to the
hook's "Take the pool."

## User flow worth showing
The player happy path, as the centerpiece:
1. **Entry** — join by 4-letter code (`AB12`) and land in the room.
2. **Key action** — the live question: countdown drains, tap the answer, correct
   flips mint with points + latency.
3. **Result** — podium: rank, "You banked ₦25,000", claim code, REDEEM.

## Tone
- Preset: `default`
- Creative direction: warm, high-energy Lagos quiz night — game-show energy with
  real money on the table.
- Interpretation: playful and confident, not corporate and not unhinged. Motion is
  springy and fast but every line gets time to read. Money moments get a beat of
  weight. Restraint on jokes — the fun comes from how seriously the playful UI
  moves real cash.

## Format: landscape — 1920x1080
## Duration: 20.2s (storyboard scenes sum to 19.3s; the outro holds to 20.2s)

## Visual identity (from the project)
- Background: `#faf3e7` canvas (warm paper)
- Cards: `#fffaf0` cream
- Accent: coral `#ff4d5f` (primary/live), violet `#6c4cf1` (host/selection)
- Money/correct: mint `#16c47f` (`mint-dark #0b9a62` on light)
- Win/prize: gold `#f6a91b`
- Text: ink `#24163f`; secondary soft `#6b5f8d`
- Display font: Fredoka (headlines, money, codes, big numbers)
- Body font: Karla
- Strongest visual elements: the deep-plum live-room card from the landing hero,
  the `AB12` code tiles, the gold countdown bar, avatars rendered as pastel lucide
  icon chips (zero image assets), the dots texture and soft blurred color glows.

## Share copy (draft)
Quiz night, except the prize is real naira. Kweeks: fund a pool, drop a 4-letter
code, and the fastest correct players get paid out from their phones.

## Audio direction
- Role: warm upbeat bed with motion-matched accents
- Music: `happy-beats-business-moves-vol-9-by-ende-dot-app.mp3` (114.84 BPM,
  mid-energy, upbeat — best fit for `default`; its strong cues are distributed
  across the planning window, see cue guidance)
- Music treatment: start at 0s at 0.34 volume, hold under the edit, short fade
  out over the final ~0.4s as the wordmark lands.
- Music cue guidance: bundled preset read —
  `assets/music/cues/happy-beats-business-moves-vol-9-by-ende-dot-app.music-cues.json`.
  Strong cues in window: 6.34s, 10.54s, 12.65s, 3.70s, 8.44s. Beat grid ~0.52s
  apart (3.18, 3.70, 4.23, 4.75, 5.28, 5.80, 6.34, …). Use 3 strong locks:
  6.34 (question scene reveal), 10.54 (correct-tap payoff), 12.65 (standings).
  Sequential code tiles snap to beats 4.23/4.75/5.28/5.80; leaderboard rows snap
  to 12.12/12.65/13.18/13.70.
- Audio-reactive treatment: subtle; use music RMS/bass to let the pool card's mint
  glow and the ink live-room card's depth breathe. No waveform/equalizer visuals.
- SFX posture: moderate, motion-matched, professional restraint.
- Audio-coupled moments: pool slam, code tiles arriving one by one, simulated
  answer tap, mint "Correct!" payoff, leaderboard rows, redeem/payout, wordmark.
- Restraint rule: no SFX on every beat and nothing over the question prompt while
  a viewer is reading it; the money reveal is the loudest moment, the wordmark
  second.

## Storyboard

### Scene 1 — Hook: the pool is real — 3.70s
Warm paper stage with the dots texture and a soft violet/coral glow. A big
`₦50,000` in Fredoka mint slams down under a small "PRIZE POOL · ON THE LINE"
label, with a mint chip glow. Beneath it the hero line lands in two steps:
"Answer fast." then "Take the pool." in the coral→violet→sky `text-pop`
gradient. Product material: the landing hero headline and the money component.
Sequential/interaction: none — the number slams, the two headline lines pop in.
Audio intent: establish warmth, then land the hook with weight.
Audio-coupled idea: pool counter/slam SFX with the figure; a second accent when
"Take the pool." pops.
Music: upbeat bed, full from 0s.
Transition mood: clean/hard → Scene 2 (hard cut on the 3.70 strong cue).

### Scene 2 — Join by code — 2.64s
The ink (deep-plum) live-room card slides in. "JOIN WITH ROOM CODE" label, then
the four cream `A B 1 2` tiles pop in one at a time with a spring, each with a
soft inner press shadow. Players join on their phones. Copy line under the card:
"One code. Any room." Product material: the landing hero live-room mock + the
instructor live-room join card.
Sequential/interaction: yes — the 4 code tiles arrive one by one.
Audio intent: light, rhythmic, satisfying — a code being assembled.
Audio-coupled idea: one soft card/drop SFX per tile, accenting the first and last.
Music: same bed.
Transition mood: clean → Scene 3.

### Scene 3 — The live question — 5.26s
Recreate the player question screen on a cream card: "Question 3 of 10", the
prompt "Which city is Nigeria's Centre of Excellence?", a gold countdown bar
draining left-to-right, and 2×2 option tiles (violet ring on hover). A simulated
cursor moves to the correct tile and taps it at ~10.54s; the tile flips mint with
a check, a coral wrong tile is shown on the alternate option for contrast, and a
mint receipt chip pops: "Correct! +500 pts in 412ms". Copy line: "Lock in. First
correct tap counts." Product material: `PlayerQuestion.tsx` layout, the gold
countdown, mint/coral answer states.
Sequential/interaction: yes — simulated cursor tap on the correct option.
Audio intent: tension (ticking countdown) resolving into a clean, bright payoff.
Audio-coupled idea: soft tick under the drain; a click on the tap; an
announcement-style hit + mint payoff SFX as "Correct!" lands on the 10.54 strong
cue.
Music: same bed.
Transition mood: hard cut → Scene 4.

### Scene 4 — Speed decides the podium — 3.16s
The standings board: "After question 3 · 6 players", wide leaderboard rows with
pastel avatar chips, correct counts in violet, and the "You" row highlighted gold.
Rows slide in one by one. Copy line: "Speed decides the podium." Product
material: `PlayerStandings` leaderboard with avatar chips.
Sequential/interaction: yes — leaderboard rows arrive one by one, the "You" row
last with a gold highlight.
Audio intent: quick, competitive, building toward the outro.
Audio-coupled idea: a short drop SFX per row; gold highlight gets the strongest
small accent.
Music: same bed.
Transition mood: soft/anticipatory → Scene 5.

### Scene 5 — Podium: real money out — 5.44s
The ink podium hero: gold crown, "1st place" chip, and "You banked ₦25,000" in
Fredoka (money mint-on-ink), confetti flecks floating. Beside/below it a gold
claim card: the per-winner claim code, a COPY chip, and the gold REDEEM button.
Then the Kweeks wordmark (ink tile + coral dot) with the line "Real money out."
Product material: `PlayerPodium.tsx` (the "You banked ₦X", claim code, REDEEM).
Sequential/interaction: yes — payout amount counts up, then claim code + REDEEM
appear, then the wordmark.
Audio intent: warm payoff, then a clean final logo land.
Audio-coupled idea: chips/coin payoff on the amount, a soft confirmation on the
claim code, and the logo hit over the fading music.
Music: same bed, short fade-out under the wordmark.
Transition mood: resolve → end card.

**Music mood for this video:** upbeat, warm, game-show
**Audio summary:** A warm upbeat bed carries a fast, readable edit; the pool slam
opens it, the code ticks in, the correct-tap lands on a strong cue, and the
podium payout is the loudest beat before the music fades under the wordmark.
