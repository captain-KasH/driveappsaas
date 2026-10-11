# Deployment runbook

Status: local development only. No staging or production exists yet (T-100).

## Local (verified 2026-10-11, macOS, Homebrew Postgres 16)
```
make test                 # creates taxios_app role + fresh taxios_test, runs all tests
make migrate              # creates taxios_dev if missing, applies migrations
TAXIOS_ADMIN_PASSWORD='…' MIGRATE_DATABASE_URL=postgres://localhost/taxios_dev \
  go -C services/api run ./cmd/taxios-api create-platform-admin you@example.com "Your Name"
make serve                # API on :8080 as taxios_app
```
`docker-compose.yml` (Postgres + PostGIS, Redis) is provided for machines with Docker; it is **not
verified** here because Docker is not installed on the dev machine.

## Roles
- Schema owner (runs `migrate`): needs CREATE on the database and CREATEROLE once (creates `taxios_app`).
- `taxios_app` (runs `serve`): LOGIN, NOSUPERUSER, NOBYPASSRLS. Set its password out of band:
  `ALTER ROLE taxios_app LOGIN PASSWORD '…'`.

## Release procedure (to be automated, T-100)
1. CI green (vet, gofmt, tests against Postgres service).
2. Back up the database; confirm the backup restores.
3. `taxios-api migrate` with the owner URL.
4. Roll the API binary; check `GET /healthz`.
5. Rollback: previous binary (migrations are backward-compatible for one release) or restore backup.
