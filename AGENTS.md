# AGENTS.md

- Repository: `modl-go` (self-hosted Go + SQLite port of `github.com/modl-gg`).
- Primary branch: `main` (pushed to `origin` on GitHub).
- Always credit original authors (`https://github.com/modl-gg`) and preserve `LICENSE` (AGPL-3.0).
- `internal/webui/dist/` holds the pre-built embedded `modl-gg/panel` SPA and MUST stay tracked in Git so `go build ./cmd/modl` and `Dockerfile` work standalone.
- Run tests with `go test ./...` before pushing changes.
