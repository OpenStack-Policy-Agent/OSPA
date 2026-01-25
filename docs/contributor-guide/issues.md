# Reporting Issues

This guide explains how to report bugs, request features, and participate in issue discussions.

## Before Opening an Issue

1. **Search existing issues** - Your issue might already be reported
2. **Check documentation** - The answer might be in the docs
3. **Try latest version** - The issue might be fixed already

## Bug Reports

### When to Report

- OSPA crashes or produces errors
- Behavior doesn't match documentation
- Unexpected results from audits
- Performance issues

### How to Report

Create a new issue with the **Bug Report** template:

```markdown
## Description
A clear description of the bug.

## Steps to Reproduce
1. Create policy file with...
2. Run command...
3. Observe error...

## Expected Behavior
What you expected to happen.

## Actual Behavior
What actually happened.

## Environment
- OSPA version: (commit hash or version)
- Go version: (go version output)
- OS: (e.g., Ubuntu 22.04)
- OpenStack version: (if relevant)

## Additional Context
- Policy file (sanitized)
- Error messages
- Log output
```

### Good Bug Report Example

```markdown
## Description
OSPA crashes when auditing security groups with more than 1000 rules.

## Steps to Reproduce
1. Create a security group with 1500 rules
2. Run: `go run ./cmd/agent --policy policy.yaml --out findings.json`
3. Agent crashes with OOM error

## Expected Behavior
Agent should audit all rules without crashing.

## Actual Behavior
Agent crashes with: "runtime: out of memory"

## Environment
- OSPA version: commit abc123
- Go version: 1.21.0
- OS: Ubuntu 22.04
- OpenStack: Zed

## Additional Context
Memory usage grows unbounded during discovery phase.
```

## Feature Requests

### When to Request

- New functionality you need
- Improvements to existing features
- New OpenStack service support
- Better integration options

### How to Request

Create a new issue with the **Feature Request** template:

```markdown
## Description
What feature do you want?

## Use Case
Why do you need this feature? What problem does it solve?

## Proposed Solution
How would this feature work?

## Alternatives Considered
What other options have you considered?

## Additional Context
Any other relevant information.
```

### Good Feature Request Example

```markdown
## Description
Add support for Glance image auditing.

## Use Case
We need to find and clean up old or unused images in our OpenStack deployment.
Currently, we have 500+ images, many of which are obsolete.

## Proposed Solution
- Add `glance` service support
- Implement `image` resource type
- Support checks: status, age_gt, visibility, unused
- Support actions: log, delete, tag

## Alternatives Considered
- Manual scripts: Hard to maintain, no policy framework
- Other tools: Don't integrate with our existing OSPA policies

## Additional Context
We would be happy to contribute to implementation.
```

## Issue Labels

| Label | Meaning |
|-------|---------|
| `bug` | Something isn't working |
| `enhancement` | New feature or improvement |
| `documentation` | Documentation improvements |
| `good first issue` | Good for newcomers |
| `help wanted` | Looking for contributors |
| `question` | Questions about usage |
| `wontfix` | Won't be addressed |
| `duplicate` | Duplicate of another issue |


## Participating in Issues

### Helpful Comments

- Provide additional context
- Confirm you can reproduce
- Suggest solutions
- Offer to help implement

### Example

```markdown
I can confirm this issue on:
- Go 1.21.1
- Ubuntu 22.04
- OpenStack Yoga

I also noticed that reducing workers to 1 prevents the crash,
suggesting a race condition.

Happy to help debug if you can point me to the relevant code.
```

### Unhelpful Comments

Avoid:

- "+1" or "me too" (use reactions instead)
- Off-topic discussion
- Demanding timeline for fixes
- Unconstructive criticism

## Security Issues

For security vulnerabilities:

1. **Do NOT open a public issue**
2. Contact maintainers privately
3. Provide details for reproduction
4. Wait for acknowledgment before disclosure

## Questions

For usage questions:

1. Check the [documentation](../index.md)
2. Search existing issues
3. Open a discussion (if available)
4. Open an issue with the `question` label

## Stale Issues

Issues without activity for 60+ days may be:

- Pinged for update
- Closed if no response
- Reopened if issue persists

Keep issues active by providing updates when requested.

