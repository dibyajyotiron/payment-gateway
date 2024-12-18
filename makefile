# Makefile for Docker, Go Tests, and Swagger

# Variables
DOCKER_COMPOSE = docker-compose
DOCKER_COMPOSE_FILE = docker-compose.yml
SWAG_CMD = swag
GO_TEST_CMD = go test -race ./...
SWAG_DIR = cmd/main.go

# Targets

# Build Docker containers
build:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build

# Start Docker containers
up:
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up

# Run Go tests with race detection
test:
	$(GO_TEST_CMD)

# Initialize Swagger documentation
swagger:
	$(SWAG_CMD) init --generalInfo $(SWAG_DIR)

# Run all steps
all: swagger build up