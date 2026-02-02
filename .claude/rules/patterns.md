# Common Patterns

## API Response Format

```typescript
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  meta?: {
    total: number
    page: number
    limit: number
  }
}
```

```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total int `json:"total"`
    Page  int `json:"page"`
    Limit int `json:"limit"`
}
```

## Custom Hooks Pattern

```typescript
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => setDebouncedValue(value), delay)
    return () => clearTimeout(handler)
  }, [value, delay])

  return debouncedValue
}
```

## Repository Pattern

```typescript
interface Repository<T> {
  findAll(filters?: Filters): Promise<T[]>
  findById(id: string): Promise<T | null>
  create(data: CreateDto): Promise<T>
  update(id: string, data: UpdateDto): Promise<T>
  delete(id: string): Promise<void>
}
```

```go
type JobRepository interface {
    FindAll(ctx context.Context, filters JobFilters) ([]Job, error)
    FindByID(ctx context.Context, id uuid.UUID) (*Job, error)
    Create(ctx context.Context, job *Job) error
    Update(ctx context.Context, job *Job) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

## Service Layer Pattern

```go
type JobService interface {
    Create(ctx context.Context, input CreateJobInput) (*Job, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
    List(ctx context.Context, userID uuid.UUID, page, limit int) (*ListResult, error)
    Cancel(ctx context.Context, id uuid.UUID) error
}

type jobService struct {
    repo   JobRepository
    logger *zap.Logger
}

func NewJobService(repo JobRepository, logger *zap.Logger) JobService {
    return &jobService{repo: repo, logger: logger}
}
```

## React Query Keys Pattern

```typescript
export const jobKeys = {
  all: ['jobs'] as const,
  lists: () => [...jobKeys.all, 'list'] as const,
  list: (filters: JobFilters) => [...jobKeys.lists(), filters] as const,
  details: () => [...jobKeys.all, 'detail'] as const,
  detail: (id: string) => [...jobKeys.details(), id] as const,
}
```

## Error Boundary Pattern

```typescript
class ErrorBoundary extends Component<Props, State> {
  state = { hasError: false, error: null }

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error }
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />
    }
    return this.props.children
  }
}
```

## Skeleton Projects

When implementing new functionality:
1. Search for battle-tested skeleton projects
2. Use parallel agents to evaluate options:
   - Security assessment
   - Extensibility analysis
   - Relevance scoring
   - Implementation planning
3. Clone best match as foundation
4. Iterate within proven structure
