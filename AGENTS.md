## gstack
Use /browse from gstack for all web browsing. Never use mcp__claude-in-chrome__* tools.
Available skills: /office-hours, /plan-ceo-review, /plan-eng-review, /plan-design-review,
/design-consultation, /design-shotgun, /design-html, /review, /ship, /land-and-deploy,
/canary, /benchmark, /browse, /open-gstack-browser, /qa, /qa-only, /design-review, /scrape,
/setup-browser-cookies, /setup-deploy, /setup-gbrain, /sync-gbrain, /retro, /investigate,
/document-release, /document-generate, /codex, /cso, /autoplan, /pair-agent, /careful, /freeze,
/guard, /unfreeze, /gstack-upgrade, /learn.

## Skill routing

When the user's request matches an available skill, invoke it via the Skill tool. When in doubt, invoke the skill.

Key routing rules:
- Product ideas/brainstorming → invoke /office-hours
- Strategy/scope → invoke /plan-ceo-review
- Architecture → invoke /plan-eng-review
- Design system/plan review → invoke /design-consultation or /plan-design-review
- Full review pipeline → invoke /autoplan
- Bugs/errors → invoke /investigate
- QA/testing site behavior → invoke /qa or /qa-only
- Code review/diff check → invoke /review
- Visual polish → invoke /design-review
- Ship/deploy/PR → invoke /ship or /land-and-deploy
- Save progress → invoke /context-save
- Resume context → invoke /context-restore
- Author a backlog-ready spec/issue → invoke /spec

# kweeks — repo essentials

Live money quiz: a Go backend serves the React SPA, `/api`, and websockets from
one binary (one origin, one port, no separate frontend server). Product brief:
`PRODUCT.md`. Design system: `DESIGN.md` is the source of truth for `web/`.
There is no README and no CI (`.github/` absent).

## Layout & boundaries
- Backend (Go 1.26, net/http, pgx): repo root. Entrypoint `cmd/kweeks-server/`.
  Hexagonal: `internal/domain` (pure) → `internal/ports` (interfaces) →
  `internal/app` (services, manual DI in `main.go`) → `internal/adapters`
  (httpapi, bmoni, store, ws, scheduler, mailer, bus, clock).
- Frontend (React 19, Vite 8, Tailwind 3.4): `web/`. Dev: Vite on :5173
  proxies `/api` (incl. ws) to the Go server on :8080. Prod has no proxy; the
  Go binary serves `web/dist` and `/api` from the same origin.

## Commands
- Backend unit tests (no DB): `go test -race ./...`
- Backend integration tests need Postgres: `make db-up` (docker) then
  `make test-integration` (DATABASE_URL is passed inline).
- `make lint` requires golangci-lint, which is NOT installed. Use
  `go vet ./...` (also part of `make fmt`).
- Frontend (from `web/`, or `npm --prefix web …`): `npm run typecheck`,
  `npm run lint`, `npm test` (vitest; single file: `npx vitest run
  src/lib/player.test.ts`), `npm run build` (= typecheck + vite build).
- Pre-commit check: `go test -race ./... && go vet ./...`, then in `web/`
  typecheck + lint + test + build.

## Config / env (`internal/config/config.go`)
- Everything from env. `DATABASE_URL` is REQUIRED or the server exits.
- `.env` is loaded from the repo root if present (godotenv); real env vars win
  over the file. Copy `.env.example` → `.env` (`make dev` does this). `.env`
  is gitignored.
- `BMONI_BASE_URL` defaults to sandbox `https://embedded-dev.bmoni.com`;
  production is `https://embedded.bmoni.com`. `BMONI_OWNER_KEY` is a hex
  secp256k1 key signing owner-proof + proposal digests (EIP-191). BMONI docs
  are queryable via the `bmoni-embedded-docs` MCP server (in `opencode.json`).
- SMTP is optional and best-effort: email is only a redemption-recovery
  artifact, never the critical path. With no SMTP configured the mailer logs
  the full redemption payload instead. `KWEEKS_PUBLIC_URL` builds `/claim`
  links in those emails.

## Backend gotchas
- Migrations are embedded (`go:embed migrations/*.sql`) and auto-applied on
  startup in filename order, tracked in `schema_migrations`. New schema →
  add `000N_*.sql`; restart to apply. No migration tool.
- ws hub, scheduler, and game pacing are in-process (no Redis/distributed
  bus). Run ONE server instance only; adding instances would split rooms.
- New feature path: domain type → ports interface → app service (+
  `internal/app/*_test.go`) → httpapi route → migration → store method.
- Routing: `net/http.ServeMux` patterns; `/api/*` and `/healthz` win, and the
  `"/"` catch-all SPA handler serves `web/dist`, falling back to
  `index.html` for client routes (deep links work only when dist exists).

## Frontend gotchas
- `BASE = import.meta.env.VITE_API_BASE ?? "/api"` (`web/src/lib/api.ts`).
  Same-origin in prod; Vite proxies in dev.
- `Modal` (`web/src/components/ui/modal.tsx`) renders via `createPortal` to
  `document.body` at `z-[100]`. Never mount `fixed`/overlay elements inside
  ancestors with `backdrop-filter`/`transform`: fixed children get trapped in
  that stacking context and paint behind the page (the header bug).
- Money is always naira green; coral `#ff4d5f` is the primary action; see
  DESIGN.md.

## Deploy (Render)
- `render.yaml` + `scripts/render-build.sh`. Build order matters: frontend
  (`npm --prefix web ci && npm --prefix web run build` → `web/dist`) BEFORE
  `go build -o app`. `web/dist` is gitignored; Render builds it, the Go
  binary serves it. If dist is missing every route 404s.
- Env vars live in the Render blueprint/dashboard, not `.env`.