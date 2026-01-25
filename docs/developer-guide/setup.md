# Development Setup

This guide covers setting up your development environment for OSPA.

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | [Download Go](https://go.dev/dl/) |
| Git | Any recent | For cloning and version control |
| OpenStack access | Any | For E2E testing (optional) |
| golangci-lint | Latest | For linting (optional) |

## Clone and Build

```bash
# Clone the repository
git clone https://github.com/OpenStack-Policy-Agent/OSPA.git
cd OSPA

# Download dependencies
go mod download

# Build the agent
go build -o ospa ./cmd/agent

# Build the scaffold tool
go build -o ospa-scaffold ./cmd/scaffold
```

## Verify Setup

```bash
# Run unit tests
go test ./...

# Check the agent works
./ospa --help

# Check the scaffold tool works
./ospa-scaffold --help
```

## Development Tools

### Go Tools

```bash
# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Install dlv (debugger)
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Editor Setup

#### VS Code

Install the Go extension and configure:

```json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v"],
  "editor.formatOnSave": true
}
```

#### GoLand

- Enable "Reformat code on save"
- Configure golangci-lint integration
- Set up run configurations for agent and tests

## Environment Setup

### For Running Agent

```bash
# OpenStack credentials
export OS_CLIENT_CONFIG_FILE=/path/to/clouds.yaml
export OS_CLOUD=mycloud
```

### For E2E Testing

```bash
# Same as above, plus:
export OSPA_E2E_POLICY=./test-policy.yaml  # Optional
```

### For Development

```bash
# Enable Go modules (default in Go 1.21+)
export GO111MODULE=on

# Set GOPATH if needed
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

## Project Structure

```
OSPA/
├── cmd/                    # CLI applications
│   ├── agent/              # Main agent
│   │   └── main.go
│   └── scaffold/           # Code generator
│       ├── main.go
│       └── internal/
├── pkg/                    # Library code
│   ├── audit/              # Auditors
│   ├── auth/               # Authentication
│   ├── discovery/          # Discoverers
│   ├── orchestrator/       # Coordination
│   ├── policy/             # Policy handling
│   ├── remediate/          # Actions
│   ├── report/             # Output
│   └── services/           # Service registry
├── e2e/                    # End-to-end tests
├── examples/               # Example policies
├── docs/                   # Documentation
├── scripts/                # Helper scripts
├── go.mod                  # Go module definition
└── go.sum                  # Dependency checksums
```

## Common Development Tasks

### Running the Agent

```bash
# From source
go run ./cmd/agent --cloud mycloud --policy examples/policies.yaml --out findings.json

# Using binary
./ospa --cloud mycloud --policy examples/policies.yaml --out findings.json
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/audit/neutron/... -v

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Formatting

```bash
# Format all files
gofmt -w .

# Use goimports (also organizes imports)
goimports -w .
```

### Building

```bash
# Build agent
go build -o ospa ./cmd/agent

# Build with version info
go build -ldflags "-X main.version=v1.0.0" -o ospa ./cmd/agent

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o ospa-linux ./cmd/agent
```

## Makefile Targets

If available:

```bash
make build        # Build binaries
make test         # Run tests
make lint         # Run linter
make coverage     # Generate coverage report
make scaffold     # Run scaffold tool
make clean        # Clean build artifacts
```

## Debugging

### Using Delve

```bash
# Debug agent
dlv debug ./cmd/agent -- --cloud mycloud --policy policy.yaml

# Debug specific test
dlv test ./pkg/audit/neutron/... -- -test.run TestSecurityGroup
```

### VS Code Launch Configuration

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Agent",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/cmd/agent",
      "args": ["--cloud", "mycloud", "--policy", "policy.yaml"],
      "env": {
        "OS_CLIENT_CONFIG_FILE": "/path/to/clouds.yaml"
      }
    },
    {
      "name": "Debug Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/pkg/audit/neutron/",
      "args": ["-test.run", "TestSecurityGroup"]
    }
  ]
}
```

## Troubleshooting

### Module Issues

```bash
# Clear cache and re-download
go clean -modcache
go mod download

# Verify modules
go mod verify

# Tidy dependencies
go mod tidy
```

### Build Errors

```bash
# Check Go version
go version

# Verify all dependencies
go build -v ./...
```

### Test Failures

```bash
# Run with verbose output
go test -v ./...

# Run specific test with extra verbosity
go test -v -run TestName ./pkg/... 2>&1 | tee test.log
```

## Next Steps

After setup:

1. Read the [Architecture](architecture.md) guide
2. Try adding a [new service](adding-services.md)
3. Learn the [testing](testing.md) approach

