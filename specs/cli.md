# Specter CLI

A Go binary for syncing markdown documents with the Specter service. Reads a `.specter` config file from the repository root and communicates with the API.

## Stack

- Go (single binary, no runtime dependencies)
- API spec synced from the `specterapp` project via Specter itself

## Config File

A `.specter` file in the repository root:

```yaml
project: my-username/my-project
paths:
  - specs/
  - docs/architecture/
exclude:
  - specs/drafts/
  - "**/_wip_*.md"
```

- **project** — the project identifier in `owner/slug` format (e.g. `hiasinho/specter`)
- **paths** — directories to sync (relative to repo root)
- **exclude** — glob patterns for files/folders to skip within synced paths

## Token

The CLI reads `SPECTER_TOKEN` from the environment. The token authenticates all API requests via the `x-specter-token` header.

## Commands

### `specter push`

Sync local documents to the service. Reads `.specter` config, detects the current git branch, and pushes all matching documents via `POST /sync/:owner/:slug`.

### `specter pull`

Sync documents from the service to local. Fetches changes via `GET /sync/:owner/:slug` and writes updated files back to their local paths. Tracks last sync timestamp in `.specter-sync`.

### `specter status`

Show what's changed on either side — documents modified locally, remotely, or both. Compares local content hashes against remote.

### `specter diff`

Show line-level differences between local and remote versions of documents.

### `specter skill`

Output an agent instruction block describing the Specter service and CLI. With `--install`, appends it to `AGENTS.md` in the repo root.

## Open Questions

- **Conflict resolution on pull** — When both local and remote changed, what's the default behavior? Show diff and prompt? Last-write-wins with warning?
- **Proposal notifications** — How does an editor know there are pending proposals? Polling, webhooks, or surfaced during `specter pull`?
