# bootstrap initial golang project setup
.PHONY: bootstrap
bootstrap:
	mkdir .vscode scripts env && touch .vscode/launch.json .vscode/settings.json env/test.env .gitignore README.md

# setup docker network for local development
.PHONY: network
network:
	@echo "Creating local-comfforts network if it doesn't exist..."
	@if ! docker network inspect local-comfforts > /dev/null 2>&1; then \
		docker network create local-comfforts; \
		echo "Network local-comfforts created."; \
	else \
		echo "Network local-comfforts already exists."; \
	fi

# start redis sentinel for local development
.PHONY: redis
redis:
	@echo "Creating Redis..."
	docker-compose -f deploy/comfforts/docker-compose.yml up -d redis-sentinel

# wait for external dependencies to initialize
.PHONY: wait-30
wait-30:
	@echo "Waiting 30 seconds..."
	sleep 30

.PHONY: wait-10
wait-10:
	@echo "Waiting 10 seconds..."
	sleep 10

# start local network and redis
.PHONY: base
base: network wait-10 redis

# stop all services and remove containers
.PHONY: down
down:
	@echo "Shutting down all services..."
	docker-compose -f deploy/comfforts/docker-compose.yml down
