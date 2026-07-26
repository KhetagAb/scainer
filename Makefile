COMPOSE ?= docker-compose
COMPOSE_PROD = -f docker-compose.yml -f docker-compose.prod.yml

.PHONY: up down rebuild logs ps up-prod down-prod

up:
	$(COMPOSE) up -d --build --force-recreate

down:
	$(COMPOSE) down

up-prod:
	$(COMPOSE) $(COMPOSE_PROD) up -d --build

down-prod:
	$(COMPOSE) $(COMPOSE_PROD) down

rebuild: down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f scainer

ps:
	$(COMPOSE) ps
