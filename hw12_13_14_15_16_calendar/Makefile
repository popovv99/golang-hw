BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) --config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

generate:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1
	oapi-codegen -config api/config.yaml api/openapi.yaml > internal/server/http/api/api.gen.go

lint:
	go vet ./...

migrate-up:
	@echo "Apply migrations manually using goose or psql"
	@echo "Example: goose -dir migrations postgres \"user=calendar password=calendar dbname=calendar host=localhost port=5432 sslmode=disable\" up"

migrate-down:
	@echo "Rollback migrations manually using goose or psql"
	@echo "Example: goose -dir migrations postgres \"user=calendar password=calendar dbname=calendar host=localhost port=5432 sslmode=disable\" down"

.PHONY: build run build-img run-img version test generate lint migrate-up migrate-down
