# Variables
APP_NAME := "md2audio"
BUILD_DIR := "."
CMD_DIR := "cmd/md2audio"

# Default recipe (shows available commands)
default:
    @just --list

# Download Go module dependencies
download:
    @echo "Downloading Go modules..."
    go mod download

# Tidy Go modules
tidy:
    @echo "Tidying Go modules..."
    go mod tidy

# Run golangci-lint (if available)
lint:
    @echo "Running golangci-lint..."
    golangci-lint run ./...

# Run go fmt on all packages
fmt:
    @echo "Formatting code..."
    go fmt ./...

# Run go vet on all packages
vet:
    @echo "Running go vet..."
    go vet ./...

# Run go-modernize (requires: go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest)
modernize:
    @echo "Running go-modernize..."
    modernize -fix -test ./...

# Run goreportcard-cli
goreportcard:
    @echo "Running goreportcard-cli..."
    goreportcard-cli -v

# Build the binary
build:
    @echo "Building {{APP_NAME}}..."
    go build -o {{APP_NAME}} ./{{CMD_DIR}}

# Build with version information
build-release version:
    @echo "Building {{APP_NAME}} v{{version}}..."
    go build -ldflags="-X main.Version={{version}}" -o {{APP_NAME}} ./{{CMD_DIR}}

# Run all tests and generate coverage report
test:
    @bash scripts/test-coverage.sh

# View the HTML coverage report in browser
_test-view-coverage:
    go tool cover -html=coverage.txt
    @echo "Coverage report displayed in your default browser."

# Run tests with verbose output
test-verbose:
    @bash scripts/test-coverage.sh -v

# Run tests and open HTML coverage report
test-coverage: test-force _test-view-coverage

# Clean test cache and run all tests
test-force:
    @echo "Cleaning test cache..."
    go clean -testcache
    @just test

# Check code quality (modernize, fmt, vet, lint, goreportcard)
check: modernize fmt vet lint goreportcard

# Install the binary using go install
install:
    @echo "Installing {{APP_NAME}}..."
    go install ./{{CMD_DIR}}

# Run the tool with demo file
demo:
    @echo "Running demo with demo_script_example.md..."
    ./{{APP_NAME}} -f demo_script_example.md -p british-female

# List available voices
voices:
    @echo "Listing available macOS voices..."
    ./{{APP_NAME}} -list-voices

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -f {{APP_NAME}}
    rm -f coverage.txt coverage.out coverage.html
    rm -rf audio_sections

# Run the tool with custom parameters
run file voice="british-female" rate="180":
    @echo "Generating audio from {{file}}..."
    ./{{APP_NAME}} -f {{file}} -p {{voice}} -r {{rate}}

# Show project structure
tree:
    @echo "Project structure:"
    @tree -I '.git|audio_sections' -L 3

# Quick build and test cycle
dev: build
    @echo "Build successful! Binary ready at ./{{APP_NAME}}"
