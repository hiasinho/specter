# Specter API

RESTful API for syncing and collaborating on markdown documents. Designed for CLI tools and AI agents.

## Base URL

`https://yentronrhnmpewiyeqxd.supabase.co/functions/v1`

## Authentication

All endpoints (except `POST /register`) require a user token via the `x-specter-token` header.

## Project identifiers

Projects are namespaced under their owner's username: `owner/slug` (e.g. `hiasinho/specter`). All endpoints that operate on a project use `:owner/:slug` in the URL path.

## Endpoints

### Users

#### `POST /register`

Create a new user. No auth required.

**Body:**
```json
{ "email": "user@example.com", "username": "my-username", "invite_code": "a1b2c3d4e5f6a7b8" }
```

`username` is required — lowercase alphanumeric + hyphens, 2-39 chars.
`invite_code` is required — a valid, unredeemed access code. On success, 3 new access codes are auto-generated for the user.

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "my-username",
  "token": "hex-token",
  "invite_codes": ["code1", "code2", "code3"]
}
```

#### `GET /me`

Get current user info. Includes the user's access codes (for inviting others to the platform).

**Response:**
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "username": "my-username",
  "invite_codes": [
    { "code": "a1b2c3d4e5f6a7b8", "redeemed": false },
    { "code": "c3d4e5f6a7b8a1b2", "redeemed": true }
  ]
}
```

### Projects

#### `POST /projects`

Create a project. The authenticated user becomes the owner. The project is namespaced under the user's username.

**Body:**
```json
{ "name": "My Project", "slug": "my-project" }
```

**Response:**
```json
{
  "id": "uuid",
  "owner": "my-username",
  "slug": "my-project",
  "full_name": "my-username/my-project",
  "name": "My Project",
  "default_branch": "main"
}
```

#### `GET /projects`

List projects the user is a member of.

**Response:**
```json
[
  {
    "id": "uuid",
    "owner": "hiasinho",
    "slug": "specter",
    "full_name": "hiasinho/specter",
    "name": "Specter",
    "role": "owner"
  }
]
```

#### `DELETE /projects/:owner/:slug`

Delete a project and all its data (branches, documents, memberships, proposals). Owner only.

### Invites

#### `POST /invites`

Create an invite for a role. Owner only. Project is specified in the body as `owner/slug`.

**Body:**
```json
{ "project": "hiasinho/specter", "role": "editor|reviewer|reader" }
```

**Response:**
```json
{ "code": "invite-code", "role": "editor" }
```

#### `POST /invites/redeem`

Redeem an invite. Adds the authenticated user as a member with the invite's role.

**Body:**
```json
{ "code": "invite-code" }
```

**Response:**
```json
{
  "project": {
    "id": "uuid",
    "owner": "hiasinho",
    "slug": "specter",
    "full_name": "hiasinho/specter",
    "name": "Specter"
  },
  "role": "editor"
}
```

### Documents

#### `GET /documents/:owner/:slug?branch=:branch`

List documents for a project and branch.

#### `GET /documents/:owner/:slug/:path?branch=:branch`

Get a single document's content.

**Response:**
```json
{ "path": "specs/foo.md", "content_md": "...", "content_hash": "sha256", "revision": 3 }
```

#### `PUT /documents/:owner/:slug/:path?branch=:branch`

Create or update a document. Requires editor role or above.

**Body:**
```json
{ "content_md": "# Document content..." }
```

#### `DELETE /documents/:owner/:slug/:path?branch=:branch`

Delete a document. Requires editor role or above.

### Sync

#### `POST /sync/:owner/:slug`

Bulk push documents. Used by the CLI's `push` command.

**Body:**
```json
{
  "branch": "main",
  "base_revision": "ISO-8601 timestamp (optional)",
  "documents": [
    { "path": "specs/foo.md", "content_md": "..." },
    { "path": "specs/bar.md", "content_md": "..." }
  ]
}
```

`base_revision` enables conflict detection. Pass the `synced_at` value from the last pull. If any document was modified on the server after this timestamp (by another user), the entire push is rejected with `409` and a list of conflicts. Omit `base_revision` to force push (last write wins, backwards compatible).

**Response (success):**
```json
{ "created": ["specs/foo.md"], "updated": ["specs/bar.md"], "unchanged": [], "synced_at": "ISO-8601 timestamp" }
```

**Response (conflict, 409):**
```json
{
  "error": "Conflicts detected",
  "conflicts": [
    { "path": "specs/foo.md", "server_revision": 5, "server_updated_at": "...", "server_hash": "..." }
  ],
  "created": [],
  "updated": [],
  "unchanged": []
}
```

#### `GET /sync/:owner/:slug?branch=:branch&since=:revision`

Get documents changed since a revision. Used by the CLI's `pull` command. Omit `since` to get all documents.

**Response:**
```json
{
  "documents": [
    { "path": "specs/foo.md", "content_md": "...", "content_hash": "sha256", "revision": 3 }
  ],
  "synced_at": "ISO-8601 timestamp"
}
```

### Document History

#### `GET /document-history/:owner/:slug/:path/history?branch=:branch&limit=10`

List revision history for a document. Returns metadata only (no content). Revisions are ordered newest first. Max 100.

**Response:**
```json
{
  "path": "specs/foo.md",
  "revisions": [
    {
      "revision": 5,
      "content_hash": "sha256",
      "author": { "id": "uuid", "username": "alice", "email": "alice@example.com" },
      "created_at": "ISO-8601"
    }
  ]
}
```

#### `GET /document-history/:owner/:slug/:path/history/:revision`

Get a specific revision's full content.

**Response:**
```json
{
  "path": "specs/foo.md",
  "revision": 4,
  "content_md": "# Foo\n\nOld content...",
  "content_hash": "sha256",
  "author": { "id": "uuid", "username": "bob", "email": "bob@example.com" },
  "created_at": "ISO-8601"
}
```

#### `GET /document-history/:owner/:slug/:path/diff?branch=:branch&from=:rev&to=:rev`

Compare two revisions. Returns a unified diff (like `git diff`). `from` is required, `to` defaults to latest. Returns 404 if a requested revision has been pruned (only the last 20 revisions per document are retained).

**Response:**
```json
{
  "path": "specs/foo.md",
  "from_revision": 4,
  "to_revision": 5,
  "diff": "--- revision 4\n+++ revision 5\n@@ -1,3 +1,3 @@\n # Foo\n-Old content\n+New content"
}
```

### Proposals

#### `GET /proposals/:owner/:slug?document=:path&status=pending`

List proposals. Filter by document path and/or status.

#### `POST /proposals/:owner/:slug`

Create a proposal. Requires reviewer role or above.

**Body:**
```json
{
  "document_path": "specs/foo.md",
  "branch": "main",
  "type": "replace|insert|delete|note",
  "anchor_content": "text snippet to anchor to",
  "anchor_line_hint": 42,
  "body": "proposed content or rationale"
}
```

#### `PATCH /proposals/:owner/:slug/:id`

Accept or reject a proposal. Requires editor role or above.

**Body:**
```json
{ "status": "accepted|rejected" }
```

## Roles

- **Owner** — full control, can invite others, manage memberships, delete the project
- **Editor** — read, write, accept/reject proposals
- **Reviewer** — read, submit proposals
- **Reader** — read only
