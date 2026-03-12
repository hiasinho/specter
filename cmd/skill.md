## Specter

This project uses [Specter](https://hiasinho.github.io/specter/) to sync documents.

- **Project:** {{ .Project }}
- **Config:** .specter (YAML)
- **Auth:** Set SPECTER_TOKEN environment variable

### Workflow

1. Run `specter pull` before starting work to get the latest documents.
2. Edit documents in the configured paths.
3. Run `specter push` after making changes to sync them.
4. Use `specter status` to check for local/remote differences.
5. Use `specter diff` to see line-level changes.
