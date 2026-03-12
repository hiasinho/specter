# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Specter?

A Go CLI tool for syncing markdown documents with the Specter service (Supabase-hosted). It pushes/pulls document collections, detects conflicts via content hashing (SHA256), and supports branch-namespaced syncing through git integration.

## Build & Test Commands

```bash
go build -o specter .          # Build binary
go test ./...                  # Run all tests
go test ./... -v               # Verbose test output
go test ./internal/api/ -run TestPush  # Run a single test
```

There is no Makefile, linter config, or CI pipeline. The pre-commit hook runs `gitleaks git --pre-commit --staged --verbose` for secret scanning.

## Architecture

```
main.go → cmd.Execute()
    │
    cmd/              CLI commands (cobra): push, pull, status, diff, propose, proposals, review, skill, init
    │
    internal/
    ├── api/          HTTP client to Supabase edge functions; types for Document, Proposal, etc.
    ├── config/       Loads .specter YAML config; finds repo root by walking up directories
    ├── documents/    Collects markdown files from configured paths, applies exclusion patterns
    ├── git/          Detects current branch name
    └── sync/         Reads/writes .specter-sync file tracking last synced revision number
```

**Key data flow:** Commands in `cmd/` load config, collect local documents via `documents.Collect()`, call `api.*` methods against the remote, and update sync state via `sync.*`.

**Configuration:** `.specter` YAML file at repo root defines `project` in `owner/slug` format (e.g. `hiasinho/specter`), `paths` to scan, and `exclude` glob patterns. Auth token comes from `SPECTER_TOKEN` env var.

**State files:** `.specter-sync` (gitignored) stores the last synced revision number for conflict detection on push/pull.

## API

- The backend is at `https://yentronrhnmpewiyeqxd.supabase.co/functions/v1`. All endpoints are documented in `API.md`. The Go client in `internal/api/` wraps these endpoints.
- NEVER CHANGE the API.md file. It gets copied over from the API layer.

## Dependencies

Minimal: `cobra` + `pflag` for CLI, `gopkg.in/yaml.v3` for config parsing. Everything else uses the standard library.

<!-- br-agent-instructions-v1 -->

---

## Beads Workflow Integration

This project uses [beads_rust](https://github.com/Dicklesworthstone/beads_rust) (`br`/`bd`) for issue tracking. Issues are stored in `.beads/` and tracked in git.

### Essential Commands

```bash
# View ready issues (unblocked, not deferred)
br ready              # or: bd ready

# List and search
br list --status=open # All open issues
br show <id>          # Full issue details with dependencies
br search "keyword"   # Full-text search

# Create and update
br create --title="..." --description="..." --type=task --priority=2
br update <id> --status=in_progress
br close <id> --reason="Completed"
br close <id1> <id2>  # Close multiple issues at once

# Sync with git
br sync --flush-only  # Export DB to JSONL
br sync --status      # Check sync status
```

### Workflow Pattern

1. **Start**: Run `br ready` to find actionable work
2. **Claim**: Use `br update <id> --status=in_progress`
3. **Work**: Implement the task
4. **Complete**: Use `br close <id>`
5. **Sync**: Always run `br sync --flush-only` at session end

### Key Concepts

- **Dependencies**: Issues can block other issues. `br ready` shows only unblocked work.
- **Priority**: P0=critical, P1=high, P2=medium, P3=low, P4=backlog (use numbers 0-4, not words)
- **Types**: task, bug, feature, epic, chore, docs, question
- **Blocking**: `br dep add <issue> <depends-on>` to add dependencies

### Session Protocol

**Before ending any session, run this checklist:**

```bash
git status              # Check what changed
git add <files>         # Stage code changes
br sync --flush-only    # Export beads changes to JSONL
git commit -m "..."     # Commit everything
git push                # Push to remote
```

### Best Practices

- Check `br ready` at session start to find available work
- Update status as you work (in_progress → closed)
- Create new issues with `br create` when you discover tasks
- Use descriptive titles and set appropriate priority/type
- Always sync before ending session

<!-- end-br-agent-instructions -->
