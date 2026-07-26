#!/usr/bin/env bash
# Перенос логинов/паролей из login_audit в teachers и очистка audit.
#
# Запуск:
#   ./scripts/import-admin-login.sh              # интерактивно (нужен TTY)
#   ./scripts/import-admin-login.sh --yes        # импортировать всех
#   ./scripts/import-admin-login.sh --only=a,b   # только указанные
#   ./scripts/import-admin-login.sh --skip=a,b   # всех, кроме указанных
#   ./scripts/import-admin-login.sh --dry-run    # показать план без изменений

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_ROOT="$(cd "$ROOT/.." && pwd)"
ENV_FILE="${ENV_FILE:-$ROOT/.env}"
DRY_RUN=0
IMPORT_ALL=0
ONLY_RAW=""
SKIP_RAW=""

usage() {
  cat >&2 <<'EOF'
usage: import-admin-login.sh [options]

  --dry-run          показать план, не менять БД
  -y, --yes          импортировать всех без вопросов
  --only=LOGIN,...   импортировать только указанные логины
  --skip=LOGIN,...   импортировать всех, кроме указанных
  -h, --help         эта справка

Без флагов скрипт спрашивает по каждому логину (нужен TTY).
Пропущенные логины остаются в login_audit.
EOF
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    -y|--yes) IMPORT_ALL=1; shift ;;
    --only=*) ONLY_RAW="${1#*=}"; shift ;;
    --only)
      [[ $# -ge 2 ]] || usage
      ONLY_RAW="$2"
      shift 2
      ;;
    --skip=*) SKIP_RAW="${1#*=}"; shift ;;
    --skip)
      [[ $# -ge 2 ]] || usage
      SKIP_RAW="$2"
      shift 2
      ;;
    -h|--help) usage ;;
    *) usage ;;
  esac
done

if [[ -n "$ONLY_RAW" && -n "$SKIP_RAW" ]]; then
  echo "import-admin-login: нельзя одновременно --only и --skip" >&2
  exit 1
fi

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

mongosh_eval() {
  docker compose exec -T mongodb mongosh \
    -u "$MONGO_INITDB_ROOT_USERNAME" \
    -p "$MONGO_INITDB_ROOT_PASSWORD" \
    --authenticationDatabase admin \
    "$DB" \
    --quiet \
    --eval "$1"
}

LOGINS=()
while IFS= read -r login; do
  [[ -n "$login" ]] && LOGINS+=("$login")
done < <(
  mongosh_eval "
const entries = db.login_audit.find().sort({ at: 1 }).toArray();
const users = {};
for (const entry of entries) {
  const login = (entry.login || '').trim();
  if (!login) continue;
  users[login] = entry.password;
}
for (const login of Object.keys(users).sort()) {
  print(login);
}
"
)

if [[ ${#LOGINS[@]} -eq 0 ]]; then
  echo "login_audit пуст — нечего импортировать"
  exit 0
fi

echo "найдено уникальных логинов в login_audit: ${#LOGINS[@]}"
for login in "${LOGINS[@]}"; do
  echo "  - $login"
done
echo

contains_login() {
  local needle="$1"
  shift
  local item
  for item in "$@"; do
    [[ "$item" == "$needle" ]] && return 0
  done
  return 1
}

split_csv() {
  local raw="$1"
  local -a out=()
  local part
  IFS=',' read -ra parts <<< "$raw"
  for part in "${parts[@]}"; do
    part="${part#"${part%%[![:space:]]*}"}"
    part="${part%"${part##*[![:space:]]}"}"
    [[ -n "$part" ]] && out+=("$part")
  done
  printf '%s\n' "${out[@]}"
}

SELECTED=()
SKIPPED=()
ONLY=()
SKIP=()

if [[ -n "$ONLY_RAW" ]]; then
  while IFS= read -r login; do
    ONLY+=("$login")
  done < <(split_csv "$ONLY_RAW")
  for login in "${ONLY[@]}"; do
    if contains_login "$login" "${LOGINS[@]}"; then
      SELECTED+=("$login")
    else
      echo "import-admin-login: логин не найден в login_audit: $login" >&2
      exit 1
    fi
  done
  for login in "${LOGINS[@]}"; do
    if ! contains_login "$login" "${SELECTED[@]}"; then
      SKIPPED+=("$login")
    fi
  done
elif [[ -n "$SKIP_RAW" ]]; then
  while IFS= read -r login; do
    SKIP+=("$login")
  done < <(split_csv "$SKIP_RAW")
  for login in "${LOGINS[@]}"; do
    if contains_login "$login" "${SKIP[@]}"; then
      SKIPPED+=("$login")
    else
      SELECTED+=("$login")
    fi
  done
elif [[ $IMPORT_ALL -eq 1 ]]; then
  SELECTED=("${LOGINS[@]}")
elif [[ -t 0 ]]; then
  for login in "${LOGINS[@]}"; do
    read -r -p "Импортировать «$login»? [Y/n] " answer </dev/tty || true
    case "${answer:-Y}" in
      n|N|no|No|NO|нет|Нет|НЕТ)
        SKIPPED+=("$login")
        ;;
      *)
        SELECTED+=("$login")
        ;;
    esac
  done
else
  echo "import-admin-login: нет TTY — укажите --yes, --only или --skip" >&2
  exit 1
fi

if [[ ${#SELECTED[@]} -eq 0 ]]; then
  echo "ничего не выбрано для импорта"
  exit 0
fi

echo
echo "импорт (${#SELECTED[@]}):"
for login in "${SELECTED[@]}"; do
  echo "  + $login"
done
if [[ ${#SKIPPED[@]} -gt 0 ]]; then
  echo "пропуск (${#SKIPPED[@]}), останутся в login_audit:"
  for login in "${SKIPPED[@]}"; do
    echo "  - $login"
  done
fi

if [[ $DRY_RUN -eq 1 ]]; then
  echo
  echo "dry-run: teachers и login_audit не изменены"
  exit 0
fi

SELECTED_JSON="$(
  python3 -c 'import json, sys; print(json.dumps(sys.argv[1:]))' "${SELECTED[@]}"
)"

docker compose exec -T \
  -e "SELECTED_JSON=${SELECTED_JSON}" \
  mongodb mongosh \
  -u "$MONGO_INITDB_ROOT_USERNAME" \
  -p "$MONGO_INITDB_ROOT_PASSWORD" \
  --authenticationDatabase admin \
  "$DB" \
  --quiet \
  --eval '
const selected = new Set(JSON.parse(process.env.SELECTED_JSON));
const entries = db.login_audit.find().sort({ at: 1 }).toArray();
const users = {};
for (const entry of entries) {
  const login = (entry.login || "").trim();
  if (!login || !selected.has(login)) continue;
  users[login] = entry.password;
}
const logins = Object.keys(users).sort();
if (logins.length === 0) {
  print("нечего импортировать");
  quit(0);
}
for (const login of logins) {
  db.teachers.replaceOne(
    { _id: login },
    { _id: login, password: users[login] },
    { upsert: true }
  );
}
const res = db.login_audit.deleteMany({ login: { $in: logins } });
print(
  "добавлено/обновлено " +
    logins.length +
    " teacher(s), удалено " +
    res.deletedCount +
    " записей из login_audit"
);
'
