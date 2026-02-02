# Agent Orchestration Rules

## Available Agents

| Agent | Purpose | When to Use |
|-------|---------|-------------|
| `planner` | Implementation planning | Complex features, refactoring |
| `architect` | System design | Architectural decisions |
| `tdd-guide` | Test-driven development | New features, bug fixes |
| `code-reviewer` | Code review | After writing code |
| `security-reviewer` | Security analysis | Before commits |
| `build-error-resolver` | Fix build errors | When build fails |
| `e2e-runner` | E2E testing | Critical user flows |
| `refactor-cleaner` | Dead code cleanup | Code maintenance |
| `doc-updater` | Documentation | Updating docs |
| `go-reviewer` | Go code review | Go projects |
| `go-build-resolver` | Go build errors | Go build failures |
| `database-reviewer` | PostgreSQL optimization | Database queries |

## Immediate Agent Usage (No Prompt Needed)

Use agents PROACTIVELY without waiting for user request:

1. **Complex feature requests** → Use **planner** agent
2. **Code just written/modified** → Use **code-reviewer** agent
3. **Bug fix or new feature** → Use **tdd-guide** agent
4. **Architectural decision** → Use **architect** agent
5. **Security-sensitive code** → Use **security-reviewer** agent
6. **Build failures** → Use **build-error-resolver** agent

## Parallel Task Execution

ALWAYS use parallel Task execution for independent operations:

```markdown
# GOOD: Parallel execution
Launch 3 agents in parallel:
1. Agent 1: Security analysis of auth.ts
2. Agent 2: Performance review of cache system
3. Agent 3: Type checking of utils.ts

# BAD: Sequential when unnecessary
First agent 1, then agent 2, then agent 3
```

## Multi-Perspective Analysis

For complex problems, use split role sub-agents:

- Factual reviewer
- Senior engineer perspective
- Security expert
- Consistency reviewer
- Redundancy checker

## Agent Selection Guidelines

**Use planner when:**
- Feature touches 3+ files
- Requirements are unclear
- Multiple implementation approaches exist

**Use tdd-guide when:**
- Writing new functionality
- Fixing bugs
- Refactoring existing code

**Use code-reviewer when:**
- After completing any code changes
- Before creating PR
- When uncertain about code quality

**Use security-reviewer when:**
- Handling user input
- Authentication/authorization changes
- API endpoint changes
- Database query changes
