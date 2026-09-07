# Kweeks Web (design → code)

TypeScript translation of the Kweeks product: a live money quiz. Instructors
fund a naira prize pool, players join a room by 4-letter code on their phones,
everyone answers the same question at the same second, and the fastest correct
players split the pool. Frontend routes all requests to the Go backend
(`internal/adapters/httpapi`) per `docs/API_CONTRACT.md` — **no Supabase**.

## Stack
- Vite 8 + React 19 + TypeScript 5.9 (strict: `noUnusedLocals`, `noUnusedParameters`)
- Tailwind CSS 3.4 — design tokens in `tailwind.config.js` + `src/index.css`
  (CSS variables as RGB triplets so alpha utilities like `bg-ink/5` work)
- React Router 7 (one route per frame, see Routes below)
- TanStack Query (server state), Zustand (auth + player session), Zod (installed)
- Fonts: Fredoka (display) + Karla (body), loaded in `index.html`

## Run
```bash
npm install
npm run dev        # http://localhost:5173  (proxies /api → :8080, ws: true)
npm run build      # typecheck + production build to dist/
npm run typecheck
npm run test       # vitest — money/split/avatar logic
npm run lint
```

The backend serves the built SPA from `web/dist` on the same origin when
`KWEEKS_WEB_ROOT=web/dist` is set — no separate reverse proxy needed.

## Routes (frame → page)
| Frame | Route | File |
|---|---|---|
| Landing Page | `/` | `src/pages/LandingPage.tsx` |
| Player · Join | `/join` | `src/pages/player/PlayerJoin.tsx` |
| Player · Lobby | `/lobby` | `src/pages/player/PlayerLobby.tsx` |
| Player · Question | `/question` | `src/pages/player/PlayerQuestion.tsx` |
| Player · Standings | `/standings` | `src/pages/player/PlayerStandings.tsx` |
| Player · Podium | `/podium` | `src/pages/player/PlayerPodium.tsx` |
| Instructor · Sign up | `/instructor/signup` | `src/pages/instructor/InstructorSignup.tsx` |
| Instructor · Log in | `/instructor/login` | `src/pages/instructor/InstructorLogin.tsx` |
| Instructor · Wallet (dashboard) | `/instructor/dashboard` | `src/pages/instructor/InstructorWallet.tsx` |
| Instructor · Fund wallet | `/instructor/fund` | `src/pages/instructor/InstructorFundWallet.tsx` |
| Instructor · Quiz Builder | `/instructor/quiz-builder` | `src/pages/instructor/InstructorQuizBuilder.tsx` |
| Instructor · Live Room | `/instructor/live-room` | `src/pages/instructor/InstructorLiveRoom.tsx` |
| Instructor · History | `/instructor/history` | `src/pages/instructor/InstructorHistory.tsx` |
| Instructor · History (empty) | `/instructor/history-empty` | `src/pages/instructor/InstructorHistoryEmpty.tsx` |

Instructor routes behind `/instructor/*` are guarded: no token → redirect to
`/instructor/login`. The quiz builder accepts `?id=…` to edit an existing quiz.

## Design tokens
Palette + type live as CSS variables in `src/index.css` and map 1:1 to
`tailwind.config.js`. Use Tailwind utility classes exclusively (no inline
styles except tokens/motion via CSS vars).

- Warm paper stage: `canvas #faf3e7`, cards `cream #fffaf0`, ink text
  `#24163f`, secondary text `soft #6b5f8d`.
- Semantic color rules (enforced across every page):
  - **Money is mint only** (`mint #16c47f` / `mint-dark` on light surfaces) —
    pools, wallet balances, payouts, claim amounts.
  - **Correct answers + secure trust = mint**; **wrong/live = coral**.
  - **Coral `#ff4d5f`** = primary action + brand dot + live.
  - **Violet `#6c4cf1`** = host identity, selection, secondary action.
  - **Gold `#f6a91b`** = win/prize/highlight + countdown fill.
  - **Sky `#38a8ff`** = info / streak.
- Typography: Fredoka for display (headlines, money figures, room codes),
  Karla for body/UI. Buttons use `press-3d` (hard shadow, raise on hover).
- Motion: `rise`/`pop` entrances, `float` decorations, `pulse-ring` waiting
  state, gold countdown bar. All gated behind `prefers-reduced-motion`.

## Backend wiring
- `src/lib/api.ts` — typed REST client + payload types, a 1:1 mirror of
  `docs/API_CONTRACT.md`. Every network call in the app goes through it.
- `src/lib/hooks.ts` — TanStack Query hooks per resource (auth, wallet, fund,
  dashboard, history, quizzes, rooms, lookup, standings, join, answer, room
  control, redeem) plus a reconnecting room WebSocket hook. Player/instructor
  pages refetch the matching REST resource on socket events.
- `src/lib/auth.ts` / `src/lib/player.ts` — Zustand stores for the instructor
  session (token in `localStorage`, `GET /api/auth/me` on boot) and the player
  room session (persisted so a mid-game refresh keeps the flow).

## Player flow
Join by code → lobby (auto-advances when the host starts) → question (lock-in
answer, countdown from server timing) → standings (auto-advances on the next
question) → podium (winners + per-winner claim code + redeem). Pacing is
server-owned (`manual`/`auto`); the frontend only reacts to the public room
state.
