# Kweeks API Contract (v1)

Single source of truth for frontend ↔ backend. Base: `/api`. All request/response
bodies are JSON. Auth uses `Authorization: Bearer <token>` on instructor routes.
Player routes are code-scoped and carry no auth (a room join token is returned).

## Error shape
Non-2xx returns `{"error": "..."}`. 401 = not authed, 403 = forbidden,
404 = not found, 400/409 = domain/business rule.

---

## Auth (instructor, multi-user)

### POST /api/auth/signup
Body: `{"firstName":"Adeola","lastName":"Peters","email":"host@kweeks.ng","phone":"+2348012345678","password":"secret"}`
Creates the instructor AND issues a NGN wallet immediately. `firstName`/`lastName`
are the host's real names as typed on the form; they are passed to BMONI
verbatim (never re-split from a single name field). `phone` is optional but
strongly recommended; it is the host's own number, normalized server-side to
E.164 (any of `+2348012345678`, `2348012345678`, `08012345678`, `8012345678`
are accepted) and becomes their distinct BMONI user identity, so two hosts
never share a wallet. Without a phone a deterministic unique phone is derived
at provisioning. An unparseable phone returns 400 `{"error":"enter a valid
phone number (E.164, e.g. +2348012345678)"}` — never the 401 "invalid email or
password".
When the money rail is configured the host's BMONI user is created
automatically (create-user). If that call fails, signup still succeeds and
returns a valid token; the account stays `unprovisioned` and the host can
retry with `POST /api/wallet/create-user`.
Returns: `{"token":"...","instructor":{"id","firstName","lastName","name","email","phone","avatar"},"wallet":{"id":"kweeks_ngn_8f2c1a","balanceNaira":"150000"}}`

### POST /api/auth/login
Body: `{"email","password"}` → same shape as signup.

### GET /api/auth/me  (Bearer)
Returns the current instructor + wallet: same shape (minus token).

---

## Wallet (instructor, Bearer)

### GET /api/wallet
Returns: `{"wallet":{"id","balanceNaira"},"transactions":[{id,kind,amountNaira,note,createdAt}]}`
`kind` ∈ fund|pool|payout|credit.

### POST /api/wallet/fund
Body: `{"amountNaira":"50000","method":"card"|"transfer"|"credit"}`
Credits the wallet. `credit` = instant platform credit (no external rail).
Returns updated `{"wallet":{"id","balanceNaira"}}`.

### GET /api/wallet/setup  (Bearer)
Wallet-setup wizard status. Returns `{"stage":"unprovisioned"|"kyc"|"wallet"|"rail"|"ready",
"bmoniUserId","kycSubmitted","bmoniWalletId","bmoniWalletAddress","railActive",
"railConfigured","depositAccount":{"accountNumber","bankName"}}`. `railConfigured`
is false when the BMONI keys are absent (no provisioning possible). `ready`
includes the NGN deposit account the host funds by bank transfer.

### POST /api/wallet/create-user  (Bearer)
Strict-flow step 1 (idempotent). Creates (or recovers) the host's BMONI user
from their signup identity and records `bmoniUserId`. Auto-run at signup when
the rail is configured; exposed so a host whose user creation failed at signup
can retry from the wizard. Returns `{"wallet":{...,"bmoniUserId"}}`.

### POST /api/wallet/kyc  (Bearer)
Strict-flow step 2. Body: `{"firstName","lastName","dateOfBirth","gender","bvn",
"street","city","state","postalCode"}` → submits the host's KYC profile.
All identity values are user-supplied (the sandbox resolves the test persona
values, e.g. Bunch Dillon / 95888168924, when entered here).

### POST /api/wallet/kyc/documents/{kind}  (Bearer)
Multipart `file` upload (JPEG/PNG). `kind` = `identification` |
`proof-of-address` | `biometric`. Optional for NGN activation.

### POST /api/wallet/create  (Bearer)
Strict-flow steps 3-4: owner-proof challenge → create-managed smart wallet.
Returns `{"wallet":{...,"bmoniWalletId","bmoniWalletAddress"}}`. Idempotent.

### POST /api/wallet/activate-rail  (Bearer)
Strict-flow step 5: `POST /onboarding/start-nigeria` with `{"bvn"}`. Marks the
wallet ready to fund and returns the NGN deposit account via `/wallet/setup`.

---

## Dashboard + history (instructor, Bearer)

### GET /api/instructor/dashboard
Returns stats for the wallet landing page:
`{"quizzesHosted":3,"playersHosted":48,"winnersPaid":9,"availableNaira":"150000","quizzes":[{id,title,poolNaira,winnerCount,pacing,questionCount,roomCode?,state?}]}`

### GET /api/instructor/history
Returns unified ledger the History page renders:
`[{id,at,type:"fund"|"quiz"|"payout"|"room",title,amountNaira?,meta?,state?}]`
Empty array when nothing happened yet (History empty state).

---

## Quizzes (instructor, Bearer)

### GET /api/quizzes   (exists)
List without answers. Each: `{id,title,poolNaira,winnerCount,pacing,questionCount}`.

### GET /api/quizzes/{id}
Full quiz for the builder/editor, answers included (instructor only).
Returns `{id,title,poolNaira,winnerCount,pacing,defaultDurationMs,questions:[{id,prompt,options[],correctIndex,durationMs}]}`.

### POST /api/quizzes   (exists — widened)
Body: `{"title","poolNaira","winnerCount","pacing","defaultDurationMs","questions":[{id,prompt,options[],correctIndex,durationMs}]}`
Returns `{"id"}`.

### PUT /api/quizzes/{id}
Update title/fields/questions (editor save). Returns `{"id"}`.

---

## Rooms (instructor open + player join by code)

### POST /api/rooms  (exists)
Body `{"quizId":"..."}` → returns `{"id":"room-hex","code":"AB12"}`.
Room gets a short human **code** generated server-side (unambiguous alphabet).

### GET /api/rooms/{id}  (public state, no correct answers)
Returns the room as a player sees it:
```
{id,code,quizId,title,poolNaira,winnerCount,pacing,state:"lobby"|"live"|"podium"|"ended",
 questionCount,currentIndex,participantCount,participants:[{id,nickname,avatar}],
 currentQuestion: null | {id,index,prompt,options[],startedAt,durationMs,remainingMs},
 winners: null | [{participantId,nickname,avatar,correctCount,totalLatencyMs,rank}],
 host:{name,email}}
```

### GET /api/lookup/{code}
Same public state, looked up by 4-letter code (player entry path).

### POST /api/rooms/{id}/join  (exists)
Body `{email,nickname,avatar}` → returns `{id,roomId,email,nickname,avatar,joinedAt}`.

### POST /api/rooms/{id}/answer  (exists)
Body `{participantId,questionId,optionIndex}` → `{id,correct,score,latencyMs}`.

### GET /api/rooms/{id}/standings  (exists)
`[{participantId,nickname,avatar,correctCount,totalLatencyMs,joinedAt}]` sorted best-first.

### POST /api/rooms/{id}/start | /next | /podium  (exists)
Start (lobby→live), next question (manual), finalize podium → winners `[...]`.

### POST /api/rooms/{id}/redeem  (exists)
Body `{email}` → `{id,amountNaira,state,claimCode}`. Creates the exactly-once
claim and emails the claim code + a `/claim` URL immediately as the recovery
artifact, so the code can never be lost.

### POST /api/claims/resolve  (new, public)
Body `{claimCode,email}` → `{claim:{id,amountNaira,state}, banks:[{code,name}]}`.
Validates the claim code + join email and returns the Nigerian bank list for
the payout form. Wrong code/email → `403`.

### POST /api/claims/payout  (new, public)
Body `{claimCode,email,accountNumber,bankCode,bankName}` → `{claim:{id,amountNaira,state,payoutRef}}`.
Verifies the winner's Nigerian bank account, registers it on the host's rail,
and offramps the prize from the host wallet. State advances
`created → bank_submitted → paying → paid` (or `failed`).

### GET /api/wallet/deposit  (new)
`{accountNumber,bankName}` — the host's NGN virtual bank account that a bank
transfer to it funds the wallet (issued during Nigeria onboarding).

---

## Realtime (WebSocket)

### GET /api/rooms/{id}/ws
Server pushes JSON frames: `{"type":"question"|"standings"|"podium"|"joined"|"ended","data":{...}}`.
- `question.data` = public room state (same shape as GET /api/rooms/{id}).
- `standings.data` = standings array.
- `podium.data` = winners array.
- `joined.data` = participant.
Frontend refetches the matching REST resource on each event for the authoritative payload.

---

## Notes / decisions
- All BMONI identity input comes from the user or the program, never from env.
  Signup creates the host's BMONI user from their real name/email/phone; the
  wallet-setup wizard collects KYC (name, DOB, BVN, address) and drives the
  strict flow (create-user → KYC → owner-proof → create-managed → start-nigeria).
  Each instructor provisions a DISTINCT BMONI user + wallet, so no two hosts
  share a money identity. In the sandbox, entering the test persona values in
  the KYC step (Bunch Dillon, 95888168924) resolves verification.
  `credit` funding is the instant local-ledger credit used by the demo;
  production funding happens by bank-transferring to the host's NGN virtual
  bank account (`GET /api/wallet/deposit`), which BMONI credits to the wallet.
- Winners are paid by the host wallet via the BMONI offramp (verify → register
  → offramp → approve → sign). The claim code is emailed immediately and is the
  only capability that authorizes the payout.
- Multi-user auth: instructors + bcrypt-hashed passwords + bearer sessions,
  stored in Postgres (memory store mirrors for tests).
- Room codes: unambiguous A-Z/2-9, no look-alikes.
- Public room state NEVER ships `correctIndex`.
