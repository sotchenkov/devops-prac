#!/usr/bin/env bash
set -Eeuo pipefail

: "${POSTGRES_DB:?POSTGRES_DB is required}"
: "${POSTGRES_USER:?POSTGRES_USER is required}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD is required}"
: "${TERRAFORM_DB_USER:?TERRAFORM_DB_USER is required}"
: "${TERRAFORM_DB_PASSWORD:?TERRAFORM_DB_PASSWORD is required}"

psql \
  --username "${POSTGRES_USER}" \
  --dbname "${POSTGRES_DB}" \
  --set=ON_ERROR_STOP=1 \
  --set=admin_user="${POSTGRES_USER}" \
  --set=admin_password="${POSTGRES_PASSWORD}" \
  --set=tf_user="${TERRAFORM_DB_USER}" \
  --set=tf_password="${TERRAFORM_DB_PASSWORD}" \
  --set=db_name="${POSTGRES_DB}" <<'SQL'
SELECT format(
  'ALTER ROLE %I WITH LOGIN PASSWORD %L',
  :'admin_user',
  :'admin_password'
)
\gexec

SELECT format('CREATE ROLE %I WITH LOGIN', :'tf_user')
WHERE NOT EXISTS (
  SELECT 1
  FROM pg_roles
  WHERE rolname = :'tf_user'
)
\gexec

SELECT format(
  'ALTER ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS',
  :'tf_user',
  :'tf_password'
)
\gexec

REVOKE ALL ON DATABASE :"db_name" FROM PUBLIC;
GRANT CONNECT ON DATABASE :"db_name" TO :"tf_user";

REVOKE ALL ON SCHEMA public FROM PUBLIC;

SELECT format(
  'CREATE SCHEMA IF NOT EXISTS %I AUTHORIZATION %I',
  schema_name,
  :'tf_user'
)
FROM (VALUES ('dev'), ('prod')) AS schemas(schema_name)
\gexec

SELECT format(
  'ALTER SCHEMA %I OWNER TO %I',
  schema_name,
  :'tf_user'
)
FROM (VALUES ('dev'), ('prod')) AS schemas(schema_name)
\gexec

SELECT format(
  'REVOKE ALL ON SCHEMA %I FROM PUBLIC',
  schema_name
)
FROM (VALUES ('dev'), ('prod')) AS schemas(schema_name)
\gexec

SELECT format(
  'GRANT USAGE, CREATE ON SCHEMA %I TO %I',
  schema_name,
  :'tf_user'
)
FROM (VALUES ('dev'), ('prod')) AS schemas(schema_name)
\gexec
SQL
