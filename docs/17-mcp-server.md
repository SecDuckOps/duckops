# File: docs/17-mcp-server.md

# Chapter 17: MCP Server

## 17.1 Overview

DuckOps implements both a **client** for consuming MCP (Model Context Protocol) servers and a **host** mode for exposing its own tools as an MCP server. This chapter covers the implementation in `internal/mcp/`, the configuration system, and the protocol flow.

### 17.1.1 What is MCP?

MCP is an open protocol by Anthropic that standardizes how AI applications connect to external tools and data sources. It uses JSON-RPC 2.0 over stdin/stdout (for subprocess servers) or HTTP SSE (for remote servers).

```
+-----------+     JSON-RPC 2.0     +-----------+
|  DuckOps  | <------------------> | MCP Server |
|  (Client)  |    stdin/stdout or   |  (Tool)    |
+-----------+     HTTP SSE         +-----------+
```

## 17.2 DuckOps MCP Client

### Client Architecture

```go
type Client struct {
    name    string
    version string
    cmd     *exec.Cmd
    stdin   io.WriteCloser
    stdout  io.ReadCloser
    mu      sync.Mutex
    msgID   int
    pending map[int]chan Response
    caps    ServerCapabilities
}
```

The client manages the full lifecycle of an MCP server subprocess:

1. **Start**: Launch the subprocess with configured command, args, and environment
2. **Initialize**: Send the `initialize` request and receive server capabilities
3. **Tool Discovery**: Call `tools/list` to register all available tools
4. **Execution**: Forward tool calls from DuckOps' agent to the MCP server
5. **Shutdown**: Gracefully terminate the subprocess on application exit

### Configuration

MCP servers are configured in `duckops.json`:

```json
{
  "mcps": {
    "my-server": {
      "command": "node",
      "args": ["/path/to/server.js"],
      "env": {
        "API_KEY": "${MY_API_KEY}"
      },
      "type": "stdio",
      "disabled": false,
      "disabled_tools": ["sensitive_tool"],
      "timeout": 30000
    }
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `command` | string | - | Executable path |
| `args` | []string | `[]` | Command-line arguments |
| `env` | map[string]string | `{}` | Environment variables (supports `${VAR}` substitution) |
| `type` | string | `"stdio"` | Transport type: `stdio` or `sse` |
| `url` | string | - | SSE endpoint URL (required for `sse` type) |
| `disabled` | bool | `false` | Skip loading this server |
| `disabled_tools` | []string | `[]` | Hide specific tools from the agent |
| `timeout` | int | `30000` | Tool execution timeout in ms |
| `headers` | map[string]string | `{}` | HTTP headers for SSE transport |

### Protocol Flow

```
Client                     MCP Server
  |                            |
  |----- initialize ---------->|  JSON-RPC Request
  |<---- initialized ----------|  Capabilities
  |                            |
  |----- tools/list ---------->|
  |<---- [Tool, ...] ---------|  Available tools
  |                            |
  |----- tools/call ---------->|  Execute with arguments
  |     (progress)            |
  |<---- result / error ------|  Tool output
  |                            |
  |----- shutdown -------------|
  |<---- exited ---------------|
```

### JSON-RPC Message Format

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "github_create_pr",
    "arguments": {
      "title": "Fix bug",
      "body": "Description"
    }
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      { "type": "text", "text": "PR #42 created" }
    ],
    "isError": false
  }
}
```

Response with progress:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "progress",
        "progress": 50,
        "total": 100,
        "message": "Processing..."
      }
    ]
  }
}
```

## 17.3 Tool Registration

When an MCP server connects, its tools are registered into DuckOps' global tool registry. Each tool is wrapped in a DuckOps-compatible interface:

```go
type Tool struct {
    Name        string
    Description string
    Schema      *jsonschema.Schema
    Execute     func(ctx context.Context, params json.RawMessage) (string, error)
}
```

The wrapping process:

1. Map MCP `inputSchema` to JSON Schema
2. Create an `Execute` closure that calls `tools/call` on the MCP server
3. Register in the agent's tool list with a `mcp_` prefix to avoid name conflicts

### Tool Name Prefixing

```
Original MCP tool:  create_pr
DuckOps tool:       mcp_github__create_pr

Original MCP tool:  read_file
DuckOps tool:       mcp_fs__read_file
```

The format is `mcp_{server_name}__{tool_name}`. This prevents name collisions between multiple MCP servers and DuckOps' built-in tools.

## 17.4 Resource Access

MCP servers can expose resources (files, data, configurations) in addition to tools:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/read",
  "params": {
    "uri": "file:///config/app.yaml"
  }
}
```

DuckOps maps this to its file tracking system. Resources read through MCP are recorded in the file version history, enabling undo/redo and context-aware editing.

## 17.5 DuckOps as an MCP Host (Server Mode)

DuckOps can run as an MCP server itself, exposing its tools to external MCP clients:

```bash
duckops mcp
```

This starts an MCP server on stdout/stdin (stdio transport) that exposes a curated subset of DuckOps tools:

| Tool Name | Description |
|-----------|-------------|
| `duckops_read_file` | Read files with version tracking |
| `duckops_edit_file` | Edit files with undo support |
| `duckops_bash` | Execute shell commands |
| `duckops_web_search` | Search the web |
| `duckops_web_fetch` | Fetch URLs |
| `duckops_list_directory` | List directory contents |
| `duckops_search_files` | Glob/search files |
| `duckops_run_linter` | Run project linter |
| `duckops_git_status` | Check git status |
| `duckops_session_get` | Get session context |

This allows other AI tools (like Cursor, VS Code extensions, etc.) to leverage DuckOps' tool ecosystem.

### Server Capabilities

```json
{
  "capabilities": {
    "tools": {
      "listChanged": true
    },
    "resources": {
      "subscribe": true
    }
  }
}
```

## 17.6 Resource Management

### Timeout Control

Each tool call has a configurable timeout. Long-running tools (like security scans) can set a higher timeout. DuckOps enforces the timeout by context cancellation:

```go
ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Timeout)*time.Millisecond)
defer cancel()
```

### Rate Limiting

MCP tool calls are rate-limited globally to prevent resource exhaustion. The rate limiter uses a token bucket algorithm:

```go
type RateLimiter struct {
    tokens  float64
    rate    float64
    burst   int
    last    time.Time
    mu      sync.Mutex
}
```

Default: 10 calls/second per server, burst 20.

### Graceful Shutdown

On application exit, DuckOps sends a `shutdown` notification to each MCP server, waits for acknowledgment, then sends `exit`. Any hanging servers are killed via context cancellation.

## 17.7 Error Handling

| Error | Behavior |
|-------|----------|
| Server not found | Log warning, skip registration |
| Timeout | Cancel context, return timeout error to agent |
| Invalid response | Retry once, then return parse error |
| Connection lost | Mark server as disconnected, return error |
| Server crash | Restart up to 3 times with exponential backoff |

## 17.8 Security

- **Environment Variable Substitution**: `${VAR}` patterns in `env`, `args`, and `command` are resolved from the host environment
- **Tool Denylist**: `disabled_tools` allows administrators to hide dangerous tools from the AI
- **Subprocess Isolation**: MCP servers run as child processes; DuckOps does not sandbox them further (admin responsibility)
- **SSE Security**: For remote MCP servers, only HTTPS URLs are allowed (enforced in config validation)

## 17.9 Built-in MCP Server List

DuckOps ships with support for these common MCP servers:

| Server | Purpose | Install |
|--------|---------|---------|
| GitHub | PRs, issues, repos | `npx @modelcontextprotocol/github` |
| Filesystem | Secure file access | `npx @modelcontextprotocol/filesystem` |
| Docker | Container management | `npx @modelcontextprotocol/docker` |
| Postgres | Database queries | `npx @modelcontextprotocol/postgres` |
| Memory | Knowledge graph | `npx @modelcontextprotocol/memory` |

---

*Next: Chapter 18 - Browser Automation*
