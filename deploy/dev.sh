#!/usr/bin/env bash
# Server DEV stack (see compose.dev.yml). Run on the server from /data/apps/taxios-ci (make dev-remote does it).
#   deploy/dev.sh up      build from the current tree, migrate, (re)start the API
#   deploy/dev.sh seed    create the test operator/driver (idempotent)
#   deploy/dev.sh logs    tail API logs
#   deploy/dev.sh down    stop (keeps data)      deploy/dev.sh reset   stop + delete dev data
set -euo pipefail
cd "$(dirname "$0")"
c() { docker compose -f compose.dev.yml "$@"; }
case "${1:-up}" in
  up)
    c build taxios-dev-migrate
    c up -d taxios-dev-db
    c run --rm taxios-dev-migrate
    c up -d --no-deps --force-recreate taxios-dev-api
    for i in $(seq 1 30); do curl -fsS 127.0.0.1:18080/healthz >/dev/null 2>&1 && break; sleep 1; done
    curl -fsS 127.0.0.1:18080/healthz && echo " — dev API on the server at 127.0.0.1:18080"
    ;;
  seed) TAXIOS_API_URL=http://127.0.0.1:18080 SEED_ADMIN_VIA=docker:taxios-dev-api python3 ../scripts/dev-seed.py ;;
  logs) docker logs --tail 100 -f taxios-dev-api ;;
  down) c down ;;
  reset) c down -v ;;
esac
