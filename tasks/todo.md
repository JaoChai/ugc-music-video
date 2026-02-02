# Current Tasks

No open tasks. All tasks completed.

---

## Completed

### [x] Add authentication guards (2026-02-02)

**Files changed:**
- Created `frontend/src/components/PrivateRoute.tsx`
- Created `frontend/src/components/index.ts`
- Updated `frontend/src/router/index.tsx` - wrap protected routes with PrivateRoute
- Updated `frontend/src/features/auth/hooks/useAuth.ts` - redirect to intended destination after login

**Changes:**
- PrivateRoute component redirects to `/login` if not authenticated
- Preserves intended destination in location state
- After login, redirects back to intended page
- Protected routes: `/`, `/jobs`, `/jobs/create`, `/jobs/:id`, `/settings`

---

### [x] Add error boundaries (2026-02-02)

**Files changed:**
- Created `frontend/src/components/ErrorBoundary.tsx`
- Updated `frontend/src/App.tsx` - wrap app with ErrorBoundary

**Changes:**
- ErrorBoundary component catches React render errors
- Shows user-friendly error page with "Try Again" and "Refresh" buttons
- Shows error details in development mode
- Logs errors to console

---

### [x] Fix route inconsistencies (2026-02-02)

**Files changed:**
- Updated `frontend/src/features/dashboard/pages/DashboardPage.tsx`

**Changes:**
- Changed `/jobs/new` → `/jobs/create` in Quick Actions link
- All routes now consistently use `/jobs/create`

---

### [x] Add DownloadButton to JobDetailPage (2026-02-02)

**Status:** Already implemented

**Verification:**
- DownloadButton component exists in `frontend/src/features/job/components/VideoPlayer.tsx`
- Already exported from components index
- Already used in JobDetailPage when video_url exists

---

### [x] Consolidate duplicate job APIs (2026-02-02)

**Files changed:**
- Deleted `frontend/src/api/jobs.ts`
- Rewrote `frontend/src/features/job/api.ts` - correct endpoints `/api/v1/jobs`
- Rewrote `frontend/src/features/dashboard/hooks/useJobs.ts` - correct endpoints
- Updated `frontend/src/features/job/types.ts` - match backend JobResponse struct
- Fixed `frontend/src/features/job/pages/JobDetailPage.tsx` - use new field names
- Fixed `frontend/src/features/job/pages/JobListPage.tsx` - use new data structure
- Fixed `frontend/src/features/dashboard/components/RecentJobsList.tsx`
- Updated `frontend/src/api/index.ts` - removed jobs export

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
