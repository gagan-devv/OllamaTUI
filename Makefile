# Variables to make the script reusable
BINARY_NAME=chat
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

# Install the binary to your system (GOPATH/bin)
install:
	@echo "Installing to your system path..."
	go install .

release:
	GOOS=windows GOARCH=amd64 go build -o chat.exe main.go
	GOOS=darwin GOARCH=arm64 go build -o chat-mac main.go

# Help command to show available options
help:
	@echo "Available commands:"
	@echo "  make build   - Compile the project"
	@echo "  make run     - Build and run immediately"
	@echo "  make clean   - Remove binary and logs"
	@echo "  make install - Move binary to your global path"