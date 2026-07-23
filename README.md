# Scainer

Система для преподавателей: отчёты о подозрительных на списывание / использование AI
посылках (формат ICPC). Не выносит вердиктов — ранжирует и объясняет; решение за
преподавателем.

## Требования

- [Go 1.26+](scainer/go.mod)
- [Node.js 18+](scainer-front/package.json) и npm
- MongoDB
- Java (JRE 17+) в `PATH` — для JPlag

## Быстрый старт (Docker Compose)

Локально:

```bash
cp scainer/.env.example scainer/.env
docker compose up --build
```

Порты (`docker-compose.override.yml`):

- UI: [http://localhost:8088](http://localhost:8088)
- Backend: [http://localhost:8080](http://localhost:8080)
- MongoDB: `localhost:27017`

В проде лучше ставить за reverse proxy (TLS снаружи). Пример:

```bash
cp scainer/.env.example scainer/.env
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

`[docker-compose.prod.yml](docker-compose.prod.yml)` публикует только nginx на `127.0.0.1:8088`. Проксируйте на него HTTPS; 

Для SSE (`/api/jobs/…`) отключите буферизацию и увеличьте `proxy_read_timeout` (см. `[scainer-front/nginx.conf](scainer-front/nginx.conf)`).

## Локальная разработка



### MongoDB

```bash
docker compose up mongodb -d
```

### Backend

```bash
cd scainer
cp .env.example .env
make setup    # bin/jplag.jar (нужен java)
make build
./bin/scainer # :8080
```

### Frontend

```bash
cd scainer-front
npm install
npm run dev   # http://localhost:5173, /api → :8080
```

Другой хост бэкенда: `VITE_API_PROXY_TARGET=http://host:8080 npm run dev`.

## Конфигурация


| Источник      | Файл                                                         |
| ------------- | ------------------------------------------------------------ |
| Дефолты       | `[scainer/configs/config.yaml](scainer/configs/config.yaml)` |
| Секреты / env | `[scainer/.env.example](scainer/.env.example)`               |
| OpenAPI       | `[scainer/api/openapi.yaml](scainer/api/openapi.yaml)`       |


Основные переменные: `ADMIN_*`, `JWT_*`, `EJUDGE_*`, `MONGO_*` / `MONGODB_*`,
`STORE_DIR`, `JPLAG_JAR_PATH`, `JOBS_MAX_CONCURRENT`, `ANALYZE_CONCURRENCY`,
опционально `AIUSAGE_ENABLED` + `OPENAI_*`.

## Команды


| Команда                    | Где              | Действие                       |
| -------------------------- | ---------------- | ------------------------------ |
| `make setup`               | `scainer/`       | Скачать `jplag.jar`            |
| `make build` / `make test` | `scainer/`       | Сборка и тесты                 |
| `make code-gen`            | `scainer/`       | OpenAPI → `generated/` |
| `make run-serve`           | `scainer/`       | HTTP на `:8080`                |
| `npm run build`            | `scainer-front/` | Production-сборка              |
| `npm run generate-client`  | `scainer-front/` | Клиент из OpenAPI              |


