# Testing Rules

## Minimum Test Coverage: 80%

Required test types:
1. **Unit Tests** - Individual functions
2. **Integration Tests** - API endpoints, database
3. **E2E Tests** - Critical user flows (Playwright)

## TDD Workflow (RED-GREEN-IMPROVE)

1. **RED**: Write failing test first
2. **GREEN**: Write minimal code to pass
3. **IMPROVE**: Refactor while maintaining coverage

```typescript
// 1. RED - Write test first
describe('JobService', () => {
  it('should create a job with valid concept', async () => {
    const result = await jobService.create({ concept: 'Test video' })
    expect(result.status).toBe('pending')
  })
})

// 2. GREEN - Minimal implementation
async create(input: CreateJobInput): Promise<Job> {
  return { id: uuid(), status: 'pending', ...input }
}

// 3. IMPROVE - Refactor if needed
```

## Go Testing Patterns

```go
// Table-driven tests
func TestJobService_Create(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateJobInput
        wantErr bool
    }{
        {
            name:    "valid concept",
            input:   CreateJobInput{Concept: "Test video about cats"},
            wantErr: false,
        },
        {
            name:    "empty concept",
            input:   CreateJobInput{Concept: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewJobService(mockRepo)
            _, err := svc.Create(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## When Tests Fail

1. Check test isolation (no shared state)
2. Verify mocks are accurate
3. **Fix implementation, NOT tests** (unless tests are wrong)
4. Re-run full test suite

## Available Agents

- **tdd-guide**: Enforces test-first for new features
- **e2e-runner**: Playwright testing specialist

## Commands

```bash
# Go
go test ./...
go test -v ./...
go test -cover ./...

# Frontend
npm run test
npm run test:coverage
npm run test:e2e
```
