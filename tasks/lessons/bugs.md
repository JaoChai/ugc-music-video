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

### [2026-02-04] Error 422 - แก้ผิดที่เพราะไม่ Trace Code Path

**Context:** Music generation ล้มเหลวด้วย Error 422 จาก Suno API

**Problem:** LLM agent เลือก model "V3.5" ที่ไม่มีอยู่จริง ทำให้ Suno API reject

**การแก้ไขที่ผิดพลาด:**
```go
// Fix #1: Force ใน handlers.go - ยังไม่หาย
job.SongPrompt.Model = "V5"

// Fix #2: ลบ Model field จาก agent output - สำเร็จ!
type SongConceptOutput struct {
    Prompt string `json:"prompt"`
    // Model string `json:"model"` // ลบออก
}

func (o *SongConceptOutput) ToSongPrompt() *models.SongPrompt {
    return &models.SongPrompt{
        Model: "V5", // Hardcode ที่นี่แทน
    }
}
```

**บทเรียน:**
- LLM ไม่ควร output ค่าที่ต้องมี exact match กับ external API
- API-specific values (model version, aspect ratio) ควร hardcode ใน code
- LLM ควรโฟกัส creative work เท่านั้น

**Prevention:**
- ไม่ให้ LLM เลือกค่าที่มี validation จาก external API
- Review agent output schema ว่ามี field ไหนที่ควร hardcode

---

### [2026-02-04] URL Allowlist - แก้ไฟล์ที่ไม่ถูกใช้งาน

**Context:** Audio URL จาก `musicfile.kie.ai` ถูก reject ด้วย "host not in allowlist"

**Problem:** แก้ไข `DefaultAllowedHosts` ใน `url_validator.go` หลายรอบ แต่ปัญหาไม่หาย

**สาเหตุที่แท้จริง:**
```go
// url_validator.go - แก้ตรงนี้
var DefaultAllowedHosts = []string{
    "musicfile.kie.ai", // เพิ่มแล้ว แต่ไม่มีผล!
}

func NewURLValidator(allowedHosts []string) *URLValidator {
    hosts := allowedHosts
    if len(hosts) == 0 {
        hosts = DefaultAllowedHosts // ใช้เมื่อ empty เท่านั้น!
    }
}

// config.go - ปัญหาอยู่ตรงนี้
viper.SetDefault("WEBHOOK_ALLOWED_HOSTS",
    "cdn1.suno.ai,cdn2.suno.ai,cdn.kie.ai,storage.kie.ai")
// ไม่มี musicfile.kie.ai!

// main.go - ส่งค่าจาก config เสมอ
urlValidator := security.NewURLValidator(cfg.Webhook.AllowedHosts)
// cfg.Webhook.AllowedHosts ไม่เคย empty → DefaultAllowedHosts ไม่ถูกใช้
```

**Solution:**
```go
// config.go - เพิ่ม hosts ครบ
viper.SetDefault("WEBHOOK_ALLOWED_HOSTS",
    "cdn1.suno.ai,cdn2.suno.ai,cdn.kie.ai,storage.kie.ai,musicfile.kie.ai,aiquickdraw.com")
```

**บทเรียน:**
- **ต้อง trace code path ทั้งหมดก่อนแก้ไข**
- Default values อาจไม่ถูกใช้ถ้ามี config override
- แก้ 10 รอบก็ไม่หาย ถ้าแก้ผิดที่

**Prevention:**
- ใช้ grep หา usage ของ function/variable ก่อนแก้
- ดู call hierarchy: ใครเรียก? ส่งค่าอะไรมา?
- ถ้าแก้แล้วไม่หาย → หยุด แล้ววิเคราะห์ใหม่

---

### [2026-02-04] Container Missing Dependency (curl)

**Context:** Video processing ล้มเหลวใน production แต่ local ทำงานได้

**Problem:**
```
Error: curl: executable file not found in $PATH
```

**สาเหตุ:**
```go
// processor.go ใช้ curl
func downloadFile(ctx context.Context, url, destPath string) error {
    cmd := exec.CommandContext(ctx, "curl", "-L", "-o", destPath, url)
    return cmd.Run()
}

// Dockerfile ไม่ได้ติดตั้ง curl
RUN apk add --no-cache ffmpeg ca-certificates tzdata
// ไม่มี curl!
```

**Solution:**
```dockerfile
RUN apk add --no-cache ffmpeg curl ca-certificates tzdata
```

**บทเรียน:**
- **Local environment ≠ Production environment**
- เมื่อ code ใช้ external command ต้องตรวจสอบ Dockerfile
- Error ที่เกิดเฉพาะ production มักเกี่ยวกับ dependencies

**Prevention:**
- เมื่อเพิ่ม `exec.Command()` ใหม่ → ตรวจสอบ Dockerfile
- ใช้ Go native library แทน external command ถ้าเป็นไปได้
- Test ใน Docker locally ก่อน deploy

---

### [2026-02-04] Frontend Progress Timeline - Terminal State Logic

**Context:** Job status "completed" แต่ progress timeline ยังแสดง spinner

**Problem:**
```typescript
// Logic เดิม
const isCompleted = currentIndex > stepIndex  // 7 > 7 = false
const isCurrent = currentStatus === step.status // true → spinner

// เมื่อ currentStatus = 'completed' (index 7)
// และ step.status = 'completed' (index 7)
// → isCompleted = false, isCurrent = true → แสดง spinner ❌
```

**Solution:**
```typescript
// Don't show spinner for terminal states
const isCurrent = currentStatus === step.status && currentStatus !== 'completed'

// สำหรับ text color
(isCompleted || (currentStatus === 'completed' && step.status === 'completed')) && 'text-green-600'
```

**บทเรียน:**
- Terminal states (completed, failed) ต้อง handle แยก
- Index comparison มี edge case ที่ boundary

**Prevention:**
- ทดสอบ UI กับทุก status รวมถึง terminal states
- เขียน test cases สำหรับ boundary conditions

---

## Add New Bug Patterns Below

_When you encounter a bug, document it here following the template above_
