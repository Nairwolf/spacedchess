# SpacedChess

SpacedChess is a tool for storing and reviewing your own chess mistakes
over time, using spaced repetition.

## Development

Only Postgres runs in Compose; the Go API is run directly on the host (see
[Running the API](#running-the-api)).

### Start and stop

```sh
docker compose up -d      # start Postgres on localhost:5432
docker compose ps         # check it's running
docker compose logs -f db # follow logs
docker compose stop       # stop, keeping the container and its data
docker compose down -v    # stop and discard the database
```

Credentials are `user` / `password`, database `spacedchess`.

**The database is throwaway.** There is no named volume, so nothing outlives
`down -v`. That's deliberate while the schema is still moving — a persistent
volume is worth adding once the backend does something worth keeping.

Always pass `-v` to `down`. The postgres image declares
`VOLUME /var/lib/postgresql`, so Docker creates an anonymous volume even though
the compose file asks for none; without `-v` those are left behind as dangling
volumes.

### Connect with psql

```sh
docker compose exec db psql -U user -d spacedchess
```

The `-d spacedchess` matters — without it psql defaults the database name to
the username and fails with `database "user" does not exist`.

Useful once you're in:

```
\dt              list tables
\d sessions      describe a table: columns, indexes, foreign keys
\l               list databases
\q               quit
```

To run a single query without opening a shell:

```sh
docker compose exec -T db psql -U user -d spacedchess -c 'SELECT * FROM users;'
```

### Schema and seed data

`./init` is mounted into the container's entrypoint directory. Its files run in
filename order whenever the database starts empty:

- `init/001_init.sql` — the schema
- `init/002_seed.sql` — dev seed data: one user and one card of each type

So every fresh start comes up already populated. Delete `002_seed.sql` if you
want an empty database.

### Reset

After editing either init file — or any time you want a clean slate:

```sh
docker compose down -v && docker compose up -d
```

There is no migration tool yet, so this is also how schema changes are applied:
throw the database away and rebuild it.

### Check the database is working

```sh
docker compose exec -T db pg_isready -U user -d spacedchess
docker compose exec -T db psql -U user -d spacedchess -c '\dt'
```

Expect `users`, `cards`, and `sessions`.

## Running the API

Two ways to run it, pick whichever suits what you're doing.

### All in Compose

The API runs as its own service, built from `Dockerfile`. Compose won't start
`api` until `db`'s healthcheck passes — the healthcheck itself connects over
TCP, since a plain `pg_isready` can report ready too early, against the
entrypoint's temporary socket-only server that runs while `./init` executes.

```sh
docker compose up -d --build
curl -s localhost:8080/health
docker compose logs -f api
```

Code changes need a rebuild to take effect — `docker compose up -d --build`
again. There's no live reload; the Dockerfile is meant to reflect the eventual
production build, not optimize the inner loop.

### On the host

Faster when iterating on `internal/` packages — no image rebuild. Needs
Postgres running in Compose (`docker compose up -d db`) and `DATABASE_URL`
pointed at `localhost` rather than the Compose service name:

```sh
export DATABASE_URL=postgres://user:password@localhost:5432/spacedchess
go run ./cmd/api
```

```sh
curl -s localhost:8080/health
```

`/health` pings Postgres, so it reports what the API can actually do:
`{"status":"ok"}` with a 200, or `{"status":"unavailable"}` with a 503 when the
database is unreachable. Try `docker compose stop db` and call it again.

### Build, test, lint

```sh
go build ./...
go vet ./...
staticcheck ./...
golangci-lint run
go test ./...
```

The `internal/store` tests need a live database and skip themselves when
`DATABASE_URL` is unset, so `go test ./...` passes without Docker running. Export
it to include them.
