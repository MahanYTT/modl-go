# modl-go (Self-Hosted Single-Binary Backend)

A lightweight, RAM-efficient Go + SQLite self-hosted implementation of the [modl.gg](https://github.com/modl-gg) Minecraft moderation platform, combining the REST/Protobuf APIs, Realtime WebSockets, and the embedded web panel UI into a single binary.

## Credits & Attribution

All original design, protocol schemas, Minecraft plugin integration, and web panel frontend belong to the **modl.gg** team:

- **Original Organization:** [github.com/modl-gg](https://github.com/modl-gg)
- **Original Java/Spring Boot Backend:** [modl-gg/backend](https://github.com/modl-gg/backend)
- **Web Panel Frontend (embedded in this binary):** [modl-gg/panel](https://github.com/modl-gg/panel)
- **Protobuf Contracts:** [modl-gg/proto](https://github.com/modl-gg/proto)
- **Minecraft Plugin:** [modl-gg/minecraft](https://github.com/modl-gg/minecraft)

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)** — see [`LICENSE`](./LICENSE) — in compliance with the upstream `modl-gg` repositories.

---

## Why `modl-go`?

Designed for self-hosting on a single VPS or alongside your Minecraft server without MongoDB, Redis, Backblaze B2, or heavy JVM memory overhead:

- **100% API compatible** with `modl-gg/panel` (embedded inside the binary) and `modl-gg/minecraft` (`v1` JSON, `v2` startup/sync, `v3` binary Protobuf, and `/v1/realtime/ws` WebSockets).
- **SQLite (WAL mode)** replaces MongoDB — zero external database required.
- **Local filesystem storage** replaces Backblaze B2 / S3 for uploaded media, server icons, and ticket/punishment evidence.
- **SMTP with log fallback** — if `SMTP_HOST` is omitted, login OTP codes print directly to server logs.
- **Optional Gemini AI** — set `GEMINI_API_KEY` to enable AI ticket analysis, or leave empty to disable.
- **Stripped SaaS / Replay bloat** — omits multi-tenant SaaS billing, beta invites, admin portal, and the heavy replay engine (`replay` / `replaylite`).

### Measured Benchmark vs. Upstream (`backend` + MongoDB)

Measured side-by-side under an identical concurrent load (10 workers: 300 logins, 180 punishments, 300 server syncs):

| Metric | `modl-go` (Go + SQLite) | Upstream (Java 21 + MongoDB) |
| :--- | :--- | :--- |
| **Cold Boot Time** | **0.75 s** | 7.06 s (~9.4x slower) |
| **Idle RAM** | **21.3 MB** | 741.6 MB (~34.8x higher) |
| **Peak Load RAM** | **33.5 MB** | 813.0 MB (~24.3x higher) |
| **Player Login** | **210.2 req/s** (47ms avg) | 113.8 req/s (87ms avg) |
| **Punishment Create** | **152.9 req/s** (65ms avg) | 122.4 req/s (81ms avg) |
| **Server Sync** | **204.6 req/s** (48ms avg) | 201.9 req/s (48ms avg) |

---

## Quick Start (Binary)

```bash
# Build single binary
go build -ldflags="-s -w" -o modl-backend ./cmd/modl

# Run (auto-provisions your server & owner account on first boot)
MODL_BOOTSTRAP_SERVER_NAME="My Minecraft Server" \
MODL_BOOTSTRAP_DOMAIN="localhost" \
MODL_BOOTSTRAP_ADMIN_EMAIL="admin@example.com" \
./modl-backend
```

Check the startup log for your generated `api_key=...` (or set `MODL_BOOTSTRAP_API_KEY` yourself) to configure the `modl-gg/minecraft` plugin, then open `http://localhost:8080` in your browser and sign in with `MODL_BOOTSTRAP_ADMIN_EMAIL` (check console logs for the 6-digit login code if SMTP is not configured).

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
