ifeq ($(OS),Windows_NT)
	EXT := .exe
else
	EXT :=
endif

BIN_CALENDAR := "./bin/calendar$(EXT)"
BIN_SCHEDULER := "./bin/calendar_scheduler$(EXT)"
BIN_STORER := "./bin/calendar_storer$(EXT)"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
BUILD_DATE := $(shell git log -1 --format="%cI")
LDFLAGS := -X main.release=develop -X main.buildDate=$(BUILD_DATE) -X main.gitHash=$(GIT_HASH)

build: build-calendar build-scheduler build-storer

build-calendar:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar

build-scheduler:
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler

build-storer:
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/calendar_storer

run: build-calendar
	$(BIN_CALENDAR) --config ./configs/calendar_config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build-calendar
	$(BIN_CALENDAR) version

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

.PHONY: build build-calendar build-scheduler build-storer run build-img run-img version test generate lint migrate-up migrate-down
