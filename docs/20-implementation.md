# File: docs/20-implementation.md

# Chapter 20: Implementation Plan

## 20.1 Overview

This chapter outlines the implementation roadmap for DuckOps, organized into milestone-based phases. Each phase builds on the previous, delivering a working increment of the system.

## 20.2 Development Phases

### Phase 1: Core Skeleton (Weeks 1-2)

| Task | Files | Dependencies |
|------|-------|-------------|
| Project scaffolding | `go.mod`, `main.go`, `Makefile` | None |
| Config system | `internal/config/` | None |
| SQLite database | `internal/db/`, `internal/sqlc/` | Config |
| CLI framework | `internal/cli/`, `cmd/duckops/` | Config |
| Logging setup | `internal/log/` | None |

**Deliverable**: `duckops` binary that loads config, initializes SQLite, and prints help.

```bash
duckops --help
# DuckOps v0.1.0 — AI-powered development toolkit
# Usage: duckops <command> [flags]
```

### Phase 2: AI Integration (Weeks 3-4)

| Task | Files | Dependencies |
|------|-------|-------------|
| Catwalk provider registry | `internal/catwalk/` | Config |
| Fantasy SDK integration | `internal/fantasy/` | Catwalk |
| Chat coordinator | `internal/coordinator/` | Fantasy |
| Session agent | `internal/agent/session.go` | Coordinator |
| Message service | `internal/service/message.go` | DB |

**Deliverable**: `duckops` can send messages to AI providers and display responses.

```bash
duckops "What is the capital of France?"
# Paris
```

### Phase 3: Tool System (Weeks 5-6)

| Task | Files | Dependencies |
|------|-------|-------------|
| Read file tool | `internal/tools/read.go` | Agent |
| Edit file tool | `internal/tools/edit.go` | Agent |
| Bash tool | `internal/tools/bash.go` | Agent |
| Search tools | `internal/tools/glob.go`, `internal/tools/grep.go` | Agent |
| Web tools | `internal/tools/webfetch.go`, `internal/tools/websearch.go` | Agent |
| Permission service | `internal/permission/` | Agent |

**Deliverable**: AI agent can read, edit, search files and execute commands with permission control.

### Phase 4: Terminal UI (Weeks 7-8)

| Task | Files | Dependencies |
|------|-------|-------------|
| Bubble Tea app | `internal/tui/` | Core, Agent |
| Chat view | `internal/tui/chat.go` | TUI framework |
| Session list | `internal/tui/sessions.go` | TUI framework |
| Key bindings | `internal/tui/keys.go` | TUI framework |
| Markdown rendering | `internal/tui/markdown.go` | Glamour |
| Input auto-suggest | `internal/tui/completion.go` | TUI framework |

**Deliverable**: Full TUI with split-pane chat, session management, and keyboard shortcuts.

### Phase 5: Client/Server (Weeks 9-10)

| Task | Files | Dependencies |
|------|-------|-------------|
| Unix socket IPC | `internal/ipc/` | Core |
| HTTP server | `internal/server/` | Core |
| SSE streaming | `internal/server/sse.go` | Server |
| Client mode | `internal/client/` | IPC |
| Daemon mode | `internal/daemon/` | Server |

**Deliverable**: `duckops daemon` runs in background; `duckops` connects to it as a thin client.

### Phase 6: MCP & LSP (Weeks 11-12)

| Task | Files | Dependencies |
|------|-------|-------------|
| MCP client | `internal/mcp/client.go` | Core |
| MCP server mode | `internal/mcp/server.go` | Core |
| LSP manager | `internal/lsp/` | Core |
| Code diagnostics | `internal/lsp/diagnostics.go` | LSP |

**Deliverable**: MCP server integration for external tools, LSP integration for code intelligence.

### Phase 7: Security & Hardening (Weeks 13-14)

| Task | Files | Dependencies |
|------|-------|-------------|
| Permission system | `internal/permission/` | Tools |
| Audit logging | `internal/audit/` | DB |
| API key management | `internal/config/keys.go` | Config |
| Context isolation | `internal/agent/isolation.go` | Agent |

**Deliverable**: Permission enforcement, audit trail, and secure key management.

### Phase 8: Advanced Features (Weeks 15-16)

| Task | Files | Dependencies |
|------|-------|-------------|
| Context compression | `internal/agent/budget.go` | Agent |
| Loop detection | `internal/agent/loop.go` | Agent |
| Tool repair | `internal/agent/repair.go` | Agent |
| Browser automation | MCP playwright | MCP |
| Docker deployment | `Dockerfile`, `docker-compose.yml` | Core |

**Deliverable**: Production-ready DuckOps with all advanced features.

## 20.3 Milestone Timeline

```
Week 2:  Core skeleton           ████████░░░░░░░░░░░░  (phase 1)
Week 4:  AI Integration           ████████████░░░░░░░░  (phase 2)
Week 6:  Tool System              ████████████████░░░░  (phase 3)
Week 8:  Terminal UI              ████████████████████  (phase 4)
Week 10: Client/Server            ████████████████████  (phase 5)
Week 12: MCP & LSP                ████████████████████  (phase 6)
Week 14: Security & Hardening     ████████████████████  (phase 7)
Week 16: Advanced Features        ████████████████████  (phase 8)
```

## 20.4 Resource Requirements

### Development Team

| Role | Count | Responsibilities |
|------|-------|-----------------|
| Go backend engineer | 2 | Core, agent, tools, server |
| TUI engineer | 1 | Bubble Tea, lipgloss, glamour |
| DevOps engineer | 1 | Docker, CI/CD, deployment |
| Security engineer | 1 | Audit, permission, threat model |
| QA engineer | 1 | Testing, integration tests |

### Infrastructure

| Resource | Specification | Purpose |
|----------|--------------|---------|
| Build server | 4 vCPU, 8GB RAM | CI/CD, cross-compilation |
| Test server | 2 vCPU, 4GB RAM | Integration tests |
| Container registry | Docker Hub / GHCR | Image distribution |

## 20.5 Testing Strategy

### Unit Tests

```bash
# Run all unit tests
go test ./...

# With race detection
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Integration tests (require network)
go test -tags=integration ./internal/agent/...

# E2E tests
go test -tags=e2e ./internal/e2e/...
```

### Coverage Targets

| Package | Current | Target |
|---------|---------|--------|
| `internal/config/` | 85% | 90% |
| `internal/agent/` | 60% | 80% |
| `internal/tools/` | 70% | 85% |
| `internal/mcp/` | 45% | 75% |
| `internal/lsp/` | 30% | 60% |
| `internal/tui/` | 15% | 40% |

## 20.6 CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go vet ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4
      - run: GOOS=linux GOARCH=amd64 go build -o duckops-linux-amd64 .
      - run: GOOS=darwin GOARCH=amd64 go build -o duckops-darwin-amd64 .
      - uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: duckops-*
```

## 20.7 Dependencies

| Dependency | Version | Purpose | License |
|------------|---------|---------|---------|
| Go | 1.26+ | Runtime | BSD |
| modernc.org/sqlite | latest | SQLite driver | MIT |
| sqlc | 1.27+ | Query codegen | MIT |
| bubbletea | v2 | TUI framework | MIT |
| lipgloss | latest | TUI styling | MIT |
| glamour | latest | Markdown rendering | MIT |
| catwalk | latest | Provider registry | MIT |
| fantasy | latest | LLM SDK | MIT |
| mvdan.cc/sh | v3 | Shell parsing | BSD |
| gorilla/websocket | latest | WebSocket (LSP) | BSD |
| rs/cors | latest | CORS middleware | MIT |

---

*Next: Chapter 21 - Deployment Guide*
