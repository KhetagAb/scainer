# Прод-контур: собранный бинарь + JRE (для JPlag) + пиннутый jplag.jar.
# Локальная разработка/тесты — на машине напрямую (см. Makefile), не через этот образ.

# --- build stage ---
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go generate ./api/ ./pkg/ejudge/
RUN CGO_ENABLED=0 go build -o /out/scainer ./cmd/scainer

# --- runtime stage ---
# JRE-базовый образ: несёт Java для JPlag-детектора (инкремент 003).
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /out/scainer /usr/local/bin/scainer
COPY testdata/ /app/testdata/

COPY bin/jplag.jar /app/bin/jplag.jar
ENV JPLAG_JAR_PATH=/app/bin/jplag.jar

EXPOSE 8080
ENTRYPOINT ["scainer"]
