# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Sclipi is a command-line tool for sending SCPI (Standard Commands for Programmable Instruments) commands to test and measurement devices. It provides an interactive shell with auto-completion, history support, and both interactive and non-interactive operation modes.

## Build System

This project uses just as its build tool. Build targets are defined in `justfile`.

### Common Commands

```bash
# Run in simulated mode (default target)
just

# Build the binary with git version
just build

# Run tests
just test

# Run tests with coverage
just cover

# Run benchmarks
just bench

# Install locally
just install

# Build for Windows
just build-windows

# Clean build artifacts
just clean
```

### Running Tests

```bash
# All tests
go test -v

# Specific test file
go test -v -run TestFunctionName

# With coverage
go test -cover
```

### Development with Nix

```bash
# Enter development shell with Go tooling
nix develop

# Build the Nix package
nix build
```

## Architecture

### Core Components

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

### Data Flow

1. User enters command in interactive prompt (powered by `go-prompt` library)
2. `scpiManager.completer()` provides suggestions by traversing the SCPI tree
3. `scpiManager.executor()` processes the complete command
4. For SCPI commands (`:` or `*` prefix), routes to `instrument.command()` or `instrument.query()`
5. Instrument sends command over TCP, reads response, checks for errors
6. Response displayed to user and stored in history

### Simulation Mode

When run with `-s/--simulate` or address "simulated", reads from `SCPI.txt` (`:SYSTem:HELP:HEADers?` format) to build the command tree without requiring a real instrument connection. Useful for testing and development.

## Version Management

Version is injected at build time via `-ldflags "-X main.version=..."` using git describe output (`git describe --always --long --dirty`).
