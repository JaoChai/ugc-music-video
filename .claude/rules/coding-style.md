# Coding Style Rules

## Immutability (CRITICAL)

ALWAYS create new objects, NEVER mutate:

```javascript
// WRONG: Mutation
function updateUser(user, name) {
  user.name = name  // MUTATION!
  return user
}

// CORRECT: Immutability
function updateUser(user, name) {
  return {
    ...user,
    name
  }
}
```

```go
// Go - return new structs instead of modifying
func UpdateJobStatus(job Job, status string) Job {
    return Job{
        ID:     job.ID,
        Status: status,
        // ... copy other fields
    }
}
```

## File Organization

**MANY SMALL FILES > FEW LARGE FILES:**

- High cohesion, low coupling
- 200-400 lines typical, 800 max
- Extract utilities from large components
- Organize by feature/domain, not by type

## Error Handling

ALWAYS handle errors comprehensively:

```typescript
try {
  const result = await riskyOperation()
  return result
} catch (error) {
  console.error('Operation failed:', error)
  throw new Error('Detailed user-friendly message')
}
```

```go
if err != nil {
    return fmt.Errorf("failed to create job: %w", err)
}
```

## Input Validation

ALWAYS validate user input:

```typescript
import { z } from 'zod'

const schema = z.object({
  email: z.string().email(),
  age: z.number().int().min(0).max(150)
})

const validated = schema.parse(input)
```

## Go-Specific Style

```go
// REQUIRED: Services must define interfaces
type JobService interface {
    Create(ctx context.Context, input CreateJobInput) (*Job, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
}

// REQUIRED: Use zap logger, never fmt.Print
logger.Info("job created", zap.String("job_id", job.ID.String()))

// REQUIRED: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create job: %w", err)
}
```

## TypeScript/React Style

```typescript
// Feature-based organization
// features/job/api.ts - API calls
// features/job/hooks/useJob.ts - React Query hooks
// features/job/components/JobCard.tsx - Components

// React Query for server state
export function useJob(id: string) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: () => fetchJob(id),
  })
}

// Zustand for auth state only
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      setAuth: (token, user) => set({ token, user }),
    }),
    { name: 'auth-storage' }
  )
)
```

## Code Quality Checklist

Before marking work complete:

- [ ] Code is readable and well-named
- [ ] Functions are small (<50 lines)
- [ ] Files are focused (<800 lines)
- [ ] No deep nesting (>4 levels)
- [ ] Proper error handling
- [ ] No console.log statements in production code
- [ ] No hardcoded values
- [ ] No mutation (immutable patterns used)
