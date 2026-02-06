# Agent Team Design

## Design Philosophy

```
โปรเจค UGC = Solo developer + External APIs หลายตัว
Agent Team = ทีมวิศวกรที่ช่วยดักจับปัญหาก่อน deploy

ข้อมูลจาก git history (57 commits):
- 44% เป็น bug fix → ต้องลดด้วยการ review ก่อน commit
- สาเหตุหลัก: แก้ผิดไฟล์, ไม่ trace code path, container deps ขาด
- ทุก feature กระทบ 3-7 ไฟล์ → ต้อง plan ก่อนเสมอ
```

---

## Team: 7 Agents, 3 Phases

```
Phase 1: PLAN         Phase 2: BUILD           Phase 3: REVIEW
┌──────────┐         ┌────────────────┐        ┌─────────────────┐
│ planner  │ ──────▶ │go-build-       │        │ go-reviewer     │
│          │         │resolver        │ ─────▶ │ code-reviewer   │
└──────────┘         │build-error-    │        │ security-       │
                     │resolver        │        │ reviewer        │
                     └────────────────┘        └────────┬────────┘
                                                        │
                                               ┌────────┴────────┐
                                               │  database-      │
                                               │  reviewer       │
                                               │  (on-demand)    │
                                               └─────────────────┘
```

---

## Phase 1 — PLAN (ก่อนเขียน code)

### `planner` — Tech Lead

**บทบาท:** วิเคราะห์ requirement, ระบุไฟล์ที่ต้องแก้, วาง implementation plan

**Trigger:**
- Feature ใหม่ที่กระทบ 3+ ไฟล์
- Bug fix ที่ยังไม่รู้ root cause
- Requirements ไม่ชัด หรือมีหลายวิธีทำ
- Refactoring ข้าม layer (handler → service → worker)

**ทำไมต้องมี:**
URL allowlist bug ถูกแก้ผิดไฟล์ 5 ครั้ง (url_validator.go แทน config.go)
เพราะไม่มีใคร trace code path ก่อน. Planner บังคับให้คิดก่อนทำ.

**ตัวอย่างจริง:**
- เพิ่ม URL domain ใหม่ → planner บอกแก้ `config.go` (viper default) ไม่ใช่ `url_validator.go`
- เพิ่ม API endpoint → planner ระบุ handler + service + worker + model ที่ต้องแก้

---

## Phase 2 — BUILD (ระหว่างเขียน code)

### `go-build-resolver` — Go Build Engineer

**บทบาท:** แก้ Go compilation errors, go vet warnings, golangci-lint errors

**Trigger:**
- `go build ./...` fail
- `make lint` fail
- Type mismatch, missing imports, interface ไม่ครบ

**ทำไมต้องมี:**
Backend มี 6-7 layers. แก้ไฟล์หนึ่งอาจทำให้อีก layer พัง.
เช่น เปลี่ยน `JobService` interface → handler, worker ต้อง update ด้วย.

### `build-error-resolver` — Frontend Build Engineer

**บทบาท:** แก้ TypeScript compilation, ESLint, Vite build errors

**Trigger:**
- `npm run build` fail
- `npm run lint` fail
- TypeScript type errors

**ทำไมต้องมี:**
React 19 + TypeScript 5.9 + Zod 4 type system เข้มงวด.
เปลี่ยน API response ฝั่ง Go จะทำให้ frontend type ไม่ match.

---

## Phase 3 — REVIEW (หลังเขียน code, ก่อน commit)

### `go-reviewer` — Senior Go Engineer

**บทบาท:** Review idiomatic Go, error handling, concurrency safety, interface design

**Trigger:** หลังแก้ Go code ทุกครั้ง (handlers, services, workers, agents)

**ทำไมต้องมี:**
Asynq workers ทำงาน concurrent, webhook callbacks มาพร้อมกันได้ (Suno ส่ง first + complete callback).
ต้องมีคน review concurrency patterns และ idempotency.

**จับอะไรได้:**
- Race conditions ใน webhook handler
- Error ที่ไม่ถูก wrap ด้วย context (`fmt.Errorf("...: %w", err)`)
- Interface ที่ไม่ครบ method
- Goroutine ที่ leak

### `code-reviewer` — Frontend Engineer

**บทบาท:** Review React/TypeScript code quality, hooks patterns, state management

**Trigger:** หลังแก้ frontend code ทุกครั้ง

**ทำไมต้องมี:**
เคยมี bug Zustand hydration (spinner stuck) เพราะ `isAuthenticated` ไม่ถูก persist.
เคยมี boundary condition bug ใน progress timeline (index 7 > 7 = false → completed แสดง spinner).

**จับอะไรได้:**
- Zustand persist ที่ลืม field ใน partialize
- React Query key ที่ไม่ unique
- Boundary conditions ใน status comparisons
- Missing loading/error states

### `security-reviewer` — Security Engineer

**บทบาท:** ตรวจ SSRF, injection, auth bypass, secret exposure, URL validation

**Trigger:**
- แก้ JWT auth, middleware, webhook handler
- เพิ่ม/แก้ URL validation
- รับ user input ใหม่ (API endpoints)
- เปลี่ยน external API integration

**ทำไมต้องมี:**
โปรเจคมี attack surface กว้าง:
- JWT auth (token generation, validation, middleware)
- URL allowlist (SSRF protection สำหรับ media downloads)
- Webhook endpoints (รับ callback จาก Suno/NanoBanana)
- API keys หลายตัว (OpenRouter, KIE, R2)
- เคยมี JWT token หลุดใน settings.local.json

**จับอะไรได้:**
- Webhook endpoint ที่ไม่มี auth check
- URL validation bypass
- Secret ที่ hardcode ใน code
- SQL injection จาก GORM raw queries

---

## On-Demand — ใช้ตามสถานการณ์

### `database-reviewer` — DBA

**บทบาท:** PostgreSQL query optimization, GORM patterns, migration review, index design

**Trigger:**
- เพิ่ม/แก้ GORM queries ที่ซับซ้อน (JOIN, subquery, aggregation)
- เขียน database migration ใหม่
- Performance issue จาก database
- เพิ่ม index

**ทำไมต้องมี:**
Job model มี status transitions หลายขั้น, query ที่ filter by status + user_id ต้องมี index เหมาะสม.
Migration ต้อง idempotent เพราะ auto-run on startup — fail = app crash.

---

## ไม่รวมในทีม (พร้อมเหตุผล)

| Agent | เหตุผล | เปิดใช้เมื่อ |
|-------|--------|-------------|
| `architect` | Architecture เสถียรแล้ว (handler→service→worker→API). ปัญหาอยู่ที่ integration ไม่ใช่ architecture. Planner ครอบคลุม 95% | ต้องเปลี่ยน architecture ใหญ่ (เช่น เพิ่ม message queue) |
| `tdd-guide` | ไม่มี test file เลย (0 `_test.go`, 0 `.test.ts`). ไม่มี Vitest/Jest config. Agent ใช้ไม่ได้จริง | Setup test infrastructure (Vitest + Go test helpers) |
| `e2e-runner` | ไม่มี Playwright setup ใน project. Agent จะ error ทันที | Install Playwright + เขียน test แรก |
| `refactor-cleaner` | 57 commits, codebase ยังเล็ก. Dead code น้อยมาก | Codebase โตจน refactor เป็นประจำ |
| `doc-updater` | มี CLAUDE.md + lessons/ directory. แก้ตรงๆ ได้ไม่ต้อง agent แยก | มี docs หลายไฟล์ที่ต้อง sync กัน |

---

## Proactive Usage Rules

ใช้ agent เชิงรุก ไม่ต้องรอ user ขอ:

| สถานการณ์ | Agent | ทันที? |
|-----------|-------|--------|
| Feature request ที่ซับซ้อน | `planner` | ใช่ |
| แก้ Go code เสร็จ | `go-reviewer` | ใช่ |
| แก้ frontend เสร็จ | `code-reviewer` | ใช่ |
| แก้ code ที่เกี่ยวกับ auth/webhook/URL | `security-reviewer` | ใช่ |
| Go build fail | `go-build-resolver` | ใช่ |
| TypeScript build fail | `build-error-resolver` | ใช่ |
| แก้ database query/migration | `database-reviewer` | ถามก่อน |

---

## Parallel Execution Patterns

รัน agent พร้อมกันเมื่อไม่มี dependency:

```
# Scenario 1: แก้ webhook handler (Go + security-sensitive)
Parallel:
  1. go-reviewer       → idempotency, error handling, race conditions
  2. security-reviewer → SSRF, auth check, input validation

# Scenario 2: แก้ทั้ง backend + frontend
Parallel:
  1. go-reviewer    → Go code quality
  2. code-reviewer  → React/TS quality

# Scenario 3: Full stack + security
Parallel:
  1. go-reviewer       → Go patterns
  2. code-reviewer     → TypeScript patterns
  3. security-reviewer → security audit ทั้ง stack

# Sequential (มี dependency):
  1. planner           → วางแผนก่อน (ต้องรู้ไฟล์ที่จะแก้)
  2. เขียน code        → ตาม plan
  3. go-build-resolver → ถ้า build fail
  4. go-reviewer + security-reviewer (parallel) → review
```

---

## Risk Zone Map (ไฟล์ไหนต้องใช้ agent ไหน)

```
RED ZONE (ต้อง review ทุกครั้ง):
├── internal/handler/webhook_handler.go  → go-reviewer + security-reviewer
├── internal/worker/tasks/handlers.go    → go-reviewer
├── internal/security/url_validator.go   → security-reviewer
└── internal/config/config.go            → planner (trace impact ก่อน)

YELLOW ZONE (review เมื่อแก้):
├── internal/agents/*.go                 → go-reviewer
├── internal/external/kie/*.go           → go-reviewer + security-reviewer
├── internal/handler/auth_handler.go     → security-reviewer
├── frontend/src/stores/auth.store.ts    → code-reviewer
└── Dockerfile                           → security-reviewer

GREEN ZONE (review ตาม standard):
├── internal/models/*.go                 → go-reviewer
├── internal/middleware/*.go             → go-reviewer
├── frontend/src/features/**/*.tsx       → code-reviewer
└── frontend/src/components/**/*.tsx     → code-reviewer
```
