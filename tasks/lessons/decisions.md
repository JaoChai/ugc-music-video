# Architecture Decision Records

Document significant architecture decisions and their rationale.

---

## ADR-001: Use Asynq for Background Jobs

**Date:** 2024

**Status:** Accepted

**Context:**
Need background job processing for video generation pipeline. Options considered:
- Temporal (workflow orchestration)
- Asynq (Redis-backed job queue)
- Custom goroutine-based worker

**Decision:**
Use Asynq with Redis for background job processing.

**Rationale:**
- Simple to set up and use
- Built-in retry logic with backoff
- Task deduplication
- Dashboard for monitoring (asynqmon)
- Scales horizontally by adding workers
- Lower complexity than Temporal for our needs

**Consequences:**
- Requires Redis infrastructure
- No built-in workflow state machine (we implement task chain pattern)
- Jobs must be idempotent for retry safety

---

## ADR-002: Feature-based Frontend Architecture

**Date:** 2024

**Status:** Accepted

**Context:**
Need to organize React frontend code. Options considered:
- Type-based organization (components/, hooks/, pages/)
- Feature-based organization (features/auth/, features/job/)

**Decision:**
Use feature-based organization with shared components/ui.

**Rationale:**
- Co-locates related code (easier navigation)
- Clear ownership boundaries
- Easier to extract features to separate packages
- Scales better as app grows

**Consequences:**
- May have some duplication across features
- Need clear guidelines on what goes in shared vs feature
- Barrel exports for public API

**Structure:**
```
src/
├── components/ui/    # Shared UI components
├── features/
│   ├── auth/        # Auth feature
│   ├── dashboard/   # Dashboard feature
│   └── job/         # Job feature
└── lib/             # Shared utilities
```

---

## ADR-003: React Query for Server State

**Date:** 2024

**Status:** Accepted

**Context:**
Need to manage server state in React. Options considered:
- useState + useEffect
- Redux + RTK Query
- React Query (TanStack Query)
- SWR

**Decision:**
Use React Query for all server state management.

**Rationale:**
- Automatic caching and cache invalidation
- Built-in loading/error states
- Automatic refetching (on focus, on mount)
- Optimistic updates
- Smaller bundle than Redux
- No boilerplate for mutations

**Consequences:**
- Server state separate from UI state
- Need Zustand or similar for non-server state (auth tokens)
- Learning curve for query keys and invalidation

---

## ADR-004: Webhook vs Polling for External Services

**Date:** 2024

**Status:** Accepted

**Context:**
Suno and NanoBanana APIs can either:
- Return immediately and callback via webhook when done
- Return immediately and we poll for completion

**Decision:**
Support both modes, prefer webhooks when configured.

**Rationale:**
- Webhooks are more efficient (no polling overhead)
- Polling works for local development without public URL
- Some deployments can't receive webhooks (firewalls)

**Implementation:**
```go
if deps.WebhookBaseURL != "" {
    req.CallBackUrl = fmt.Sprintf("%s/webhooks/suno/%s", deps.WebhookBaseURL, jobID)
    return nil  // Webhook will trigger next task
}

// Fallback to polling
taskResp, err := deps.SunoClient.WaitForCompletion(ctx, taskID, 10*time.Minute)
```

**Consequences:**
- Need webhook handler implementation
- Need timeout handling for polling mode
- Environment variable to configure mode

---

## ADR-005: Zustand for Auth State

**Date:** 2024

**Status:** Accepted

**Context:**
Need to persist authentication state (JWT token, user info) across page refreshes.

**Decision:**
Use Zustand with persist middleware for auth state only.

**Rationale:**
- Lightweight (2KB vs 42KB for Redux)
- Built-in persist middleware
- Simple API (no boilerplate)
- Works well alongside React Query

**Implementation:**
```typescript
export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            token: null,
            user: null,
            setAuth: (token, user) => set({ token, user }),
            logout: () => set({ token: null, user: null }),
        }),
        { name: 'auth-storage' }
    )
)
```

**Consequences:**
- Auth state in localStorage (consider security implications)
- Clear separation: Zustand for auth, React Query for data

---

## Template for New Decisions

```markdown
## ADR-XXX: Title

**Date:** YYYY-MM-DD

**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-XXX

**Context:**
What is the issue that we're seeing that is motivating this decision?

**Decision:**
What is the change that we're proposing and/or doing?

**Rationale:**
Why is this the best choice? What alternatives were considered?

**Consequences:**
What becomes easier or more difficult because of this decision?
```

---

## Add New Decisions Below

_When making significant architecture decisions, document them here_
