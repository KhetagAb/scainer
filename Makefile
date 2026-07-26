COMPOSE ?= docker-compose
COMPOSE_PROD = -f docker-compose.yml -f docker-compose.prod.yml
IMPORT_LOGINS_SCRIPT = scainer/scripts/import-admin-login.sh

.PHONY: up down rebuild logs ps up-prod down-prod logins

up:
	$(COMPOSE) up -d --build --force-recreate

down:
	$(COMPOSE) down

up-prod:
	$(COMPOSE) $(COMPOSE_PROD) up -d --build

down-prod:
	$(COMPOSE) $(COMPOSE_PROD) down

logins:
	bash $(IMPORT_LOGINS_SCRIPT)

rebuild: down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f scainer

ps:
	$(COMPOSE) ps
