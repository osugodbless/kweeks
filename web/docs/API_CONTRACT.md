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
Body: `{"name":"Adeola Peters","email":"host@kweeks.ng","password":"secret"}`
Creates the instructor AND issues a NGN wallet immediately.
Returns: `{"token":"...","instructor":{"id","name","email","avatar"},"wallet":{"id":"kweeks_ngn_8f2c1a","balanceNaira":"150000"}}`

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
Card/transfer settle through the money rail when configured; failures return 502.
Returns updated `{"wallet":{"id","balanceNaira"}}`.

### POST /api/wallet/provision  (Bearer)
Provisions a real BMONI user + CNGN smart wallet for the instructor on the
configured rail (NGN persona). Idempotent. Requires a configured BMONI persona;
otherwise returns 400. Returns `{"wallet":{"id","balanceNaira","bmoniUserId",
"bmoniWalletId","bmoniWalletAddress"}}`. Wallet provisioning also runs
automatically at signup when `BMONI_PROVISION_ON_SIGNUP=true`.

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
Body `{email}` → `{id,amountNaira,state}`. Claim code returned in `claimCode` for the winner's own session.

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
- Wallet is a real instructor-scoped ledger. `credit` funding is the instant
  platform credit used by the demo (BMONI sandbox test-credit); card/transfer
  settle via the money rail when BMONI is configured and otherwise return a
  clear 502 so the UI can show a recoverable error.
- Multi-user auth: instructors + bcrypt-hashed passwords + bearer sessions,
  stored in Postgres (memory store mirrors for tests).
- Room codes: unambiguous A-Z/2-9, no look-alikes.
- Public room state NEVER ships `correctIndex`.
