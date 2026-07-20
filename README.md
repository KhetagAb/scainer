# Scainer

Информационная система для преподавателей: отчёты о подозрительных на списывание /
использование AI посылках (формат ICPC). Система **не выносит вердиктов** — ранжирует и
объясняет; финальное решение за преподавателем.

## Требования

**Обязательно:**

- [Go 1.26+](scainer/go.mod) — бэкенд
- [Node.js 18+](scainer-front/package.json) и **npm** — фронтенд
- **MongoDB**
- **Java (JRE 17+)** в `PATH` — для детектора JPlag

**Для сборки бэкенда** (codegen):

```bash
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0
```

**Опционально:** Docker и Docker Compose — запуск всего стека одной командой.

## Быстрый старт (Docker Compose, локально)

```bash
cp scainer/.env.example scainer/.env   # заполнить секреты и MONGO_INITDB_ROOT_PASSWORD
docker compose up --build
```

Порты с хоста задаёт [`docker-compose.override.yml`](docker-compose.override.yml)
(только для локальной разработки — **не** для школьного сервера):

- UI (nginx): **http://localhost:8088**
- Backend: **http://localhost:8080**
- MongoDB: `localhost:27017` (с auth из `.env`)

## Школьный сервер (LAN + reverse proxy)

TLS терминируется на **внешнем** reverse proxy; внутри стека остаётся HTTP.

```bash
cp scainer/.env.example scainer/.env   # сильные ADMIN_PASSWORD, JWT_SECRET, Mongo-пароль
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

[`docker-compose.prod.yml`](docker-compose.prod.yml) публикует только nginx на
`127.0.0.1:8088`; Mongo и API **без** host-портов (не видны из LAN).

**Чеклист reverse proxy → `http://127.0.0.1:8088`:**

- Снаружи только HTTPS; до прокси не отдавать `:8088`/`:8080`/`:27017` в LAN.
- Не логировать заголовок `Authorization`.
- Для SSE прогресса (`/api/jobs/…`): `proxy_buffering off`, длинный `proxy_read_timeout`
  (как во внутреннем [`scainer-front/nginx.conf`](scainer-front/nginx.conf)).
- Прокидывать `X-Forwarded-For` / `X-Real-IP` (rate limit login смотрит на IP клиента).

**MongoDB auth:** `MONGO_INITDB_*` применяются только при **первом** старте с пустым
volume. Если volume уже без пароля — удалить volume (`docker compose down -v`, данные
пропадут) или вручную создать пользователя в Mongo и обновить пароль в `.env`.

## Локальная разработка

### MongoDB

```bash
docker compose up mongodb -d
```

Порт `27017` на хосте. Для локального бэкенда в `scainer/.env`:

```
MONGO_INITDB_ROOT_USERNAME=scainer
MONGO_INITDB_ROOT_PASSWORD=…
MONGODB_HOST=localhost
MONGODB_DATABASE=scainer
```

### Backend

```bash
cd scainer
cp .env.example .env          # ADMIN_*, JWT_*, EJUDGE_API_KEY, Mongo …
make setup                    # тянет bin/jplag.jar (нужен java)
make build                    # codegen + go build
./bin/scainer                 # слушает :8080
# или: make run-serve
```

Реестр контестов стартует пустым — контесты добавляются через UI / `POST /api/contests`.

### Frontend

```bash
cd scainer-front
npm install
npm run dev                   # http://localhost:5173, прокси /api → :8080
```

Другой хост бэкенда:

```bash
VITE_API_PROXY_TARGET=http://host:8080 npm run dev
```

## Конфигурация

| Источник | Файл | Примеры |
|----------|------|---------|
| Секреты и env | [`scainer/.env.example`](scainer/.env.example) | JWT, admin, ejudge, MongoDB, concurrency |
| OpenAPI | [`scainer/api/openapi.yaml`](scainer/api/openapi.yaml) | HTTP API |

Основные переменные (см. `.env.example`):

- `ADMIN_USERNAME` / `ADMIN_PASSWORD` / `JWT_SECRET` / `JWT_TTL` (по умолчанию `12h`)
- `EJUDGE_BASE_URL` / `EJUDGE_API_KEY` / `EJUDGE_TIMEOUT`
- `MONGO_INITDB_ROOT_USERNAME` / `MONGO_INITDB_ROOT_PASSWORD` / `MONGODB_HOST` / `MONGODB_DATABASE`
- `STORE_DIR` — персистентность посылок (по умолчанию `./data`)
- `JPLAG_JAR_PATH` — путь к jar (по умолчанию `bin/jplag.jar` после `make setup`)
- JPlag CLI: `--normalize` (cpp/java), `-n -1` (все сравнения), `--cluster-skip`
- `JOBS_MAX_CONCURRENT` / `ANALYZE_CONCURRENCY` — пул импорта и потолок JVM/JPlag

## Полезные команды

| Команда | Где | Действие |
|---------|-----|----------|
| `make setup` | `scainer/` | Скачать пиннутый `jplag.jar` |
| `make build` / `make test` | `scainer/` | Сборка и тесты (с codegen) |
| `make code-gen` | `scainer/` | OpenAPI → `internal/generated` |
| `make run-serve` | `scainer/` | HTTP-сервис на `:8080` |
| `npm run build` | `scainer-front/` | Production-сборка |
| `npm run generate-client` | `scainer-front/` | Клиент из OpenAPI → `src/client` |

## Документация

- [docs/requirements.md](docs/requirements.md) — требования
- [docs/principles.md](docs/principles.md) — инварианты
- [docs/specs/](docs/specs/) — спеки инкрементов
