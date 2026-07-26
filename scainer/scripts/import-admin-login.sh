#!/usr/bin/env bash
# Перенос логинов/паролей из login_audit в teachers и очистка audit.
# Запуск: make import-logins (из корня scainer/) или ./scripts/import-admin-login.sh [--dry-run]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_ROOT="$(cd "$ROOT/.." && pwd)"
ENV_FILE="${ENV_FILE:-$ROOT/.env}"
DRY_RUN=0

usage() {
  echo "usage: $0 [--dry-run]" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage ;;
    *) usage ;;
  esac
done

if [[ ! -f "$ENV_FILE" ]]; then
  echo "import-admin-login: .env не найден: $ENV_FILE" >&2
  exit 1
fi

read_env() {
  local key="$1"
  local line value
  line="$(grep -E "^${key}=" "$ENV_FILE" | tail -n1 || true)"
  if [[ -z "$line" ]]; then
    return 1
  fi
  value="${line#*=}"
  printf '%s' "$value"
}

MONGO_INITDB_ROOT_USERNAME="$(read_env MONGO_INITDB_ROOT_USERNAME || true)"
MONGO_INITDB_ROOT_PASSWORD="$(read_env MONGO_INITDB_ROOT_PASSWORD || true)"
MONGODB_DATABASE="$(read_env MONGODB_DATABASE || true)"

: "${MONGO_INITDB_ROOT_USERNAME:?MONGO_INITDB_ROOT_USERNAME не задан в $ENV_FILE}"
: "${MONGO_INITDB_ROOT_PASSWORD:?MONGO_INITDB_ROOT_PASSWORD не задан в $ENV_FILE}"
DB="${MONGODB_DATABASE:-scainer}"

cd "$COMPOSE_ROOT"

docker compose exec -T mongodb mongosh \
  -u "$MONGO_INITDB_ROOT_USERNAME" \
  -p "$MONGO_INITDB_ROOT_PASSWORD" \
  --authenticationDatabase admin \
  "$DB" \
  --quiet \
  --eval "const dryRun = $DRY_RUN === 1;

const entries = db.login_audit.find().sort({ at: 1 }).toArray();
if (entries.length === 0) {
  print('login_audit пуст — нечего импортировать');
  quit(0);
}

const users = {};
for (const entry of entries) {
  const login = (entry.login || '').trim();
  if (!login) continue;
  users[login] = entry.password;
}

const logins = Object.keys(users).sort();
print('найдено ' + entries.length + ' записей в login_audit, уникальных логинов: ' + logins.length);
for (const login of logins) {
  print('  - ' + login);
}

if (dryRun) {
  print('dry-run: teachers и login_audit не изменены');
  quit(0);
}

for (const login of logins) {
  db.teachers.replaceOne(
    { _id: login },
    { _id: login, password: users[login] },
    { upsert: true }
  );
}

const res = db.login_audit.deleteMany({});
print('добавлено/обновлено ' + logins.length + ' teacher(s), удалено ' + res.deletedCount + ' записей из login_audit');"
