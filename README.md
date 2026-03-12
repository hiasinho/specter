# Specter

A CLI for syncing markdown documents with the Specter service. Single Go binary, git-aware, branch-namespaced.

## Setup

Create a `.specter` config in your repo root:

```yaml
project: my-project-slug
paths:
  - specs/
  - docs/architecture/
exclude:
  - specs/drafts/
  - "**/_wip_*.md"
```

Set your token:

```bash
export SPECTER_TOKEN="your-hex-token"
```

## Usage

### Sync

```bash
specter pull                  # fetch latest from remote
specter push                  # push local changes
specter status                # show local/remote differences
specter diff                  # line-level diffs (all files)
specter diff specs/foo.md     # diff a specific file
```

Pull detects local conflicts — files with unpushed changes are skipped with a warning. Use `--force` to overwrite:

```bash
specter pull --force
```

### Proposals

```bash
# List pending proposals
specter proposals
specter proposals --status=accepted
specter proposals --document=specs/foo.md

# Create a proposal
specter propose specs/foo.md \
  --type replace \
  --anchor "text to anchor to" \
  --line 42 \
  --body "proposed replacement text"

# Accept or reject
specter review <proposal-id> accept
specter review <proposal-id> reject
```

Pending proposals are surfaced automatically during `pull` and `status`.

### Agent instructions

```bash
specter skill             # print to stdout
specter skill --install   # append to AGENTS.md
```

## Build

```bash
go build -o specter .
```
