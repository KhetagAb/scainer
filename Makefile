COMPOSE ?= podman-compose
COMPOSE_PROD = -f docker-compose.yml -f docker-compose.prod.yml
IMPORT_LOGINS_SCRIPT = scainer/scripts/import-admin-login.sh

.PHONY: up down rebuild logs ps up-prod down-prod import-logins import-logins-dry

up:
	$(COMPOSE) up -d --build --force-recreate

down:
	$(COMPOSE) down

up-prod:
	$(COMPOSE) $(COMPOSE_PROD) up -d --build

down-prod:
	$(COMPOSE) $(COMPOSE_PROD) down

import-logins:
	bash $(IMPORT_LOGINS_SCRIPT)

import-logins-dry:
	bash $(IMPORT_LOGINS_SCRIPT) --dry-run

rebuild: down
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f scainer

ps:
	$(COMPOSE) ps
