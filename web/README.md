# Kweeks Web (design → code)

Pixel-faithful React + Tailwind translation of the pen.dev frames in
`designs/design.pen`. Each pen frame is one route/page.

## Stack
- Vite 5 + React 18 + TypeScript 5 (strict)
- Tailwind CSS 3.4 (design tokens in `tailwind.config.js` + `src/index.css`)
- React Router 6 (one route per frame)
- TanStack Query (configured), Zustand (installed), Zod (installed) — ready for
  wiring the Go API in `internal/adapters/httpapi`
- Fonts: Bricolage Grotesque (display) + Work Sans (body), loaded in
  `index.html`

## Run
```bash
npm install
npm run dev        # http://localhost:5173
npm run build      # typecheck + production build to dist/
npm run typecheck
```

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

## Design tokens
Palette + type live as CSS variables in `src/index.css` and map 1:1 to the pen
`SetVariables` (bg #171220, surface #211B2C, gold #FFC53D, naira #2ED08F, …).
Money = naira green only; scores/actions use paper/gold; red = live/wrong.
Use Tailwind utility classes exclusively (no inline styles).
