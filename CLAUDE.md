# UGC - AI Video Generator

AI-powered short-form video content generator using LLMs, Suno (music), NanoBanana (images), and FFmpeg.

---

## Architecture Overview

### System Flow

```
User Request (concept)
        │
        ▼
┌───────────────────┐
│   API Gateway     │  (Go + Chi Router)
│   /api/jobs       │
└────────┬──────────┘
         │
         ▼
┌───────────────────┐     ┌───────────────────┐
│   Job Service     │────▶│   Asynq Worker    │
│   (creates job)   │     │   (background)    │
└───────────────────┘     └────────┬──────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
        ▼                          ▼                          ▼
┌───────────────┐         ┌───────────────┐         ┌───────────────┐
│ OpenRouter    │         │   Suno (KIE)  │         │ NanoBanana    │
│ (LLM agents)  │         │   (music)     │         │ (images)      │
└───────────────┘         └───────────────┘         └───────────────┘
                                   │
                                   ▼
                          ┌───────────────┐
                          │    FFmpeg     │
                          │ (video proc)  │
                          └───────┬───────┘
                                  │
                                  ▼
                          ┌───────────────┐
                          │  Cloudflare   │
                          │     R2        │
                          └───────────────┘
```

### Job Status Flow

```
pending → analyzing → generating_music → selecting_song → generating_image → processing_video → uploading → completed
                │              │               │                 │                  │             │
                └──────────────┴───────────────┴─────────────────┴──────────────────┴─────────────┘
                                              ▼
                                           failed
```

### Directory Structure

```
ugc/
├── cmd/ugc/main.go           # Entry point, DI setup
├── internal/
│   ├── agents/               # LLM agents (BaseAgent, SongConcept, ImageConcept, SongSelector)
│   ├── config/               # Environment config
│   ├── database/             # GORM setup, migrator
│   ├── external/             # External service clients
│   │   ├── kie/              # Suno, NanoBanana clients
│   │   ├── openrouter/       # LLM client
│   │   └── r2/               # Cloudflare R2 storage
│   ├── ffmpeg/               # Video processing
│   ├── handler/              # HTTP handlers (auth, job, webhook)
│   ├── middleware/           # Auth, CORS, logging
│   ├── models/               # Domain models (User, Job)
│   ├── repository/           # Data access layer
│   ├── service/              # Business logic
│   └── worker/               # Asynq background workers
│       └── tasks/            # Task handlers (REAL implementations)
├── pkg/
│   ├── errors/               # Custom error types
│   └── response/             # API response helpers
└── frontend/                 # React + TypeScript
    └── src/
        ├── api/              # Legacy API functions
        ├── components/ui/    # Reusable UI components
        ├── features/         # Feature modules
        │   ├── auth/         # Login, Register
        │   ├── dashboard/    # Main dashboard
        │   ├── job/          # Job CRUD, status
        │   └── settings/     # User settings
        ├── hooks/            # Custom React hooks
        ├── lib/              # Axios, utils
        ├── providers/        # React Query provider
        ├── router/           # React Router config
        └── stores/           # Zustand stores
```

---

## Development Guidelines

### Plan Mode Rules

**ALWAYS enter Plan Mode for:**
- New feature implementation
- Refactoring across multiple files
- Database schema changes
- API contract changes
- Worker task modifications

**Plan Mode Process:**
1. Read relevant files to understand current state
2. List all files that will be modified
3. Describe the changes for each file
4. Identify potential breaking changes
5. Get user approval before implementing

### Code Quality Standards

#### Go Backend

```go
// REQUIRED: Services must define interfaces
type JobService interface {
    Create(ctx context.Context, input CreateJobInput) (*Job, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
}

// REQUIRED: Use zap logger, never fmt.Print
logger.Info("job created", zap.String("job_id", job.ID.String()))

// REQUIRED: Agents extend BaseAgent
type SongConceptAgent struct {
    *agents.BaseAgent
}

// REQUIRED: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create job: %w", err)
}
```

#### TypeScript/React

```typescript
// REQUIRED: Feature-based organization
// features/job/api.ts - API calls
// features/job/hooks/useJob.ts - React Query hooks
// features/job/components/JobCard.tsx - Components
// features/job/pages/JobDetailPage.tsx - Pages

// REQUIRED: React Query for server state
export function useJob(id: string) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: () => fetchJob(id),
  })
}

// REQUIRED: Zustand for auth state only
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

// REQUIRED: react-hook-form + Zod for forms
const schema = z.object({
  concept: z.string().min(10, 'Concept must be at least 10 characters'),
})
```

---

## Verification Checklist

### Before Committing Go Code

```bash
# MUST pass all checks
make lint           # golangci-lint
make test           # go test ./...
go build ./...      # Compilation check
```

**Check for:**
- [ ] All services have interfaces
- [ ] No `fmt.Print` - use zap logger
- [ ] Errors wrapped with `fmt.Errorf`
- [ ] Context passed to all database/API calls
- [ ] No hardcoded credentials

### Before Committing Frontend Code

```bash
# MUST pass all checks
npm run lint        # ESLint
npm run build       # TypeScript compilation
npm run test        # Vitest (if tests exist)
```

**Check for:**
- [ ] No `any` types - use proper TypeScript types
- [ ] React Query for server state (not useState for API data)
- [ ] Proper error handling in API calls
- [ ] Loading states handled

### Pull Request Checklist

- [ ] All CI checks pass
- [ ] No new lint warnings
- [ ] Database migrations included (if schema changed)
- [ ] Environment variables documented (if new ones added)
- [ ] Breaking changes noted in PR description

---

## Known Issues & Technical Debt

### Critical

| Issue | Location | Impact | Solution |
|-------|----------|--------|----------|
| Worker handlers not wired | `worker/worker.go:183-353` | Tasks run but do nothing | Wire `tasks/handlers.go` functions to worker.go registrations |
| Duplicate job APIs | `api/jobs.ts` vs `features/job/api.ts` | Confusion, inconsistent behavior | Consolidate to `features/job/api.ts`, delete `api/jobs.ts` |

### High Priority

| Issue | Location | Impact | Solution |
|-------|----------|--------|----------|
| Webhook handlers stub | `handler/webhook_handler.go` | Suno/KIE callbacks fail | Implement actual webhook processing |
| Missing DownloadButton | `features/job/pages/JobDetailPage.tsx` | Users can't download videos | Add DownloadButton component |
| Route inconsistency | `router/index.tsx` | `/jobs/new` vs `/jobs/create` | Standardize to `/jobs/create` |

### Improvements

| Issue | Location | Impact | Solution |
|-------|----------|--------|----------|
| No auth guards | `router/index.tsx` | Protected routes accessible | Add PrivateRoute component |
| No error boundaries | `App.tsx` | Unhandled errors crash app | Add React ErrorBoundary |

---

## Quick Reference

### Commands

```bash
# Backend
make dev              # Hot reload with air
make test             # Run all tests
make test-v           # Verbose tests
make lint             # Run golangci-lint
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations

# Frontend
cd frontend
npm run dev           # Dev server :5173
npm run build         # Production build
npm run lint          # ESLint
npm run preview       # Preview production build

# Git (gh CLI)
gh pr create --fill   # Create PR with auto-fill
gh pr checkout <num>  # Checkout PR branch
gh pr merge --squash  # Squash merge
gh issue list         # List issues
gh issue create       # Create issue
```

### Environment Variables

**Backend (required):**
```env
DATABASE_URL=postgres://user:pass@localhost:5432/ugc
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
OPENROUTER_API_KEY=sk-or-xxx
KIE_API_KEY=xxx
KIE_BASE_URL=https://api.kie.ai
```

**Backend (optional):**
```env
R2_ACCOUNT_ID=xxx
R2_ACCESS_KEY_ID=xxx
R2_SECRET_ACCESS_KEY=xxx
R2_BUCKET_NAME=ugc-assets
R2_PUBLIC_URL=https://cdn.example.com
WEBHOOK_BASE_URL=https://api.example.com  # Empty to use polling
```

**Frontend:**
```env
VITE_API_BASE_URL=http://localhost:8080
```

### Job Status Values

| Status | Description |
|--------|-------------|
| `pending` | Job created, waiting for worker |
| `analyzing` | LLM analyzing concept |
| `generating_music` | Suno generating songs |
| `selecting_song` | LLM selecting best song |
| `generating_image` | NanoBanana generating image |
| `processing_video` | FFmpeg combining audio + image |
| `uploading` | Uploading to R2 |
| `completed` | Job finished successfully |
| `failed` | Job failed (check error_message) |

---

## Task Management

Tasks are tracked in `tasks/` directory:

```
tasks/
├── todo.md           # Current sprint tasks
├── backlog.md        # Future work
└── lessons/
    ├── bugs.md       # Bug patterns and fixes
    ├── patterns.md   # Useful coding patterns
    └── decisions.md  # Architecture decision records
```

### Task Format

```markdown
## [Status] Task Title

**Priority:** Critical | High | Medium | Low
**Files:** list of affected files

### Description
What needs to be done

### Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
```

---

## Autonomous Bug Fixing Protocol

When encountering an error:

1. **Read the error message carefully**
2. **Identify the file and line number**
3. **Read the relevant code**
4. **Check for common patterns below**
5. **Fix and verify with lint/build**

### Common Go Errors

```go
// Error: undefined: SomeType
// Fix: Check imports, add missing import
import "github.com/jaochai/ugc/internal/models"

// Error: cannot use x (type A) as type B
// Fix: Check interface implementation, add missing methods

// Error: sql: no rows in result set
// Fix: Use GetByID that returns (*Model, error), check for nil
job, err := repo.GetByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("job not found: %w", err)
}
```

### Common React/TypeScript Errors

```typescript
// Error: Type 'X' is not assignable to type 'Y'
// Fix: Check the type definitions, ensure they match

// Error: Cannot find module '@/xxx'
// Fix: Check tsconfig.json paths, ensure file exists

// Error: 'X' is possibly 'undefined'
// Fix: Add optional chaining or null check
const name = user?.name ?? 'Unknown'
```

---

## Self-Improvement Loop

After fixing a bug or learning something new:

1. **Document in `tasks/lessons/bugs.md`** if it's a bug pattern
2. **Document in `tasks/lessons/patterns.md`** if it's a useful pattern
3. **Update `tasks/lessons/decisions.md`** if it's an architecture decision
4. **Update this CLAUDE.md** if it affects development workflow

### Lesson Format

```markdown
## [Date] Title

**Context:** Why did this happen?
**Problem:** What was the issue?
**Solution:** How was it fixed?
**Prevention:** How to avoid in future?
```

---

## Git Workflow

### Branch Naming
- `feature/xxx` - New features
- `fix/xxx` - Bug fixes
- `refactor/xxx` - Code refactoring
- `docs/xxx` - Documentation

### Commit Messages (Conventional Commits)
```
feat: add job creation endpoint
fix: handle null audio_url in video processing
refactor: extract agent base class
docs: update CLAUDE.md with architecture
```

### PR Process
1. Create feature branch from `master`
2. Make changes with atomic commits
3. Run verification checklist
4. Create PR with `gh pr create --fill`
5. Squash merge after approval

---

## API Endpoints

### Auth
- `POST /api/auth/register` - Create account
- `POST /api/auth/login` - Get JWT token

### Jobs
- `GET /api/jobs` - List user's jobs (paginated)
- `POST /api/jobs` - Create new job
- `GET /api/jobs/:id` - Get job details
- `POST /api/jobs/:id/cancel` - Cancel job

### Webhooks (internal)
- `POST /webhooks/suno/:job_id` - Suno callback
- `POST /webhooks/nano/:job_id` - NanoBanana callback
