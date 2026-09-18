<div align="center">

# Cline2API

Cline API reverse proxy · multi-account rotation · triple protocol · dynamic model sync · desktop app

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](#build)

**🌏 中文: [中文 README](README.md)**

</div>

---

## Introduction

Cline2API is a reverse proxy for the Cline API featuring multi-account rotation, triple protocol support (OpenAI Chat, Anthropic Messages, and OpenAI Responses APIs), API key authentication, and a bilingual admin panel (English/Chinese, auto-detected from your browser language with a manual toggle in the sidebar). A single-file cross-platform desktop app (Windows / macOS / Linux) is included — just download and run.

**Built with**: Go (backend + proxy + desktop shell), HTML/CSS/JS (embedded single-page admin frontend).

## Features

- **Triple protocol support**: serves `/v1/chat/completions` (OpenAI Chat), `/v1/messages` (Anthropic Messages API), and `/v1/responses` (OpenAI Responses API, with full support for streaming SSE and third-party clients like Cherry Studio)
- **Multi-account rotation**: load-balances across Cline accounts (`round_robin` / `fill` / `random`)
- **Bilingual admin panel**: `/admin/` manages accounts, API keys, model lists, headers, and proxy settings; auto-detects browser language with manual toggle
- **Dual-engine dynamic model sync**:
  - **Sync from Cline**: fetches the official Cline recommended-models endpoint on startup (free / cline-pass / recommended); notifies you on model changes and supports manual on-demand sync
  - **Sync from opencode**: background scheduler synchronizes open models from opencode every 10 minutes automatically
- **Egress proxy support**:
  - **Cline official proxy**: configure an HTTP / HTTPS / SOCKS5 forward proxy with credentials for requests to `api.cline.bot` directly in the settings panel with one-click connectivity test and instant hot-reloading
  - **opencode proxy pool**: supports proxy rotation with automatic rate-limit cooldown
- **Custom models**: add/remove model IDs manually and pick a default model (falls back to the first free model automatically)
- **API key auth**: protects proxy endpoints; generate/delete multiple API keys
- **System Prompt override**: place `override.md` in your data directory to replace the system prompt for all client requests
- **Account import/export**: OAuth login, manual tokens, batch file import, and cross-device export
- **Request logs**: per-request token usage, latency, TPS, and monitoring
- **Desktop app**: single-file cross-platform desktop application (Wails v2); closing the window stops the service

## Quick Start

### Option 1: Desktop app (recommended for sharing)

Download the executable for your platform from [Releases](https://github.com/luawei1/cline2api/releases) and double-click it.

> On Windows, the SmartScreen "Windows protected your PC" warning is normal because no code-signing certificate is purchased. Click "More info → Run anyway".

| Platform | File | Notes |
|----------|------|-------|
| Windows x64 | `cline-proxy-desktop.exe` | WebView2 built into Win10/11 |
| macOS Apple Silicon | `cline-proxy-desktop-darwin-arm64` | Requires Xcode CLT |
| macOS Intel | `cline-proxy-desktop-darwin-amd64` | Requires Xcode CLT |
| Linux x64 | `cline-proxy-desktop-linux-amd64` | Requires GTK3 + WebKit2GTK |

### Option 2: Command line

```bash
go build -o cline-proxy .
./cline-proxy              # default port 3457
./cline-proxy -port 8080   # custom port
```

Then open http://127.0.0.1:3457/admin/ for the admin panel.

### Option 3: Docker Deployment (AMD64 & ARM64 Multi-Arch)

Recommended `docker-compose.yml` with persistent data directory volume:

```yaml
services:
  cline-proxy:
    image: 2casiku/cline2api:latest
    container_name: cline-proxy
    restart: unless-stopped
    ports:
      - "3457:3457"
    volumes:
      # Recommended directory mount: automatically persists all configs & accounts
      - ./data:/app/data
    environment:
      - PORT=3457
      - CLINE_PROXY_HOST=0.0.0.0
      - DATA_DIR=/app/data
```

```bash
docker compose up -d      # start container
docker compose logs -f    # view logs
docker compose down       # stop container
```

> **Notice**: Using `DATA_DIR=/app/data` with `./data:/app/data` ensures that all configuration files are auto-initialized on first startup. Avoid binding non-existent single files directly to prevent Docker from creating directories unexpectedly.

## Usage Guide

### 1. Add a Cline account

In the admin panel, go to **Accounts → Import**:

- **OAuth browser login**: starts the device-authorization flow; complete login in your system browser
- **Manual token**: paste an existing refreshToken
- **Batch import**: upload a JSON file or paste text (one token per line, or JSON array `[{refreshToken, email}]`)

### 2. Configure your client

```
Base URL: http://127.0.0.1:3457/v1
API Key:  <key generated in the admin panel>
Model:    cline-free/glm-5.2 (or any available model)
```

Supported protocol endpoints:
1. **OpenAI Chat**: `/v1/chat/completions`
2. **Anthropic Claude**: `/v1/messages`
3. **OpenAI Responses**: `/v1/responses` (ideal for Cherry Studio and modern clients)

### 3. Account export/import (device migration)

- **Export**: click "Export" on the Accounts page to download `cline-accounts-export.json`
- **Import**: upload that file via "Import from File" on another device
- The export format is fully compatible with batch import

### 4. System Prompt override

Create `override.md` in your data directory; its content automatically replaces the system prompt for all client requests.

### 5. Listen address & access settings (LAN / multi-NIC)

By default the proxy listens on `127.0.0.1` (local only). The **Access Settings** section of the admin panel lets you:

- **Choose a listen address**: `127.0.0.1` (local) / `0.0.0.0` (all interfaces) / detected local IPs; saving restarts the listener immediately
- **Admin password**: none by default; once set, `/admin/` requires a password (session cookie, 24h); save an empty field to clear it

Command line flag:
```bash
./cline-proxy -host 0.0.0.0
# Or via environment variable:
CLINE_PROXY_HOST=0.0.0.0 ./cline-proxy
```

## Available Models & Synchronization

The admin panel supports both automated and on-demand model synchronization:

### 1. Sync Models from Cline
- **Mechanism**: probes Cline's official recommended-models endpoint (`https://api.cline.bot/api/v1/ai/cline/recommended-models`) to fetch free, subscription (clinePass), and recommended models.
- **Auto-alignment**: compares differences against the local model pool, automatically registering newly introduced models and removing delisted ones, then saving to `.cline-accounts.json`.
- **Manual trigger**: click **"Sync Models from Cline"** above the model list to sync on demand.

### 2. Sync Models from opencode
- **Mechanism**: fetches available models from opencode API, identifies free tier models (`-free` suffix or whitelist), and applies 200K context limits.
- **Scheduled sync**: runs a background ticker that **automatically synchronizes every 10 minutes** to keep the list up to date.
- **Manual trigger**: click **"Sync Models from opencode"** in the admin panel to trigger an instant update.

## Forward Egress Proxy

For network environments requiring a proxy to reach external APIs:

- **Cline Official Egress Proxy**: located inside the "Proxy Config" card; enter the proxy URL for `api.cline.bot` (supports `http://user:pass@ip:port` or `socks5://user:pass@ip:port`). Includes a "Test Connection" button for live latency probing and applies immediately without restarts.
- **opencode Proxy Pool**: configure multiple proxies for round-robin rotation with cooldown upon hitting rate limits.

## Data Files & Persistence

Data files lookup order:
1. Directory specified by `DATA_DIR` / `CLINE_DATA_DIR` (recommended for Docker mounts)
2. Executable directory
3. Current working directory
4. User home directory `~/.cline2api/`

| File | Purpose |
|------|---------|
| `.cline-accounts.json` | Account pool, API keys, models configuration |
| `.cline-config.json` | System settings (rotation strategy, headers, Cline egress proxy) |
| `.cline-zen.json` | opencode free models configuration and proxy pool |
| `.cline-request-logs.json` | Request logs and metrics |
| `override.md` | System Prompt override (optional) |

## Build & CI/CD

### Local Desktop Build
```bash
# Windows
./desktop/build.sh
# macOS
xcode-select --install && ./desktop/build.sh
# Linux
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev && ./desktop/build.sh
```

### Multi-Arch Docker Build
Pushing to the `dev` branch or publishing a `v*` tag triggers GitHub Actions to automatically build and push multi-arch **AMD64 & ARM64** images to Docker Hub.

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Project Structure

```
├── main.go              CLI entry (go build .)
├── desktop_main.go      Desktop entry (go build -tags desktop)
├── proxy.go             HTTP server, API routes, protocol conversion, SSE
├── responses.go         OpenAI Responses API protocol adapter & streaming SSE
├── admin.go             Admin REST API & configuration persistence
├── admin_html.go        Admin frontend (embedded single-page app)
├── models_sync.go       Cline recommended models fetching & auto-sync
├── zen.go               opencode models adapter, sync scheduler & routing
├── zen_proxy.go         opencode dedicated proxy pool & uTLS simulation
├── http.go              Global HTTP client & Cline dynamic egress proxy
├── pool.go              Account pool, multi-location lookup & initialization
├── request_logs.go      Request logs and metrics
├── desktop/             Desktop build scripts, docs, icon generator
├── Dockerfile           Multi-arch cross-compiled Docker build
├── docker-compose.yml   Docker Compose template
└── .github/workflows/   CI pipelines (Multi-Arch Docker & Tri-Platform Release)
```

## License

[MIT License](LICENSE) © 2026 [luawei1](https://github.com/luawei1)
