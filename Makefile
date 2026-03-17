# Variables to make the script reusable
BINARY_NAME=ollama-go
GO_FILES=main.go

# The default 'make' command runs this
all: build

# Build the binary
build:
	@echo "Building the binary..."
	go build -o $(BINARY_NAME) $(GO_FILES)

# Run the application
run: build
	./$(BINARY_NAME)

# Clean up binaries and logs
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f chat_history.md
	rm -rf test_output

# Install the binary to your system (GOPATH/bin)
install:
	@echo "Installing to your system path..."
	go install .

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	GOOS=windows GOARCH=amd64 go build -o ollama-go.exe main.go
	GOOS=darwin GOARCH=arm64 go build -o ollama-go-mac-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o ollama-go-mac-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o ollama-go-linux main.go

# Help command to show available options
help:
	@echo "Available commands:"
	@echo "  make build          - Compile the project"
	@echo "  make run            - Build and run immediately"
	@echo "  make clean          - Remove binary, logs, and test output"
	@echo "  make install        - Move binary to your global path"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make release        - Build binaries for multiple platforms"