# File: docs/16-docker-deployment.md

# Chapter 16: Docker Deployment

## 16.1 Overview

DuckOps uses a multi-stage Docker build to produce minimal, secure, production-ready images. The containerization strategy follows the DuckOps Dockerization SOP (Standard Operating Procedure), optimized for both client and server deployment modes.

### 16.1.1 Architecture

```
+----------------------------------------------+
|              Docker Host                       |
|                                                |
|  +--------------------------------------+     |
|  |       duckops:latest Container        |     |
|  |                                        |     |
|  |  +----------+  +------------------+   |     |
|  |  |  duckops  |  |    SQLite DB     |   |     |
|  |  |  binary   |  |  /data/duckops.db|   |     |
|  |  +----------+  +------------------+   |     |
|  |                                        |     |
|  |  +----------+  +------------------+   |     |
|  |  |  Config   |  |   Skills/Agents  |   |     |
|  |  | duckops.js|  |   (optional)     |   |     |
|  |  +----------+  +------------------+   |     |
|  +--------------------------------------+     |
+----------------------------------------------+
```

## 16.2 Multi-Stage Dockerfile

```dockerfile
# ---- Build Stage ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o duckops ./cmd/duckops

# ---- Runtime Stage ----
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata bash openssh-client git curl

RUN adduser -D -h /home/duckops duckops

WORKDIR /home/duckops

COPY --from=builder /build/duckops /usr/local/bin/duckops

RUN mkdir -p /data /config /skills
VOLUME ["/data", "/config"]

USER duckops
ENV HOME=/home/duckops

ENTRYPOINT ["duckops"]
CMD ["server"]
```

### Stage Breakdown

| Stage | Base Image | Purpose | Size Impact |
|-------|-----------|---------|-------------|
| `builder` | `golang:1.26-alpine` | Compile static binary | ~800MB (ephemeral) |
| `runtime` | `alpine:3.21` | Minimal runtime | ~20MB + binary |

### Build Commands

```bash
# Production build
docker build -t duckops:latest .

# With build args
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg COMMIT=$(git rev-parse HEAD) \
  -t duckops:1.0.0 .

# Multi-platform build
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t duckops:latest .
```

## 16.3 Docker Compose

```yaml
version: "3.9"

services:
  duckops:
    image: duckops:latest
    build:
      context: .
      dockerfile: Dockerfile
    container_name: duckops-server
    volumes:
      - duckops-data:/data
      - ./duckops.json:/config/duckops.json:ro
      - $HOME/.ssh:/home/duckops/.ssh:ro
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - DUCKOPS_DATA_DIR=/data
      - DUCKOPS_CONFIG_DIR=/config
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "duckops", "health"]
      interval: 30s
      timeout: 5s
      retries: 3

volumes:
  duckops-data:
    driver: local
```

## 16.4 Volume Mounts

| Mount Point | Purpose | Required | Security |
|------------|---------|----------|----------|
| /data | SQLite database, session data | Yes | DuckOps user writes only |
| /config/duckops.json | Configuration file | Yes | Read-only recommended |
| /home/duckops/.ssh | SSH keys for git operations | Optional | Read-only recommended |
| /skills | Custom skills directory | Optional | DuckOps user reads only |

## 16.5 Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| OPENAI_API_KEY | Conditional | OpenAI provider key |
| ANTHROPIC_API_KEY | Conditional | Anthropic provider key |
| DUCKOPS_DATA_DIR | No | Override data directory (default: /data) |
| DUCKOPS_CONFIG_DIR | No | Override config directory (default: /config) |
| DUCKOPS_DEBUG | No | Enable debug logging (true/false) |

API keys should be provided through environment variables (referenced in duckops.json):

```json
{
  "providers": [
    {
      "name": "openai",
      "api_key": "${OPENAI_API_KEY}",
      "models": [{ "name": "gpt-4o" }]
    }
  ]
}
```

## 16.6 Docker Compose with MCP Servers

```yaml
services:
  duckops:
    image: duckops:latest
    volumes:
      - duckops-data:/data
      - ./duckops.json:/config/duckops.json:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    ports:
      - "8080:8080"

  mcp-github:
    image: ghcr.io/github/mcp-server:latest
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
    networks:
      - duckops-net

  mcp-filesystem:
    image: ghcr.io/modelcontextprotocol/filesystem:latest
    volumes:
      - /workspace:/workspace:ro
    networks:
      - duckops-net

networks:
  duckops-net:
    driver: bridge
```

## 16.7 Security Considerations

### Read-Only Root Filesystem

```yaml
services:
  duckops:
    image: duckops:latest
    read_only: true
    tmpfs:
      - /tmp
    volumes:
      - duckops-data:/data
```

### Dropping Capabilities

```yaml
services:
  duckops:
    image: duckops:latest
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
```

### Non-Root User

The Dockerfile creates a duckops user (UID 1000). The binary runs as this user, not root. All data volumes should be owned by this UID:

```bash
chown -R 1000:1000 /data
```

### Resource Limits

```yaml
services:
  duckops:
    image: duckops:latest
    deploy:
      resources:
        limits:
          cpus: "2"
          memory: "2G"
        reservations:
          cpus: "0.5"
          memory: "512M"
```

## 16.8 Image Optimization

| Technique | Before | After | Reduction |
|-----------|--------|-------|-----------|
| Multi-stage build | 800MB | 25MB | 97% |
| -ldflags="-s -w" | 35MB | 18MB | 49% |
| Alpine base | 50MB | 25MB | 50% |
| .dockerignore | 30MB | 25MB | 17% |

### .dockerignore

```
.git/
docs/
testdata/
*.md
.gitignore
*.log
```

## 16.9 Docker Compose Profiles

```yaml
services:
  duckops:
    image: duckops:latest
    profiles: ["core"]

  duckops-server:
    image: duckops:latest
    command: server --port 8080
    profiles: ["server"]
    ports:
      - "8080:8080"

  duckops-dev:
    image: duckops:latest
    command: dev
    profiles: ["dev"]
    volumes:
      - $PWD:/workspace
    stdin_open: true
    tty: true
```

Usage:

```bash
docker compose up duckops
docker compose --profile server up
docker compose --profile dev run duckops-dev
```

## 16.10 Production Deployment Checklist

1. Build image with specific version tag, not latest
2. Read-only config mount duckops.json as read-only
3. Persistent volume for SQLite data (backup regularly)
4. Health check configured
5. Resource limits set (CPU, memory)
6. Non-root user verified
7. Secrets via environment or Docker secrets, not config files
8. Network isolate in a dedicated Docker network
9. Logging configure Docker logging driver
10. Backup automate SQLite backup

---

*Next: Chapter 17 - MCP Server*
