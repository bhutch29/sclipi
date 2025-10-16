# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sclipi is a tool for sending SCPI (Standard Commands for Programmable Instruments) commands to test and measurement devices. The project consists of two main components:

1. **CLI (`cmd/cli/`)**: An interactive command-line shell with auto-completion, history support, and both interactive and non-interactive operation modes
2. **Web Server (`cmd/server/`)**: An HTTP server that serves content/behavior supporting the web interface for remote SCPI device interaction

### Project Structure

```
sclipi/
├── cmd/
│   ├── cli/          # Command-line interface application
│   └── server/       # HTTP server to support Angular frontend
├── web/              # Angular web application
│   ├── src/
│   │   ├── app/      # Angular components and services
│   │   ├── index.html
│   │   ├── main.ts
│   │   └── styles.scss
│   └── package.json
├── justfile          # Build configuration
├── SCPI.txt          # Sample SCPI command tree for simulation
└── CLAUDE.md         # This file
```

## Build System

This project uses just as its build tool. Build targets are defined in `justfile`.

### Common Commands

#### Build Commands

```bash
# Run CLI in simulated mode (default target)
just

# Build CLI binary
just build-cli

# Build server binary
just build-server

# Build Angular application
just build-web

# Build all projects (CLI + server + web)
just build

# Build for Windows
just build-cli-windows
just build-server-windows
just build-windows

# Clean build artifacts
just clean
```

#### Development Commands

```bash
# Run CLI in simulated mode
just simulate

# Run HTTP server
just run-server

# Install web dependencies
just install-web

# Serve Angular app in development mode
just serve-web
```

#### Testing Commands

```bash
# Run all tests (Go + web)
just test

# Run Go tests only
just test-go

# Run web tests only
just test-web

# Run benchmarks
just bench
```

### Running Tests Manually

```bash
# All tests (Go + web)
just test

# Go tests only
just test-go

# Angular tests only
just test-web

# Specific Go test
go test -v -run TestFunctionName ./cmd/cli
```

### Development with Nix

```bash
# Enter development shell with Go tooling
nix develop

# Build the Nix package
nix build
```

## Architecture

### CLI Application (`cmd/cli/`)

The CLI application provides an interactive terminal interface for SCPI device control.

#### Core Components

**`main.go`**: Entry point that orchestrates the application flow:
- Parses command-line arguments
- Prompts for instrument address (interactive mode)
- Establishes instrument connection with progress bar
- Initializes the SCPI manager and interactive prompt

**`instrument.go`**: Defines the `instrument` interface and two implementations:
- `scpiInstrument`: Real TCP/IP connection to instruments via port 5025
- `simInstrument`: Simulated mode that reads commands from `SCPI.txt`

Key protocol features:
- Handles SCPI block data format (`#` prefix with length encoding)
- Implements error checking via `SYST:ERR?` queries after commands
- Supports configurable timeouts with progress bars for long queries

**`scpiManager.go`**: Command execution and completion engine:
- `executor()`: Routes input to handlers (SCPI commands, dash commands, shell passthrough)
- `completer()`: Provides context-aware auto-completion using the SCPI command tree
- Manages command history and clipboard operations
- Implements special commands: `-history`, `-copy`, `-save_script`, `-run_script`, `-set_timeout`, `-reconnect`

**`scpiParser.go`**: SCPI command tree builder:
- Parses `:SYSTem:HELP:HEADers?` format into a navigable tree structure
- Handles SCPI syntax features:
  - Optional segments in `[]` brackets
  - Alternative options with `|` (bar)
  - Numeric suffixes like `{1:16}` for indexed commands
  - Query-only (`?/qonly/`) and no-query (`/nquery/`) variants
- Builds `scpiNode` tree for efficient auto-completion

**`argParser.go`**: Command-line argument parsing using `akamensky/argparse`:
- Connection options: `-a/--address`, `-p/--port`, `-t/--timeout`
- Operation modes: `-s/--simulate`, `-c/--command`, `-f/--file`
- UI options: `-q/--quiet`, various color customization flags

**`ipCompleter.go`**: Auto-completion for IP address input, suggesting local network interface prefixes.

**`history.go`**: Session history management, storing commands and responses for recall and export.

**`helpers.go`**: Utility functions for file I/O and common operations.

#### Data Flow

1. User enters command in interactive prompt (powered by `go-prompt` library)
2. `scpiManager.completer()` provides suggestions by traversing the SCPI tree
3. `scpiManager.executor()` processes the complete command
4. For SCPI commands (`:` or `*` prefix), routes to `instrument.command()` or `instrument.query()`
5. Instrument sends command over TCP, reads response, checks for errors
6. Response displayed to user and stored in history

#### Simulation Mode

When run with `-s/--simulate` or address "simulated", reads from `SCPI.txt` (`:SYSTem:HELP:HEADers?` format) to build the command tree without requiring a real instrument connection. Useful for testing and development.

### Web Application (`web/`)

The web application provides a browser-based interface for SCPI device control, built with Angular 20.

#### Technology Stack

- **Angular 20**: Frontend framework with standalone components
- **TypeScript 5.9**: Type-safe JavaScript
- **RxJS 7.8**: Reactive programming for async operations
- **SCSS**: Styling
- **Karma + Jasmine**: Testing framework

#### Core Files

**`src/main.ts`**: Angular application bootstrap entry point

**`src/app/app.ts`**: Root application component

**`src/app/app.config.ts`**: Application-level configuration and providers

**`src/app/app.routes.ts`**: Client-side routing configuration

**`src/index.html`**: Main HTML template with base href for routing

**`package.json`**: Node dependencies and npm scripts

#### Development Workflow

```bash
# Install dependencies
just install-web

# Run development server (http://localhost:4200)
just serve-web

# Build for production
just build-web

# Run tests
just test-web
```

### HTTP Server (`cmd/server/`)

A lightweight Go HTTP server that serves the Angular application and provides backend API endpoints.

#### Core Components

**`main.go`**: HTTP server implementation
- Serves on port 8080
- `/health` endpoint for health checks
- Graceful shutdown support (SIGINT/SIGTERM)
- 10-second shutdown timeout

The server is designed to:
1. Serve the compiled Angular static assets from `web/dist/`
2. Provide REST API endpoints for SCPI device communication
3. Handle WebSocket connections for real-time device interaction

## Version Management

Version is injected at build time via `-ldflags "-X main.version=..."` using git describe output (`git describe --always --long --dirty`) for both CLI and server binaries.

## Development Setup

### Go Development

The project uses Go modules. All Go code is located under `cmd/`:
- `cmd/cli/` - CLI application
- `cmd/server/` - HTTP server

### Angular Development

The Angular application is in `web/`:

```bash
# First time setup
just install-web

# Development
just serve-web   # Dev server with hot reload
just test-web    # Run tests
just build-web   # Production build
```

### Integrated Development

To run the full stack locally:

```bash
# Terminal 1: Run the Go server
just run-server

# Terminal 2: Run Angular dev server with proxy
just serve-web

# Or build everything together
just build
```
