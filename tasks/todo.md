# Current Tasks

## Critical

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

### [x] Wire worker handlers to real implementations (2026-02-02)

**Files changed:**
- `internal/worker/worker.go` - Removed stub handlers, wired real handlers from tasks package
- `cmd/ugc/main.go` - Updated Dependencies to use JobRepo instead of JobService

**Changes:**
- Removed 170+ lines of stub handler code
- Added import for `tasks` package
- Changed `Dependencies.JobService` → `Dependencies.JobRepo`
- Added `Dependencies.AsynqClient` and `Dependencies.WebhookBaseURL`
- Registered 6 real handlers: AnalyzeConcept, GenerateMusic, SelectSong, GenerateImage, ProcessVideo, UploadAssets
