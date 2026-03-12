#!/usr/bin/env node

import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, statSync } from "node:fs";
import { resolve, relative, dirname, join } from "node:path";
import { execSync } from "node:child_process";
import { createHash } from "node:crypto";

// ── Config ──────────────────────────────────────────────────────────────────

const API_URL = "https://yentronrhnmpewiyeqxd.supabase.co/functions/v1";

function getToken() {
  const token = process.env.SPECTER_TOKEN;
  if (!token) {
    console.error("Error: SPECTER_TOKEN not set. Export it or add to ~/.secrets");
    process.exit(1);
  }
  return token;
}

function getRepoRoot() {
  try {
    return execSync("git rev-parse --show-toplevel", { encoding: "utf-8" }).trim();
  } catch {
    console.error("Error: not inside a git repository");
    process.exit(1);
  }
}

function getGitBranch() {
  try {
    return execSync("git rev-parse --abbrev-ref HEAD", { encoding: "utf-8" }).trim();
  } catch {
    return "main";
  }
}

// ── Simple YAML parser (handles our .specter format) ────────────────────────

function parseSpecterConfig(text) {
  const config = { project: "", paths: [], exclude: [] };
  let currentKey = null;

  for (const line of text.split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;

    const kvMatch = trimmed.match(/^(\w+):\s*(.+)$/);
    if (kvMatch) {
      config[kvMatch[1]] = kvMatch[2];
      currentKey = null;
      continue;
    }

    const listKeyMatch = trimmed.match(/^(\w+):$/);
    if (listKeyMatch) {
      currentKey = listKeyMatch[1];
      if (!Array.isArray(config[currentKey])) config[currentKey] = [];
      continue;
    }

    const itemMatch = trimmed.match(/^-\s+(.+)$/);
    if (itemMatch && currentKey) {
      const val = itemMatch[1].replace(/^["']|["']$/g, "");
      config[currentKey].push(val);
    }
  }

  return config;
}

function loadConfig() {
  const root = getRepoRoot();
  const configPath = join(root, ".specter");

  if (!existsSync(configPath)) {
    console.error("Error: no .specter config file found in repository root");
    process.exit(1);
  }

  const config = parseSpecterConfig(readFileSync(configPath, "utf-8"));

  if (!config.project) {
    console.error("Error: .specter config missing 'project' field");
    process.exit(1);
  }

  if (!config.paths || config.paths.length === 0) {
    console.error("Error: .specter config missing 'paths' field");
    process.exit(1);
  }

  return { ...config, root };
}

// ── File discovery ──────────────────────────────────────────────────────────

function findMarkdownFiles(root, paths, exclude = []) {
  const files = [];

  for (const p of paths) {
    const dir = resolve(root, p);
    if (!existsSync(dir)) continue;
    collectMd(dir, files);
  }

  return files.filter((f) => {
    const rel = relative(root, f.abs);
    return !exclude.some((ex) => minimatch(rel, ex));
  });
}

function collectMd(dir, results) {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    try {
      const stat = statSync(full);
      if (stat.isDirectory()) {
        collectMd(full, results);
      } else if (entry.endsWith(".md")) {
        results.push({ abs: full });
      }
    } catch { /* skip */ }
  }
}

function minimatch(filepath, pattern) {
  const re = pattern
    .replace(/\./g, "\\.")
    .replace(/\*\*/g, "{{GLOBSTAR}}")
    .replace(/\*/g, "[^/]*")
    .replace(/\{\{GLOBSTAR\}\}/g, ".*");
  return new RegExp(`^${re}$`).test(filepath);
}

// ── API calls ───────────────────────────────────────────────────────────────

async function api(method, path, body = null) {
  const token = getToken();
  const opts = {
    method,
    headers: {
      "x-specter-token": token,
      "Content-Type": "application/json",
    },
  };
  if (body) opts.body = JSON.stringify(body);

  const res = await fetch(`${API_URL}${path}`, opts);
  const data = await res.json();

  if (!res.ok) {
    console.error(`API error (${res.status}): ${data.error || JSON.stringify(data)}`);
    process.exit(1);
  }

  return data;
}

// ── Commands ────────────────────────────────────────────────────────────────

async function push() {
  const config = loadConfig();
  const branch = getGitBranch();
  const files = findMarkdownFiles(config.root, config.paths, config.exclude);

  if (files.length === 0) {
    console.log("No markdown files found to push.");
    return;
  }

  const documents = files.map((f) => ({
    path: relative(config.root, f.abs),
    content_md: readFileSync(f.abs, "utf-8"),
  }));

  console.log(`Pushing ${documents.length} file(s) to ${config.project}/${branch}...`);

  const result = await api("POST", `/sync/${config.project}`, {
    branch,
    documents,
  });

  if (result.created?.length) console.log(`  created: ${result.created.join(", ")}`);
  if (result.updated?.length) console.log(`  updated: ${result.updated.join(", ")}`);
  if (result.unchanged?.length) console.log(`  unchanged: ${result.unchanged.length} file(s)`);

  console.log("Done.");
}

async function pull() {
  const config = loadConfig();
  const branch = getGitBranch();

  const syncFile = join(config.root, ".specter-sync");
  const since = existsSync(syncFile) ? readFileSync(syncFile, "utf-8").trim() : null;

  const params = new URLSearchParams({ branch });
  if (since) params.set("since", since);

  console.log(`Pulling from ${config.project}/${branch}${since ? ` (since ${since})` : ""}...`);

  const result = await api("GET", `/sync/${config.project}?${params}`);

  if (!result.documents || result.documents.length === 0) {
    console.log("No changes.");
  } else {
    for (const doc of result.documents) {
      const filePath = resolve(config.root, doc.path);
      const dir = dirname(filePath);
      if (!existsSync(dir)) mkdirSync(dir, { recursive: true });

      if (existsSync(filePath)) {
        const local = readFileSync(filePath, "utf-8");
        if (local === doc.content_md) {
          console.log(`  unchanged: ${doc.path}`);
          continue;
        }
      }

      writeFileSync(filePath, doc.content_md);
      console.log(`  updated: ${doc.path}`);
    }
  }

  if (result.synced_at) {
    writeFileSync(syncFile, result.synced_at);
  }

  console.log("Done.");
}

async function status() {
  const config = loadConfig();
  const branch = getGitBranch();

  const localFiles = findMarkdownFiles(config.root, config.paths, config.exclude);

  const params = new URLSearchParams({ branch });
  const result = await api("GET", `/sync/${config.project}?${params}`);

  const remoteDocs = new Map();
  for (const doc of result.documents || []) {
    remoteDocs.set(doc.path, doc);
  }

  const localOnly = [];
  const modified = [];
  const synced = [];

  for (const f of localFiles) {
    const rel = relative(config.root, f.abs);
    const content = readFileSync(f.abs, "utf-8");
    const hash = createHash("sha256").update(content).digest("hex");
    const remote = remoteDocs.get(rel);

    if (!remote) {
      localOnly.push(rel);
    } else if (remote.content_hash !== hash) {
      modified.push(rel);
    } else {
      synced.push(rel);
    }
    remoteDocs.delete(rel);
  }

  const remoteOnly = [...remoteDocs.keys()];

  console.log(`Project: ${config.project} | Branch: ${branch}\n`);

  if (localOnly.length) {
    console.log("Local only (not pushed):");
    localOnly.forEach((f) => console.log(`  + ${f}`));
  }
  if (remoteOnly.length) {
    console.log("Remote only (not pulled):");
    remoteOnly.forEach((f) => console.log(`  - ${f}`));
  }
  if (modified.length) {
    console.log("Modified (local differs from remote):");
    modified.forEach((f) => console.log(`  ~ ${f}`));
  }
  if (synced.length) {
    console.log(`In sync: ${synced.length} file(s)`);
  }
  if (!localOnly.length && !remoteOnly.length && !modified.length) {
    console.log("Everything is in sync.");
  }
}

async function diff() {
  const config = loadConfig();
  const branch = getGitBranch();

  const localFiles = findMarkdownFiles(config.root, config.paths, config.exclude);

  const params = new URLSearchParams({ branch });
  const result = await api("GET", `/sync/${config.project}?${params}`);

  const remoteDocs = new Map();
  for (const doc of result.documents || []) {
    remoteDocs.set(doc.path, doc);
  }

  let hasDiff = false;

  for (const f of localFiles) {
    const rel = relative(config.root, f.abs);
    const localContent = readFileSync(f.abs, "utf-8");
    const localHash = createHash("sha256").update(localContent).digest("hex");
    const remote = remoteDocs.get(rel);

    if (!remote) {
      hasDiff = true;
      console.log(`\n--- /dev/null`);
      console.log(`+++ ${rel}`);
      console.log(`@@ new file @@`);
      localContent.split("\n").forEach((line) => console.log(`+${line}`));
      continue;
    }

    if (remote.content_hash !== localHash) {
      hasDiff = true;
      console.log(`\n--- remote:${rel}`);
      console.log(`+++ local:${rel}`);
      showSimpleDiff(remote.content_md, localContent);
    }

    remoteDocs.delete(rel);
  }

  for (const [path, doc] of remoteDocs) {
    hasDiff = true;
    console.log(`\n--- remote:${path}`);
    console.log(`+++ /dev/null`);
    console.log(`@@ deleted locally @@`);
    doc.content_md.split("\n").forEach((line) => console.log(`-${line}`));
  }

  if (!hasDiff) {
    console.log("No differences.");
  }
}

function showSimpleDiff(a, b) {
  const aLines = a.split("\n");
  const bLines = b.split("\n");
  const max = Math.max(aLines.length, bLines.length);

  for (let i = 0; i < max; i++) {
    const aLine = aLines[i];
    const bLine = bLines[i];

    if (aLine === bLine) continue;

    if (aLine !== undefined && bLine !== undefined) {
      console.log(`-${aLine}`);
      console.log(`+${bLine}`);
    } else if (aLine !== undefined) {
      console.log(`-${aLine}`);
    } else {
      console.log(`+${bLine}`);
    }
  }
}

// ── Skill ───────────────────────────────────────────────────────────────────

function skill() {
  const root = getRepoRoot();
  const configPath = join(root, ".specter");
  const hasConfig = existsSync(configPath);
  const config = hasConfig ? parseSpecterConfig(readFileSync(configPath, "utf-8")) : null;

  const projectLine = config
    ? `This repo syncs to the Specter project \`${config.project}\`. The synced paths are: ${config.paths.map((p) => `\`${p}\``).join(", ")}.`
    : "This repo is not yet configured. Run `specter init` or create a `.specter` file.";

  const skillText = `# Specter

${projectLine}

Specter is a sync service for markdown documents (specs, docs, ADRs). It keeps documents in sync across repos and teams via a central API.

## When to use

- **Before starting work**: run \`specter pull\` to get the latest documents from the service.
- **After changing docs**: run \`specter push\` to sync your changes.
- **To check sync state**: run \`specter status\` to see what's changed locally or remotely.

## CLI commands

\`\`\`
specter push      # Push local markdown documents to the service
specter pull      # Pull remote changes to local files
specter status    # Show sync status (local vs remote)
specter diff      # Show line-level differences between local and remote
\`\`\`

Requires \`SPECTER_TOKEN\` in the environment. The CLI auto-detects the current git branch.

## API (for direct access)

Base URL: \`${API_URL}\`
Auth: \`x-specter-token\` header with the user's token.

### Key endpoints

- \`GET /sync/{project}?branch={branch}\` — fetch all documents (or changes since a timestamp with \`&since=\`)
- \`POST /sync/{project}\` — bulk push documents: \`{ "branch": "main", "documents": [{ "path": "...", "content_md": "..." }] }\`
- \`GET /projects/{slug}/proposals?status=pending\` — list pending proposals
- \`POST /projects/{slug}/proposals\` — create a proposal: \`{ "document_path": "...", "type": "replace|insert|delete|note", "anchor_content": "...", "body": "..." }\`
- \`PATCH /projects/{slug}/proposals/{id}\` — accept or reject: \`{ "status": "accepted|rejected" }\`

## Proposals

To suggest changes to a document without editing it directly, create a proposal:

- **replace** — replace the anchored text with new content
- **insert** — insert content after the anchor
- **delete** — remove the anchored content
- **note** — leave an observation, no change

Proposals use \`anchor_content\` (a text snippet from the document) as a stable identifier.
`;

  const flag = process.argv[3];

  if (flag === "--install") {
    const target = join(root, "AGENTS.md");
    const existing = existsSync(target) ? readFileSync(target, "utf-8") : "";

    if (existing.includes("# Specter")) {
      // Replace existing Specter section
      const replaced = existing.replace(/# Specter\n[\s\S]*?(?=\n# [^\n]|\Z)/, skillText.trim());
      writeFileSync(target, replaced);
    } else {
      // Append
      writeFileSync(target, existing ? existing.trimEnd() + "\n\n" + skillText : skillText);
    }
    console.log(`Skill installed to ${relative(root, target)}`);
  } else {
    process.stdout.write(skillText);
  }
}

// ── Main ────────────────────────────────────────────────────────────────────

const command = process.argv[2];

const commands = { push, pull, status, diff, skill };

if (!command || !commands[command]) {
  console.log(`Usage: specter <command>

Commands:
  push      Push local documents to the service
  pull      Pull remote documents to local
  status    Show sync status (local vs remote)
  diff      Show differences between local and remote
  skill     Output agent skill (add --install to write to AGENTS.md)`);
  process.exit(command ? 1 : 0);
}

commands[command]();
