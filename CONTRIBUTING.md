# Contributing to OSPA

Thanks for helping improve OSPA. This guide covers the basics for contributors.

## Prerequisites

- Go version compatible with `go.mod`
- OpenStack access for e2e tests (optional)

## Development Setup

```bash
git clone <repo>
cd OSPA
go mod download
go test ./...
```

## Running Locally

```bash
go run ./cmd/agent --cloud "$OS_CLOUD" --policy ./examples/policies.yaml --out findings.jsonl
```

## Adding a Service

- Use `go run ./cmd/scaffold --list` to see supported services/resources
- Use `go run ./cmd/scaffold --service <name> --resources <r1,r2> --type <serviceType>`
- See `docs/DEVELOPMENT.md` for details

## Tests

```bash
go test ./...
go test ./pkg/... -count=1
go test ./cmd/scaffold/... -count=1
```

## Pull Requests

- Keep PRs focused and small
- Add tests when behavior changes
- Update docs for user-visible changes
- Ensure `go test ./...` passes

