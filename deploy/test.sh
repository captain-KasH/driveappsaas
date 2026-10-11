#!/usr/bin/env bash
# Run the full suite in Docker against a fresh PostGIS. Works on any Docker host.
#   deploy/test.sh                 # all tests
#   TEST_ARGS="-run Booking -v" deploy/test.sh
set -euo pipefail
cd "$(dirname "$0")"
trap 'docker compose -f compose.test.yml down -v --remove-orphans >/dev/null 2>&1' EXIT
docker compose -f compose.test.yml run --rm test
