.PHONY: test test-verbose test-coverage build clean run


BINARY_NAME=wisdom-gate
CLIENT_BINARY_NAME=client
BUILD_DIR=build

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wisdom-gate.go

build-client:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(CLIENT_BINARY_NAME) ../client/main.go

build-all: build build-client

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html


run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Docker
docker-build:
	docker build -f docker/Dockerfile -t wisdom-gate:latest .

docker-run:
	docker run -p 8080:8080 wisdom-gate:latest
