# Локальная разработка (на своей машине). Прод-контур — через Docker (см. Dockerfile).

BIN        := bin/scainer
JPLAG_VER  := 5.1.0
JPLAG_JAR  := bin/jplag.jar

.PHONY: all code-gen build vet test run-serve setup docker-build docker-run fmt clean

all: vet test build

code-gen:
	go generate ./api/ ./pkg/ejudge/

build: code-gen
	go build -o $(BIN) ./cmd/scainer

vet: code-gen
	go vet ./...

test: code-gen
	go test -race ./...

# HTTP-сервис (нужны переменные окружения из .env — см. .env.example).
run-serve: code-gen
	go run ./cmd/scainer

# Тянет пиннутый jplag.jar (обязателен для старта сервера — детектор JPlag).
setup:
	mkdir -p bin
	curl -fsSL -o $(JPLAG_JAR) \
		https://github.com/jplag/JPlag/releases/download/v$(JPLAG_VER)/jplag-$(JPLAG_VER)-jar-with-dependencies.jar

fmt:
	gofmt -w .

docker-build: setup
	docker build -t scainer:local .

docker-run:
	docker run --rm --env-file .env -p 8080:8080 scainer:local

clean:
	rm -rf bin
