# Contributing

This guide explains how to contribute to OSPA effectively.

## Before You Start

1. **Read the Code of Conduct** - Be respectful and constructive
2. **Check existing issues** - Your idea might already be discussed
3. **Open an issue first** - For significant changes, discuss before coding

## Development Setup

See the [Development Setup](../developer-guide/setup.md) guide for detailed instructions.

### Quick Setup

```bash
# Clone the repository
git clone https://github.com/OpenStack-Policy-Agent/OSPA.git
cd OSPA

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./cmd/agent
```

## Contribution Workflow

### 1. Fork and Clone

```bash
# Fork on GitHub, then clone your fork
git clone https://github.com/YOUR-USERNAME/OSPA.git
cd OSPA

# Add upstream remote
git remote add upstream https://github.com/OpenStack-Policy-Agent/OSPA.git
```

### 2. Create a Branch

```bash
# Update main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-feature
```

Branch naming conventions:

| Prefix | Use For |
|--------|---------|
| `feature/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation |
| `refactor/` | Code refactoring |
| `test/` | Test additions/fixes |

### 3. Make Changes

- Write clean, idiomatic Go code
- Follow existing code style
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run unit tests
go test ./...

# Run with race detector
go test -race ./...

# Check code formatting
gofmt -d .

# Run linter (if installed)
golangci-lint run
```

### 5. Commit

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add support for Glance image auditing

- Add GlanceImageDiscoverer for image discovery
- Add ImageAuditor for policy checks
- Add unit tests for new components
- Update documentation"
```

Commit message format:

```
<type>: <short summary>

<longer description if needed>

<references to issues if applicable>
```

### 6. Push and Create PR

```bash
git push origin feature/my-feature
```

Then create a pull request on GitHub.

## Code Style

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Keep functions focused and small
- Add comments for exported functions

### Example

```go
// ImageAuditor audits Glance images against policy rules.
// It implements the audit.Auditor interface.
type ImageAuditor struct{}

// ResourceType returns the resource type this auditor handles.
func (a *ImageAuditor) ResourceType() string {
    return "image"
}

// Check evaluates an image against the given policy rule.
func (a *ImageAuditor) Check(ctx context.Context, resource interface{}, rule *policy.Rule) (*audit.Result, error) {
    image, ok := resource.(images.Image)
    if !ok {
        return nil, fmt.Errorf("expected images.Image, got %T", resource)
    }
    
    // ... implementation
}
```

### Error Handling

```go
// Good: Return errors with context
if err != nil {
    return nil, fmt.Errorf("failed to list images: %w", err)
}

// Bad: Panic
if err != nil {
    panic(err)
}
```

### Testing

```go
func TestImageAuditor_Check(t *testing.T) {
    tests := []struct {
        name     string
        image    images.Image
        rule     *policy.Rule
        want     bool
        wantErr  bool
    }{
        {
            name: "matches active status",
            image: images.Image{
                ID:     "img-1",
                Status: "active",
            },
            rule: &policy.Rule{
                Check: policy.CheckConditions{Status: "active"},
            },
            want: false, // not compliant = flagged
        },
        // ... more cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            a := &ImageAuditor{}
            got, err := a.Check(context.Background(), tt.image, tt.rule)
            if (err != nil) != tt.wantErr {
                t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got.Compliant != tt.want {
                t.Errorf("Check() compliant = %v, want %v", got.Compliant, tt.want)
            }
        })
    }
}
```

## Documentation

### When to Update Docs

- Adding new features
- Changing existing behavior
- Adding new services or resources
- Fixing unclear documentation

### Documentation Files

| Location | Content |
|----------|---------|
| `docs/` | User and developer documentation |
| `examples/` | Example policies and guides |
| Code comments | API documentation |

### Documentation Style

- Use clear, simple language
- Include code examples
- Keep paragraphs short
- Use tables for structured data

### Feature Contributions

For larger features:

1. Open an issue to discuss the approach
2. Wait for maintainer feedback
3. Implement according to agreed design
4. Include tests and documentation

### Service Contributions

Adding new OpenStack services:

1. Use the scaffold tool
2. Implement discoverers and auditors
3. Add comprehensive tests
4. Document the new service

See [Adding Services](../developer-guide/adding-services.md) for details.

## Questions?

- **GitHub Discussions** - For general questions
- **Issue Tracker** - For bugs and features
- **Pull Request Comments** - For PR-specific questions

Thank you for contributing to OSPA!

