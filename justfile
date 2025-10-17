# Default target: run CLI in simulated mode
default: simulate

# Get git version string
git-version := `git describe --always --long --dirty`

# Build CLI binary with git version
build-cli:
    go build -ldflags "-X main.version={{git-version}}" -o sclipi ./cmd/cli

# Build server binary with git version
build-server:
    go build -ldflags "-X main.version={{git-version}}" -o scpi-server ./cmd/server

# Build Angular application
build-web:
    cd web && npm run build

# Build all projects/binaries
build: build-cli build-server build-web

# Build CLI for Windows
build-cli-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version={{git-version}}" -o sclipi.exe ./cmd/cli

# Build server for Windows
build-server-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version={{git-version}}" -o scpi-server.exe ./cmd/server

# Build both for Windows
build-windows: build-cli-windows build-server-windows

test-go:
    go test -v ./...

test-web:
    cd web && npm test

test: test-go test-web

bench:
    go test -bench . ./...

# Run CLI in simulated mode
simulate:
    go run ./cmd/cli -s -q

run-server:
    go run ./cmd/server

watch-server:
    fd -e go | entr -r just run-server

install-web:
    cd web && npm install

# Serve Angular application in development mode
serve-web:
    cd web && npm start

# Clean build artifacts
clean:
    rm -f sclipi sclipi.exe scpi-server scpi-server.exe
    rm -rf web/dist/
