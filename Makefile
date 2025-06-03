# Project variables
APP_NAME := identity-service
BUILD_DIR := app
DEPLOY_DIR := /var/deployment/identity-service
SERVICE_NAME := identity-service

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building..."
	go build -o $(BUILD_DIR)/$(APP_NAME) main.go

# Run the app (local/dev)
.PHONY: run
run:
	@echo "Running locally..."
	swag init
	go run main.go

# Test the code
.PHONY: test
test:
	@echo "Running tests..."
	go test ./...

# Clean build output
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BUILD_DIR)/$(APP_NAME)

# Tidy dependencies
.PHONY: tidy
tidy:
	go mod tidy

# Deploy to systemd directory and restart the service
.PHONY: deploy
deploy: build
	@echo "Deploying..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) $(DEPLOY_DIR)/
	sudo chown identity-service:texkin $(DEPLOY_DIR)/$(APP_NAME)
	sudo chmod 750 $(DEPLOY_DIR)/$(APP_NAME)
	sudo systemctl restart $(SERVICE_NAME)
	@echo "Deployment complete."

