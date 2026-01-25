# Developer Guide

This guide covers everything you need to know about developing OSPA, from initial setup to adding new OpenStack services and resources.

## Overview

OSPA follows a plugin-based architecture that makes it easy to extend:

- **Services** - Top-level OpenStack services (Nova, Neutron, Cinder)
- **Resources** - Types within services (instance, security_group, volume)
- **Discoverers** - Find resources in OpenStack
- **Auditors** - Evaluate resources against policies
- **Validators** - Validate policy syntax

## Quick Navigation

<div class="grid cards" markdown>

- **Architecture**

    ---

    Understand OSPA's design and data flow.

    [→ Architecture](architecture.md)

- **Development Setup**

    ---

    Set up your development environment.

    [→ Setup Guide](setup.md)

- **Scaffold Tool**

    ---

    Generate code for new services automatically.

    [→ Scaffold Guide](scaffold.md)

- **Adding Services**

    ---

    Add support for new OpenStack services.

    [→ Adding Services](adding-services.md)

- **Adding Resources**

    ---

    Add resource types to existing services.

    [→ Adding Resources](adding-resources.md)

- **Testing**

    ---

    Unit tests, integration tests, and E2E tests.

    [→ Testing Guide](testing.md)

- **Troubleshooting**

    ---

    Common issues and solutions.

    [→ Troubleshooting](troubleshooting.md)

</div>

## Project Structure

```
OSPA/
├── cmd/
│   ├── agent/              # Main CLI agent
│   └── scaffold/           # Code generation tool
├── pkg/
│   ├── audit/              # Auditor implementations
│   │   ├── interface.go    # Auditor interface
│   │   ├── nova/           # Nova auditors
│   │   ├── neutron/        # Neutron auditors
│   │   └── cinder/         # Cinder auditors
│   ├── auth/               # OpenStack authentication
│   ├── discovery/          # Resource discovery
│   │   ├── interface.go    # Discoverer interface
│   │   └── services/       # Service discoverers
│   ├── orchestrator/       # Worker coordination
│   ├── policy/             # Policy loading and validation
│   │   ├── policy.go       # Policy structures
│   │   ├── validator.go    # Main validator
│   │   └── validation/     # Service validators
│   ├── remediate/          # Remediation actions
│   ├── report/             # Output formatting
│   └── services/           # Service registry
│       ├── interface.go    # Service interface
│       ├── registry.go     # Service registry
│       └── services/       # Service implementations
├── e2e/                    # End-to-end tests
├── examples/               # Example policies
└── docs/                   # Documentation
```

## Development Workflow

Follow these steps to add a new service, resource, or feature to OSPA:

1. **Identify the Service/Resource**
   - Decide which OpenStack service (e.g., Nova, Neutron, Cinder) or resource you want to support.

2. **Use the Scaffold Tool (Recommended)**
   - Run the code generation tool to quickly set up new modules, interfaces, and boilerplate based on existing patterns.
   - _Or_, you may create files manually if you prefer.

3. **Customize the Generated Code**
   - Modify the scaffolded files to fit your new service or feature requirements.

4. **Implement the Discoverer**
   - Write code that can list/discover the relevant OpenStack resources.

5. **Implement the Auditor**
   - Add logic for evaluating resources against policy rules.

6. **Add Validation**
   - Ensure policy schemas and rule definitions for the new service are properly validated.

7. **Write Unit Tests**
   - Add tests for your discoverer, auditor, and validation logic.

8. **Write E2E Tests**
   - Ensure your new functionality works as expected in integration/end-to-end scenarios.

9. **Test with Agent**
   - Run the OSPA agent and test your changes in a real or test OpenStack environment.

10. **Submit a Pull Request**
    - Open a PR following the [contribution guidelines](../contributor-guide/index.md).

**Tips:**

- New features should follow the structure and style of existing services (see `pkg/audit/nova`, `pkg/discovery/services`, etc.).
- Keep PRs focused and well-documented.
- Write clear commit messages describing the reason for changes.

If you have questions, check the [Contributor Guide](../contributor-guide/index.md) or open a discussion.

## Best Practices

1. **Use the scaffold tool** - Ensures consistency and saves time
2. **Write tests first** - TDD approach catches issues early
3. **Follow existing patterns** - Look at Nova/Neutron/Cinder implementations
4. **Handle errors gracefully** - Return errors, don't panic
5. **Respect context cancellation** - Check `ctx.Done()` in loops
6. **Document your code** - Add comments to exported functions

