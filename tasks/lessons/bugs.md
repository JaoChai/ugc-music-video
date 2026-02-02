# Bug Patterns & Fixes

Document recurring bugs and their solutions to prevent future occurrences.

---

## Go Backend

### [Template] Nil Pointer Dereference on Optional Fields

**Context:** Job model has optional fields like `AudioURL *string`

**Problem:** Accessing `*job.AudioURL` when it's nil causes panic

**Solution:**
```go
// Always check for nil before dereferencing
if job.AudioURL != nil && *job.AudioURL != "" {
    // Safe to use *job.AudioURL
}

// Or use helper function
func SafeString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

**Prevention:**
- Add nil checks before dereferencing pointers
- Consider using sql.NullString for database fields

---

### [Template] Context Cancellation Not Handled

**Context:** Long-running operations like FFmpeg processing

**Problem:** Operation continues even after client disconnects

**Solution:**
```go
func (p *Processor) CreateVideo(ctx context.Context, ...) error {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // For exec commands, use CommandContext
    cmd := exec.CommandContext(ctx, "ffmpeg", args...)
}
```

**Prevention:**
- Always use `ctx` parameter
- Use `exec.CommandContext` for external commands
- Check `ctx.Done()` in loops

---

## React Frontend

### [Template] Stale Closure in useEffect

**Context:** Using state variables inside useEffect callbacks

**Problem:** Callback captures old state value

**Solution:**
```typescript
// Bad - stale closure
useEffect(() => {
    const interval = setInterval(() => {
        console.log(count); // Always logs initial value
    }, 1000);
    return () => clearInterval(interval);
}, []); // Missing dependency

// Good - include dependency
useEffect(() => {
    const interval = setInterval(() => {
        console.log(count);
    }, 1000);
    return () => clearInterval(interval);
}, [count]); // Recreates interval when count changes

// Better - use functional update
useEffect(() => {
    const interval = setInterval(() => {
        setCount(c => c + 1); // No dependency needed
    }, 1000);
    return () => clearInterval(interval);
}, []);
```

**Prevention:**
- Use ESLint exhaustive-deps rule
- Prefer functional updates for setState in effects
- Use useCallback for stable function references

---

### [Template] Race Condition in API Calls

**Context:** User quickly navigates between pages

**Problem:** Old API response overwrites newer data

**Solution:**
```typescript
// Use React Query - handles this automatically
const { data } = useQuery({
    queryKey: ['job', id],
    queryFn: () => fetchJob(id),
});

// Or with useEffect, use abort controller
useEffect(() => {
    const controller = new AbortController();

    fetchJob(id, { signal: controller.signal })
        .then(setData)
        .catch(err => {
            if (err.name !== 'AbortError') {
                setError(err);
            }
        });

    return () => controller.abort();
}, [id]);
```

**Prevention:**
- Prefer React Query for data fetching
- Always cleanup async operations in useEffect
- Use abort controllers for fetch requests

---

## Database

### [Template] N+1 Query Problem

**Context:** Loading jobs with related data

**Problem:** Separate query for each job's user

**Solution:**
```go
// Bad - N+1 queries
jobs, _ := jobRepo.List(ctx)
for _, job := range jobs {
    user, _ := userRepo.GetByID(ctx, job.UserID) // N queries
}

// Good - Preload/Join
func (r *JobRepository) ListWithUser(ctx context.Context) ([]JobWithUser, error) {
    var jobs []JobWithUser
    err := r.db.WithContext(ctx).
        Table("jobs").
        Select("jobs.*, users.email as user_email").
        Joins("LEFT JOIN users ON users.id = jobs.user_id").
        Find(&jobs).Error
    return jobs, err
}
```

**Prevention:**
- Use GORM's Preload for relationships
- Monitor query count in development
- Consider adding query logging in development

---

### [2026-02-02] Zustand Persist Hydration Race Condition

**Context:** Using Zustand with persist middleware and checking hydration state in PrivateRoute

**Problem:** Spinner stuck after login because `useSyncExternalStore` didn't trigger re-render when hydration completed between `getSnapshot()` and `subscribe()` calls.

**Root Cause:**
```typescript
// BAD - Race condition in useSyncExternalStore
export function useHasHydrated(): boolean {
  return useSyncExternalStore(
    (onStoreChange) => {
      // BUG: If hydrated, returns empty function without calling onStoreChange
      if (useAuthStore.persist.hasHydrated()) {
        return () => {}  // No re-render triggered!
      }
      return useAuthStore.persist.onFinishHydration(onStoreChange)
    },
    () => useAuthStore.persist.hasHydrated(),
    () => false
  )
}
```

Timeline:
1. `getSnapshot()` returns `false` (hydration not done)
2. Component renders with spinner
3. Hydration completes (between render and subscribe)
4. `subscribe()` sees `hasHydrated() = true`, returns empty function
5. No listener registered → No re-render → Spinner stuck forever

**Solution:**
```typescript
// GOOD - useState + useEffect with race condition handling
export function useHasHydrated(): boolean {
  const [hasHydrated, setHasHydrated] = useState(
    useAuthStore.persist.hasHydrated()
  )

  useEffect(() => {
    // Skip subscription if already hydrated
    if (hasHydrated) return

    const unsubscribe = useAuthStore.persist.onFinishHydration(() => {
      setHasHydrated(true)
    })

    // Handle race condition: hydration may have completed between
    // useState initialization and this effect running
    if (useAuthStore.persist.hasHydrated()) {
      queueMicrotask(() => setHasHydrated(true))
    }

    return unsubscribe
  }, [hasHydrated])

  return hasHydrated
}
```

**Why this works:**
- `useState` captures initial value synchronously
- `useEffect` subscribes to future changes
- Double-check in effect catches race condition
- `queueMicrotask` avoids synchronous setState in effect body (ESLint rule)
- Early return when hydrated skips unnecessary subscription

**Prevention:**
- Avoid `useSyncExternalStore` for Zustand persist hydration state
- Always double-check async state in useEffect after subscribing
- Use `queueMicrotask` or `setTimeout` for deferred setState in effects
- Test login flow with CPU throttling to catch race conditions

**Related Files:**
- `frontend/src/hooks/useHasHydrated.ts`
- `frontend/src/components/PrivateRoute.tsx`
- `frontend/src/stores/auth.store.ts`

---

## Add New Bug Patterns Below

_When you encounter a bug, document it here following the template above_
