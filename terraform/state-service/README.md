# Local Terraform state service

> **LLM-generated:** this service, its scripts, and this documentation were
> generated with an LLM for a local Terraform learning environment. Review and
> runtime-test it before relying on it. It is not a production-ready database
> deployment.

The service provides a PostgreSQL backend for local Terraform state. One
database stores two isolated backend schemas:

- `dev` for the development Terraform root;
- `prod` for the production-like Terraform root.

PostgreSQL is exposed only on `127.0.0.1`, and its data is kept in the named
Docker volume `terraform-state-db-data`. The service has its own lifecycle and
must not be managed by either Terraform root whose state it stores.

## Start

From this directory:

1. Copy `.env.example` to `.env`.
2. Replace both password placeholders with different values of at least 24
   characters.
3. Validate and start the service:

```bash
./scripts/state-service.sh validate
./scripts/state-service.sh up
```

Every `up` reconciles the administrator password, the restricted Terraform
role, both schemas, and the database objects required by Terraform's `pg`
backend. Existing state data in the named volume is preserved.

## Select a Terraform state

A successful `up` creates two ignored, permission-restricted environment
files. Source exactly one of them before running Terraform:

```bash
source .runtime/dev.env
# or
source .runtime/prod.env
```

These files provide the standard PostgreSQL connection variables, select the
backend schema through `PG_SCHEMA_NAME`, and tell Terraform to use the tables
and indexes pre-created by the service. The corresponding Terraform root only
needs an empty `backend "pg" {}` configuration. Pre-creating the shared
`public.global_states_id_seq` lets the runtime role stay restricted instead of
granting it broad `CREATE` access to the `public` schema.

Check `PG_SCHEMA_NAME` before `terraform init`, migration, or any state-changing
operation. If both files are sourced in one shell, the last one wins.

## Operations

```text
./scripts/state-service.sh status
./scripts/state-service.sh backup
./scripts/state-service.sh down
./scripts/state-service.sh destroy --delete-data
```

- `status` verifies the container and the restricted role's access.
- `backup` creates and validates a custom-format dump in `backups/`.
- `down` stops PostgreSQL while preserving its named volume.
- `destroy --delete-data` requires the explicit flag, creates a verified backup,
  and then deletes the container and volume.

The `.env`, `.runtime/`, and `backups/` paths are excluded from Git. Backups can
contain sensitive Terraform state and must be protected like credentials.

## Local-lab limitations

This setup is a single PostgreSQL container without TLS, replication,
high availability, automated off-host backups, or a managed secret store.
Those constraints are intentional for the current local exercise and make the
service unsuitable for shared or production infrastructure.
