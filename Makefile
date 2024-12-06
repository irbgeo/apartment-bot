
include .env

SERVER_NAME=apartment-bot-server
CLIENT_NAME=apartment-bot-client

env:
	$(.env)
	@echo SERVER_VERSION:$(SERVER_VERSION)
	@echo CLIENT_VERSION:$(CLIENT_VERSION)
	@echo MESSAGE_VERSION:$(MESSAGE_VERSION)
	@echo AVAILABLE_CITIES:$(AVAILABLE_CITIES)
	@echo TELEGRAM_BOT_SECRET:$(TELEGRAM_BOT_SECRET)
	@echo MONGO_URL:$(MONGO_URL)
	@echo MONGO_INITDB_ROOT_PASSWORD:$(MONGO_INITDB_ROOT_PASSWORD)
	@echo "\n"
	
generate:
	go generate ./...

lint:
	go fmt ./...
	protolint lint -fix .
	
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --fix

test: generate lint
	go test -v ./...

build-and-push-service:
	docker build  -f ./cmd/$(SERVICE)/Dockerfile -t irbgeo/$(SERVICE):$(VERSION) .
	docker push irbgeo/$(SERVICE):$(VERSION)

build-and-push-all: env
	$(MAKE) build-and-push-service SERVICE=$(SERVER_NAME) VERSION=$(SERVER_VERSION)
	$(MAKE) build-and-push-service SERVICE=$(CLIENT_NAME) VERSION=$(CLIENT_VERSION)

run-test: env 
	docker compose -f docker-compose.test.yaml up -d --build

stop-test: env
	docker compose -f docker-compose.test.yaml down

run: env
	docker compose up -d

setup:
	ansible-playbook -i deploy/inventory/server deploy/playbooks/setup.yaml



