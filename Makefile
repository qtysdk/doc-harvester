# Makefile for DocHarvester

# Variables
BINARY_NAME=bin/harvester

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd

# Run the binary with example arguments
run:
	./$(BINARY_NAME) --explore-only https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview

# Clean up build artifacts
clean:
	rm -f $(BINARY_NAME)

# Test the binary with a download example
test:
	./$(BINARY_NAME) --xml-output ./output/site-docs.xml https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview
