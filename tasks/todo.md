# Current Tasks

## High Priority

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

### [x] Consolidate duplicate job APIs (2026-02-02)

**Files changed:**
- Deleted `frontend/src/api/jobs.ts`
- Rewrote `frontend/src/features/job/api.ts` - correct endpoints `/api/v1/jobs`
- Rewrote `frontend/src/features/dashboard/hooks/useJobs.ts` - correct endpoints
- Updated `frontend/src/features/job/types.ts` - match backend JobResponse struct
- Fixed `frontend/src/features/job/pages/JobDetailPage.tsx` - use new field names
- Fixed `frontend/src/features/job/pages/JobListPage.tsx` - use new data structure
- Fixed `frontend/src/features/dashboard/components/RecentJobsList.tsx`

**Changes:**
- Removed duplicate api/jobs.ts file
- All API calls now use `/api/v1/jobs` endpoints
- Frontend Job type now matches backend JobResponse
- Fixed field names: `created` → `created_at`, `updated` → `updated_at`, `model` → `llm_model`
- Fixed song structure: `song` → `song_prompt` and `generated_songs`
- npm run build passes

---

### [x] Wire webhook handlers (2026-02-02)

**Files changed:**
- `cmd/ugc/main.go` - Used WebhookHandler.RegisterRoutes() instead of inline stubs

**Changes:**
- Webhook handlers were already implemented in `internal/handler/webhook_handler.go`
- Just needed to be wired in main.go

---

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
