# File: docs/25-appendix.md

# Chapter 25: Appendix

## 25.1 Architecture Diagrams (Full)

### Complete System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         User Interface Layer                         │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────────┐  │
│  │ Bubble Tea │  │   CLI     │  │  REST API  │  │  MCP Client  │  │
│  │    TUI     │  │  Commands │  │  (Server)  │  │  (Consumer)  │  │
│  └────────────┘  └────────────┘  └────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────────┐
│                         Application Layer                            │
│  ┌────────────────────────────────────────────────────────────┐     │
│  │                    Coordinator                              │     │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐   │     │
│  │  │ SessionAgent  │ │ SessionAgent │ │ SessionAgent     │   │     │
│  │  │ (Session 1)  │ │ (Session 2)  │ │ (Session N)      │   │     │
│  │  └──────┬───────┘ └──────┬───────┘ └──────┬───────────┘   │     │
│  └─────────┼────────────────┼────────────────┼───────────────┘     │
│            │                │                │                     │
│  ┌─────────┴────────────────┴────────────────┴───────────────┐    │
│  │                   Tool Engine                              │    │
│  │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐  │    │
│  │  │ Read │ │ Edit │ │ Bash │ │ Glob │ │ Web  │ │ MCP  │  │    │
│  │  │ File │ │ File │ │      │ │      │ │Fetch │ │ Proxy│  │    │
│  │  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘  │    │
│  └───────────────────────────────────────────────────────────┘    │
│                                                                   │
│  ┌───────────────────────────────────────────────────────────┐    │
│  │                   Provider Router                          │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────────────┐  │    │
│  │  │ OpenAI  │ │Anthropic│ │OpenRouter│ │ Ollama (Local)│  │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └───────────────┘  │    │
│  └───────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────────┐
│                         Service Layer                                │
│  ┌──────────────┐ ┌──────────────┐ ┌────────────┐ ┌────────────┐  │
│  │   Session    │ │   Message    │ │   File     │ │ Permission │  │
│  │   Service    │ │   Service    │ │  History   │ │  Service   │  │
│  └──────────────┘ └──────────────┘ └────────────┘ └────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────────┐
│                         Data Layer                                   │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │                      SQLite                               │      │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │      │
│  │  │ sessions │ │ messages │ │  files   │ │  todos   │    │      │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘    │      │
│  └──────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

## 25.2 Data Flow Diagrams

### Configuration Merge Order

```
Global (~/.duckops/duckops.json)
    ↓ (base)
Project (.duckops/duckops.json)
    ↓ (override)
Workspace (./duckops.json)
    ↓ (override)
CLI Flags (--model, --provider)
    ↓ (override)
Environment Variables ($OPENAI_API_KEY)
    ↓
Final Config
```

### Request Lifecycle

```
User Input
    ↓
TUI captures text
    ↓
App.SaveMessage(msg) → SQLite
    ↓
Coordinator.Run(sessionID, msg)
    ↓
SessionAgent.Run(params)
    ↓
Budget.Check() → OK?
    ├── No → Summarize, trim context
    └── Yes → Continue
    ↓
Build provider request (system + messages + tools)
    ↓
Provider.Send(messages, tools)
    ↓
[Loop: Stream response]
    ↓
Text token → Forward to UI
Tool call → Execute → Result → Send back to provider
Finish → Save assistant message → Update session stats
    ↓
UI displays complete response → Ready for next input
```

## 25.3 Package Dependency Map

```
cmd/duckops/
    ↓
internal/cli/
    ├── internal/config/
    ├── internal/agent/
    │   ├── internal/db/  (via internal/service/)
    │   ├── internal/tools/
    │   └── internal/permission/
    ├── internal/tui/
    ├── internal/server/
    ├── internal/client/
    └── internal/lsp/
```

## 25.4 Tool Reference

### Built-in Tools

| Tool | File | Description |
|------|------|-------------|
| `read` | `internal/tools/read.go` | Read file contents with line numbers |
| `edit` | `internal/tools/edit.go` | Edit file with exact string replacement |
| `write` | `internal/tools/write.go` | Write new file |
| `create` | `internal/tools/create.go` | Create file with initial content |
| `glob` | `internal/tools/glob.go` | Search files by glob pattern |
| `grep` | `internal/tools/grep.go` | Search file contents by regex |
| `bash` | `internal/tools/bash.go` | Execute shell commands safely |
| `list_dir` | `internal/tools/list_dir.go` | List directory contents |
| `web_search` | `internal/tools/web_search.go` | Search the web |
| `web_fetch` | `internal/tools/web_fetch.go` | Fetch and extract URL content |
| `file_tracker` | `internal/tools/file_tracker.go` | Track file read/write operations |
| `todowrite` | `internal/tools/todowrite.go` | Create/update todo lists |

### Security Tools

| Tool | Package | Description |
|------|---------|-------------|
| `nmap` | `internal/tools/nmap.go` | Network scanning |
| `nuclei` | `internal/tools/nuclei.go` | Vulnerability scanning |
| `subfinder` | `internal/tools/subfinder.go` | Subdomain enumeration |
| `httpx` | `internal/tools/httpx.go` | HTTP probing |
| `ffuf` | `internal/tools/ffuf.go` | Web fuzzing |
| `katana` | `internal/tools/katana.go` | Web crawling |
| `naabu` | `internal/tools/naabu.go` | Port scanning |
| `sqlmap` | `internal/tools/sqlmap.go` | SQL injection testing |
| `semgrep` | `internal/tools/semgrep.go` | Static analysis |

## 25.5 Configuration Reference

### Full Config Structure

```json
{
  "$schema": "https://github.com/SecDuckOps/duckops/config.schema.json",
  "models": {
    "large": { "provider": "openai", "model": "gpt-4o" },
    "small": { "provider": "openai", "model": "gpt-4o-mini" },
    "local": { "provider": "ollama", "model": "llama3" }
  },
  "providers": [{ "name": "openai", "type": "openai", "api_key": "${OPENAI_API_KEY}" }],
  "mcps": {},
  "lsps": {},
  "options": { "debug": false },
  "permissions": { "default": "ask" }
}
```

## 25.6 Environment Variables Reference

| Variable | Type | Description |
|----------|------|-------------|
| `OPENAI_API_KEY` | string | OpenAI API key |
| `ANTHROPIC_API_KEY` | string | Anthropic API key |
| `GEMINI_API_KEY` | string | Google Gemini API key |
| `OPENROUTER_API_KEY` | string | OpenRouter API key |
| `GITHUB_TOKEN` | string | GitHub personal access token |
| `DUCKOPS_DEBUG` | bool | Enable debug logging |
| `DUCKOPS_DATA_DIR` | path | Override data directory |
| `DUCKOPS_CONFIG_DIR` | path | Override config directory |
| `DUCKOPS_NO_COLOR` | bool | Disable colored output |
| `HOME` | path | User home directory (for config path) |
| `XDG_CONFIG_HOME` | path | XDG config directory (fallback) |
| `XDG_DATA_HOME` | path | XDG data directory (fallback) |

## 25.7 SQLite Schema

```sql
-- Migrations tracking
CREATE TABLE goose_db_version (
    id INTEGER PRIMARY KEY,
    version_id INTEGER NOT NULL,
    is_applied INTEGER NOT NULL,
    tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sessions
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    parent_session_id TEXT,
    title TEXT NOT NULL DEFAULT '',
    message_count INTEGER NOT NULL DEFAULT 0,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    cost REAL NOT NULL DEFAULT 0,
    summary_message_id TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Messages
CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    model TEXT,
    provider TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Files (version history)
CREATE TABLE files (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    hash TEXT NOT NULL,
    operation TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- Todos
CREATE TABLE todos (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    active_form TEXT,
    created_at INTEGER NOT NULL
);

-- Indices
CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_files_session_id ON files(session_id);
CREATE INDEX idx_files_path ON files(path);
CREATE INDEX idx_todos_session_id ON todos(session_id);
CREATE INDEX idx_sessions_updated_at ON sessions(updated_at);
```

---

*Next: Chapter 26 - Glossary*
