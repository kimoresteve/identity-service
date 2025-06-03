# Project variables
APP_NAME := identity-service
BUILD_DIR := app

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building..."
	go build -o $(BUILD_DIR)/$(APP_NAME) main.go

# Run the app
.PHONY: run
run:
	@echo "Running..."
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
