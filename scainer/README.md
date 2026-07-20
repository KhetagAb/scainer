# scainer

Информационная система для преподавателей: готовит отчёты о подозрительных на списывание /
использование AI посылках школьников (формат ICPC). Система **не выносит вердиктов** — только
ранжирует и объясняет, финальное решение за преподавателем.

> Статус: HTTP-сервис с JPlag-детектором. CLI-режимов (`analyze`/`eval`) больше нет — единственная
> точка входа это веб-сервер + UI (`scainer-front`), контесты добавляются через API.

## Быстрый старт (локально)

scainer — это **HTTP-сервис**. Реестр контестов стартует пустым, контесты
добавляются через веб-интерфейс (`POST /api/contests`). Настройки — через переменные окружения
(см. [.env.example](.env.example)): `ADMIN_USERNAME`/`ADMIN_PASSWORD`/`JWT_SECRET`/`JWT_TTL`,
`EJUDGE_BASE_URL`/`EJUDGE_API_KEY`/`EJUDGE_TIMEOUT`, `JPLAG_JAR_PATH`.

```bash
make test                 # go test ./...
make setup                # тянет bin/jplag.jar (нужен java в PATH)
make run-serve             # HTTP-сервис на :8080
# или напрямую:
go run ./cmd/scainer --addr :8080
```

Прод-контур (Docker):
```bash
make docker-build && make docker-run
```

## Архитектура (кратко)
```
HTTP API (POST /api/contests) → Importer(реестр источников) → Store(канон)
  → Selector[U] → Detector[U] (сигналы) → scoring (находки) → transport (view-модель для GET /findings)
```
- **Модульность:** детектор параметризован типом юнита (`Detector[U Unit]`); новый вид анализа =
  новый юнит + селектор + детектор, без правки ядра. Сторонние инструменты (JPlag/MOSS) — тем же контрактом.
- **Юниты-выборки:** `StandaloneUnit`, `PairUnit`, `ProblemUnit`, `ProblemParticipantUnit`,
  `ParticipantUnit`. Пара/кластер — гранулярность выхода (`Signal.Subject`), не входа.
- **Абстракция от judge:** источники транслируются в канонический `domain.Submission` до аналитики.

## Структура
```
cmd/scainer/          entrypoint HTTP-сервера (флаги, env, graceful shutdown)
internal/domain/      канонические типы + семейство юнитов
internal/importer/    источники (folder — для тестов, ejudge — прод) + реестр
internal/store/       каноническое хранилище (FS; Mem — только тесты)
internal/detect/      ядро: Detector[U], Selector[U], Stage, AnalysisPolicy
  detect/dummy/        два dummy-детектора (для тестов ядра)
  detect/jplag/        JPlag-детектор (обязателен для serve — fail-fast при старте, если недоступен)
internal/scoring/     агрегация сигналов → находки
internal/contests/    реестр контестов + мутации (единая точка входа: и HTTP, и внутренние вызовы)
internal/transport/   HTTP-хендлеры + view-модель JSON-ответов (view.go)
internal/pipeline/    оркестрация прогона
api/detector/         wire-контракт out-of-process детекторов (дизайн)
docs/                 требования, принципы, SDD-спеки
```

## Документация (SDD)
- [docs/requirements.md](docs/requirements.md) — требования.
- [docs/principles.md](docs/principles.md) — инварианты проекта.
- [docs/specs/](docs/specs/) — спеки инкрементов (spec → design → tasks).

Разработка ведётся **step-by-step**: пакет за пакетом, с локальной проверкой и ревью.
