# File: docs/10-system-design.md

# Chapter 10: System Design

## 10.1 High-Level System Architecture

### 10.1.1 Architecture Overview

DuckOps follows a **layered architecture** with clear separation of concerns. Each layer communicates through well-defined interfaces, enabling independent testing, modification, and extension of components.

```mermaid
graph TB
    subgraph "Presentation Layer"
        CLI["CLI (Cobra)"]
        TUI["TUI (Bubble Tea)"]
        API["HTTP API (Unix Socket)"]
    end

    subgraph "Application Layer"
        APP["App Controller"]
        COORD["Agent Coordinator"]
        BACKEND["Backend Service"]
    end

    subgraph "Domain Layer"
        SESSION["Session Service"]
        MESSAGE["Message Service"]
        HISTORY["File History Service"]
        CONFIG["Config Store"]
        PERM["Permission Service"]
        FILETRACK["File Tracker"]
    end

    subgraph "Infrastructure Layer"
        DB[("SQLite Database")]
        PROVIDER["Provider Registry"]
        SHELL["Shell Executor"]
        LSP["LSP Client"]
        MCP["MCP Client"]
        TOOLS["Tool Executor"]
    end

    subgraph "External"
        EXT_AI["AI Providers"]
        EXT_WEB["Web Services"]
        EXT_MCP["MCP Servers"]
        EXT_LSP["LSP Servers"]
        EXT_DOCKER["Docker Engine"]
    end

    CLI --> APP
    TUI --> APP
    API --> BACKEND --> APP
    APP --> COORD
    COORD --> TOOLS
    COORD --> PROVIDER
    PROVIDER --> EXT_AI
    TOOLS --> SHELL
    TOOLS --> MCP --> EXT_MCP
    TOOLS --> EXT_WEB
    TOOLS --> EXT_DOCKER
    APP --> SESSION
    APP --> MESSAGE
    APP --> HISTORY
    APP --> CONFIG
    APP --> PERM
    APP --> FILETRACK
    SESSION --> DB
    MESSAGE --> DB
    HISTORY --> DB
    LSP --> EXT_LSP
```

### 10.1.2 Architecture Principles

| Principle | Description | Implementation |
|-----------|-------------|----------------|
| **Single Responsibility** | Each package has one clearly defined purpose | 43 packages, each with focused scope |
| **Interface Segregation** | Components depend on interfaces, not implementations | Querier interface, Service interfaces |
| **Dependency Inversion** | High-level modules don't depend on low-level modules | Domain services depend on db.Querier interface |
| **Separation of Concerns** | UI, business logic, and data access are separated | ui/ -> app/ -> db/ layering |
| **Event-Driven Communication** | Components communicate via events | pubsub.Broker for cross-component events |

## 10.2 Component Architecture

### 10.2.1 Core Components

| Component | Responsibility | Key Interfaces | Dependencies |
|-----------|---------------|----------------|--------------|
| **Config Store** | Load, merge, validate, persist configuration | VariableResolver | None |
| **Session Service** | CRUD for chat sessions | session.Service | db.Querier |
| **Message Service** | CRUD for messages, content parts | message.Service | db.Querier |
| **History Service** | File version tracking | history.Service | db.Querier |
| **Agent Coordinator** | AI provider orchestration, tool execution | agent.Coordinator | All services |
| **Permission Service** | Tool execution authorization | permission.Service | None |
| **File Tracker** | Track file read/write operations | filetracker.Service | db.Querier |

### 10.2.2 Component Interaction Diagram

```mermaid
sequenceDiagram
    participant UI as TUI
    participant APP as App Controller
    participant CORD as Agent Coordinator
    participant SVC as Core Services
    participant DB as SQLite
    participant AI as AI Provider

    UI->>APP: User input
    APP->>SVC: Save message
    SVC->>DB: INSERT message
    APP->>CORD: Submit to agent
    CORD->>CORD: Build context
    CORD->>AI: LLM request
    AI-->>CORD: Stream response
    CORD->>CORD: Process tool calls
    CORD-->>APP: Response tokens
    APP-->>UI: Render output
```

## 10.3 Low-Level Design

### 10.3.1 Package Dependency Graph

```mermaid
graph TB
    MAIN["main.go"] --> CMD["cmd/"]
    CMD --> WORKSPACE["workspace/"]
    CMD --> APP["app/"]
    APP --> AGENT["agent/"]
    APP --> SESSION["session/"]
    APP --> MESSAGE["message/"]
    APP --> HISTORY["history/"]
    APP --> CONFIG["config/"]
    APP --> FILETRACK["filetracker/"]
    APP --> PERM["permission/"]
    APP --> LSP["lsp/"]
    AGENT --> TOOLS["agent/tools/"]
    AGENT --> PROMPTS["agent/prompts/"]
    AGENT --> PROV["provider/"]
    SESSION --> DB["db/"]
    MESSAGE --> DB
    HISTORY --> DB
    FILETRACK --> DB
    TOOLS --> SHELL["shell/"]
    TOOLS --> MCP["mcp/"]
    TOOLS --> FSEXT["fsext/"]
    TOOLS --> WEB["webfetch"]
    CONFIG --> CSYNC["csync/"]
    CONFIG --> HOME["home/"]
    SERVER["server/"] --> BACKEND["backend/"]
    BACKEND --> APP
    CLIENT["client/"] --> BACKEND
```

### 10.3.2 Module Design: Config Store

**Package**: `internal/config/`

```go
// ConfigStore is the single entry point for all configuration access.
type ConfigStore struct {
    mu           sync.RWMutex
    cfg          *Config
    workingDir   string
    resolver     VariableResolver
    globalPath   string
    workspacePath string
    sessionOverrides map[string]SessionOverride
}

// Config holds all user-configurable settings.
type Config struct {
    Schema       string                          `json:"$schema,omitempty"`
    Models       map[SelectedModelType]ModelRef   `json:"models,omitempty"`
    Providers    *csync.Map[string, ProviderConfig] `json:"providers,omitempty"`
    MCP          MCPs                            `json:"mcp,omitempty"`
    LSP          LSPs                            `json:"lsp,omitempty"`
    Options      *Options                        `json:"options,omitempty"`
    Permissions  *Permissions                    `json:"permissions,omitempty"`
    Tools        Tools                           `json:"tools,omitzero"`
    Hooks        map[string][]HookConfig         `json:"hooks,omitempty"`
}
```

**Design Decisions**:

1. **Multiple Config Sources**: Config is loaded from global (~/.duckops/), project (.duckops/), and workspace locations, merged with last-write-wins semantics.
2. **Environment Variable Resolution**: All string values can reference environment variables via $VAR or ${VAR} syntax.
3. **Concurrent Access**: Provider config uses csync.Map for safe concurrent read/write from multiple goroutines.
4. **JSON Schema**: Config includes jsonschema tags for automatic validation and IDE support.

### 10.3.3 Module Design: Agent Coordinator

**Package**: `internal/agent/`

```go
// Coordinator manages agent lifecycle across sessions.
type Coordinator interface {
    Run(ctx context.Context, req Request) error
    Cancel(sessionID string)
    CancelAll()
    IsSessionBusy(sessionID string) bool
    Summarize(ctx context.Context, sessionID string) error
    Model() *Model
    UpdateModels(ctx context.Context, large, small, local *ModelRef)
}

// SessionAgent handles a single conversation session.
type SessionAgent interface {
    Run(ctx context.Context, req Request) error
    SetModels(large, small, local *fantasy.LanguageModel)
    SetTools(tools []fantasy.Tool)
    SetSystemPrompt(prompt string)
    Cancel()
    Summarize(ctx context.Context) error
}
```

**Design Decisions**:

1. **Coordinator/SessionAgent Split**: Coordinator manages multiple sessions; SessionAgent handles one conversation.
2. **Context Budgeting**: Before each request, the agent checks context window usage and auto-summarizes if needed.
3. **Tool Loop Detection**: After 50 consecutive tool calls without user input, the agent breaks the loop.
4. **Tool Call Repair**: Malformed tool calls are automatically corrected based on error analysis before retry.

### 10.3.4 Module Design: Message System

**Package**: `internal/message/`

```go
// ContentPart is a discriminated union of message content types.
type ContentPart interface {
    isPart()
}

type TextContent struct { Text string }
type ReasoningContent struct { Reasoning string }
type ImageURLContent struct { ImageURL string; Detail string }
type BinaryContent struct { Data []byte; MimeType string }
type ToolCall struct { ID string; Name string; Input json.RawMessage }
type ToolResult struct { ToolCallID string; Content string; IsError bool }
type Finish struct { FinishReason string }

// MarshalParts serializes ContentPart slice to JSON with type discriminator.
func MarshalParts(parts []ContentPart) (string, error)

// UnmarshalParts deserializes JSON to ContentPart slice.
func UnmarshalParts(data string) ([]ContentPart, error)
```

**Design Decisions**:

1. **Discriminated Union**: Content parts use a `type` discriminator field for JSON serialization/deserialization.
2. **Type Safety**: The `isPart()` marker method prevents external packages from implementing ContentPart.
3. **Minimal Allocation**: Binary content uses byte slices to avoid unnecessary copying.
4. **Provider Agnostic**: Content parts abstract provider-specific formats (Anthropic's content blocks, OpenAI's content array).

## 10.4 Concurrency Design

### 10.4.1 Concurrency Model

DuckOps uses Go's goroutine-based concurrency model:

```mermaid
graph TB
    subgraph "Main Goroutine"
        MAIN["main()"]
        TUI_G["TUI Event Loop"]
    end

    subgraph "Worker Goroutines"
        AGENT_G["Agent Processor"]
        STREAM_G["Stream Handler"]
        TOOL_G["Tool Executor"]
        SUMM_G["Summarization Worker"]
    end

    subgraph "Background Goroutines"
        LSP_G["LSP Connection"]
        MCP_G["MCP Connections"]
        WATCH_G["File Watcher"]
        RELOAD_G["Config Reload"]
    end

    MAIN --> TUI_G
    TUI_G --> AGENT_G
    AGENT_G --> STREAM_G
    AGENT_G --> TOOL_G
    AGENT_G --> SUMM_G
    TUI_G --> LSP_G
    TUI_G --> MCP_G
    TUI_G --> WATCH_G
    TUI_G --> RELOAD_G
```

### 10.4.2 Synchronization Strategy

| Resource | Synchronization Mechanism | Purpose |
|----------|---------------------------|---------|
| Configuration | csync.Map (concurrent map) | Thread-safe provider/hook config |
| Session State | sync.Mutex per session | Prevent concurrent modification |
| Database | SQLite WAL mode + transactions | Concurrent reads, serialized writes |
| Tool Execution | Context cancellation + timeout | Safe termination |
| Event Bus | pubsub.Broker with channels | Publisher/subscriber pattern |
| Stream Output | io.Pipe + channel | Non-blocking token streaming |

## 10.5 Error Handling Design

### 10.5.1 Error Hierarchy

```go
// Sentinel errors (defined once, compared by identity)
var (
    ErrSessionNotFound    = errors.New("session not found")
    ErrProviderNotFound   = errors.New("provider not configured")
    ErrModelNotFound      = errors.New("model not found")
    ErrPermissionDenied   = errors.New("permission denied")
    ErrToolExecutionFailed = errors.New("tool execution failed")
    ErrLoopDetected       = errors.New("infinite loop detected")
    ErrContextCancelled   = errors.New("context cancelled")
)

// Wrapped errors (add context)
if err != nil {
    return fmt.Errorf("saving session %s: %w", id, err)
}
```

### 10.5.2 Error Recovery Strategy

| Error Type | Detection | Recovery Action |
|------------|-----------|-----------------|
| Provider API Error | HTTP status code 4xx/5xx | Retry with backoff (3 attempts) |
| Rate Limit | HTTP 429 | Wait and retry with exponential backoff |
| Network Error | Connection refused/timeout | Retry with backoff, switch provider |
| Tool Timeout | Context deadline exceeded | Kill process, return partial output |
| Invalid Tool Call | Schema validation failure | Auto-repair and retry (1 attempt) |
| Permission Error | Permission check fail | Prompt user for decision |

---

**END OF CHAPTER 10**
