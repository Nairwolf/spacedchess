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
docker compose stop       # stop, keeping data
```

Credentials are `user` / `password`, database `spacedchess`.

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

### Reset the schema

`./init` is mounted into the container's entrypoint directory, so those SQL
files run **only against an empty data volume**. After editing them you have to
destroy the volume — there is no migration tool yet, so every schema change
costs you the data:

```sh
docker compose down -v && docker compose up -d
```

### Seed data

`scripts/seed.sql` creates one user and one card of each type. It is not run
automatically:

```sh
docker compose exec -T db psql -U user -d spacedchess < scripts/seed.sql
```

### Check the database is working

```sh
docker compose exec -T db pg_isready -U user -d spacedchess
docker compose exec -T db psql -U user -d spacedchess -c '\dt'
```

Expect `users`, `cards`, and `sessions`.
