# Taxios — local developer commands. Requires Go 1.26+ and a local Postgres 16
# where the current OS user can create databases and roles.
API := services/api
PGHOST ?= localhost
TEST_DB := taxios_test
DEV_DB := taxios_dev
# Local-only password for the taxios_app role.
APP_PW := taxios_app_dev

.PHONY: app-role test test-db dev-db migrate serve vet test-remote

vet:
	cd $(API) && go vet ./... && test -z "$$(gofmt -l .)"

app-role:
	@psql -h $(PGHOST) -d postgres -qtc "SELECT 1 FROM pg_roles WHERE rolname='taxios_app'" | grep -q 1 \
	  || psql -h $(PGHOST) -d postgres -qc "CREATE ROLE taxios_app"
	@psql -h $(PGHOST) -d postgres -qc "ALTER ROLE taxios_app LOGIN NOSUPERUSER NOBYPASSRLS PASSWORD '$(APP_PW)'"

test-db: app-role
	@dropdb -h $(PGHOST) --if-exists $(TEST_DB) && createdb -h $(PGHOST) $(TEST_DB)

test: vet test-db
	cd $(API) && TAXIOS_TEST_MIGRATE_URL="postgres://$(PGHOST)/$(TEST_DB)" \
	  TAXIOS_TEST_DATABASE_URL="postgres://taxios_app:$(APP_PW)@$(PGHOST)/$(TEST_DB)" \
	  go test -count=1 ./...

dev-db: app-role
	@psql -h $(PGHOST) -d postgres -qtc "SELECT 1 FROM pg_database WHERE datname='$(DEV_DB)'" | grep -q 1 || createdb -h $(PGHOST) $(DEV_DB)

migrate: dev-db
	cd $(API) && MIGRATE_DATABASE_URL="postgres://$(PGHOST)/$(DEV_DB)" go run ./cmd/taxios-api migrate

serve:
	cd $(API) && DATABASE_URL="postgres://taxios_app:$(APP_PW)@$(PGHOST)/$(DEV_DB)" go run ./cmd/taxios-api serve

# Run the suite on the packline server (Docker + PostGIS) from the local working tree.
SERVER := appuser@164.68.97.219
REMOTE_CI := /data/apps/taxios-ci
test-remote:
	rsync -az --delete --exclude .git --exclude 'deploy/.env' ./ $(SERVER):$(REMOTE_CI)/
	ssh $(SERVER) 'cd $(REMOTE_CI) && TEST_ARGS="$(TEST_ARGS)" deploy/test.sh'
