
include .env

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
	docker build --platform linux/amd64 -f ./cmd/$(SERVICE)/Dockerfile -t irbgeo/$(SERVICE):$(VERSION) .
	docker push irbgeo/$(SERVICE):$(VERSION)

build-and-push-all: env
	$(MAKE) build-and-push-service SERVICE="server" VERSION=$(SERVER_VERSION)
	$(MAKE) build-and-push-service SERVICE="client" VERSION=$(CLIENT_VERSION)
	$(MAKE) build-and-push-service SERVICE="message" VERSION=$(MESSAGE_VERSION)

run-test: env 
	docker compose -f ./docker-compose.test.yaml up -d --build

stop-test: env
	docker compose -f ./docker-compose.test.yaml down

run:
	docker compose up -d


