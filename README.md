# modl-go (Self-Hosted Single-Binary Backend)

A lightweight, RAM-efficient Go + SQLite port of the [modl.gg](https://github.com/modl-gg) moderation platform, combining the REST/Protobuf APIs, Realtime WebSockets, and the embedded `modl-gg/panel` web UI into a single self-hostable binary.

## What's Included vs. Upstream

- **100% compatible** with `modl-gg/panel` (embedded inside the binary) and the `modl-gg/minecraft` plugin (`v1` JSON, `v2` startup/sync, `v3` binary Protobuf, and `/v1/realtime/ws` WebSockets).
- **SQLite (WAL mode)** replaces MongoDB — zero external database required.
- **Local filesystem storage** replaces Backblaze B2/S3 for uploaded media, server icons, and ticket/punishment evidence.
- **SMTP with stdout fallback** — if `SMTP_HOST` is omitted, login verification codes print directly to server logs.
- **Optional Gemini AI** — set `GEMINI_API_KEY` to enable AI ticket analysis, or leave empty to disable.
- **Stripped SaaS / Replay bloat** — no billing, beta invites, SaaS admin portal, or heavy Minecraft replay engine (`replay` / `replaylite`), keeping idle RAM around **~15–35 MB**.

## Quick Start (Binary)

```bash
# Build
go build -o modl-backend ./cmd/modl

# Run (auto-provisions your server & admin account on first boot)
MODL_BOOTSTRAP_SERVER_NAME="My Minecraft Server" \
MODL_BOOTSTRAP_DOMAIN="localhost" \
MODL_BOOTSTRAP_ADMIN_EMAIL="admin@example.com" \
./modl-backend
```

Check the startup log for your generated `api_key=...` (or set `MODL_BOOTSTRAP_API_KEY` yourself) to configure the `modl-gg/minecraft` plugin, then open `http://localhost:8080` in your browser and sign in with your `MODL_BOOTSTRAP_ADMIN_EMAIL` (check console logs for the 6-digit login code if SMTP is not configured).

## Quick Start (Docker Compose)

```bash
docker compose up -d --build
docker compose logs -f modl
```

Persistent SQLite database and uploaded files are stored in the `modl_data` volume (`/app/data`).

## Environment Variables

See [`.env.example`](./.env.example) for all configuration options.

## Running Tests

```bash
go test ./...
```
