
COMPOSE_FILE := docker-compose.yml
GO_CMD := go
GO_RUN := $(GO_CMD) run
MAIN_FILE := cmd/api/main.go

.PHONY: all
all: db-up run

.PHONY: db-up
db-up:
	docker-compose -f $(COMPOSE_FILE) up -d

.PHONY: db-down
db-down:
	docker-compose -f $(COMPOSE_FILE) down

.PHONY: db-clean
db-clean:
	docker-compose -f $(COMPOSE_FILE) down -v

.PHONY: run
run:
	$(GO_RUN) $(MAIN_FILE)

.PHONY: build
build:
	$(GO_CMD) build -o osquery-mvp $(MAIN_FILE)

.PHONY: clean
clean:
	rm -f osquery-mvp
	go clean

.PHONY: reset
reset: db-clean clean
