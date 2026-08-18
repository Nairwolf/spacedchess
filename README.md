# SpacedChess

SpacedChess is a tool for storing and reviewing your own chess mistakes
over time, using spaced repetition.

## Development

Only Postgres runs in Compose for now; the Go API is run directly on the host.

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
