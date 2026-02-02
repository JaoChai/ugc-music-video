# Git Workflow Rules

## Commit Message Format (Conventional Commits)

```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code refactoring
- `docs` - Documentation
- `test` - Adding tests
- `chore` - Maintenance
- `perf` - Performance improvement
- `ci` - CI/CD changes

**Examples:**
```
feat: add job creation endpoint
fix: handle null audio_url in video processing
refactor: extract agent base class
docs: update CLAUDE.md with architecture
```

## Branch Naming

- `feature/xxx` - New features
- `fix/xxx` - Bug fixes
- `refactor/xxx` - Code refactoring
- `docs/xxx` - Documentation

## PR Process

Before creating PR:

1. **Analyze full commit history** (not just latest commit)
2. Use `git diff [base-branch]...HEAD` to see all changes
3. Draft comprehensive summary
4. Include test plan

**PR Template:**
```markdown
## Summary
<1-3 bullet points>

## Test plan
- [ ] Test criterion 1
- [ ] Test criterion 2

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Feature Development Workflow

1. **Planning Phase**
   - Use planner agent to identify dependencies
   - Create task list with milestones

2. **TDD Phase**
   - Use tdd-guide agent
   - RED-GREEN-IMPROVE cycle
   - Target 80%+ coverage

3. **Review Phase**
   - Use code-reviewer agent
   - Address critical and high severity issues
   - Security check before merge

## Git Safety Protocol

- NEVER update git config
- NEVER run destructive commands (push --force, reset --hard) without explicit request
- NEVER skip hooks (--no-verify) unless explicitly requested
- ALWAYS create NEW commits rather than amending after hook failures
- Prefer staging specific files over `git add -A`

## Commands

```bash
# Create PR
gh pr create --fill

# Checkout PR
gh pr checkout <num>

# Squash merge
gh pr merge --squash

# List issues
gh issue list
```
