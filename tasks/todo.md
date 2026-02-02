# Current Tasks

## Critical

### [ ] Wire worker handlers to real implementations

**Priority:** Critical
**Files:** `internal/worker/worker.go`

#### Description
The `worker.go` file registers task handlers using stub functions (lines 183-353) that only log and return nil. The real implementations exist in `internal/worker/tasks/handlers.go` but are not wired up.

#### Current State
```go
// worker.go - STUBS (do nothing)
mux.HandleFunc(TypeAnalyzeConcept, newAnalyzeConceptHandler(deps))

// tasks/handlers.go - REAL implementations
func HandleAnalyzeConcept(deps *Dependencies) asynq.HandlerFunc { ... }
```

#### Solution
Replace worker.go stub registrations with tasks/handlers.go functions:
```go
import "github.com/jaochai/ugc/internal/worker/tasks"

// Create tasks.Dependencies
taskDeps := &tasks.Dependencies{...}

// Register real handlers
mux.HandleFunc(tasks.TypeAnalyzeConcept, tasks.HandleAnalyzeConcept(taskDeps))
```

#### Acceptance Criteria
- [ ] All 8 task types use real handlers from tasks/handlers.go
- [ ] Dependencies properly passed to handlers
- [ ] `make test` passes
- [ ] Manual test: create job, verify it progresses through statuses

---

### [ ] Consolidate duplicate job APIs

**Priority:** Critical
**Files:** `frontend/src/api/jobs.ts`, `frontend/src/features/job/api.ts`

#### Description
Two different job API implementations exist with different endpoint patterns:
- `api/jobs.ts`: Uses `/api/jobs` (matches Go backend)
- `features/job/api.ts`: Uses `/api/collections/jobs/records` (PocketBase style - WRONG)

#### Solution
1. Delete `frontend/src/api/jobs.ts`
2. Update `features/job/api.ts` to use correct endpoints (`/api/jobs`)
3. Update all imports to use `features/job/api.ts`

#### Acceptance Criteria
- [ ] Only one job API file exists
- [ ] Endpoints match Go backend (`/api/jobs`)
- [ ] All components use the correct API
- [ ] `npm run build` passes

---

## High Priority

### [ ] Implement webhook handlers

**Priority:** High
**Files:** `internal/handler/webhook_handler.go`

#### Description
Suno and NanoBanana send callbacks when generation completes, but handlers are not implemented.

#### Acceptance Criteria
- [ ] `POST /webhooks/suno/:job_id` updates job with generated songs
- [ ] `POST /webhooks/nano/:job_id` updates job with image URL
- [ ] Both enqueue next task in pipeline
- [ ] Error handling for invalid payloads

---

### [ ] Add DownloadButton to JobDetailPage

**Priority:** High
**Files:** `frontend/src/features/job/pages/JobDetailPage.tsx`

#### Description
Completed jobs have `video_url` but no way to download.

#### Acceptance Criteria
- [ ] DownloadButton component shows when job.status === 'completed'
- [ ] Button triggers file download
- [ ] Proper loading state during download

---

### [ ] Fix route inconsistencies

**Priority:** High
**Files:** `frontend/src/router/index.tsx`

#### Description
Routes use inconsistent patterns: `/jobs/new` vs `/jobs/create`

#### Acceptance Criteria
- [ ] Standardize to `/jobs/create`
- [ ] Update all links and navigation
- [ ] Add redirect from old route if needed

---

## Improvements

### [ ] Add authentication guards

**Priority:** Medium
**Files:** `frontend/src/router/index.tsx`

#### Description
Protected routes (dashboard, jobs) should redirect to login if not authenticated.

#### Acceptance Criteria
- [ ] PrivateRoute component wraps protected routes
- [ ] Redirects to `/login` if no token
- [ ] Preserves intended destination for redirect after login

---

### [ ] Add error boundaries

**Priority:** Medium
**Files:** `frontend/src/App.tsx`, `frontend/src/components/ErrorBoundary.tsx`

#### Description
Unhandled React errors crash the entire app.

#### Acceptance Criteria
- [ ] ErrorBoundary component catches render errors
- [ ] Shows user-friendly error message
- [ ] Provides "Try Again" action
- [ ] Logs error to console (or error service)

---

## Completed

_Move completed tasks here with completion date_
