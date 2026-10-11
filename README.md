# Taxios

White-label, multi-tenant ride-hailing SaaS for transport operators — Sri Lanka first.

- **Status:** Stage 1 foundation (identity, tenants, isolation, audit). See [docs/PROJECT_STATUS.md](docs/PROJECT_STATUS.md).
- **Architecture:** Go modular monolith + PostgreSQL row-level security. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
- **Contributing / AI sessions:** read [CLAUDE.md](CLAUDE.md) first.

## Quick start
Requires Go 1.26+ and PostgreSQL 16 where your OS user can create databases and roles.
```
make test      # run the full suite against a fresh local database
make migrate   # set up taxios_dev
make serve     # API on http://localhost:8080
```
