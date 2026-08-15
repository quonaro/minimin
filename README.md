<p align="center">
  <img src="docs/logo.png" alt="MiniMin Logo" width="462">
</p>

<p align="center">
  <a href="https://github.com/quonaro/minimin/actions/workflows/docker-build.yml">
    <img src="https://github.com/quonaro/minimin/actions/workflows/docker-build.yml/badge.svg" alt="Docker Build and Push">
  </a>
  <img src="https://img.shields.io/badge/Go-1.26-blue?logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
</p>

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
- **Minimin Sync** — Pair with the [desktop client](https://github.com/quonaro/minimin-sync) so players can auto-sync modpacks
- **Modrinth Integration** — Search and download mods directly from Modrinth

## Architecture

- **Backend**: Go + Docker SDK — orchestrates containers, exposes REST API and WebSocket endpoints
- **Frontend**: Nuxt 4 + Vue 3 + TailwindCSS — SPA dashboard served by Caddy
- **Reverse Proxy**: Caddy — serves static frontend and proxies `/api/*` and `/ws/*` to the backend

## Quick Start

### Requirements

- Docker & Docker Compose
- Docker socket access (for orchestrating Minecraft containers)

### Run

```bash
mkdir minimin && cd minimin

# Download compose file and env template
curl -O https://raw.githubusercontent.com/quonaro/minimin/main/docker-compose.yml
curl -o .env https://raw.githubusercontent.com/quonaro/minimin/main/.env.example

# Edit .env — set ORCHESTRATOR_API_KEY and MC_SERVERS_HOST_DIR
nano .env

# Start the stack
docker compose up -d

# Open http://localhost:8081
```

### Docker Compose

| Service    | Description                                                                |
| ---------- | -------------------------------------------------------------------------- |
| `minimin`  | Single container with Go backend + embedded frontend (port 8081)           |
| `volumes`  | `./data` — state file (`instance.yml`), `${MC_SERVERS_HOST_DIR}` — servers |
| `ports`    | `8081:8081` — web UI and API                                               |
| `env_file` | `.env` — all configuration in one file                                     |

### Environment Variables

All variables live in `.env` (see `.env.example` for the full template):

| Variable                 | Default                  | Description                                         |
| ------------------------ | ------------------------ | --------------------------------------------------- |
| `ORCHESTRATOR_API_KEY`   | _(required)_             | Secret key for authentication                       |
| `ORCHESTRATOR_LOG_LEVEL` | `info`                   | Backend log level: `debug`, `info`, `warn`, `error` |
| `MC_SERVERS_DIR`         | `/app/servers`           | Directory for server data inside the container      |
| `MC_SERVERS_HOST_DIR`    | _(required)_             | Absolute host path that maps to `MC_SERVERS_DIR`    |
| `MC_INSTANCE_FILE`       | `/app/data/instance.yml` | Path to state file                                  |
| `MODRINTH_CUSTOM_URL`    | _(empty)_                | Optional custom Modrinth API URL                    |

## Development

Use `dev.yml` for hot-reload development.

```bash
# 1. Edit dev.yml — set your MC_SERVERS_HOST_DIR and ORCHESTRATOR_API_KEY

# 2. Start dev stack
docker compose -f dev.yml up

# 3. Open frontend at http://localhost:3000
#    Backend API is available at http://localhost:8081
```

## Building

```bash
docker build -t minimin .
```

Multi-stage build:

1. `frontend-builder` — installs npm deps and generates static SPA
2. `backend-builder` — compiles Go binary
3. `runtime` — Alpine Linux with Caddy + supervisor

## Screenshots

> Commit screenshots to `docs/screenshots/` (or use GitHub drag-n-drop upload) and replace the paths below.

### Dashboard

![Dashboard](docs/screenshots/dashboard.png)
Main dashboard — server cards with status, online players, quick actions.

### Server Console

![Server Console](docs/screenshots/server-console.png)
Server console tab — real-time logs + RCON command input.

### File Manager

![File Manager](docs/screenshots/server-files.png)
File manager — browsing server directories, editing config files.

### Mod Management

![Mod Management](docs/screenshots/server-mods.png)
Mod management — installed mods list, toggle / remove / add from Modrinth.

### Resources

![Resources](docs/screenshots/server-resources.png)
Resource packs & shaders management.

### Client Archive

![Client Archive](docs/screenshots/client-archive.png)
Client mod archive — generating a public download link for players.

### Players

![Players](docs/screenshots/player-list.png)
Online players, bans, ops, and whitelist view.

## License

Open source. See repository for details.
