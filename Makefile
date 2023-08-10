# Go parameters
GO := /usr/local/go/bin/go
BINARY_NAME := webrtcgameserver
BINARY_FILE := ./webrtcgameserver
MAIN_FILE := ./cmd/main.go

.PHONY: all
all: clean preview

.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

.PHONY: build
build: clean
	@echo "Building binary..."
	@$(GO) build -o $(BINARY_NAME) $(MAIN_FILE)

.PHONY: run
run:
	@echo "Running application..."
	@$(GO) run $(MAIN_FILE)


.PHONY: preview
preview: build
	@echo "Running application..."
	@$(BINARY_FILE)
