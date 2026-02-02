# Performance Rules

## Model Selection

| Model | Use Case | Notes |
|-------|----------|-------|
| **Haiku** | Lightweight agents, pair programming | 90% of Sonnet capability, 3x cost savings |
| **Sonnet** | Primary development, orchestration | Best coding model |
| **Opus** | Architectural decisions, research | Deepest reasoning |

## Context Window Management

**High Context Sensitivity** (Avoid final 20%):
- Large-scale refactoring
- Multi-file feature implementation

**Low Context Sensitivity** (Safe):
- Single-file edits
- Documentation updates
- Bug fixes

## Strategic Compaction

Consider manual compaction at logical intervals:
- After completing a major feature
- Before starting a new task area
- When context is full of completed work

## Parallel Execution

ALWAYS parallelize independent operations:

```markdown
# Good - parallel
Task.spawn([
  { agent: 'security-reviewer', file: 'auth.ts' },
  { agent: 'code-reviewer', file: 'api.ts' },
  { agent: 'tdd-guide', file: 'service.ts' }
])

# Bad - sequential
await securityReview('auth.ts')
await codeReview('api.ts')
await tddGuide('service.ts')
```

## Build Error Resolution

When builds fail:
1. Use **build-error-resolver** agent
2. Analyze error messages systematically
3. Apply incremental fixes
4. Verify after each fix

## Advanced Techniques

For demanding projects:
- Use "ultrathink" with Plan Mode enabled
- Incorporate multiple critique rounds
- Deploy specialized sub-agents for varied perspectives

## Caching Strategy

- Cache expensive computations
- Use React Query for server state
- Avoid redundant API calls
- Implement proper cache invalidation

## Database Performance

- Always use indexes for frequently queried fields
- Use `EXPLAIN ANALYZE` before deploying queries
- Batch operations when possible
- Avoid N+1 query patterns
