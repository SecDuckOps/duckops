# DuckOps Documentation — Part 2: Installation Guide & API Documentation

---

# 4. Installation Guide

## 4.1 Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| **Go** | 1.26+ | Required for building from source |
| **Git** | 2.0+ | Required for `go install` and cloning |
| **Docker** | 20.10+ | Optional, for containerized deployment |
| **AI Provider API Key** | — | At least one: OpenAI, Anthropic, Google, etc. |

## 4.2 Installation Methods

### Binary Install (Recommended)

```bash
# Linux/macOS — via install script
curl -fsSL https://raw.githubusercontent.com/SecDuckOps/duckops/main/install.sh | bash

# Manual binary placement
chmod +x duckops
sudo mv duckops /usr/local/bin/
```

### From Source

```bash
# Using go install (latest release)
go install github.com/SecDuckOps/duckops@latest

# Clone and build from source
git clone https://github.com/SecDuckOps/duckops.git
cd duckops
go mod download
go build -ldflags "-X github.com/SecDuckOps/duckops/internal/version.Version=v3.0.0" -o duckops .
sudo mv duckops /usr/local/bin/
```

### Docker

```bash
# Build image
docker build -t duckops .

# Run interactive mode
docker run --rm -it \
  -v "$(pwd):/workspace" \
  -v "$HOME/.duckops:/root/.duckops" \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  duckops

# Run non-interactive
docker run --rm \
  -v "$(pwd):/workspace" \
  -e ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY}" \
  duckops run "find security vulnerabilities"
```

#### Dockerfile Explanation

```dockerfile
## Builder Stage — compiles a static binary
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o duckops ./main.go

## Runtime Stage — minimal Alpine with non-root user
FROM alpine:3.20 AS runtime
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/duckops /usr/local/bin/duckops
USER appuser
ENTRYPOINT ["duckops"]
CMD ["--help"]
```

> [!IMPORTANT]
> The Docker build uses `CGO_ENABLED=0` for a fully static binary. SQLite is provided via `ncruces/go-sqlite3` (WASM-based) which does **not** require CGo.

## 4.3 Environment Variables

### Required (at least one AI provider key)

| Variable | Provider | Example |
|----------|----------|---------|
| `OPENAI_API_KEY` | OpenAI | `sk-proj-...` |
| `ANTHROPIC_API_KEY` | Anthropic | `sk-ant-api03-...` |
| `GOOGLE_API_KEY` | Google Gemini | `AIza...` |
| `AWS_ACCESS_KEY_ID` + `AWS_SECRET_ACCESS_KEY` | AWS Bedrock | Standard AWS credentials |
| `OPENROUTER_API_KEY` | OpenRouter | `sk-or-...` |

### Optional

| Variable | Default | Description |
|----------|---------|-------------|
| `DUCKOPS_PROFILE` | unset | Set to `1` to enable pprof on `localhost:6060` |
| `DUCKOPS_CLIENT_SERVER` | `0` | Set to `1` to enable client/server architecture |
| `DUCKOPS_DISABLE_METRICS` | `0` | Set to `1` to disable anonymous telemetry |
| `DO_NOT_TRACK` | `0` | Standard opt-out for telemetry |
| `DUCKOPS_SHORT_TOOL_DESCRIPTIONS` | `1` | Set to `0` for full tool descriptions |
| `DUCKOPS_DISABLE_ANTHROPIC_CACHE` | `0` | Disable Anthropic prompt caching |

## 4.4 Configuration

DuckOps uses a **multi-layer JSON configuration** system with the following precedence (highest first):

1. **CLI flags** (`--debug`, `--cwd`, `--data-dir`)
2. **Environment variables** (provider API keys)
3. **Project config** (`.duckops/duckops.json` in working directory)
4. **Global config** (`~/.config/duckops/duckops.json`)
5. **Built-in defaults** (embedded provider definitions via Catwalk)

### Config File Location

```
~/.duckops/duckops.json          # Global user config
./.duckops/duckops.json          # Project-specific config (gitignored)
./duckops.json                   # Project-level LSP/tool config (committed)
```

### Minimal Configuration Example

```json
{
  "$schema": "https://secduckops.dev/duck.json",
  "models": {
    "large": {
      "model": "claude-sonnet-4-20250514",
      "provider": "anthropic"
    },
    "small": {
      "model": "claude-haiku-4-20250514",
      "provider": "anthropic"
    }
  }
}
```

### Full Configuration Example

```json
{
  "$schema": "https://secduckops.dev/duck.json",
  "models": {
    "large": {
      "model": "claude-sonnet-4-20250514",
      "provider": "anthropic",
      "max_tokens": 16384,
      "think": true
    },
    "small": {
      "model": "claude-haiku-4-20250514",
      "provider": "anthropic"
    }
  },
  "providers": {
    "anthropic": {
      "api_key": "$ANTHROPIC_API_KEY"
    },
    "openai": {
      "api_key": "$OPENAI_API_KEY"
    }
  },
  "mcp": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/workspace"]
    },
    "remote-server": {
      "type": "http",
      "url": "http://localhost:3000/mcp",
      "headers": {
        "Authorization": "Bearer $MCP_TOKEN"
      }
    }
  },
  "lsp": {
    "gopls": {
      "command": "gopls",
      "args": ["serve"],
      "filetypes": ["go", "mod"],
      "root_markers": ["go.mod"],
      "options": {
        "staticcheck": true,
        "gofumpt": true
      }
    }
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "bash",
        "command": "/usr/local/bin/validate-command.sh",
        "timeout": 10
      }
    ]
  },
  "options": {
    "data_directory": ".duckops",
    "debug": false,
    "disable_auto_summarize": false,
    "disable_metrics": true,
    "auto_lsp": true,
    "tui": {
      "compact_mode": false,
      "diff_mode": "unified"
    },
    "attribution": {
      "trailer_style": "co-authored-by",
      "generated_with": true
    }
  },
  "permissions": {
    "allowed_tools": ["view", "glob", "grep", "ls"]
  },
  "tools": {
    "ls": { "max_depth": 5, "max_items": 2000 },
    "grep": { "timeout": "10s" }
  }
}
```

## 4.5 Troubleshooting Installation

| Issue | Solution |
|-------|----------|
| `SQLITE_NOTADB` on startup | Delete `.duckops/duckops.db` — WAL desync from crash. Data is recoverable. |
| Provider "unauthorized" | Verify API key env var is exported: `echo $OPENAI_API_KEY` |
| Server socket stale | Remove `/tmp/duckops-*.sock` and restart |
| Version mismatch (client/server) | DuckOps auto-detects and restarts the server. Force: `duckops server --host unix:///tmp/duckops.sock` |
| Docker: permission denied | Ensure volume mounts are readable by UID 100 (appuser) |

---

# 5. API Documentation

DuckOps exposes a REST API over Unix socket (or TCP with `--host tcp://host:port`) when running in server mode. The API is documented via Swagger at `/v1/docs/`.

## 5.1 Base URL & Transport

```
# Unix socket (default)
curl --unix-socket /tmp/duckops-$(id -u).sock http://localhost/v1/health

# TCP (explicit)
curl http://localhost:8080/v1/health
```

## 5.2 Health & System Endpoints

### `GET /v1/health`

Check server liveness.

```bash
curl --unix-socket /tmp/duckops-1000.sock http://localhost/v1/health
# Response: 200 OK
```

### `GET /v1/version`

```json
{
  "version": "DuckOps v3.0.0",
  "commit": "abc1234",
  "go_version": "go1.26.3",
  "platform": "linux/amd64"
}
```

### `GET /v1/config`

Returns server-level configuration.

### `POST /v1/control`

Server control commands (e.g., shutdown).

## 5.3 Workspace Endpoints

### `POST /v1/workspaces` — Create Workspace

Creates a new workspace with its own config, database, and agent.

**Request Body:**
```json
{
  "path": "/home/user/project",
  "data_dir": ".duckops",
  "debug": false,
  "duck": false,
  "version": "DuckOps v3.0.0",
  "env": ["OPENAI_API_KEY=sk-..."]
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "path": "/home/user/project",
  "config": { "..." }
}
```

### `GET /v1/workspaces` — List Workspaces

### `GET /v1/workspaces/{id}` — Get Workspace

### `DELETE /v1/workspaces/{id}` — Delete Workspace

Shuts down the workspace. If it's the last workspace, the server shuts down.

## 5.4 Session Endpoints

### `GET /v1/workspaces/{id}/sessions` — List Sessions

### `POST /v1/workspaces/{id}/sessions` — Create Session

### `GET /v1/workspaces/{id}/sessions/{sid}` — Get Session

```json
{
  "id": "uuid",
  "title": "Fix authentication bug",
  "message_count": 12,
  "prompt_tokens": 45000,
  "completion_tokens": 8500,
  "cost": 0.0342,
  "todos": [
    {"content": "Update auth middleware", "status": "completed"},
    {"content": "Add tests", "status": "pending"}
  ],
  "created_at": 1715000000,
  "updated_at": 1715003600
}
```

### `PUT /v1/workspaces/{id}/sessions/{sid}` — Update Session

### `DELETE /v1/workspaces/{id}/sessions/{sid}` — Delete Session

Cascading delete: removes all messages and files associated with the session.

### `GET /v1/workspaces/{id}/sessions/{sid}/messages` — List Messages

### `GET /v1/workspaces/{id}/sessions/{sid}/history` — Get File History

## 5.5 Agent Endpoints

### `POST /v1/workspaces/{id}/agent` — Run Agent

Sends a prompt to the AI agent and begins streaming.

**Request Body:**
```json
{
  "session_id": "uuid",
  "prompt": "Find and fix the SQL injection in auth.go",
  "attachments": []
}
```

### `POST /v1/workspaces/{id}/agent/init` — Initialize Agent

### `POST /v1/workspaces/{id}/agent/update` — Update Models

### `GET /v1/workspaces/{id}/agent/sessions/{sid}` — Get Agent Session State

### `POST /v1/workspaces/{id}/agent/sessions/{sid}/cancel` — Cancel Agent

### `POST /v1/workspaces/{id}/agent/sessions/{sid}/summarize` — Force Summarize

## 5.6 Configuration Endpoints

### `POST /v1/workspaces/{id}/config/set` — Set Config Value

### `POST /v1/workspaces/{id}/config/model` — Switch Model

### `POST /v1/workspaces/{id}/config/provider-key` — Set Provider API Key

### `POST /v1/workspaces/{id}/config/import-copilot` — Import GitHub Copilot

## 5.7 MCP Endpoints

### `GET /v1/workspaces/{id}/mcp/states` — MCP Server States

### `POST /v1/workspaces/{id}/mcp/refresh-tools` — Refresh MCP Tools

### `POST /v1/workspaces/{id}/mcp/read-resource` — Read MCP Resource

## 5.8 LSP Endpoints

### `GET /v1/workspaces/{id}/lsps` — List LSP Servers

### `GET /v1/workspaces/{id}/lsps/{lsp}/diagnostics` — Get Diagnostics

### `POST /v1/workspaces/{id}/lsps/start` — Start LSP Server

### `POST /v1/workspaces/{id}/lsps/stop` — Stop All LSP Servers

## 5.9 Permission Endpoints

### `GET /v1/workspaces/{id}/permissions/skip` — Check Skip Status

### `POST /v1/workspaces/{id}/permissions/skip` — Toggle Skip

### `POST /v1/workspaces/{id}/permissions/grant` — Grant Permission

## 5.10 Events Endpoint

### `GET /v1/workspaces/{id}/events` — Server-Sent Events

Long-lived SSE connection for real-time workspace events including messages, session updates, permission requests, LSP diagnostics, and MCP state changes.

```bash
curl -N --unix-socket /tmp/duckops-1000.sock \
  http://localhost/v1/workspaces/{id}/events
```

## 5.11 Complete Endpoint Reference

| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/v1/health` | Health check |
| `GET` | `/v1/version` | Version info |
| `GET` | `/v1/config` | Server config |
| `POST` | `/v1/control` | Server control |
| `GET` | `/v1/workspaces` | List workspaces |
| `POST` | `/v1/workspaces` | Create workspace |
| `GET` | `/v1/workspaces/{id}` | Get workspace |
| `DELETE` | `/v1/workspaces/{id}` | Delete workspace |
| `GET` | `/v1/workspaces/{id}/config` | Workspace config |
| `GET` | `/v1/workspaces/{id}/events` | SSE event stream |
| `GET` | `/v1/workspaces/{id}/providers` | List providers |
| `GET` | `/v1/workspaces/{id}/sessions` | List sessions |
| `POST` | `/v1/workspaces/{id}/sessions` | Create session |
| `GET` | `/v1/workspaces/{id}/sessions/{sid}` | Get session |
| `PUT` | `/v1/workspaces/{id}/sessions/{sid}` | Update session |
| `DELETE` | `/v1/workspaces/{id}/sessions/{sid}` | Delete session |
| `GET` | `/v1/workspaces/{id}/sessions/{sid}/history` | File history |
| `GET` | `/v1/workspaces/{id}/sessions/{sid}/messages` | List messages |
| `GET` | `/v1/workspaces/{id}/sessions/{sid}/messages/user` | User messages |
| `GET` | `/v1/workspaces/{id}/messages/user` | All user messages |
| `GET` | `/v1/workspaces/{id}/sessions/{sid}/filetracker/files` | Tracked files |
| `POST` | `/v1/workspaces/{id}/filetracker/read` | Record file read |
| `GET` | `/v1/workspaces/{id}/filetracker/lastread` | Last read files |
| `GET` | `/v1/workspaces/{id}/lsps` | List LSPs |
| `GET` | `/v1/workspaces/{id}/lsps/{lsp}/diagnostics` | LSP diagnostics |
| `POST` | `/v1/workspaces/{id}/lsps/start` | Start LSP |
| `POST` | `/v1/workspaces/{id}/lsps/stop` | Stop all LSPs |
| `GET` | `/v1/workspaces/{id}/permissions/skip` | Check auto-approve |
| `POST` | `/v1/workspaces/{id}/permissions/skip` | Toggle auto-approve |
| `POST` | `/v1/workspaces/{id}/permissions/grant` | Grant permission |
| `GET` | `/v1/workspaces/{id}/agent` | Agent state |
| `POST` | `/v1/workspaces/{id}/agent` | Run agent |
| `POST` | `/v1/workspaces/{id}/agent/init` | Init agent |
| `POST` | `/v1/workspaces/{id}/agent/update` | Update models |
| `GET` | `/v1/workspaces/{id}/agent/sessions/{sid}` | Agent session |
| `POST` | `/v1/workspaces/{id}/agent/sessions/{sid}/cancel` | Cancel |
| `POST` | `/v1/workspaces/{id}/agent/sessions/{sid}/summarize` | Summarize |
| `GET` | `/v1/workspaces/{id}/agent/sessions/{sid}/prompts/queued` | Queued count |
| `GET` | `/v1/workspaces/{id}/agent/sessions/{sid}/prompts/list` | Queued list |
| `POST` | `/v1/workspaces/{id}/agent/sessions/{sid}/prompts/clear` | Clear queue |
| `GET` | `/v1/workspaces/{id}/agent/default-small-model` | Default small |
| `POST` | `/v1/workspaces/{id}/config/set` | Set config |
| `POST` | `/v1/workspaces/{id}/config/remove` | Remove config |
| `POST` | `/v1/workspaces/{id}/config/model` | Set model |
| `POST` | `/v1/workspaces/{id}/config/compact` | Compact config |
| `POST` | `/v1/workspaces/{id}/config/provider-key` | Set provider key |
| `POST` | `/v1/workspaces/{id}/config/import-copilot` | Import Copilot |
| `POST` | `/v1/workspaces/{id}/config/refresh-oauth` | Refresh OAuth |
| `GET` | `/v1/workspaces/{id}/project/needs-init` | Check init |
| `POST` | `/v1/workspaces/{id}/project/init` | Init project |
| `GET` | `/v1/workspaces/{id}/project/init-prompt` | Get init prompt |
| `POST` | `/v1/workspaces/{id}/mcp/refresh-tools` | Refresh MCP tools |
| `POST` | `/v1/workspaces/{id}/mcp/read-resource` | Read MCP resource |
| `POST` | `/v1/workspaces/{id}/mcp/get-prompt` | Get MCP prompt |
| `GET` | `/v1/workspaces/{id}/mcp/states` | MCP states |
| `POST` | `/v1/workspaces/{id}/mcp/refresh-prompts` | Refresh prompts |
| `POST` | `/v1/workspaces/{id}/mcp/refresh-resources` | Refresh resources |
| `POST` | `/v1/workspaces/{id}/mcp/docker/enable` | Enable Docker MCP |
| `POST` | `/v1/workspaces/{id}/mcp/docker/disable` | Disable Docker MCP |
| `GET` | `/v1/docs/*` | Swagger UI |
