BIN_NAME ?= subscription-service
MAIN_PATH ?= ./cmd/main.go
SWAG_ARGS ?= -o docs -d cmd,internal/models,internal/api
DOCKER_IMAGE ?= subscription-service:latest
LOG_FILE_PREFIX ?= app
LOG_DIR ?= logs

.PHONY: build run test swagger docker compose deploy clean

build:
	go build -o $(BIN_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test ./...

swagger:
	swag init $(SWAG_ARGS)

docker:
	docker build -t $(DOCKER_IMAGE) .

compose:
	docker compose up -d

deploy:
	@echo "//TODO: _"

clean:
	rm -rf -- $(BIN_NAME) $(LOG_FILE_PREFIX)*.log $(LOG_DIR)/
