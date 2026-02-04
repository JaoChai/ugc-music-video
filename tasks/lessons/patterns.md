# Coding Patterns

Useful patterns discovered while working on this codebase.

---

## Go Backend

### Agent Pattern with BaseAgent

All LLM agents extend BaseAgent for consistent JSON chat functionality.

```go
// internal/agents/base.go
type BaseAgent struct {
    client   *openrouter.Client
    model    string
    logger   *zap.Logger
}

func (b *BaseAgent) ChatJSON(ctx context.Context, systemPrompt string, userMessage string, result interface{}) error {
    // Handles JSON response parsing with retry logic
}

// Usage in specific agent
type SongConceptAgent struct {
    *BaseAgent
}

func NewSongConceptAgent(client *openrouter.Client, model string, logger *zap.Logger) *SongConceptAgent {
    return &SongConceptAgent{
        BaseAgent: &BaseAgent{client: client, model: model, logger: logger},
    }
}

func (a *SongConceptAgent) Analyze(ctx context.Context, input SongConceptInput) (*SongConceptOutput, error) {
    var output SongConceptOutput
    err := a.ChatJSON(ctx, systemPrompt, input.String(), &output)
    return &output, err
}
```

---

### Repository Pattern with Interface

Services depend on interfaces, not concrete implementations.

```go
// internal/repository/job_repo.go
type JobRepository interface {
    Create(ctx context.Context, job *models.Job) error
    GetByID(ctx context.Context, id uuid.UUID) (*models.Job, error)
    Update(ctx context.Context, job *models.Job) error
    List(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Job, int64, error)
}

type jobRepository struct {
    db *gorm.DB
}

func NewJobRepository(db *gorm.DB) JobRepository {
    return &jobRepository{db: db}
}

// internal/service/job_service.go
type jobService struct {
    jobRepo JobRepository  // Interface, not *jobRepository
    client  *asynq.Client
    logger  *zap.Logger
}

func NewJobService(jobRepo JobRepository, client *asynq.Client, logger *zap.Logger) JobService {
    return &jobService{jobRepo: jobRepo, client: client, logger: logger}
}
```

---

### Error Handling with Context

Always wrap errors with context for debugging.

```go
// Bad
if err != nil {
    return err
}

// Good
if err != nil {
    return fmt.Errorf("failed to create job for user %s: %w", userID, err)
}

// For repository layer
if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil, fmt.Errorf("job %s not found: %w", id, err)
}
```

---

### Task Chain Pattern (Asynq)

Each task handler processes, updates state, and enqueues next task.

```go
func HandleAnalyzeConcept(deps *Dependencies) asynq.HandlerFunc {
    return func(ctx context.Context, task *asynq.Task) error {
        // 1. Parse payload
        payload, err := UnmarshalTaskPayload(task.Payload())
        if err != nil {
            return fmt.Errorf("failed to unmarshal: %w", err)
        }

        // 2. Load current state
        job, err := deps.JobRepo.GetByID(ctx, payload.JobID)
        if err != nil {
            return markJobFailed(ctx, deps, payload.JobID, "failed to load job")
        }

        // 3. Update status
        job.Status = models.StatusAnalyzing
        deps.JobRepo.Update(ctx, job)

        // 4. Do work
        result, err := doWork(ctx, job)
        if err != nil {
            return markJobFailed(ctx, deps, payload.JobID, err.Error())
        }

        // 5. Save result
        job.SongPrompt = result
        deps.JobRepo.Update(ctx, job)

        // 6. Enqueue next task
        nextPayload, _ := (&TaskPayload{JobID: payload.JobID}).Marshal()
        deps.AsynqClient.Enqueue(asynq.NewTask(TypeGenerateMusic, nextPayload))

        return nil
    }
}
```

---

## React Frontend

### Feature Module Structure

Each feature is self-contained with its own api, hooks, components, pages.

```
features/job/
├── api.ts              # API calls + React Query hooks
├── types.ts            # TypeScript types
├── components/
│   ├── index.ts        # Barrel export
│   ├── JobCard.tsx
│   └── JobStatusBadge.tsx
├── pages/
│   ├── JobListPage.tsx
│   ├── JobDetailPage.tsx
│   └── CreateJobPage.tsx
└── index.ts            # Public API
```

```typescript
// features/job/index.ts - Public API
export { JobListPage, JobDetailPage, CreateJobPage } from './pages'
export { JobCard, JobStatusBadge } from './components'
export { useJobs, useJob, useCreateJob } from './api'
export type { Job, CreateJobRequest } from './types'
```

---

### React Query with Factory Keys

Consistent query key structure for cache management.

```typescript
// features/job/api.ts
export const jobKeys = {
    all: ['jobs'] as const,
    lists: () => [...jobKeys.all, 'list'] as const,
    list: (filters: Record<string, unknown>) => [...jobKeys.lists(), filters] as const,
    details: () => [...jobKeys.all, 'detail'] as const,
    detail: (id: string) => [...jobKeys.details(), id] as const,
}

// Usage
const { data } = useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: () => fetchJob(id),
})

// Invalidation
queryClient.invalidateQueries({ queryKey: jobKeys.lists() })  // All lists
queryClient.invalidateQueries({ queryKey: jobKeys.all })       // Everything
```

---

### Form Pattern with react-hook-form + Zod

Type-safe forms with validation.

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const createJobSchema = z.object({
    concept: z.string()
        .min(10, 'Concept must be at least 10 characters')
        .max(500, 'Concept must be less than 500 characters'),
})

type CreateJobInput = z.infer<typeof createJobSchema>

function CreateJobForm() {
    const { register, handleSubmit, formState: { errors } } = useForm<CreateJobInput>({
        resolver: zodResolver(createJobSchema),
    })

    const createJob = useCreateJob()

    const onSubmit = (data: CreateJobInput) => {
        createJob.mutate(data)
    }

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <textarea {...register('concept')} />
            {errors.concept && <span>{errors.concept.message}</span>}
            <button type="submit" disabled={createJob.isPending}>
                {createJob.isPending ? 'Creating...' : 'Create Job'}
            </button>
        </form>
    )
}
```

---

### Polling with React Query

Auto-refresh for in-progress jobs.

```typescript
function useJobWithPolling(id: string) {
    const query = useQuery({
        queryKey: jobKeys.detail(id),
        queryFn: () => fetchJob(id),
        // Only poll if job is not in terminal state
        refetchInterval: (query) => {
            const job = query.state.data
            if (!job) return false
            if (['completed', 'failed'].includes(job.status)) return false
            return 3000 // Poll every 3 seconds
        },
    })
    return query
}
```

---

### Systematic Debugging Pattern (เมื่อ fix ไม่ได้ผล)

เมื่อแก้ไขแล้วปัญหาไม่หาย ให้หยุดแก้และทำตามขั้นตอนนี้:

```
1. STOP - หยุดแก้ไข
2. TRACE - trace code path จาก entry point ถึง error
3. VERIFY - ตรวจสอบว่า code ที่แก้ถูกใช้จริงหรือไม่
4. ANALYZE - ใช้ Sequential Thinking วิเคราะห์อย่างเป็นระบบ
5. FIX - แก้ไขที่ root cause
```

**ตัวอย่าง: URL Allowlist Debug**
```bash
# TRACE: หา usage ของ DefaultAllowedHosts
grep -r "DefaultAllowedHosts" --include="*.go"
# ผลลัพธ์: ใช้ใน NewURLValidator เมื่อ allowedHosts empty

# VERIFY: ดูว่าใครเรียก NewURLValidator
grep -r "NewURLValidator" --include="*.go"
# ผลลัพธ์: main.go เรียก พร้อม cfg.Webhook.AllowedHosts

# ANALYZE: cfg.Webhook.AllowedHosts มาจากไหน?
grep -r "AllowedHosts" --include="*.go"
# ผลลัพธ์: config.go มี viper.SetDefault → ไม่เคย empty!

# FIX: แก้ที่ config.go ไม่ใช่ url_validator.go
```

---

### LLM Agent Output Design Pattern

**หลักการ:** LLM ควร output creative content เท่านั้น ไม่ควร output API-specific values

```go
// ❌ BAD - LLM ต้องรู้ว่า Suno มี model อะไรบ้าง
type SongConceptOutput struct {
    Prompt string `json:"prompt"`
    Model  string `json:"model"` // LLM อาจเลือก "V3.5" ที่ไม่มี
}

// ✅ GOOD - LLM โฟกัส creative work
type SongConceptOutput struct {
    Prompt string `json:"prompt"`
    Style  string `json:"style"`
    Title  string `json:"title"`
    // Model, AspectRatio, Resolution → hardcode ใน code
}

func (o *SongConceptOutput) ToSongPrompt() *models.SongPrompt {
    return &models.SongPrompt{
        Prompt: o.Prompt,
        Model:  "V5", // Hardcode - LLM ไม่ต้องรู้
    }
}
```

**เหตุผล:**
- LLM ไม่มี knowledge ที่ up-to-date เกี่ยวกับ API versions
- API validation errors ยากต่อการ debug
- Hardcode ใน code ง่ายต่อการ maintain และ update

---

### Configuration Single Source of Truth Pattern

**หลักการ:** Configuration ควรมี single source of truth ไม่ใช่หลายที่

```go
// ❌ BAD - มี default หลายที่
// url_validator.go
var DefaultAllowedHosts = []string{"cdn1.suno.ai", ...}

// config.go
viper.SetDefault("WEBHOOK_ALLOWED_HOSTS", "cdn1.suno.ai,...")
// → สับสนว่าต้องแก้ที่ไหน

// ✅ GOOD - Single source of truth
// config.go เป็น source of truth
viper.SetDefault("WEBHOOK_ALLOWED_HOSTS",
    "cdn1.suno.ai,cdn2.suno.ai,musicfile.kie.ai,aiquickdraw.com")

// url_validator.go ไม่มี default
func NewURLValidator(allowedHosts []string) *URLValidator {
    // ใช้ค่าที่ส่งมาเสมอ ไม่มี fallback
}
```

---

### Container Dependency Verification Pattern

เมื่อ code ใช้ external command ต้องตรวจสอบ:

```go
// เมื่อเพิ่ม exec.Command ใหม่
cmd := exec.CommandContext(ctx, "curl", args...)  // <- ต้องมี curl ใน container

// Checklist:
// 1. ตรวจสอบ Dockerfile
// 2. ทดสอบใน Docker locally
// 3. Document dependency ใน code comment
```

```dockerfile
# Dockerfile - ระบุ dependencies ให้ชัด
# Install dependencies for video processing
# - ffmpeg: video encoding
# - curl: downloading media files from external URLs
RUN apk add --no-cache ffmpeg curl ca-certificates tzdata
```

---

### Railway Deployment Verification Pattern

หลังจาก push code ให้ตรวจสอบ:

```bash
# 1. ดู deployment status
railway status

# 2. ดู logs ว่า container ใหม่เริ่มทำงาน
railway logs -s backend | head -10
# ดู timestamp ว่าใหม่กว่า commit ที่ push

# 3. ถ้า auto-deploy ไม่ทำงาน
railway redeploy -s backend --yes

# 4. รอ container เริ่มใหม่แล้วทดสอบ
sleep 60 && curl https://api.example.com/health
```

---

## Add New Patterns Below

_When you discover a useful pattern, document it here_
