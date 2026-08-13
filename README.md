# SpacedChess

SpacedChess is a tool for storing and reviewing your own chess mistakes
over time, using spaced repetition.

## Development

Start the docker compose:

  docker compose up -d

Connect with psql:

  docker compose exec db psql -U user

Remove docker volume and start with a fresh start

  docker volume rm spacedchess_db-data


Populate spacedchess database
  docker compose exec -T db psql -U user -d spacedchess < scripts/seed.sql
