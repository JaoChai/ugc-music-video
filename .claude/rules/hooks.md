# Hooks System

## Hook Types

- **PreToolUse**: Before tool execution (validation, parameter modification)
- **PostToolUse**: After tool execution (auto-format, checks)
- **Stop**: When session ends (final verification)

## Recommended Hooks

### PreToolUse Hooks

**Long-running command reminder:**
```json
{
  "type": "preToolUse",
  "tool": "Bash",
  "pattern": "npm|pnpm|yarn|cargo|go build",
  "action": "suggest_tmux"
}
```

**Git push review:**
```json
{
  "type": "preToolUse",
  "tool": "Bash",
  "pattern": "git push",
  "action": "prompt_review"
}
```

### PostToolUse Hooks

**Auto-format after edit:**
```json
{
  "type": "postToolUse",
  "tool": "Edit",
  "pattern": "\\.(ts|tsx|js|jsx)$",
  "command": "prettier --write"
}
```

**TypeScript check:**
```json
{
  "type": "postToolUse",
  "tool": "Edit",
  "pattern": "\\.(ts|tsx)$",
  "command": "tsc --noEmit"
}
```

**Console.log warning:**
```json
{
  "type": "postToolUse",
  "tool": "Edit",
  "pattern": "\\.(ts|tsx|js|jsx)$",
  "check": "console.log",
  "action": "warn"
}
```

### Stop Hooks

**Final audit:**
```json
{
  "type": "stop",
  "checks": [
    "no_console_log",
    "no_hardcoded_secrets",
    "tests_passing"
  ]
}
```

## Auto-Accept Permissions

Use with caution:
- ✅ Enable for trusted, well-defined plans
- ❌ Disable for exploratory work
- ❌ Never use `dangerously-skip-permissions` flag
- ✅ Configure `allowedTools` in `~/.claude.json` instead

## TodoWrite Best Practices

Use TodoWrite tool to:
- Track progress on multi-step tasks
- Verify understanding of instructions
- Enable real-time steering
- Show granular implementation steps

Todo list reveals:
- Out of order steps
- Missing items
- Extra unnecessary items
- Wrong granularity
- Misinterpreted requirements

## Current Project Hooks

Located in `.claude/settings.local.json`:
- Memory persistence hooks
- Strategic compaction suggestions
- Auto-format on save
