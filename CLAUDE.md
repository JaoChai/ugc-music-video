# UGC - AI Video Generator

AI-powered short-form video content generator using LLMs, Suno (music), NanoBanana (images), FFmpeg.

## Build Commands

### Backend (Go)
```bash
make dev            # Hot reload with air
make test           # Run tests
make lint           # Run linter
make migrate-up     # Database migrations
```

### Frontend
```bash
cd frontend && npm run dev    # Dev server :5173
npm run build                 # Production build
npm run lint                  # ESLint
```

## Code Style

### Go
- Services MUST define interfaces for testability
- Use zap logger, never fmt.Print
- Agents extend BaseAgent for ChatJSON()
- Error handling: wrap with pkg/errors

### TypeScript/React
- Feature-based organization in `features/`
- Server state: React Query (auto-refetch, caching)
- Auth state: Zustand with localStorage
- Forms: react-hook-form + Zod validation

## Testing
- Prefer running single tests: `make test -- -run TestName`
- Frontend: `npm test -- --testNamePattern="pattern"`

## Git Workflow (gh)
```bash
gh pr create --fill          # Create PR
gh pr checkout <number>      # Checkout PR
gh pr merge --squash         # Merge with squash
gh issue list                # List issues
gh issue create              # Create issue
```

- Branch naming: `feature/xxx`, `fix/xxx`, `refactor/xxx`
- Commit: ใช้ conventional commits (feat:, fix:, refactor:, docs:)
- PR: squash merge เป็นหลัก

## Environment Variables
Required: `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`, `OPENROUTER_API_KEY`, `KIE_API_KEY`, `KIE_BASE_URL`

Frontend: `VITE_API_BASE_URL`

## Current TODO
- [ ] Worker task handlers (8 tasks)
- [ ] Agent implementations (concept analyzer, song selector)
- [ ] Webhook endpoints for Suno/KIE callbacks
- [ ] Video processing logic with FFmpeg
