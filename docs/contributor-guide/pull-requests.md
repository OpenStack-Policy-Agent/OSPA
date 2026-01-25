# Pull Requests

This guide explains how to submit and review pull requests effectively.

## Before Submitting

1. **Open an issue first** - For significant changes, discuss the approach
2. **Follow contribution guide** - Set up development environment properly
3. **Run tests** - Ensure all tests pass locally
4. **Update documentation** - Include docs for new features

## Creating a Pull Request

### 1. Prepare Your Branch

```bash
# Update your fork
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/my-feature

# Make changes
# ... edit files ...

# Run tests
go test ./...

# Commit
git add .
git commit -m "Add feature description"

# Push
git push origin feature/my-feature
```

### 2. Open the PR

1. Go to your fork on GitHub
2. Click "Compare & pull request"
3. Fill in the PR template
4. Submit the pull request

### 3. PR Template

```markdown
## Description
Brief description of what this PR does.

## Related Issue
Fixes #123

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Changes Made
- Added X
- Modified Y
- Fixed Z

## Testing
- [ ] Unit tests pass
- [ ] E2E tests pass (if applicable)
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style
- [ ] Self-reviewed the code
- [ ] Added tests for new functionality
- [ ] Updated documentation
- [ ] No breaking changes (or documented)
```

## PR Guidelines

### Size

- **Keep PRs small** - Under 400 lines changed is ideal
- **Split large changes** - Break into smaller, focused PRs
- **One concern per PR** - Don't mix features with refactoring

### Title

Use clear, descriptive titles:

```
# Good
Add Glance image discovery and auditing
Fix race condition in worker pool
Update documentation for security groups

# Bad
Update code
Fix bug
WIP
```

### Description

Include:

- What the PR does
- Why it's needed
- How it was tested
- Any breaking changes

### Commits

- **Meaningful messages** - Describe what and why
- **Atomic commits** - Each commit should be self-contained
- **Squash if needed** - Combine fixup commits before merge

## Review Process

### What Reviewers Look For

| Area | Questions |
|------|-----------|
| **Correctness** | Does it work? Are there edge cases? |
| **Design** | Is the approach sound? Does it fit the codebase? |
| **Tests** | Are there adequate tests? Do they cover edge cases? |
| **Documentation** | Is it documented? Are comments clear? |
| **Style** | Does it follow project conventions? |
| **Performance** | Any performance concerns? |

### Responding to Reviews

- **Be responsive** - Address feedback promptly
- **Be open** - Consider suggestions objectively
- **Ask questions** - If feedback is unclear, ask for clarification
- **Explain decisions** - If you disagree, explain your reasoning

### Making Changes

```bash
# Make requested changes
git add .
git commit -m "Address review feedback"
git push origin feature/my-feature
```

For significant changes, consider force-pushing a clean history:

```bash
# Rebase and squash commits
git rebase -i main
git push --force-with-lease origin feature/my-feature
```

## Review Checklist

When reviewing others' PRs:

### Code Quality

- [ ] Code is readable and well-organized
- [ ] Functions are focused and not too long
- [ ] Error handling is appropriate
- [ ] No obvious bugs or edge cases

### Testing

- [ ] Tests cover the changes
- [ ] Tests are meaningful (not just coverage)
- [ ] Edge cases are tested
- [ ] Tests pass in CI

### Documentation

- [ ] Public APIs are documented
- [ ] Complex logic has comments
- [ ] User docs updated if needed
- [ ] Examples included where helpful

### Style

- [ ] Follows Go conventions
- [ ] Consistent with existing code
- [ ] No unnecessary changes

## Review Etiquette

### For Reviewers

- **Be constructive** - Suggest improvements, don't just criticize
- **Be specific** - Point to exact lines, suggest alternatives
- **Be timely** - Review within a reasonable timeframe
- **Approve when ready** - Don't nitpick indefinitely

### Example Comments

```markdown
# Good
This could be simplified using `strings.Builder`:
```go
var b strings.Builder
for _, s := range items {
    b.WriteString(s)
}
return b.String()
```

# Bad
This is wrong.

# Good
Consider adding a test case for when the list is empty.
We've had bugs in this area before.

# Bad
Add more tests.
```

### For Authors

- **Be gracious** - Thank reviewers for their time
- **Be responsive** - Address feedback promptly
- **Be open** - Consider suggestions fairly
- **Ask for clarification** - If you don't understand feedback

## Merging

### Requirements

Before merging:

- [ ] All CI checks pass
- [ ] At least one approval from maintainer
- [ ] No unresolved review comments
- [ ] No merge conflicts

### Merge Strategy

We use **squash and merge** to keep history clean:

1. Reviewer approves the PR
2. Author or maintainer clicks "Squash and merge"
3. Final commit message is edited if needed
4. PR is merged

## After Merging

- **Delete your branch** - Clean up feature branches
- **Update related issues** - Close fixed issues
- **Monitor CI** - Ensure no regressions on main

## Stale PRs

PRs without activity for 30+ days may be:

- Pinged for update
- Closed if no response

Keep PRs active by:

- Responding to review comments
- Rebasing on latest main
- Providing progress updates

