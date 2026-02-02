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

### [2026-02-02] Zustand Persist Hydration - Spinner Stuck After Login/Refresh

**Context:** Using Zustand with persist middleware and checking auth state in PrivateRoute

**Problem:** Spinner stuck indefinitely after login or page refresh because hydration state tracking was overly complex.

**Root Cause:**
The original approach tried to track hydration state separately using hooks like `useHasHydrated()` with `useSyncExternalStore` or `useState/useEffect`. These approaches had race conditions:

1. `isAuthenticated` was NOT persisted to localStorage (only `user` and `token`)
2. After page refresh, `isAuthenticated` defaulted to `false`
3. `onRehydrateStorage` callback was supposed to set `isAuthenticated: true` but didn't fire reliably
4. Various hydration tracking hooks had race conditions between React lifecycle and Zustand persist

```typescript
// BAD - Complex hydration tracking that doesn't work reliably
partialize: (state) => ({ user: state.user, token: state.token }), // Missing isAuthenticated!

onRehydrateStorage: () => (state) => {
  // This callback doesn't fire reliably in all scenarios
  useAuthStore.setState({ isAuthenticated: !!state?.token })
},
```

**Solution:**
**Simply persist `isAuthenticated` directly!** No need for hydration tracking at all.

```typescript
// GOOD - Persist isAuthenticated directly
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      login: (user, token) => {
        localStorage.setItem('auth_token', token)
        set({ user, token, isAuthenticated: true })
      },
      logout: () => {
        localStorage.removeItem('auth_token')
        set({ user: null, token: null, isAuthenticated: false })
      },
    }),
    {
      name: 'auth-storage',
      // Key fix: persist isAuthenticated so it's available immediately after hydration
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated  // <-- This is the fix!
      }),
    }
  )
)

// PrivateRoute becomes simple - no hydration check needed
export function PrivateRoute({ children }: PrivateRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}
```

**Why this works:**
- `isAuthenticated` is saved to localStorage on login
- After page refresh, Zustand hydrates `isAuthenticated` directly from localStorage
- No race conditions because value is immediately available
- No need for `onRehydrateStorage`, `_hasHydrated`, or `useHasHydrated` hooks

**Prevention:**
- When using Zustand persist, include ALL auth-related state in `partialize`
- Avoid complex hydration tracking - persist the values you need directly
- Don't rely on `onRehydrateStorage` callback for critical state updates
- Keep auth logic simple: if it needs to survive page refresh, persist it

**Related Files:**
- `frontend/src/stores/auth.store.ts`
- `frontend/src/components/PrivateRoute.tsx`

---

## Add New Bug Patterns Below

_When you encounter a bug, document it here following the template above_
