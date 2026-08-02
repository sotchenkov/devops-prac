#!/usr/bin/env bash
set -Eeuo pipefail

umask 077

readonly ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly COMPOSE_FILE="${ROOT_DIR}/docker-compose.yml"
readonly ENV_FILE="${ROOT_DIR}/.env"
readonly RUNTIME_DIR="${ROOT_DIR}/.runtime"
readonly BACKUP_DIR="${ROOT_DIR}/backups"
readonly SERVICE_NAME="terraform-state-db"

DOCKER_CLI=""
POSTGRES_IMAGE="postgres:17-bookworm"
POSTGRES_ADMIN_PASSWORD=""
TERRAFORM_DB_USER=""
TERRAFORM_DB_PASSWORD=""
TERRAFORM_STATE_DB_PORT=""

die() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage: state-service.sh <command>

Commands:
  validate               Validate .env and the rendered Compose configuration
  up                     Start PostgreSQL, reconcile backend objects, verify access
  status                 Show container status and verify runtime access
  backup                 Write and verify a timestamped database backup
  down                   Stop the service while preserving its named volume
  destroy --delete-data  Back up, then remove the service and its named volume
EOF
}

load_config() {
  [[ -f "${ENV_FILE}" ]] || die "Missing ${ENV_FILE}; copy .env.example to .env and replace both password placeholders."

  local line key value
  local seen_admin=0
  local seen_user=0
  local seen_runtime=0
  local seen_port=0
  local seen_image=0

  while IFS= read -r line || [[ -n "${line}" ]]; do
    line="${line%$'\r'}"

    [[ -z "${line}" || "${line}" =~ ^[[:space:]]*# ]] && continue
    [[ "${line}" == *=* ]] || die "Invalid line in ${ENV_FILE}: expected KEY=VALUE."

    key="${line%%=*}"
    value="${line#*=}"

    [[ -n "${value}" ]] || die "${key} must not be empty."
    [[ ! "${value}" =~ [[:space:]#] ]] || die "${key} must be unquoted and contain no whitespace or '#'."

    case "${key}" in
      POSTGRES_IMAGE)
        ((seen_image == 0)) || die "POSTGRES_IMAGE is defined more than once."
        POSTGRES_IMAGE="${value}"
        seen_image=1
        ;;
      POSTGRES_ADMIN_PASSWORD)
        ((seen_admin == 0)) || die "POSTGRES_ADMIN_PASSWORD is defined more than once."
        POSTGRES_ADMIN_PASSWORD="${value}"
        seen_admin=1
        ;;
      TERRAFORM_DB_USER)
        ((seen_user == 0)) || die "TERRAFORM_DB_USER is defined more than once."
        TERRAFORM_DB_USER="${value}"
        seen_user=1
        ;;
      TERRAFORM_DB_PASSWORD)
        ((seen_runtime == 0)) || die "TERRAFORM_DB_PASSWORD is defined more than once."
        TERRAFORM_DB_PASSWORD="${value}"
        seen_runtime=1
        ;;
      TERRAFORM_STATE_DB_PORT)
        ((seen_port == 0)) || die "TERRAFORM_STATE_DB_PORT is defined more than once."
        TERRAFORM_STATE_DB_PORT="${value}"
        seen_port=1
        ;;
      *)
        die "Unsupported key ${key} in ${ENV_FILE}."
        ;;
    esac
  done < "${ENV_FILE}"

  ((seen_admin == 1)) || die "POSTGRES_ADMIN_PASSWORD is missing from ${ENV_FILE}."
  ((seen_user == 1)) || die "TERRAFORM_DB_USER is missing from ${ENV_FILE}."
  ((seen_runtime == 1)) || die "TERRAFORM_DB_PASSWORD is missing from ${ENV_FILE}."
  ((seen_port == 1)) || die "TERRAFORM_STATE_DB_PORT is missing from ${ENV_FILE}."

  [[ "${POSTGRES_ADMIN_PASSWORD}" != *replace-with* ]] || die "Replace POSTGRES_ADMIN_PASSWORD in ${ENV_FILE}."
  [[ "${TERRAFORM_DB_PASSWORD}" != *replace-with* ]] || die "Replace TERRAFORM_DB_PASSWORD in ${ENV_FILE}."
  ((${#POSTGRES_ADMIN_PASSWORD} >= 24)) || die "POSTGRES_ADMIN_PASSWORD must contain at least 24 characters."
  ((${#TERRAFORM_DB_PASSWORD} >= 24)) || die "TERRAFORM_DB_PASSWORD must contain at least 24 characters."
  [[ "${POSTGRES_ADMIN_PASSWORD}" =~ ^[A-Za-z0-9_.,:@%+/=~!-]+$ ]] || die "POSTGRES_ADMIN_PASSWORD contains characters that are unsafe in an unquoted Compose .env value."
  [[ "${TERRAFORM_DB_PASSWORD}" =~ ^[A-Za-z0-9_.,:@%+/=~!-]+$ ]] || die "TERRAFORM_DB_PASSWORD contains characters that are unsafe in an unquoted Compose .env value."
  [[ "${POSTGRES_ADMIN_PASSWORD}" != "${TERRAFORM_DB_PASSWORD}" ]] || die "Administrator and Terraform passwords must differ."
  [[ "${TERRAFORM_DB_USER}" =~ ^[a-z_][a-z0-9_]{0,62}$ ]] || die "TERRAFORM_DB_USER must be a valid unquoted PostgreSQL role name."
  [[ "${TERRAFORM_STATE_DB_PORT}" =~ ^[0-9]+$ ]] || die "TERRAFORM_STATE_DB_PORT must be numeric."
  ((10#${TERRAFORM_STATE_DB_PORT} >= 1 && 10#${TERRAFORM_STATE_DB_PORT} <= 65535)) || die "TERRAFORM_STATE_DB_PORT must be between 1 and 65535."
}

resolve_docker() {
  if [[ -n "${DOCKER_BIN:-}" ]]; then
    [[ -x "${DOCKER_BIN}" ]] || die "DOCKER_BIN is not executable: ${DOCKER_BIN}"
    DOCKER_CLI="${DOCKER_BIN}"
  elif command -v docker >/dev/null 2>&1; then
    DOCKER_CLI="$(command -v docker)"
  elif [[ -x /Applications/Docker.app/Contents/Resources/bin/docker ]]; then
    DOCKER_CLI="/Applications/Docker.app/Contents/Resources/bin/docker"
  else
    die "Docker CLI was not found. Start Docker/Colima and optionally set DOCKER_BIN to the Docker CLI path."
  fi
}

compose() {
  "${DOCKER_CLI}" compose \
    --project-directory "${ROOT_DIR}" \
    --env-file "${ENV_FILE}" \
    --file "${COMPOSE_FILE}" \
    "$@"
}

validate_config() {
  compose config --quiet
}

service_is_running() {
  [[ -n "$(compose ps --quiet --status running "${SERVICE_NAME}")" ]]
}

verify_runtime_access() {
  service_is_running || die "${SERVICE_NAME} is not running."

  compose exec --no-TTY "${SERVICE_NAME}" bash -ceu '
    export PGPASSWORD="${TERRAFORM_DB_PASSWORD}"
    access_ok="$(
      psql \
        --host=127.0.0.1 \
        --username="${TERRAFORM_DB_USER}" \
        --dbname="${POSTGRES_DB}" \
        --tuples-only \
        --no-align \
        --set=ON_ERROR_STOP=1 \
        --command="
          SELECT (
            (SELECT count(*) = 2
             FROM pg_namespace
             WHERE nspname IN ('\''dev'\'', '\''prod'\'')
               AND nspowner = (SELECT oid FROM pg_roles WHERE rolname = current_user))
            AND has_schema_privilege(current_user, '\''public'\'', '\''USAGE'\'')
            AND has_sequence_privilege(current_user, '\''public.global_states_id_seq'\'', '\''USAGE'\'')
            AND has_table_privilege(current_user, '\''dev.states'\'', '\''SELECT,INSERT,UPDATE,DELETE'\'')
            AND has_table_privilege(current_user, '\''prod.states'\'', '\''SELECT,INSERT,UPDATE,DELETE'\'')
          )
        "
    )"
    [[ "${access_ok}" == "t" ]]
  '
}

write_backend_env() {
  local schema="$1"
  local target="${RUNTIME_DIR}/${schema}.env"
  local temporary="${target}.tmp.$$"

  {
    printf 'unset PG_CONN_STR\n'
    printf 'export PGHOST=%q\n' '127.0.0.1'
    printf 'export PGPORT=%q\n' "${TERRAFORM_STATE_DB_PORT}"
    printf 'export PGDATABASE=%q\n' 'terraform_backend'
    printf 'export PGUSER=%q\n' "${TERRAFORM_DB_USER}"
    printf 'export PGPASSWORD=%q\n' "${TERRAFORM_DB_PASSWORD}"
    printf 'export PG_SCHEMA_NAME=%q\n' "${schema}"
    printf 'export PG_SKIP_SCHEMA_CREATION=%q\n' 'true'
    printf 'export PG_SKIP_TABLE_CREATION=%q\n' 'true'
    printf 'export PG_SKIP_INDEX_CREATION=%q\n' 'true'
    printf 'export PGSSLMODE=%q\n' 'disable'
    printf 'export PGCONNECT_TIMEOUT=%q\n' '5'
  } > "${temporary}"

  chmod 600 "${temporary}"
  mv "${temporary}" "${target}"
}

write_backend_envs() {
  mkdir -p "${RUNTIME_DIR}"
  chmod 700 "${RUNTIME_DIR}"
  write_backend_env dev
  write_backend_env prod
}

start_service() {
  chmod 600 "${ENV_FILE}"
  validate_config
  compose up --detach --wait --wait-timeout 90 --remove-orphans
  compose exec --no-TTY "${SERVICE_NAME}" bash /usr/local/bin/bootstrap-state-backend.sh
  verify_runtime_access
  write_backend_envs

  printf 'Terraform state service is ready at 127.0.0.1:%s.\n' "${TERRAFORM_STATE_DB_PORT}"
  printf 'Backend environments: %s and %s.\n' "${RUNTIME_DIR}/dev.env" "${RUNTIME_DIR}/prod.env"
}

show_status() {
  compose ps
  verify_runtime_access
  printf 'Runtime role can access both backend schemas.\n'
}

backup_service() {
  service_is_running || die "Start the service before creating a backup."

  mkdir -p "${BACKUP_DIR}"
  chmod 700 "${BACKUP_DIR}"

  local timestamp target temporary
  timestamp="$(date -u '+%Y%m%dT%H%M%SZ')"
  target="${BACKUP_DIR}/terraform_backend_${timestamp}.dump"
  temporary="${target}.tmp.$$"

  compose exec --no-TTY "${SERVICE_NAME}" \
    pg_dump \
    --username=postgres \
    --dbname=terraform_backend \
    --format=custom > "${temporary}"

  if ! compose exec --no-TTY "${SERVICE_NAME}" pg_restore --list < "${temporary}" >/dev/null; then
    rm -f "${temporary}"
    die "PostgreSQL could not read the generated backup archive."
  fi

  chmod 600 "${temporary}"
  mv "${temporary}" "${target}"
  printf 'Verified backup: %s\n' "${target}"
}

destroy_service() {
  [[ "${1:-}" == "--delete-data" ]] || die "Refusing to delete the state volume without: destroy --delete-data"
  service_is_running || die "Start the service first so a verified backup can be created before deletion."

  backup_service
  compose down --volumes --remove-orphans
  rm -f "${RUNTIME_DIR}/dev.env" "${RUNTIME_DIR}/prod.env"
  printf 'Removed the PostgreSQL container and terraform-state-db-data volume. The verified backup was preserved.\n'
}

main() {
  local command="${1:-}"

  [[ -n "${command}" ]] || {
    usage
    exit 1
  }

  case "${command}" in
    -h|--help|help)
      usage
      exit 0
      ;;
  esac

  load_config
  resolve_docker

  case "${command}" in
    validate)
      validate_config
      printf 'State service configuration is valid.\n'
      ;;
    up)
      start_service
      ;;
    status)
      show_status
      ;;
    backup)
      backup_service
      ;;
    down)
      compose down --remove-orphans
      printf 'State service stopped; named volume preserved.\n'
      ;;
    destroy)
      destroy_service "${2:-}"
      ;;
    *)
      usage
      die "Unknown command: ${command}"
      ;;
  esac
}

main "$@"
