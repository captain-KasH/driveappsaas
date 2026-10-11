# Deployment runbook

## Production: https://taxios.packline.lk
Hosted on the packline server (Contabo, `164.68.97.219`) as a **separate Docker stack**. Server
access is **only ever** `ssh appuser@164.68.97.219` — never root, never sudo.

| Piece | Where |
|---|---|
| Code checkout | `/data/apps/taxios` (git clone of `captain-KasH/driveappsaas`, `main`) |
| Stack | `deploy/compose.yml`, compose project `taxios`: `taxios-db` (postgis/postgis:16-3.4, volume `taxios_db`), `taxios-migrate` (one-shot), `taxios-api`. **Every name is `taxios-`-prefixed** (D-026): the server also runs packline and its voice AI |
| Secrets | `/data/apps/taxios/deploy/.env` (chmod 600, generated on first deploy, never printed or committed) |
| Edge | packline's `packlink-nginx` container. Vhost in the **packline repo** `ops/nginx/nginx.conf` (`taxios.packline.lk` block). The API joins network `packlink_packlink` as `taxios-api:8080`; no host ports published |
| DNS / TLS | Cloudflare (packline.lk zone, account kavindaroc@): `A taxios → 164.68.97.219`, **Proxied**. SSL mode **Full** → origin uses packline's self-signed cert; Cloudflare's edge cert covers `*.packline.lk` |
| Rate limits | nginx: `/v1/auth/*` + passenger sign-up → `auth` zone (10/min/IP, burst 5); rest → `api_general` (20/s, burst 40) |

### Deploy (manual — no CI)
```
ssh appuser@164.68.97.219
cd /data/apps/taxios && ./deploy/deploy.sh
```
`deploy.sh`: `git pull --ff-only` → generate `.env` if missing → **run the full test suite** in a throwaway
PostGIS container (aborts on failure) → build image → start db → run migrations → recreate api →
wait for container health → smoke `GET https://taxios.packline.lk/healthz` through nginx →
prints `Deploy succeeded: <sha>`.

### Test without deploying
From the dev machine: `make test-remote` (rsyncs the working tree to `/data/apps/taxios-ci` and
runs `deploy/test.sh`). Add `TEST_ARGS="-run Name -v"` to narrow.

### Bootstrap a platform admin (once)
```
cd /data/apps/taxios
docker compose -f deploy/compose.yml --env-file deploy/.env run --rm \
  -e TAXIOS_ADMIN_PASSWORD='…' migrate create-platform-admin you@example.com "Your Name"
```
(The `migrate` service has the owner DB URL; the password is read from the env, never a flag.)

### Changing the nginx vhost
Edit `ops/nginx/nginx.conf` **in the packline repo**, commit only that file, push, then on the
server `cd /data/apps/packlink && git pull --ff-only`. The file is bind-mounted read-only as a
single file, so the running container keeps the old inode (K-011): the change goes live on
the next packline deploy (it recreates nginx) or with
`docker compose up -d --no-build --force-recreate nginx` from `/data/apps/packlink` — only when no
packline deploy is running (`pgrep -f deploy-manual.sh`). Always `docker exec packlink-nginx nginx -t` first.

### Rollback
`git checkout <previous sha>` in `/data/apps/taxios`, then `deploy.sh` without pull
(migrations are forward-only and stay backward-compatible for one release). For data loss:
restore the `taxios_db` volume from backup (backups: T-100, **not yet configured**).

### Logs and health
```
docker compose -f deploy/compose.yml --env-file deploy/.env logs -f taxios-api   # JSON, one line per request
docker inspect -f '{{.State.Health.Status}}' taxios-api
```

## Local development
`make test` needs a local Postgres 16 **with PostGIS**; the original dev Mac doesn't have it (K-001),
so use `make test-remote`. Pure unit tests (`go test ./internal/fare/`) run anywhere.
