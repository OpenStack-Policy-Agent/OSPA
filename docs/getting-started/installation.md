# Installation

This guide covers how to install OSPA on your system.

## Prerequisites

Before installing OSPA, ensure you have:

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.21+ | [Download Go](https://go.dev/dl/) |
| Git | Any recent | For cloning the repository |
| OpenStack access | Any | Valid credentials and API access |

## Installation Methods

### From Source (Recommended)

Clone the repository and build:

```bash
# Clone the repository
git clone https://github.com/OpenStack-Policy-Agent/OSPA.git
cd OSPA

# Download dependencies
go mod download

# Build the agent
go build -o ospa ./cmd/agent

# Build the scaffold tool (optional)
go build -o ospa-scaffold ./cmd/scaffold
```

### Verify Installation

```bash
# Check the agent
./ospa --help

# Run unit tests
go test ./...
```

## Directory Structure

After cloning, you'll see:

```
OSPA/
├── cmd/
│   ├── agent/          # Main agent CLI
│   └── scaffold/       # Code generation tool
├── pkg/
│   ├── audit/          # Auditor implementations
│   ├── auth/           # OpenStack authentication
│   ├── discovery/      # Resource discovery
│   ├── orchestrator/   # Worker coordination
│   ├── policy/         # Policy loading and validation
│   ├── remediate/      # Remediation actions
│   ├── report/         # Output formatting
│   └── services/       # Service registry
├── e2e/                # End-to-end tests
├── examples/           # Example policies
└── docs/               # Documentation
```

## Updating

To update to the latest version:

```bash
cd OSPA
git pull origin main
go mod download
go build -o ospa ./cmd/agent
```

## Troubleshooting

### Go Version Issues

If you see Go version errors:

```bash
# Check your Go version
go version

# Should be 1.21 or higher
```

### Module Download Fails

If `go mod download` fails:

```bash
# Clear module cache and retry
go clean -modcache
go mod download
```

### Build Fails

If the build fails, ensure all dependencies are available:

```bash
# Verify dependencies
go mod verify

# Tidy up modules
go mod tidy
```

## Next Steps

After installation:

1. [Configure your OpenStack credentials](configuration.md)
2. [Run your first audit](quickstart.md)

