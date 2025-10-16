# Default target: run in simulated mode
default: simulate

# Get git version string
git-version := `git describe --always --long --dirty`

# Build the binary with git version
build:
    go build -ldflags "-X main.version={{git-version}}" .

# Build for Windows
build-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version={{git-version}}" .

# Install locally
install:
    go install -v -ldflags "-X main.version={{git-version}}" .

# Run tests
test:
    go test -v

# Run benchmarks
bench:
    go test -bench .

# Run tests with coverage
cover:
    go test -cover

# Run in simulated mode
simulate:
    go run . -s -q

# Clean build artifacts
clean:
    rm -f sclipi sclipi.exe
