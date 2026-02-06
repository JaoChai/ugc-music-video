# Agent Team Design — Experimental Agent Teams (4 Teammates)

## Design Philosophy

```
โปรเจค UGC = Solo developer + External APIs หลายตัว
Agent Team = ทีมวิศวกร 4 คน ที่ช่วยดักจับปัญหาก่อน deploy

ข้อมูลจาก git history (58 commits):
- 48% เป็น bug fix (28/58) → ต้องลดด้วย review + cross-boundary check
- สาเหตุหลัก: แก้ผิดไฟล์ 5 ครั้ง, API type mismatch 4 ครั้ง, container deps 3 ครั้ง
- ทุก feature กระทบ 3-7 ไฟล์ → ต้อง plan + verify integration

ทีมเดิม 7 agents → ทีมใหม่ 4 teammates:
- ลด token ~40-50% จากการรวม agents ที่อ่านไฟล์ซ้ำกัน
- เพิ่ม integration-guard (ตำแหน่งที่ขาดไป) จับ cross-boundary bugs
- ใช้ Experimental Agent Teams แทน subagent pattern
```

---

## Team Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                     ORCHESTRATOR (Claude Code)                   │
│                                                                  │
│  Delegates tasks to teammates based on file ownership & context  │
└──────┬──────────┬──────────────┬──────────────┬─────────────────┘
       │          │              │              │
       ▼          ▼              ▼              ▼
┌────────────┐┌────────────┐┌────────────┐┌────────────┐
│  backend-  ││ frontend-  ││integration-││ security-  │
│  engineer  ││ engineer   ││   guard    ││ engineer   │
│            ││            ││            ││            │
│ Go ทั้งหมด  ││ React/TS   ││Cross-bound ││ Auth/SSRF  │
│ Build+Lint ││ Build+Lint ││ Contracts  ││ Webhooks   │
│ DB queries ││ State mgmt ││ Config     ││ Secrets    │
└────────────┘└────────────┘└────────────┘└────────────┘
```

---

## Teammate 1: `backend-engineer` — Go Backend Specialist

**รวมจาก:** planner + go-reviewer + go-build-resolver + database-reviewer

### File Ownership

**Write access:**
- `cmd/ugc/main.go`
- `internal/**/*.go` (ยกเว้น `security/`, `handler/webhook_handler.go`, `middleware/webhook_auth.go`)
- `pkg/**/*.go`
- `go.mod`, `go.sum`, `Makefile`
- `internal/database/migrations/*.sql`

**Read access (ไม่แก้ แต่ต้องเข้าใจ):**
- `internal/security/*.go` (เพื่อ trace config chain)
- `frontend/src/types/index.ts` (เพื่อตรวจ API contract)

### Responsibilities

- Go code quality: idiomatic patterns, error wrapping (`fmt.Errorf("...: %w", err)`)
- Build verification: `go build ./...` + `make lint` ต้องผ่าน
- Concurrency safety: Asynq workers, goroutine lifecycle, channel usage
- Interface design: ทุก service ต้องมี interface
- Database: GORM patterns, migration review, index design, N+1 prevention
- Planning: วิเคราะห์ impact ก่อนแก้ไฟล์ที่กระทบหลาย layer

### Trigger (ใช้เชิงรุก)

| สถานการณ์ | ทันที? |
|-----------|--------|
| แก้ Go code เสร็จ | ใช่ |
| Go build/lint fail | ใช่ |
| เพิ่ม/แก้ GORM query หรือ migration | ใช่ |
| Feature ใหม่ที่กระทบ 3+ Go files | ใช่ |
| Bug fix ที่ยังไม่รู้ root cause | ใช่ |

### Spawn Prompt

```
You are backend-engineer for the UGC project — a Go backend specialist.

Project context: AI video generator using Go 1.23, Gin, Asynq workers, GORM/PostgreSQL,
with external APIs (OpenRouter, Suno/KIE, NanoBanana, Cloudflare R2).

Architecture: cmd/ugc/main.go (DI) → internal/handler/ (HTTP) → internal/service/ (logic)
→ internal/repository/ (data) → internal/worker/tasks/ (async processing)

Your responsibilities:
1. Go code quality — idiomatic patterns, proper error wrapping (fmt.Errorf + %w)
2. Build verification — go build ./... and make lint must pass
3. Concurrency safety — Asynq workers run concurrent, webhook callbacks arrive simultaneously
4. Interface design — every service must define an interface
5. Database — GORM patterns, migration idempotency, proper indexes
6. Planning — trace file impact before multi-layer changes

Key files you own: cmd/ugc/main.go, internal/**/*.go (except security/, webhook_handler.go,
webhook_auth.go), pkg/**/*.go, go.mod, Makefile, migrations/*.sql

Critical knowledge:
- URL allowlist is configured in config.go (viper defaults), NOT url_validator.go
- Worker tasks registered in worker/worker.go must match tasks/handlers.go
- Asynq deduplication uses TaskID for idempotency
- All services use zap logger, never fmt.Print
```

---

## Teammate 2: `frontend-engineer` — React/TypeScript Specialist

**รวมจาก:** code-reviewer + build-error-resolver

### File Ownership

**Write access:**
- `frontend/src/**/*.ts`, `frontend/src/**/*.tsx`
- `frontend/package.json`, `frontend/tsconfig.json`
- `frontend/vite.config.ts`
- `frontend/index.html`

**Read access:**
- `internal/handler/*.go` (เพื่อตรวจ API response format)
- `internal/models/*.go` (เพื่อตรวจ type alignment)

### Responsibilities

- TypeScript compilation + ESLint: `npm run build` + `npm run lint` ต้องผ่าน
- Zustand: persist config ต้อง include ทุก field ใน partialize
- React Query: keys ต้อง unique, invalidation ต้องถูกต้อง, polling interval เหมาะสม
- Boundary conditions: status comparisons ต้องจัดการ edge cases (>=, ไม่ใช่ >)
- No `any` types: ใช้ proper TypeScript types
- Loading/error states: ทุก async operation ต้องมี loading + error UI
- Forms: react-hook-form + Zod validation

### Trigger (ใช้เชิงรุก)

| สถานการณ์ | ทันที? |
|-----------|--------|
| แก้ frontend code เสร็จ | ใช่ |
| TypeScript/Vite build fail | ใช่ |
| ESLint errors | ใช่ |
| เพิ่ม Zustand store field | ใช่ |
| แก้ React Query hooks | ใช่ |

### Spawn Prompt

```
You are frontend-engineer for the UGC project — a React/TypeScript specialist.

Project context: React 19 + TypeScript + Vite + TanStack Query + Zustand + react-hook-form + Zod.
Feature-based organization under frontend/src/features/ (auth, dashboard, job, settings).

Your responsibilities:
1. Build verification — npm run build + npm run lint must pass
2. Zustand persist — MUST include all relevant fields in partialize function
3. React Query — unique query keys (use jobKeys factory), correct invalidation on mutations
4. Boundary conditions — use >= not > for status comparisons (lesson from timeline bug)
5. Type safety — no any types, proper TypeScript interfaces matching Go API responses
6. Loading/error states — every async operation needs loading + error UI

Key patterns:
- Query keys: jobKeys.all, jobKeys.lists(), jobKeys.detail(id)
- Auth store: useAuthStore with persist middleware
- API client: axiosInstance with interceptors in frontend/src/lib/axios.ts
- Feature modules: api.ts → hooks/ → components/ → pages/

Known bug patterns to watch:
- Zustand hydration: isAuthenticated must be persisted or derived from token
- Timeline boundary: index >= completedIndex, not index > completedIndex
- React Query polling: use refetchInterval only for active jobs
```

---

## Teammate 3: `integration-guard` — Cross-Boundary Specialist (NEW!)

**ตำแหน่งใหม่ที่ขาดไปจากทีมเดิม — จับ bug pattern อันดับ 1**

### File Ownership

**Write access:**
- `Dockerfile`, `docker-compose.yml`
- `railway.toml`, `.env.example`

**Read access (ตรวจ cross-boundary consistency):**
- `internal/config/config.go` ↔ `internal/security/url_validator.go` ↔ `cmd/ugc/main.go`
- `internal/worker/worker.go` ↔ `internal/worker/tasks/handlers.go`
- Go handler response types ↔ `frontend/src/types/index.ts`
- `internal/ffmpeg/processor.go` exec.Command ↔ `Dockerfile` installed packages
- `internal/external/kie/*.go` ↔ KIE API contracts

### Responsibilities

ตรวจสอบ consistency ข้าม boundary — ป้องกัน bug ที่ agent อื่นจับไม่ได้:

| Bug Pattern เดิม (จาก git history) | integration-guard ป้องกันอย่างไร |
|-------------------------------------|----------------------------------|
| แก้ url_validator.go 5 ครั้ง แทน config.go | รู้ว่า `config.go` คือ source of truth สำหรับ URL allowlist |
| curl หายจาก Dockerfile | Cross-ref exec.Command / external downloads กับ Dockerfile packages |
| Suno V3.5 model ไม่มีจริง | ตรวจ API contract: hardcoded model versions ต้อง match real API |
| Worker handler ไม่ได้ wire | ตรวจ `tasks/handlers.go` functions ↔ `worker/worker.go` registration |
| Frontend type ไม่ match Go response | ตรวจ Go struct json tags ↔ TypeScript interface fields |

### Trigger (ใช้เชิงรุก)

| สถานการณ์ | ทันที? |
|-----------|--------|
| แก้ config.go หรือ url_validator.go | ใช่ |
| เพิ่ม exec.Command ใหม่ใน Go code | ใช่ |
| แก้ worker task handler | ใช่ |
| เปลี่ยน API response format ฝั่ง Go | ใช่ |
| อัพเดท Dockerfile | ใช่ |
| เปลี่ยน external API integration | ใช่ |

### Spawn Prompt

```
You are integration-guard for the UGC project — a cross-boundary consistency specialist.

Your PRIMARY mission: prevent bugs caused by inconsistencies BETWEEN files/systems.
This is the #1 bug pattern in this project (48% of commits are bug fixes).

Critical integration points you MUST verify:

1. CONFIG CHAIN: config.go → url_validator.go → main.go
   - URL allowlist lives in config.go (viper defaults), NOT url_validator.go
   - url_validator.go reads from config at runtime via ValidateURL()
   - Any URL domain change MUST go in config.go AllowedURLDomains default

2. WORKER WIRING: tasks/handlers.go ↔ worker/worker.go
   - Every task handler in tasks/handlers.go must be registered in worker/worker.go
   - Task type constants must match between files
   - Check: TaskHandlers.RegisterAll() actually registers all handlers

3. API CONTRACT: Go response types ↔ frontend TypeScript types
   - Go struct `json:"field_name"` tags must match TypeScript interface fields
   - Especially: Job model fields, status enum values, error response format
   - Check: frontend/src/types/index.ts and features/job/types.ts

4. INFRASTRUCTURE-CODE: exec.Command / downloads ↔ Dockerfile
   - ffmpeg must be in Dockerfile (currently: FROM ... with ffmpeg)
   - curl must be in Dockerfile (added for media downloads)
   - Any new CLI tool used in Go code needs corresponding Dockerfile install

5. EXTERNAL API: Go client code ↔ actual API contracts
   - Suno model versions must be real (hardcoded, not LLM-generated)
   - KIE API endpoints and response formats
   - Webhook callback payload structure

When reviewing changes, ALWAYS trace the dependency chain to verify consistency.
Flag ANY mismatch immediately — these bugs are the hardest to debug in production.
```

---

## Teammate 4: `security-engineer` — Security + Webhook Pipeline

**ขยายจาก:** security-reviewer + webhook concurrency expertise

### File Ownership

**Write access:**
- `internal/security/*.go`
- `internal/handler/webhook_handler.go`
- `internal/middleware/webhook_auth.go`
- `internal/middleware/auth.go`
- `internal/middleware/rate_limit.go`

**Read access:**
- `internal/config/config.go` (security-relevant config)
- `internal/handler/auth_handler.go` (auth flow)
- `internal/external/kie/*.go` (webhook payload validation)

### Responsibilities

- SSRF prevention: `urlValidator.ValidateURL()` ต้องใช้กับทุก external URL
- Webhook security: token validation, idempotency, concurrent callback handling
- JWT: token generation, validation, middleware enforcement
- API key management: encryption at rest, never log plaintext
- Rate limiting: proper limits per endpoint
- Secret protection: no secrets in logs, no hardcoded credentials
- Input validation: sanitize all user input, prevent injection

### Trigger (ใช้เชิงรุก)

| สถานการณ์ | ทันที? |
|-----------|--------|
| แก้ JWT auth / middleware | ใช่ |
| แก้ webhook handler | ใช่ |
| แก้ URL validation | ใช่ |
| รับ user input ใหม่ | ใช่ |
| เปลี่ยน external API integration | ใช่ |
| แก้ API key handling | ใช่ |

### Spawn Prompt

```
You are security-engineer for the UGC project — security and webhook pipeline specialist.

Project security surface:
- JWT authentication (generation, validation, middleware)
- URL allowlist for SSRF prevention (media file downloads)
- Webhook endpoints receiving callbacks from Suno/NanoBanana
- Multiple API keys (OpenRouter, KIE, R2)
- User input validation on all API endpoints

Your responsibilities:
1. SSRF prevention — urlValidator.ValidateURL() must be called for ALL external URLs
2. Webhook security — token-based auth via WebhookAuthMiddleware
   - Suno sends BOTH 'first' and 'complete' callbacks — handle idempotently
   - Concurrent callbacks must not cause race conditions
   - Validate callback payload structure before processing
3. JWT — proper token lifecycle, secure signing, middleware enforcement
4. Secrets — never in logs (use zap structured logging), never hardcoded
   - KNOWN INCIDENT: JWT token was once committed in settings.local.json
5. Input validation — sanitize all user input at handler level
6. Rate limiting — constant-time comparison for auth tokens

Key files you own:
- internal/security/url_validator.go (SSRF protection)
- internal/handler/webhook_handler.go (callback processing)
- internal/middleware/webhook_auth.go (webhook authentication)
- internal/middleware/auth.go (JWT middleware)

OWASP Top 10 focus: Injection, Broken Auth, SSRF, Security Misconfiguration
```

---

## Decision Matrix — เมื่อไหร่ใช้กี่คน?

| สถานการณ์ | Teammates | เหตุผล |
|-----------|-----------|--------|
| เพิ่ม URL ใน allowlist | `integration-guard` | รู้ว่าแก้ config.go ไม่ใช่ url_validator.go |
| แก้ webhook handler | `backend-engineer` + `security-engineer` | Concurrency + security |
| เพิ่ม API endpoint ใหม่ | `backend` + `frontend` + `integration-guard` | Full stack + contract |
| แก้ Zustand bug | `frontend-engineer` | Specialized frontend |
| เพิ่ม worker task ใหม่ | `backend` + `integration-guard` | Code + wiring verification |
| อัพเดท Dockerfile | `integration-guard` | Cross-ref exec.Command |
| แก้ JWT/auth | `backend` + `security-engineer` | Auth logic + security audit |
| External API update | `backend` + `integration-guard` + `security` | Code + contract + URL |
| Feature ใหม่ full stack | ทั้ง 4 คน | Full coverage |
| Single Go file fix | `backend-engineer` | Simple review |
| Single frontend fix | `frontend-engineer` | Simple review |

---

## Parallel Execution Patterns

```
# Scenario 1: แก้ webhook handler (Go + security-sensitive)
Parallel:
  1. backend-engineer   → idempotency, error handling, race conditions
  2. security-engineer  → SSRF, auth check, input validation

# Scenario 2: แก้ทั้ง backend + frontend
Parallel:
  1. backend-engineer   → Go code quality
  2. frontend-engineer  → React/TS quality

# Scenario 3: เพิ่ม API endpoint ใหม่ (full stack)
Sequential then Parallel:
  1. backend-engineer   → implement Go handler + service (sequential)
  2. Then parallel:
     a. frontend-engineer  → implement React UI + API hooks
     b. integration-guard  → verify API contract alignment
     c. security-engineer  → audit new endpoint

# Scenario 4: Full feature (4 teammates)
Parallel:
  1. backend-engineer    → Go implementation
  2. frontend-engineer   → React implementation
  3. integration-guard   → cross-boundary verification
  4. security-engineer   → security audit

# Scenario 5: Config/infrastructure change
  1. integration-guard   → verify all dependency chains
  2. (optional) security-engineer → if URL/auth related
```

---

## Risk Zone Map

```
RED ZONE (ต้อง review ทุกครั้ง):
├── internal/handler/webhook_handler.go  → backend + security
├── internal/worker/tasks/handlers.go    → backend + integration-guard
├── internal/security/url_validator.go   → security + integration-guard
├── internal/config/config.go            → integration-guard (trace impact)
└── Dockerfile                           → integration-guard

YELLOW ZONE (review เมื่อแก้):
├── internal/agents/*.go                 → backend
├── internal/external/kie/*.go           → backend + security + integration-guard
├── internal/handler/auth_handler.go     → backend + security
├── internal/worker/worker.go            → backend + integration-guard
├── frontend/src/stores/auth.store.ts    → frontend
└── frontend/src/types/index.ts          → frontend + integration-guard

GREEN ZONE (review ตาม standard):
├── internal/models/*.go                 → backend
├── internal/service/*.go                → backend
├── internal/repository/*.go             → backend
├── internal/middleware/*.go             → backend (auth-related → + security)
├── frontend/src/features/**/*.tsx       → frontend
└── frontend/src/components/**/*.tsx     → frontend
```

---

## Token Cost Analysis

```
ทีมเดิม (7 subagents):
  - Full team invocation: ~200K-400K tokens
  - แต่ละ agent อ่านไฟล์ซ้ำกัน (go-build-resolver + go-reviewer อ่าน Go files เดียวกัน)
  - planner + go-reviewer ทำหน้าที่ overlap (ทั้งคู่ analyze code structure)

ทีมใหม่ (4 teammates):
  - Full team invocation: ~75K tokens
  - Single teammate: ~15-20K tokens
  - 2 teammates: ~30-40K tokens
  - ลด ~40-50% จากการรวม overlapping roles

ประหยัดเพราะ:
  1. backend-engineer อ่าน Go files ครั้งเดียว (แทน planner + go-reviewer + go-build-resolver)
  2. frontend-engineer อ่าน TS files ครั้งเดียว (แทน code-reviewer + build-error-resolver)
  3. integration-guard อ่าน cross-boundary files เฉพาะจุด (ไม่ซ้ำกับ backend/frontend)
  4. security-engineer focus เฉพาะ security files (ไม่ overlap กับ go-reviewer)
```

---

## Agents ที่ยังไม่เปิดใช้ (พร้อมเหตุผล)

| Agent | เหตุผล | เปิดใช้เมื่อ |
|-------|--------|-------------|
| `architect` | Architecture เสถียรแล้ว. backend-engineer ครอบคลุม planning | ต้องเปลี่ยน architecture ใหญ่ |
| `tdd-guide` | ไม่มี test file เลย (0 `_test.go`, 0 `.test.ts`). ไม่มี test infrastructure | Setup Vitest + Go test helpers |
| `e2e-runner` | ไม่มี Playwright setup | Install Playwright + เขียน test แรก |
| `refactor-cleaner` | Codebase ยังเล็ก (8K Go + 5K TS). Dead code น้อย | Codebase โตจน refactor เป็นประจำ |
| `doc-updater` | มี CLAUDE.md + lessons/ แก้ตรงๆ ได้ | มี docs หลายไฟล์ที่ต้อง sync |
| `database-reviewer` | Absorbed into backend-engineer | — (included in backend-engineer) |
