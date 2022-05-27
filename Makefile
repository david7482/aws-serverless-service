all: build

GIT_BRANCH=$(shell git branch | grep \* | cut -d ' ' -f2)
GIT_REV=$(shell git rev-parse HEAD | cut -c1-7)
BUILD_DATE=$(shell date +%Y-%m-%d.%H:%M:%S)
EXTRA_LD_FLAGS=-X main.AppVersion=${GIT_BRANCH}-${GIT_REV} -X main.AppBuild=${BUILD_DATE}

GOLANGCI_LINT_VERSION="v1.42.1"

DATABASE_DSN="postgresql://chatbot_test:chatbot_test@localhost:5432/chatbot_test?sslmode=disable"
AWS_EVENTBRIDGE_NAME="message-bus-production"

# Setup test packages
TEST_PACKAGES = ./internal/...

clean:
	rm -rf bin/ cover.out

test:
	go vet $(TEST_PACKAGES)
	go test -race -cover -coverprofile cover.out $(TEST_PACKAGES)
	go tool cover -func=cover.out | tail -n 1

lint:
	@if [ ! -f ./bin/golangci-lint ]; then \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s $(GOLANGCI_LINT_VERSION); \
	fi;
	@echo "golangci-lint checking..."
	@./bin/golangci-lint -v run $(TEST_PACKAGES) ./cmd/...

mock:
	@which mockgen > /dev/null || (echo "No mockgen installed. Try: go install github.com/golang/mock/mockgen@v1.6.0"; exit 1)
	@echo "generating mocks..."
	@go generate ./...

build:
	go build -ldflags '${EXTRA_LD_FLAGS}' -o bin/chatbot        ./cmd/chatbot-service
	go build -ldflags '${EXTRA_LD_FLAGS}' -o bin/chatbot-worker ./cmd/chatbot-worker

run: build
	./bin/chatbot \
	--database_dsn=$(DATABASE_DSN) \
	--aws_eventbridge_name=$(AWS_EVENTBRIDGE_NAME) \

# Migrate db up to date
migrate-db:
	docker run --rm -v $(shell pwd)/migration:/migration --network host migrate/migrate -verbose -path=/migration/ -database=$(DATABASE_DSN) up

docker:
	docker buildx build \
	--platform="linux/amd64,linux/arm64" \
	-t 553321195691.dkr.ecr.us-west-2.amazonaws.com/chatbot-serverless-service:production \
	-t 553321195691.dkr.ecr.us-west-2.amazonaws.com/chatbot-serverless-service:production-${GIT_REV} \
	-f Dockerfile . \
	--push