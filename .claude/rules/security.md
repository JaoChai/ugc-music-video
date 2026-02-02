# Security Rules

## Mandatory Checks Before Every Commit

- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] All user inputs validated
- [ ] No SQL injection vulnerabilities
- [ ] No XSS vulnerabilities
- [ ] No command injection risks
- [ ] Error messages don't leak sensitive info
- [ ] Authentication properly enforced
- [ ] Authorization checks in place

## Secret Management

```go
// WRONG: Hardcoded secrets
apiKey := "sk-abc123..."

// CORRECT: Environment variables
apiKey := os.Getenv("OPENROUTER_API_KEY")
if apiKey == "" {
    return fmt.Errorf("OPENROUTER_API_KEY not configured")
}
```

```typescript
// WRONG
const apiKey = "sk-abc123..."

// CORRECT
const apiKey = process.env.OPENAI_API_KEY
if (!apiKey) {
  throw new Error("OPENAI_API_KEY not configured")
}
```

## Input Validation

Always validate and sanitize user input:

```go
// Go - validate with custom validation
func (h *JobHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateJobInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid JSON")
        return
    }

    if len(input.Concept) < 10 {
        response.Error(w, http.StatusBadRequest, "Concept too short")
        return
    }
}
```

```typescript
// TypeScript - use Zod
import { z } from 'zod'

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
})

const validated = schema.parse(input)
```

## Incident Response Protocol

When a security vulnerability is found:

1. **STOP** all other work immediately
2. **USE** security-reviewer agent to assess severity
3. **FIX** critical issues before any other commits
4. **ROTATE** any potentially compromised credentials
5. **AUDIT** codebase for similar vulnerabilities

## OWASP Top 10 Awareness

Always check for:
- Injection (SQL, NoSQL, Command)
- Broken Authentication
- Sensitive Data Exposure
- XML External Entities (XXE)
- Broken Access Control
- Security Misconfiguration
- Cross-Site Scripting (XSS)
- Insecure Deserialization
- Using Components with Known Vulnerabilities
- Insufficient Logging & Monitoring
