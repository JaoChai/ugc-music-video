# Backlog

Future work items that are not prioritized for current sprint.

---

## Features

### User Settings for LLM Model Selection

**Description:** Allow users to select their preferred LLM model in settings.

**Notes:**
- User model already has `openrouter_model` field
- Need settings page UI
- Need API endpoint to update user settings

---

### Job Retry Functionality

**Description:** Allow users to retry failed jobs.

**Notes:**
- Need "Retry" button on failed jobs
- Should reset status to pending and re-enqueue
- Consider which step to retry from (start or failed step)

---

### Job History/Timeline View

**Description:** Show detailed timeline of job processing steps.

**Notes:**
- Track timestamps for each status change
- Show which agent/service processed each step
- Display any intermediate outputs (song options, image prompt)

---

### Batch Job Creation

**Description:** Create multiple jobs from a list of concepts.

**Notes:**
- CSV upload or multi-line text input
- Queue management to avoid overwhelming services
- Progress tracking for batch

---

## Technical Improvements

### Add Request Rate Limiting

**Description:** Prevent API abuse with rate limiting.

**Notes:**
- Use Redis for rate limit counters
- Per-user limits for job creation
- Global limits for external API calls

---

### Add Structured Logging with Correlation IDs

**Description:** Track requests across services with correlation IDs.

**Notes:**
- Generate correlation ID in middleware
- Pass through context to all services
- Include in external API calls

---

### Add Metrics/Monitoring

**Description:** Prometheus metrics for monitoring.

**Notes:**
- Job processing duration
- Success/failure rates
- External API latency
- Queue depth

---

### Add Integration Tests

**Description:** End-to-end tests for job pipeline.

**Notes:**
- Mock external services (Suno, NanoBanana)
- Test full job lifecycle
- Test error handling at each step

---

## Technical Debt

### Refactor Worker Dependencies

**Description:** Worker has two different Dependencies structs.

**Location:** `worker/worker.go` vs `worker/tasks/types.go`

**Notes:**
- Consolidate to single Dependencies struct
- Move to tasks package
- Update worker initialization

---

### Add Database Indexes

**Description:** Optimize queries with proper indexes.

**Tables:** jobs, users

**Notes:**
- Index on jobs.user_id for listing
- Index on jobs.status for filtering
- Index on jobs.created_at for sorting

---

### Extract API Response Types

**Description:** Shared response types for frontend/backend.

**Notes:**
- Consider OpenAPI spec generation
- Or shared TypeScript definitions
- Ensure consistency between API and frontend types
