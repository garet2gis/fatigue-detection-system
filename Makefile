ifeq ($(DEBUG_DEPLOY_FOLDER),)
	DEBUG_DEPLOY_FOLDER := ./deploys/debug/docker-compose.yaml
endif

.PHONY: start-debug-containers
start-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" up --build

.PHONY: stop-debug-containers
stop-debug-containers:
	docker-compose -f "$(DEBUG_DEPLOY_FOLDER)" down