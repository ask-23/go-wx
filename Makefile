.PHONY: build run clean docker docker-compose test

# Build the application
build:
	go build -o bin/go-wx ./cmd/go-wx

# Run the application
run: build
	./bin/go-wx

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f go-wx

# Build Docker image
docker:
	docker build -t go-wx .

# Run with Docker Compose
docker-compose:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Help command
help:
	@echo "Available commands:"
	@echo "  make build            - Build the application"
	@echo "  make run              - Run the application"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make docker           - Build Docker image"
	@echo "  make docker-compose   - Run with Docker Compose"
	@echo "  make docker-compose-down - Stop Docker Compose services"
	@echo "  make test             - Run tests"
	@echo "  make deps             - Install dependencies"
	@echo "  make fmt              - Format code"
	@echo "  make lint             - Run linter" 