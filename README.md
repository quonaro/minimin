# MiniMin

A lightweight, web-based control panel for managing multiple Minecraft servers via Docker.

## Features

- **Multi-Server Management** — Deploy, start, stop, and monitor multiple Minecraft servers from a single dashboard
- **Docker Orchestration** — Each server runs in its own isolated Docker container
- **File Management** — Browse, upload, download, and edit server files directly in the browser
- **Mod Management** — Install, remove, and toggle server-side and client-side mods
- **RCON Console** — Send commands to running servers via WebSocket
- **Real-Time Logs** — Stream server logs in real-time through WebSocket
- **Player Management** — View online players, bans, ops, and whitelist
- **Client Mod Archive** — Generate and share client mod packages with public download links
- **Modrinth Integration** — Search and download mods directly from Modrinth

## Architecture

- **Backend**: Go + Docker SDK — orchestrates containers, exposes REST API and WebSocket endpoints
- **Frontend**: Nuxt 4 + Vue 3 + TailwindCSS — SPA dashboard served by Caddy
- **Reverse Proxy**: Caddy — serves static frontend and proxies `/api/*` and `/ws/*` to the backend
- **Process Manager**: supervisord — runs backend and Caddy in a single container

## Quick Start

### Requirements

- Docker & Docker Compose
- Docker socket access (for orchestrating Minecraft containers)

### Run

```bash
# 1. Clone the repository
git clone https://github.com/quonaro/minimin.git
cd minimin

# 2. Set a strong API key
cp backend/.env.example .env
# Edit .env and set ORCHESTRATOR_API_KEY

# 3. Start the stack
docker compose up -d

# 4. Open http://localhost:8080
```

### Docker Compose

| Service | Description |
|---------|-------------|
| `minimin` | Main container with Caddy (port 80) + Go backend (port 8081) |
| `volumes` | `instance` — state file (`instance.yml`), `./servers` — server data |
| `ports` | `8080:80` — web UI and API |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ORCHESTRATOR_API_KEY` | *(required)* | Secret key for authentication |
| `ORCHESTRATOR_LOG_LEVEL` | `info` | Backend log level: `debug`, `info`, `warn`, `error` |
| `MC_SERVERS_DIR` | `/app/servers` | Directory for server data |
| `MC_INSTANCE_FILE` | `/app/instance.yml` | Path to state file |

## Development

### Backend

```bash
cd backend
go run ./cmd/main.go
```

Requires Go 1.26+ and Docker.

### Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

Requires Node.js 22+ and pnpm.

## Building

```bash
docker build -t minimin .
```

Multi-stage build:
1. `frontend-builder` — installs npm deps and generates static SPA
2. `backend-builder` — compiles Go binary
3. `runtime` — Alpine Linux with Caddy + supervisor

## License

Open source. See repository for details.
