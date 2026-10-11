#!/usr/bin/env bash
# Run the full suite in Docker against a fresh PostGIS. Works on any Docker host.
#   deploy/test.sh                 # all tests
#   TEST_ARGS="-run Booking -v" deploy/test.sh
#   SKIP_WEB=1 deploy/test.sh       # Go only (API iteration)
# Project name: `taxios-test` here (deploy.sh); `make test-remote` overrides COMPOSE_PROJECT_NAME=taxios-ci so the
# two never share a stack. Never run two of the SAME project at once: the exit trap tears it down.
set -euo pipefail
cd "$(dirname "$0")"
# Keep the cache volumes (gomod/gobuild/webmods/webnext); only the throwaway DB goes.
trap 'docker compose -f compose.test.yml down --remove-orphans >/dev/null 2>&1' EXIT
docker compose -f compose.test.yml run --rm taxios-test-go
if [ "${SKIP_WEB:-}" != "1" ]; then
  docker compose -f compose.test.yml run --rm --no-deps taxios-test-web
fi
