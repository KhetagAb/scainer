COMPOSE_PROD = -f docker-compose.yml -f docker-compose.prod.yml

.PHONY: up down rebuild destroy logs ps up-prod down-prod

up:
	podman-compose up -d --build --force-recreate

down:
	podman-compose down --remove-orphans

destroy:
	-podman-compose down -v --remove-orphans 2>/dev/null
	@for rt in podman docker; do \
		if command -v $$rt >/dev/null 2>&1; then \
			$$rt ps -a --format '{{.Names}}' 2>/dev/null | grep -E '^scainer[-_]' | while read -r n; do \
				$$rt rm -f "$$n" 2>/dev/null || true; \
			done; \
			$$rt pod rm -f pod_scainer 2>/dev/null || true; \
			$$rt network rm scainer_default 2>/dev/null || true; \
		fi; \
	done

rebuild: down
	podman-compose build --no-cache scainer nginx
	podman-compose up -d --force-recreate

up-prod:
	docker-compose $(COMPOSE_PROD) up -d --build

down-prod:
	docker-compose $(COMPOSE_PROD) down --remove-orphans

logs:
	podman-compose logs -f scainer

ps:
	podman-compose ps
