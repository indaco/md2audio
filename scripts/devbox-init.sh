#!/usr/bin/env sh
# devbox-init.sh — one-time setup for md2audio development environment
set -eu

# -------- Load logging utilities --------
. "$(cd "$(dirname "$0")" && pwd)/lib/logger.sh"

# -------- Welcome --------
h1 'Welcome to the md2audio devbox!'

log_default ""

# -------- Go Environment --------
log_info 'Setting up Go environment...'
if command_exists go; then
  export GOROOT=$(go env GOROOT)
  log_success "GOROOT set to: $GOROOT"

  # Download Go modules
  if [ -f "go.mod" ]; then
    log_info 'Downloading Go modules...'
    if go mod download; then
      log_success 'Go modules downloaded'
    else
      log_warning 'Failed to download Go modules'
    fi
  else
    log_warning 'go.mod not found'
  fi

  # Install development tools
  log_info 'Installing Go development tools...'
  if go install golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest; then
    log_success 'go-modernize installed'
  else
    log_warning 'Failed to install go-modernize'
  fi
else
  log_error 'Go not found! Please install Go 1.25 or later'
  exit 1
fi

log_default ""

# -------- Git Hooks (prek) --------
log_info 'Setting up prek hooks...'
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  # Make custom hook scripts executable
  if [ -d "scripts/githooks" ]; then
    find scripts/githooks -type f -exec chmod +x {} \;
  fi

  # Install prek hooks
  if command_exists prek; then
    if prek install; then
      log_success 'prek hooks installed'
    else
      log_warning 'Failed to install prek hooks'
    fi
  else
    log_warning 'prek not found — install with: go install github.com/j178/prek@latest'
  fi
else
  log_warning 'Not a git repository — skipping hooks installation'
fi

log_default ""

# -------- Build Test --------
log_info 'Testing build...'
if go build -o md2audio ./cmd/md2audio; then
  log_success 'Build successful!'
  rm -f md2audio  # Clean up test binary
else
  log_warning 'Build failed — check your code'
fi

log_default ""

# -------- Helpful Commands --------
h3 'Available just commands:'
cat <<'TXT'
 Development:
    just build          - Build the binary
    just dev            - Quick build and test cycle
    just fmt            - Format code
    just vet            - Run go vet
    just modernize      - Run go-modernize
    just check          - Run all quality checks (modernize, fmt, vet, lint)

 Running:
    just demo           - Run with demo_script_example.md
    just voices         - List available macOS voices
    just run FILE       - Run with custom file
    just run FILE voice="us-female" rate="170"  - With custom parameters

 Testing:
    just test           - Run tests with coverage report (shows total %)
    just test-verbose   - Run tests with verbose output
    just test-coverage  - Run tests and open HTML coverage in browser
    just test-force     - Clean cache and run tests

 Maintenance:
    just download       - Download Go modules
    just tidy           - Tidy Go modules
    just clean          - Remove build artifacts
    just install        - Install binary to $GOPATH/bin

 Project Info:
    just tree           - Show project structure
    just --list         - Show all available commands

 Quick start: `just build && just demo` to build and test with demo file!

 Stack:
  - Language: Go 1.25+
  - Platform: macOS (uses `say`, `afinfo`, `afconvert`)
  - Task Runner: just
  - Git Hooks: prek
  - Architecture: Modular internal packages
TXT

# -------- End --------
log_default ""
h2 'Setup complete!'
log_default "Run 'just build' to get started or 'just demo' to see it in action"
log_default ""
