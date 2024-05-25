ifeq ($(DEBUG_DEPLOY_FOLDER),)
	DEBUG_DEPLOY_FOLDER := ./deploys/debug/docker-compose.yaml
endif

ifeq ($(FULL_DEPLOY_FOLDER),)
	FULL_DEPLOY_FOLDER := ./deploys/full/docker-compose.yaml
endif

ifeq ($(LOAD_TEST_DEPLOY_FOLDER),)
	LOAD_TEST_DEPLOY_FOLDER := ./deploys/load_testing/docker-compose.yaml
endif

.PHONY: start-debug-containers
start-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" up --build

.PHONY: stop-debug-containers
stop-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" down

.PHONY: start-full-containers
start-full-containers:
	docker-compose -f "$(FULL_DEPLOY_FOLDER)" up --build

.PHONY: stop-full-containers
stop-full-containers:
	docker-compose -f "$(FULL_DEPLOY_FOLDER)" down

.PHONY: start-test-containers
start-test-containers:
	docker-compose -f "$(LOAD_TEST_DEPLOY_FOLDER)" up --build

.PHONY: stop-test-containers
stop-test-containers:
	docker-compose -f "$(LOAD_TEST_DEPLOY_FOLDER)" down