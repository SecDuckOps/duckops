# DuckOps MCP + Docker Security Scan Integration Report

## Goal

Connect an AI agent to the existing DuckOps security tooling through MCP, while executing scanners inside the `duckops:latest` Docker sandbox. The target flow is:

```text
AI Agent
  -> MCP stdio server: duckops mcp-tool-server
  -> HTTP sandbox bridge: http://127.0.0.1:48081
  -> Docker container: duckops:latest
  -> Security tools: bandit, gosec, semgrep, trivy, gitleaks, trufflehog
```

This keeps the AI-facing interface MCP-native, while keeping scanner execution isolated inside Docker.

## Current Implementation

### 1. Docker Security Sandbox

The security image is built from:

```text
containers/Dockerfile
```

Run command:

```bash
docker build -f containers/Dockerfile -t duckops:latest .

docker run --rm --name duckops-security \
  -p 48081:48081 \
  -e TOOL_SERVER_TOKEN=duckops-local-token \
  -v "$(pwd):/workspace/$(basename "$PWD")" \
  -w "/workspace/$(basename "$PWD")" \
  duckops:latest
```

The container starts:

```bash
duckops tool-server --port 48081 --timeout "$TOOL_SERVER_TIMEOUT" --token "$TOOL_SERVER_TOKEN"
```

The entrypoint supports externally supplied `TOOL_SERVER_TOKEN`, so the host-side MCP adapter can authenticate predictably.

### 2. HTTP Tool Server

Implemented in:

```text
internal/cmd/toolserver.go
```

Available endpoints:

```text
GET  /health
POST /execute
POST /scan
POST /register_agent
```

Authentication:

```http
Authorization: Bearer duckops-local-token
```

`/health` is intentionally unauthenticated so orchestration can check readiness.

### 3. Direct Scan Endpoint

Endpoint:

```text
POST http://127.0.0.1:48081/scan
```

Example:

```bash
curl -s http://127.0.0.1:48081/scan \
  -H 'Authorization: Bearer duckops-local-token' \
  -H 'Content-Type: application/json' \
  -d "{\"agent_id\":\"manual\",\"path\":\"/workspace/$(basename "$PWD")\"}"
```

Request shape:

```json
{
  "agent_id": "manual",
  "path": "/workspace/duck",
  "target": "/workspace/duck",
  "tools": [
    "sast__bandit",
    "sast__gosec",
    "sast__semgrep",
    "sca__trivy_fs",
    "secrets__gitleaks",
    "secrets__trufflehog"
  ]
}
```

If `tools` is omitted, the default scan preset runs:

```text
sast__bandit
sast__gosec
sast__semgrep
sca__trivy_fs
secrets__gitleaks
secrets__trufflehog
```

Response shape:

```json
{
  "agent_id": "manual",
  "results": [
    {
      "tool": "sast__semgrep",
      "result": "..."
    },
    {
      "tool": "sca__trivy_fs",
      "error": "Tool execution error: ..."
    }
  ]
}
```

### 4. Single Tool Endpoint

Endpoint:

```text
POST http://127.0.0.1:48081/execute
```

Example:

```bash
curl -s http://127.0.0.1:48081/execute \
  -H 'Authorization: Bearer duckops-local-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "agent_id": "manual",
    "tool_name": "sast__semgrep",
    "kwargs": {
      "target": "/workspace"
    }
  }'
```

Supported tool names:

```text
sast__bandit
sast__gosec
sast__semgrep
sast__eslint
sast__retire
sca__syft_sbom
sca__trivy_fs
secrets__gitleaks
secrets__trufflehog
```

## MCP Adapter

Implemented in:

```text
internal/cmd/mcp_tool_server.go
```

Command:

```bash
duckops mcp-tool-server
```

The adapter is a stdio MCP server. It exposes AI-callable MCP tools and forwards requests to the Docker sandbox HTTP server.

MCP tools:

```text
duckops_security_scan
duckops_execute_tool
```

Environment variables:

```bash
DUCKOPS_TOOL_SERVER_URL=http://127.0.0.1:48081
DUCKOPS_TOOL_SERVER_TOKEN=duckops-local-token
```

## DuckOps Project MCP Config

Configured in:

```text
duckops.json
```

Current config:

```json
{
  "mcp": {
    "duckops-security": {
      "type": "stdio",
      "command": "duckops",
      "args": ["mcp-tool-server"],
      "env": {
        "DUCKOPS_TOOL_SERVER_URL": "http://127.0.0.1:48081",
        "DUCKOPS_TOOL_SERVER_TOKEN": "duckops-local-token",
        "DUCKOPS_CONTAINER_WORKSPACE": "/workspace/duck"
      },
      "workspace_mount": "/workspace/duck",
      "path_mapping": {
        "/run/media/h3ckt0r/apps/Work/duck": "/workspace/duck"
      },
      "timeout": 180
    }
  }
}
```

This means the AI agent should see the `duckops-security` MCP server and can call:

```text
duckops_security_scan
duckops_execute_tool
```

## AI Discussion Prompt

Use this prompt with another AI agent:

```text
We have a Go project called DuckOps. It now has a Docker-based security sandbox image called duckops:latest. The container runs an HTTP tool server on port 48081.

The AI agent should not execute scanners directly. It should call a stdio MCP server command:

duckops mcp-tool-server

That MCP server exposes:
- duckops_security_scan: runs the default security scan preset
- duckops_execute_tool: runs one scanner by tool name

The MCP adapter forwards calls to:
- POST /scan
- POST /execute

The Docker container is started with:

docker run --rm --name duckops-security \
  -p 48081:48081 \
  -e TOOL_SERVER_TOKEN=duckops-local-token \
  -v "$(pwd):/workspace/$(basename "$PWD")" \
  -w "/workspace/$(basename "$PWD")" \
  duckops:latest

The DuckOps project config has an MCP entry called duckops-security pointing to:

DUCKOPS_TOOL_SERVER_URL=http://127.0.0.1:48081
DUCKOPS_TOOL_SERVER_TOKEN=duckops-local-token

Please review the architecture and suggest improvements for:
1. secure token handling
2. scan job lifecycle and cancellation
3. report format normalization
4. mapping host paths to /workspace
5. making scanner results easier for an AI agent to consume
6. whether /scan should become async with job IDs
```

## Recommended Next Improvements

1. Replace the fixed local token in `duckops.json` with an environment variable such as `$DUCKOPS_TOOL_SERVER_TOKEN`.
2. Add async scan jobs:

```text
POST /scan/jobs
GET  /scan/jobs/{id}
DELETE /scan/jobs/{id}
```

3. Normalize scanner output into a common finding schema:

```json
{
  "tool": "semgrep",
  "severity": "high",
  "title": "SQL injection risk",
  "file": "/workspace/internal/api/user.go",
  "line": 42,
  "confidence": "medium",
  "raw": {}
}
```

4. Add a summary endpoint for AI consumption:

```text
POST /scan/summary
```

5. Add Docker healthcheck or readiness wait in local orchestration.
6. Add tests for `/scan`, MCP forwarding, and missing kwargs handling.

## Validation Done

Commands that passed:

```bash
go test ./internal/cmd
go build ./...
```

Known unrelated test issue:

```text
go test ./internal/config
```

This currently fails because existing config tests expect older built-in tool lists and because default Ollama model configuration is incomplete in those tests. That failure is not caused by the MCP Docker security scan bridge.
