# File: docs/21-deployment.md

# Chapter 21: Deployment Guide

## 21.1 Overview

DuckOps supports multiple deployment modes to accommodate different use cases — from a local developer workstation to a team server running in Docker.

## 21.2 Deployment Modes

### Mode 1: Local Binary (Developer Workstation)

The simplest deployment. Run `duckops` directly on the developer machine.

```bash
# Install
curl -fsSL https://github.com/SecDuckOps/duckops/internalreleases/latest/download/duckops-linux-amd64 -o /usr/local/bin/duckops
chmod +x /usr/local/bin/duckops

# Run with TUI
duckops

# Run non-interactive prompt
duckops "Explain Go interfaces"
```

**Pros**: Zero dependencies, offline-capable, full TUI experience
**Cons**: No daemon mode, single-user

### Mode 2: Client/Server (Daemon)

Run `duckops daemon` in the background; connect via `duckops`.

```bash
# Start daemon
duckops daemon --port 8080 --socket /tmp/duckops.sock

# Connect (same machine)
duckops "Explain Go interfaces"

# Connect (different machine)
duckops --server http://remote-server:8080 --api-key "sk-xxx"
```

**Pros**: Persistent sessions, multi-client, can run on server hardware
**Cons**: Network dependency for remote connections

### Mode 3: Docker (Server)

```bash
docker run -d \
  --name duckops \
  -v duckops-data:/data \
  -v ./duckops.json:/config/duckops.json:ro \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -p 8080:8080 \
  duckops:latest
```

**Pros**: Isolated, portable, easy deployment with Docker Compose
**Cons**: Slightly more resource usage than bare metal

### Mode 4: Kubernetes (Team)

For teams, deploy DuckOps as a Kubernetes StatefulSet:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: duckops
spec:
  replicas: 1
  selector:
    matchLabels:
      app: duckops
  template:
    metadata:
      labels:
        app: duckops
    spec:
      containers:
      - name: duckops
        image: duckops:latest
        args: ["server", "--port", "8080"]
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: data
          mountPath: /data
        - name: config
          mountPath: /config/duckops.json
          subPath: duckops.json
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: duckops-secrets
              key: openai-api-key
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
```

**Pros**: Centralized, persistent, team access
**Cons**: Requires Kubernetes cluster

## 21.3 Installation Methods

### From GitHub Releases

```bash
# Linux AMD64
curl -LO https://github.com/SecDuckOps/duckops/internalreleases/latest/download/duckops-linux-amd64.tar.gz
tar xzf duckops-linux-amd64.tar.gz
sudo mv duckops /usr/local/bin/

# macOS ARM64
curl -LO https://github.com/SecDuckOps/duckops/internalreleases/latest/download/duckops-darwin-arm64.tar.gz
tar xzf duckops-darwin-arm64.tar.gz
sudo mv duckops /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/SecDuckOps/duckops.git
cd duckops
make build
sudo make install
```

### Via Homebrew (macOS)

```bash
brew tap SecDuckOps/duckops
brew install duckops
```

## 21.4 Configuration Setup

### First Run

```bash
# Initialize default config
duckops init

# Set API keys
duckops config set openai.api_key "sk-..."
duckops config set anthropic.api_key "sk-ant-..."

# Set preferred model
duckops config set model.large "{provider:'openai',model:'gpt-4o'}"
```

### Configuration File Locations

| Location | Priority | Description |
|----------|----------|-------------|
| `~/.duckops/duckops.json` | Lowest | Global user config |
| `.duckops/duckops.json` | Medium | Project config |
| `./duckops.json` | Highest | Workspace config |

DuckOps merges configs in order: global → project → workspace, with later files overriding earlier ones.

## 21.5 Environment Setup

### Required Environment Variables

```bash
# AI Provider Keys
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GEMINI_API_KEY="AIza..."

# DuckOps Options
export DUCKOPS_DEBUG="true"
export DUCKOPS_DATA_DIR="$HOME/.duckops/data"

# MCP Server Keys (if using MCP servers)
export GITHUB_TOKEN="ghp_..."
```

### Shell Integration

Add to `~/.zshrc` or `~/.bashrc`:

```bash
# DuckOps alias
alias duckops='duckops'

# Auto-start daemon
if [[ -z "$DUCKOPS_DAEMON_PID" ]]; then
  duckops daemon > /dev/null 2>&1 &
  export DUCKOPS_DAEMON_PID=$!
fi
```

## 21.6 Security Hardening

### API Key Protection

```bash
# Set restrictive permissions
chmod 600 ~/.duckops/duckops.json
chmod 700 ~/.duckops/

# Use environment variables instead of file
export OPENAI_API_KEY="sk-..."
# In duckops.json: "api_key": "${OPENAI_API_KEY}"
```

### Network Security

```bash
# Bind to localhost only (server mode)
duckops server --bind 127.0.0.1

# Enable API key authentication for remote clients
duckops server --api-key "sk-xxxx"

# Use TLS
duckops server --tls-cert server.crt --tls-key server.key
```

## 21.7 Monitoring & Observability

### Health Check

```bash
duckops health
# {"status":"ok","version":"1.0.0","uptime":"2h30m","sessions":5,"memory":"45MB"}
```

### Metrics

DuckOps exposes Prometheus metrics at `/metrics` in server mode:

| Metric | Type | Description |
|--------|------|-------------|
| `duckops_sessions_total` | Counter | Total sessions created |
| `duckops_messages_total` | Counter | Total messages sent |
| `duckops_tokens_total` | Counter | Total tokens processed |
| `duckops_tools_calls_total` | Counter | Total tool executions |
| `duckops_errors_total` | Counter | Total errors |
| `duckops_request_duration_ms` | Histogram | Request latency |
| `duckops_active_sessions` | Gauge | Current active sessions |

### Logging

```bash
# Debug mode
duckops --debug

# JSON log format (for log aggregation)
duckops --log-format json

# Log to file
duckops --log-file /var/log/duckops.log
```

## 21.8 Backup & Restore

### Database Backup

```bash
# Backup
sqlite3 ~/.duckops/data/duckops.db ".backup /backup/duckops-$(date +%Y%m%d).db"

# Automated backup cron
0 3 * * * sqlite3 ~/.duckops/data/duckops.db ".backup /backup/duckops-$(date +%%Y%%m%%d).db"
```

### Restore

```bash
# Stop duckops
pkill duckops

# Restore database
cp /backup/duckops-20260101.db ~/.duckops/data/duckops.db

# Start duckops
duckops
```

## 21.9 Upgrade Procedure

```bash
# 1. Backup database
sqlite3 ~/.duckops/data/duckops.db ".backup /tmp/duckops-pre-upgrade.db"

# 2. Download new binary
curl -LO https://github.com/SecDuckOps/duckops/internalreleases/latest/download/duckops-linux-amd64.tar.gz

# 3. Replace binary
sudo mv duckops /usr/local/bin/

# 4. Run database migrations (automatic on first start)
duckops --migrate

# 5. Verify
duckops version
```

## 21.10 Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/duckops

# Remove data (optional)
rm -rf ~/.duckops

# Remove Docker image
docker rmi duckops:latest

# Remove Kubernetes resources
kubectl delete statefulset duckops
kubectl delete pvc data-duckops-0
kubectl delete secret duckops-secrets
```

---

*Next: Chapter 22 - Maintenance Guide*
