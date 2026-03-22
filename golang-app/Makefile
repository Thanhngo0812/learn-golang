
build:
	go build -o bin/server cmd/server/main.go

run:
	go run cmd/server/main.go

# Docker Automation
COMPOSE_FILE := deployments/docker/docker-compose.yml

docker-up:
	docker-compose -f $(COMPOSE_FILE) up -d --build

docker-down:
	docker-compose -f $(COMPOSE_FILE) down

docker-clean:
	docker-compose -f $(COMPOSE_FILE) down -v

docker-db:
	docker-compose -f $(COMPOSE_FILE) up -d db-golang

docker-app:
	docker-compose -f $(COMPOSE_FILE) up -d app-golang

stop-db:
	docker-compose -f $(COMPOSE_FILE) stop db-golang

stop-app:
	docker-compose -f $(COMPOSE_FILE) stop app-golang

docker-prune:
	docker builder prune -a -f
	docker system prune -f
