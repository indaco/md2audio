# Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Development Environment Setup

### Using Devbox (Recommended)

To set up a development environment for this repository, you can use [devbox](https://www.jetify.com/devbox) along with the provided `devbox.json` configuration file.

1. Install devbox by following [these instructions](https://www.jetify.com/devbox/docs/installing_devbox/).
2. Clone this repository to your local machine.

   ```bash
   git clone https://github.com/indaco/md2audio.git
   cd md2audio
   ```

3. Run `devbox install` to install all dependencies specified in `devbox.json`.
4. Enter the environment with `devbox shell --pure`.
5. Start developing, testing, and contributing!

### Manual Setup

If you prefer not to use Devbox, ensure you have the following tools installed:

- [Go 1.25+](https://go.dev/dl/): The project requires Go 1.25 or later
- [golangci-lint](https://golangci-lint.run/): For linting Go code
- [just](https://github.com/casey/just): Task runner for project tasks
- [prek](https://github.com/j178/prek): Git hooks framework for code quality (Rust-based `pre-commit` alternative)
- [modernize](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize): Run the modernizer analyzer to simplify code by using modern constructs
- macOS: Required for `say`, `afinfo`, and `afconvert` commands

## Setting Up Git Hooks

This project uses [prek](https://github.com/j178/prek) to manage Git hooks for enforcing code quality and commit message format. The hooks are configured in `.pre-commit-config.yaml`.

### Using Devbox

If using `devbox`, prek hooks are automatically installed when you run `devbox shell`. The `devbox-init.sh` script handles the setup. No further action is required.

### Manual Setup

For users not using `devbox`, ensure you have [prek](https://github.com/j178/prek) installed (see prerequisites above), then:

1. Clone the repository:

   ```bash
   git clone https://github.com/indaco/md2audio.git
   cd md2audio
   ```

2. Install the prek hooks:

   ```bash
   prek install
   ```

3. (Optional) Run hooks manually against all files:

   ```bash
   prek run --all-files
   ```

### Configured Hooks

The project uses the following hooks:

- **commit-msg**: Validates commit message format using `scripts/githooks/commit-msg`
- **pre-commit/pre-push**: Runs code quality checks via `just check` (modernize, fmt, vet, lint)

## Running Tasks

This project uses [just](https://github.com/casey/just) as a task runner. A `justfile` is provided with common development tasks.

### View all available tasks

```bash
just --list
```

### Common tasks

```bash
# Development
just build              # Build the binary
just dev                # Quick build and test cycle
just fmt                # Format code with go fmt
just vet                # Run go vet
just lint               # Run golangci-lint
just modernize          # Run go-modernize
just check              # Run all quality checks (modernize, fmt, vet, lint)

# Running
just demo               # Run with demo_script_example.md
just voices             # List available macOS voices
just run FILE           # Run with custom file
just run FILE voice="us-female" rate="170"  # With custom parameters

# Testing
just test               # Run tests with coverage report (shows total %)
just test-verbose       # Run tests with verbose output
just test-coverage      # Run tests and open HTML coverage in browser
just test-force         # Clean cache and run tests

# Maintenance
just download           # Download Go modules
just tidy               # Tidy Go modules
just clean              # Remove build artifacts
just install            # Install binary to $GOPATH/bin

# Project Info
just tree               # Show project structure
```
