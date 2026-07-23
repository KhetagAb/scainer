COMPOSE ?= podman-compose

.PHONY: up down rebuild logs ps

up:
	$(COMPOSE) up -d --build --force-recreate

down:
	$(COMPOSE) down

rebuild: down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f scainer

ps:
	$(COMPOSE) ps
