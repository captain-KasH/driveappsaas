#!/bin/sh
# Runs once on first database init (postgres image entrypoint).
# Creates the runtime role the API connects as: no superuser, cannot bypass RLS.
set -eu
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
  -v pw="$APP_DB_PASSWORD" <<'SQL'
CREATE ROLE taxios_app LOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE PASSWORD :'pw';
SQL
