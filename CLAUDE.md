# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Role: Active implementer, incremental delivery

Claude builds the project directly. Work is done feature by feature on
short-lived branches; nothing lands on `main` until the slice is working.

### Rules
- Build incrementally — one well-scoped slice at a time, per user instructions.
- No large batch changes. If a task feels too big, split it and confirm scope first.
- Follow the architecture and constraints documented below; don't introduce
  patterns not already decided.
- Write tests alongside implementation, not after.
- Run tests, builds, and linters before reporting a task done.
- **Write code as a human developer would:** terse, direct, idiomatic. No over-engineering, no elaborate abstractions beyond what the task needs, no comment blocks explaining obvious code, practical variable names. Avoid the "AI-generated" tell of excessive structure and ceremony.
- **Git is the user's responsibility.** Claude writes code; the user reviews the diff, commits, and merges. Claude may suggest a commit message but never runs git commands unless explicitly asked.
- **Merges to `main` are fast-forward only.** Never squash, never `--no-ff` — the
  individual small commits are the history the user wants to keep. If `main` has
  moved, rebase the branch onto it first, then `git merge --ff-only`.

## Project state

SpacedChess is in early implementation. The repository has design documentation,
a throwaway Postgres dev environment that seeds itself on every start, and a Go
API wired to Postgres via a pgx pool. Auth is in place — `/auth/register`,
`/auth/login`, `/auth/logout`, `/auth/me`, and session-cookie middleware — but
no product endpoints yet beyond `/health`. Both services run in Compose; `db`'s
healthcheck gates `api`'s startup, so the API never races Postgres coming up.

Sessions are opaque `crypto/rand` tokens stored in the `sessions` table; the
cookie is `session` (HttpOnly, SameSite=Lax, `Secure` from `COOKIE_SECURE`, 7
days). Passwords are bcrypt. Registration is open — anyone can create an
account; a duplicate username is a 409. Error responses are `{"error": "..."}`.

Schema so far: `users` and `sessions` are real; the `cards` table in
`init/001_init.sql` is still the original placeholder and gets redesigned in
slice 4.

## Build order

Each slice is one short-lived feature branch merged to `main` when working and tested.
Mark slices done as they land.

- [x] **Slice 1 — Auth schema** (`feat/db-schema`): `users` + `sessions` tables in `init/001_init.sql`. Remaining tables land in the slices that need them.
- [x] **Slice 2 — Go project structure** (`feat/go-skeleton`): `go.mod`, `cmd/api/main.go`, `internal/` layout, env-var config, DB connection pool, health endpoint wired properly.
- [x] **Slice 3 — Auth** (`feat/auth`): register, login, logout endpoints; session-cookie middleware; password hashing.
- [ ] **Slice 4 — Card CRUD** (`feat/card-crud`): `cards` table (real schema, replacing the placeholder); full CRUD behind auth middleware; list filterable by type/tag/set.
- [ ] **Slice 5 — Tags & Sets** (`feat/tags-sets`): `tags`, `card_tags`, `sets`, `set_cards` tables; tag and set CRUD; join management.
- [ ] **Slice 6 — Review engine** (`feat/review-engine`): `review_state` and `review_log` tables; `GET /review/due`, `POST /review/:id/grade`; SM-2 logic.
- [ ] **Slice 7 — React SPA scaffold** (`feat/frontend-scaffold`): Vite + TypeScript, router, auth flow (login page, protected routes), served by Go binary.
- [ ] **Slice 8 — Card creation UI** (`feat/frontend-cards`): FEN paste + type selection + type-specific fields; chessground position preview.
- [ ] **Slice 9 — Review UI** (`feat/frontend-review`): per-card-type review flows with chessground board, grading, sequential reveal for Blunder.

## Commands

```sh
docker compose up -d --build      # Postgres + API (5432, 8080); waits for db healthy
docker compose exec db psql -U user -d spacedchess
curl -s localhost:8080/health     # {"status":"ok"} — 503 if Postgres is down

# or run the API on the host, against Compose's Postgres — faster inner loop:
docker compose up -d db
export DATABASE_URL=postgres://user:password@localhost:5432/spacedchess
go run ./cmd/api                  # API on :8080, override with PORT
                                  # COOKIE_SECURE=true behind TLS; off by default

# the seeded user is nairwolf / password
curl -si -c /tmp/c -X POST localhost:8080/auth/login \
  -d '{"username":"nairwolf","password":"password"}'
curl -s -b /tmp/c localhost:8080/auth/me

go build ./... && go vet ./... && staticcheck ./... && golangci-lint run
go test ./...                     # store tests skip unless DATABASE_URL is set
```

Code changes to the API need `docker compose up -d --build` to take effect —
there is no live reload. `Dockerfile` is a multi-stage build (`golang:1.26` →
`distroless/static-debian12`); `db`'s healthcheck runs `pg_isready -h
127.0.0.1` specifically, since a bare `pg_isready` can pass against the
entrypoint's temporary socket-only server before Postgres is really listening.

The database is **throwaway** — there is no named volume. `./init` is mounted
into the container's entrypoint dir and its files run in filename order on every
fresh start: `001_init.sql` (schema) then `002_seed.sql` (dev data). After
editing either:

```sh
docker compose down -v && docker compose up -d
```

The `-v` matters: the postgres image declares `VOLUME /var/lib/postgresql`, so
Docker creates an anonymous volume regardless. Without `-v` those accumulate as
dangling volumes and the reset is not clean.

Frontend commands do not exist yet; add them here when that part lands.

## Documentation is the source of truth

`docs/` is unusually complete and is written to be binding, not aspirational.
Read the relevant document before implementing anything, and prefer following it
over re-deriving decisions:

- `docs/SPEC.md` — product behavior: the three card types, grading, tags, sets,
  review flow, SM-2, and an explicit non-goals list.
- `docs/ARCHITECTURE.md` — stack, entity model, API surface, deployment.
- `docs/DESIGN.md` — visual/UX direction and design tokens.
- `docs/PITCH.md` — the product's rationale, useful for judgment calls.

Both ARCHITECTURE.md and DESIGN.md explicitly state that where they are silent,
the implementer should extrapolate from the principles given rather than
introduce new patterns or fall back on library defaults.

## Architecture in brief

Go REST JSON API + PostgreSQL + a React SPA (TypeScript preferred), session-cookie
auth, deployed via Docker Compose. The board is `chessground`; move validation
uses `chess.js`/`chessops` on the frontend and a Go chess library on the backend.

The domain shape that spans multiple files: a `cards` table holds fields common
to all three card types (type, FEN, source note) with type-specific fields in a
`jsonb` details column (or three typed tables — implementer's call). SRS
scheduling lives in a separate `review_state` table, one row per card,
deliberately isolated so SM-2 can later be swapped for FSRS without touching card
content; `review_log` is append-only history that a future FSRS migration would
be trained on.

## Constraints that are easy to violate accidentally

These are decided, not open questions:

- **Multi-user, every account private.** Registration is open. No sharing,
  public cards, or community features — one user's data is never visible to
  another. Scope every query by `user_id`, always.
- **No engine (Stockfish) dependency** and **no game import** (Lichess/chess.com)
  in v1.
- **Grading is binary or self-assessed only.** Do not build three-way or
  partial-credit grading UI or scoring.
- **Nothing is auto-detected or auto-classified.** The user authors every card,
  question, answer, and tag by hand — that is the product's whole premise.
- New categories of mistake are **tags**, never new card types. There are exactly
  three card types.
- Configuration comes from environment variables, never hardcoded.

## Code Style Philosophy

Write the simplest possible code that solves the problem. Prioritize
readability and directness over cleverness, abstraction, or premature
generalization.

### Simplicity rules
- Prefer the most boring, obvious solution. Do not introduce a design
  pattern, abstraction layer, or indirection unless it's solving a
  problem that exists *right now* in this codebase.
- No speculative generality. Don't add config options, interfaces,
  or extension points for hypothetical future needs. YAGNI, hard.
- Avoid unnecessary layers: no wrapper classes, no factories, no
  dependency-injection containers, no "manager"/"service" objects
  wrapping a single external call — call it directly.
- Flat over nested. Prefer early returns / guard clauses over deep
  if/else nesting.
- Don't add error handling, logging, or validation beyond what the
  task actually needs. No defensive code for cases that can't happen.
- Function and variable names should be plain and literal, not
  abstract or "enterprise" (e.g. `parseInvoice`, not
  `AbstractInvoiceProcessorFactory`).

### Comments and structure
- Comment only where the *why* isn't obvious from the code. Do not
  add comments that restate what the line does.
- No boilerplate header comments, no "Author:", no change-log
  comments, no section-divider comments (`// ---- Helpers ----`).

### Avoid AI "tells"
- Don't add TODO comments unless explicitly asked.
- Don't wrap the whole implementation in a try/catch "just in case."
- Don't over-explain in commit messages or PR descriptions — write
  them the way a developer would, tersely, describing what changed
  and why.
- Use standard Go formatting (`gofmt`) and idiomatic Go conventions,
  not a "cleaner" personal style imposed on top.
- When in doubt between two approaches, pick the one a competent
  engineer would write fastest by hand, not the one that looks most
  thorough in a diff.

### Before finishing
- Re-read the diff and remove anything that isn't load-bearing:
  unused imports, unused parameters, redundant checks, extra
  abstraction that isn't used more than once.
- If you've created a small hierarchy of tiny functions each called
  once, inline them.

## Design constraints

Dark-first, Lichess-like: the board is always the largest, highest-contrast
object on any screen it appears on, and nothing else competes with it. Compact
spacing, 2–4px radii, hairline borders, monospace for chess notation and FEN.
No gamification language or visuals (streaks, XP, badges, confetti, animated
progress bars), no hero illustrations or mascots, no glossy consumer-SaaS look.
Copy is plain and second-person, and assumes the reader already knows chess
terminology.
