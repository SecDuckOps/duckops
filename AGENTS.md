# Agent Guidance for DuckOps Repository

## Quick Commands
- **Run interactive UI**: `duckops`
- **Run non‑interactive task**: `duckops run "<prompt>"`
- **Build binary**: `go build ./...`
- **Run tests**: `go test ./...`
- **Docker sandbox**:
  ```bash
  docker build -f containers/Dockerfile -t duckops:latest .
  docker run --rm -it -p 48081:48081 \
    -e TOOL_SERVER_TOKEN=duckops-local-token \
    -v "$(pwd):/workspace/$(basename "$PWD")" \
    -w "/workspace/$(basename "$PWD")" duckops:latest
  ```
- **MCP scan** (inside sandbox): `curl -s http://192.168.1.14:48081/scan -H 'Authorization: Bearer duckops-local-token' -H 'Content-Type: application/json' -d '{"agent_id":"manual","path":"/workspace/$(basename "$PWD")"}'`

## Repository Layout
- `internal/` – core implementation (agents, skills, server, DB, UI, LSP, etc.)
- `cmd/` – Cobra CLI commands (`run`, `login`, `logout`, `projects`, `stats`, etc.)
- `ui/` – Bubble‑Tea TUI, `model/` contains the top‑level UI model, `chat/` renders tool messages, `dialog/` for overlays.
- `docs/` – design and architecture documentation.
- `containers/` – Docker image definition for the security sandbox.
- `agent/` – container‑related agent code.
- `skills/` – built‑in skill definitions (`.agents/skills/` for custom skills).
- `graphx/` – knowledge‑graph engine used for code‑base analysis.

## Architectural Highlights
- **Agent System** – `internal/agent` implements the AI‑agent loop, tool orchestration, context budgeting, and loop‑detection.
- **Tool Set** – 35+ built‑in tools (file ops, shell, fetch, LSP, diagnostics, MCP scanners) exposed via the `tool` protocol.
- **MCP (Model Context Protocol)** – client/server communication for large‑scale context handling (`internal/mcp` and `duckops_security_scan` tools).
- **Skill Framework** – declarative skill registration (`internal/skills`) with automatic loading of `.agents/skills/*`.
- **UI Rendering** – hybrid approach: Ultraviolet screen buffer for layout + string‑based components; all state changes go through the single `model/UI` `Update` method.
- **Hooks** – pre/post‑execution customization (`internal/hooks`).

## Naming & Style Conventions
- Go packages use lower‑case, no underscores.
- Files are `snake_case.go`; test files mirror with `_test.go`.
- Struct and function names follow Go `PascalCase`.
- UI component files live under `internal/ui/model/` (models) and `internal/ui/chat/` (renderers).
- Skills live under `internal/skills/` or `.agents/skills/`.
- Environment variables are upper‑case with underscores (e.g., `OPENAI_API_KEY`).

## Testing Approach
- Unit tests: `go test ./...` (covers most packages). Many integration tests require a running DuckOps server; they are guarded by build tags.
- UI tests use the Bubble‑Tea `Test` utilities; focus on state transitions, not rendering.
- Security tool tests are located under `internal/platform` and `internal/skills`.

## Gotchas & Non‑Obvious Details
- **Loop Detection** – the agent will abort tool calls that could cause infinite recursion; see `internal/agent/loop_detection.go`.
- **Tool Permissions** – `--yolo` flag auto‑accepts all permission prompts; avoid in production.
- **MCP Token** – sandbox requires `TOOL_SERVER_TOKEN` env var; mismatch leads to auth errors.
- **UI Threading** – all side‑effects must be performed via `tea.Cmd`; never mutate model state directly in background goroutine.
- **ANSI Handling** – use `github.com/SecDuckOps/x/ansi` helpers; raw byte manipulation breaks width calculations.
- **Database** – DuckOps uses SQLite (`internal/db`); concurrency is managed via file‑level locks (`internal/lock`).
- **Configuration Load** – early logs are discarded until logger is set up (see `cmd/root.go` comment).
- **Docker Sandbox** – the sandbox runs tools as a non‑root user; volume mounts must map the workspace correctly.

## Helpful References
- UI dev guide: `internal/ui/AGENTS.md`
- Skill authoring: `.agents/skills/README.md`
- Architecture overview: `docs/11-architecture.md`
- Security design: `docs/14-security-design.md`
- MCP spec: `doc/01_executive_summary_and_architecture.md`
