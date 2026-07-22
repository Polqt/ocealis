# Ocealis

Anonymous message-in-a-bottle world. See `README.md`, `docs/prd-v1.md`, `docs/backend-architecture.md`, and `CONTEXT.md` (product language) for context.

## Cursor Cloud specific instructions

### Services

| Service | Path | Dev command | Port |
| --- | --- | --- | --- |
| API (Go / Fiber v3) | `server/` | `go run .` | 8080 |
| Client (SolidStart / vinxi) | `client/` | `pnpm dev` | 3000 |
| PostgreSQL 16 | system service | `sudo pg_ctlcluster 16 main start` | 5432 |

The update script only refreshes dependencies (`go mod download`, `pnpm install`). It does NOT start Postgres or the app services — start those manually as above.

### Database

- The API reads `DATABASE_URL` from `server/.env` (loaded via `godotenv`). Local value: `postgres://ocealis:ocealis@localhost:5432/ocealis?sslmode=disable`. Without it the server exits with `DATABASE_URL not set`.
- Postgres is installed but its service is not auto-started on boot; run `sudo pg_ctlcluster 16 main start` before starting the API. Check with `sudo pg_lsclusters`.
- Schema source of truth is `server/db/schema.sql` (matches the sqlc-generated models). The `server/db/migrations/*.sql` goose files are outdated/incomplete — apply `schema.sql` directly, not the migrations. To (re)create the DB: `createdb -O ocealis ocealis` then `psql -d ocealis -f server/db/schema.sql`.
- Anonymous casts insert `sender_id = 0`, and `bottles.sender_id` has an FK to `users(id)`. A placeholder row `users(id=0, nickname='anonymous')` must exist or every cast fails with a 500 (`could not cast bottle`). Seed it once: `psql -d ocealis -c "INSERT INTO users (id, nickname) VALUES (0, 'anonymous') ON CONFLICT (id) DO NOTHING;"`. This seed data persists in the VM snapshot.

### API routing caveat

- Fiber runs with `StrictRouting: true`, so trailing slashes matter. `POST /api/v1/bottles/` works but `POST /api/v1/bottles` returns 404. When testing with curl, match the registered path (e.g. the bottles collection endpoint needs the trailing slash).
- Known pre-existing app bug (not an environment issue): the client calls `POST /api/v1/bottles` (no trailing slash) in `client/src/lib/api.ts`, so the browser "Cast" flow gets a 404. The full stack works when the correct route is used.

### Testing / build

- Go: `cd server && go test ./...` (tests are self-contained, no DB needed), `go build ./...`.
- Client: `cd client && pnpm build`. There is no client lint or test script defined.
- The client points at the API via `VITE_API_URL` (defaults to `http://localhost:8080`).
- Turnstile abuse-check is disabled in dev when `TURNSTILE_SECRET` is empty (any non-empty `turnstile_token` is accepted).

### Notes

- `server/.air.toml` is present but configured with Windows paths (`tmp\main.exe`); prefer `go run .` for the dev server on Linux.
- Cast/re-release apply a "Mystery Delay" — new bottles have status `scheduled` and only become visible/`drifting` after the in-process scheduler (drift tick every 15 min) flips them.
