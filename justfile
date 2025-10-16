# Default target: run CLI in simulated mode
default: simulate

# Get git version string
git-version := `git describe --always --long --dirty`

# Build CLI binary with git version
build-cli:
    go build -ldflags "-X main.version={{git-version}}" -o sclipi ./cmd/cli

# Build server binary with git version
build-server:
    go build -ldflags "-X main.version={{git-version}}" -o scwipi ./cmd/server

# Build both binaries
build: build-cli build-server

# Build CLI for Windows
build-cli-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version={{git-version}}" -o sclipi.exe ./cmd/cli

# Build server for Windows
build-server-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version={{git-version}}" -o scwipi.exe ./cmd/server

# Build both for Windows
build-windows: build-cli-windows build-server-windows

# Install CLI locally
install-cli:
    go install -v -ldflags "-X main.version={{git-version}}" ./cmd/cli

# Install server locally
install-server:
    go install -v -ldflags "-X main.version={{git-version}}" ./cmd/server

# Install both locally
install: install-cli install-server

# Run tests
test:
    go test -v ./...

# Run benchmarks
bench:
    go test -bench . ./...

# Run tests with coverage
cover:
    go test -cover ./...

# Run CLI in simulated mode
simulate:
    go run ./cmd/cli -s -q

# Run server
run-server:
    go run ./cmd/server

# Clean build artifacts
clean:
    rm -f sclipi sclipi.exe scwipi scwipi.exe
