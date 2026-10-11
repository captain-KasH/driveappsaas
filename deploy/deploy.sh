#!/usr/bin/env bash
# Deploy Taxios on the packline server. Run as appuser from /data/apps/taxios:
#   ./deploy/deploy.sh
# pull → test → build → migrate → recreate api → smoke test through packline nginx.
set -euo pipefail
cd "$(dirname "$0")/.."

git pull --ff-only

if [ ! -f deploy/.env ]; then
  umask 077
  printf 'OWNER_DB_PASSWORD=%s\nAPP_DB_PASSWORD=%s\n' "$(openssl rand -hex 24)" "$(openssl rand -hex 24)" > deploy/.env
  echo "generated deploy/.env (new secrets, not printed)"
fi

deploy/test.sh

compose() { docker compose -f deploy/compose.yml --env-file deploy/.env "$@"; }
compose build migrate
compose up -d db
compose run --rm migrate
compose up -d --no-deps --force-recreate --remove-orphans taxios-api

for i in $(seq 1 30); do
  status=$(docker inspect -f '{{.State.Health.Status}}' taxios-api 2>/dev/null || echo starting)
  [ "$status" = healthy ] && break
  sleep 2
done
[ "$status" = healthy ] || { echo "api not healthy: $status"; compose logs --tail 50 taxios-api; exit 1; }

docker exec packlink-nginx curl -fsS -k --resolve taxios.packline.lk:443:127.0.0.1 https://taxios.packline.lk/healthz \
  && echo && echo "Deploy succeeded: $(git rev-parse --short HEAD)"
