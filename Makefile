# Define the binary name and output directory
BINARY_NAME=fs
OUTPUT_DIR=bin

# Default target to build the project
build:
	@go build -o $(OUTPUT_DIR)/$(BINARY_NAME)

run: build
	@./$(OUTPUT_DIR)/$(BINARY_NAME)

test: 
	@go test ./... -v