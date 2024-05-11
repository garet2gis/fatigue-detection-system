MIGRATION_FOLDER=$(CURDIR)/migrations

ifeq ($(DEBUG_DEPLOY_FOLDER),)
	DEBUG_DEPLOY_FOLDER := ./deploys/debug/docker-compose.yaml
endif

ifeq ($(TEST_DEPLOY_FOLDER),)
	TEST_DEPLOY_FOLDER := ./deploys/test/docker-compose.yaml
endif

ifeq ($(POSTGRES_SETUP_DEBUG),)
	POSTGRES_SETUP_DEBUG := user=debug password=debug dbname=debug host=0.0.0.0 port=5436 sslmode=disable
endif

ifeq ($(POSTGRES_SETUP_TEST),)
	POSTGRES_SETUP_TEST := user=test password=test dbname=test host=0.0.0.0 port=5440 sslmode=disable
endif

.PHONY: migration-create
migration-create:
	goose -dir "$(MIGRATION_FOLDER)" create "$(name)" go

.PHONY: start-debug-containers
start-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" up --build

.PHONY: stop-debug-containers
stop-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" down

.PHONY: start-test-containers
start-test-containers:
	docker-compose -f "$(TEST_DEPLOY_FOLDER)" up --build

.PHONY: stop-test-containers
stop-test-containers:
	docker-compose -f "$(TEST_DEPLOY_FOLDER)" down

.PHONY: debug-migration-up
debug-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_DEBUG)" up

.PHONY: test-migration-up
test-migration-up:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" up

.PHONY: debug-migration-down
debug-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_DEBUG)" down

.PHONY: test-migration-down
test-migration-down:
	goose -dir "$(MIGRATION_FOLDER)" postgres "$(POSTGRES_SETUP_TEST)" down

.PHONY: swagger
swagger:
	swag fmt && \
	swag init -g cmd/app/main.go --parseInternal --md docs/

.PHONY: markdown-docs
markdown-docs:
	swagger generate markdown -f docs/swagger.json --with-expand --output docs/docs.md

.PHONY: docs
docs: swagger markdown-docs

.PHONY: int-tests
int-tests:
	go test -p 1 --tags=integration -coverprofile=coverage.out ./...

.PHONY: html-cover
html-cover:
	go tool cover -html=coverage.out

.PHONY: int-tests/cover
int-tests/cover: int-tests html-cover
