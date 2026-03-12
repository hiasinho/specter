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
