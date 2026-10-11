#!/usr/bin/env bash
# Local Taxios for device testing: a throwaway PostGIS cluster (port 5433, data under .dev/) + the API on :8080.
# Never touches the Postgres already running on the Mac (port 5432) or anything on the server.
#   scripts/dev-local.sh up      # init (first time) + start db + migrate + start API (logs: .dev/api.log)
#   scripts/dev-local.sh down    # stop API + db
#   scripts/dev-local.sh reset   # down + delete .dev/ data
set -euo pipefail
export LC_ALL=en_US.UTF-8 LANG=en_US.UTF-8
cd "$(dirname "$0")/.."
DEV=.dev; DATA=$DEV/pg; PORT=5433
PGBIN=/opt/homebrew/opt/postgresql@17/bin   # PostGIS bottles target 17/18; keg-only, no background service
[ -x "$PGBIN/postgres" ] || { echo "need: brew install postgis postgresql@17"; exit 1; }
APP_PW=taxios_dev_app; OWNER=taxios_owner; OWNER_PW=taxios_dev_owner

case "${1:-up}" in
up)
  mkdir -p $DEV
  if [ ! -d $DATA ]; then
    "$PGBIN/initdb" -D $DATA -U $OWNER --pwfile=<(echo $OWNER_PW) -A scram-sha-256 >/dev/null
  fi
  "$PGBIN/pg_ctl" -D $DATA -o "-p $PORT -k /tmp" -l $DEV/pg.log status >/dev/null 2>&1 || "$PGBIN/pg_ctl" -D $DATA -o "-p $PORT -k /tmp" -l $DEV/pg.log -w start >/dev/null
  export PGPASSWORD=$OWNER_PW
  "$PGBIN/psql" -h 127.0.0.1 -p $PORT -U $OWNER -d postgres -qtc "SELECT 1 FROM pg_database WHERE datname='taxios'" | grep -q 1 \
    || "$PGBIN/createdb" -h 127.0.0.1 -p $PORT -U $OWNER taxios
  "$PGBIN/psql" -h 127.0.0.1 -p $PORT -U $OWNER -d taxios -qc "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='taxios_app') THEN CREATE ROLE taxios_app LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD '$APP_PW'; END IF; END \$\$;"
  (cd services/api && go build -o ../../$DEV/taxios-api ./cmd/taxios-api)
  MIGRATE_DATABASE_URL="postgres://$OWNER:$OWNER_PW@127.0.0.1:$PORT/taxios?sslmode=disable" $DEV/taxios-api migrate
  pkill -f "$DEV/taxios-api serve" 2>/dev/null || true
  DATABASE_URL="postgres://taxios_app:$APP_PW@127.0.0.1:$PORT/taxios?sslmode=disable" ADDR=":8080" \
    TAXIOS_FILES_DIR=$DEV/files TAXIOS_PUBLIC_URL=http://localhost:8080 TAXIOS_DEV_TOKENS=1 \
    nohup $DEV/taxios-api serve > $DEV/api.log 2>&1 &
  for i in $(seq 1 30); do curl -fsS localhost:8080/healthz >/dev/null 2>&1 && break; sleep 0.5; done
  curl -fsS localhost:8080/healthz && echo " — local API on http://localhost:8080 (db :$PORT)"
  ;;
down)
  pkill -f "$DEV/taxios-api serve" 2>/dev/null || true
  "$PGBIN/pg_ctl" -D $DATA -o "-p $PORT -k /tmp" stop >/dev/null 2>&1 || true
  echo stopped
  ;;
reset)
  "$0" down; rm -rf $DEV; echo "removed $DEV"
  ;;
esac
