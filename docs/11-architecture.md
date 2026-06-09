# File: docs/11-architecture.md

# Chapter 11: Architecture

## 11.1 High-Level Architecture

### 11.1.1 Architecture Overview

DuckOps follows a **Hexagonal Architecture (Ports and Adapters)** pattern at its core, combined with **Layered Architecture** for the internal structure. This hybrid approach provides:

1. **Testability**: Core business logic can be tested without external dependencies
2. **Flexibility**: Adapters can be swapped (e.g., different SQLite drivers)
3. **Maintainability**: Clear boundaries between layers
4. **Extensibility**: New providers, tools, and skills can be added via ports

```mermaid
graph TB
    subgraph "Ports (Interfaces)"
        DB_PORT[("DB Querier")]
        PROV_PORT[("Provider Registry")]
        TOOL_PORT[("Tool Interface")]
        EVT_PORT[("Event Bus")]
    end

    subgraph "Core Domain"
        SESSION["Session Domain"]
        MESSAGE["Message Domain"]
        AGENT["Agent Domain"]
        CONFIG["Config Domain"]
    end

    subgraph "Adapters (Implementations)"
        DB_ADAP["SQLite Adapter"]
        PROV_ADAP["Catwalk Provider"]
        TOOL_ADAP["Tool Executor"]
        UI_ADAP["Bubble Tea TUI"]
        CLI_ADAP["Cobra CLI"]
        API_ADAP["HTTP API"]
    end

    subgraph "External"
        EXT_DB[("SQLite File")]
        EXT_AI[("AI Providers")]
        EXT_TOOLS[("Security Tools")]
        EXT_USER[("User")]
    end

    SESSION --> DB_PORT
    MESSAGE --> DB_PORT
    AGENT --> PROV_PORT
    AGENT --> TOOL_PORT
    CONFIG --> DB_PORT

    DB_PORT --> DB_ADAP --> EXT_DB
    PROV_PORT --> PROV_ADAP --> EXT_AI
    TOOL_PORT --> TOOL_ADAP --> EXT_TOOLS
    UI_ADAP --> EXT_USER
    CLI_ADAP --> EXT_USER
    API_ADAP --> EXT_USER
```

## 11.2 Package Architecture

### 11.2.1 Package Map

```mermaid
graph TB
    subgraph "Entry"
        MAIN["main.go"]
    end

    subgraph "Commands"
        CMD["cmd/"]
    end

    subgraph "App Layer"
        APP["app/"]
        BACKEND["backend/"]
        WORKSPACE["workspace/"]
    end

    subgraph "Domain Services"
        SESSION["session/"]
        MESSAGE["message/"]
        HISTORY["history/"]
        CONFIG["config/"]
        PERM["permission/"]
        FILETRACK["filetracker/"]
    end

    subgraph "Infrastructure"
        DB["db/"]
        SHELL["shell/"]
        LSP["lsp/"]
        SERVER["server/"]
        CLIENT["client/"]
        HOME["home/"]
        FSEXT["fsext/"]
        CSYNC["csync/"]
        PUBSUB["pubsub/"]
    end

    subgraph "Agent System"
        AGENT["agent/"]
        TOOLS["agent/tools/"]
        MCP["agent/tools/mcp/"]
        PROMPT["agent/prompts/"]
        HYPER["agent/hyper/"]
    end

    subgraph "UI"
        UI["ui/"]
        CHAT["ui/chat/"]
        DIALOG["ui/dialog/"]
        MODEL["ui/model/"]
        STYLES["ui/styles/"]
    end

    subgraph "Protocol"
        PROTO["proto/"]
    end

    subgraph "Security"
        SEC["security/"]
        GRAPHX["graphx/"]
    end

    MAIN --> CMD
    CMD --> APP
    CMD --> WORKSPACE --> APP
    CMD --> BACKEND --> APP
    APP --> AGENT
    APP --> SESSION
    APP --> MESSAGE
    APP --> HISTORY
    APP --> CONFIG
    APP --> PERM
    APP --> FILETRACK
    APP --> LSP
    AGENT --> TOOLS
    AGENT --> PROMPT
    AGENT --> HYPER
    TOOLS --> MCP
    TOOLS --> SHELL
    TOOLS --> FSEXT
    SESSION --> DB
    MESSAGE --> DB
    HISTORY --> DB
    CONFIG --> CSYNC
    CONFIG --> HOME
    APP --> PUBSUB
    SERVER --> BACKEND
    CLIENT --> SERVER
    BACKEND --> PROTO
    SESSION --> PROTO
    MESSAGE --> PROTO
    UI --> MODEL
    MODEL --> CHAT
    MODEL --> DIALOG
    MODEL --> STYLES
    SEC --> GRAPHX
```

## 11.3 Component Architecture

### 11.3.1 Core Components

| Layer | Component | Package | Responsibility |
|-------|-----------|---------|----------------|
| **Entry** | Main | main.go | Application startup, CLI execution |
| **Commands** | CLI | cmd/ | Cobra command definitions, flags |
| **Application** | App Controller | app/ | Service wiring, lifecycle management |
| **Application** | Backend | backend/ | Multi-workspace management (server mode) |
| **Application** | Workspace | workspace/ | Local/remote workspace abstraction |
| **Domain** | Session Service | session/ | Session CRUD, todo management |
| **Domain** | Message Service | message/ | Message CRUD, content part handling |
| **Domain** | Config Store | config/ | Configuration loading, resolving, persisting |
| **Domain** | Permission | permission/ | Tool execution authorization |
| **Domain** | File Tracker | filetracker/ | File read/write tracking |
| **Agent** | Coordinator | agent/ | Multi-session agent orchestration |
| **Agent** | Session Agent | agent/ | Single-session AI conversation |
| **Agent** | Tools | agent/tools/ | 35+ built-in tool implementations |
| **Infrastructure** | Database | db/ | SQLite queries (sqlc-generated) |
| **Infrastructure** | Shell | shell/ | Command execution and parsing |
| **Infrastructure** | LSP | lsp/ | Language Server Protocol client |
| **Infrastructure** | Server | server/ | HTTP server over Unix socket |
| **Infrastructure** | Client | client/ | HTTP client for remote server |
| **UI** | Main Model | ui/model/ | Bubble Tea application model |
| **UI** | Chat | ui/chat/ | Message rendering components |
| **UI** | Dialog | ui/dialog/ | Overlay dialog components |
| **UI** | Styles | ui/styles/ | Style definitions and themes |
| **Protocol** | Proto | proto/ | Wire-format type definitions |

### 11.3.2 Component Responsibilities

**App Controller** (`internal/app/app.go`):
- Initializes all services (session, message, history, permission, file tracker)
- Creates the agent coordinator with coder agent
- Starts MCP connections and LSP manager
- Manages graceful shutdown (cancel agents, kill shells, close LSP)
- Bridges application events to the Bubble Tea TUI

**Agent Coordinator** (`internal/agent/coordinator.go`):
- Manages multiple session agents
- Discovers and registers skills
- Handles model configuration updates
- Provides session queuing and cancellation

**Session Agent** (`internal/agent/agent.go`):
- Manages a single conversation session
- Handles context budgeting and auto-summarization
- Executes the agent loop (message -> LLM -> tool -> LLM -> ...)
- Detects and breaks infinite loops
- Auto-repairs malformed tool calls

**Config Store** (`internal/config/store.go`):
- Single entry point for all configuration access
- Loads config from multiple sources with merging
- Provides session-level configuration overrides
- Persists runtime configuration changes
- Handles provider API key management and OAuth

## 11.4 Layered Architecture Details

### 11.4.1 Layer Separation

```mermaid
graph TB
    subgraph "Layer 1: Presentation"
        TUI["Bubble Tea TUI"]
        CLI["Cobra CLI"]
        API["HTTP REST API"]
    end

    subgraph "Layer 2: Application"
        APP["App Controller"]
        BACKEND["Backend"]
        COORD["Agent Coordinator"]
    end

    subgraph "Layer 3: Domain"
        SVC_DOMAIN["Session / Message / History"]
        CFG_DOMAIN["Configuration"]
        PERM_DOMAIN["Permissions"]
    end

    subgraph "Layer 4: Infrastructure"
        DB["Database (SQLite)"]
        SHELL["Shell Executor"]
        PROV["Provider Registry"]
        TOOL["Tool Engine"]
        MCP_CLIENT["MCP Client"]
        LSP_CLIENT["LSP Client"]
    end

    subgraph "Layer 5: External"
        EXT_AI["AI Provider API"]
        EXT_WEB["Web Services"]
        EXT_MCP["MCP Servers"]
        EXT_DOCKER["Docker Engine"]
    end

    Layer1 --> Layer2
    Layer2 --> Layer3
    Layer3 --> Layer4
    Layer4 --> Layer5
```

### 11.4.2 Layer Rules

| Rule | Description | Enforcement |
|------|-------------|-------------|
| **L1** | Presentation depends only on Application | Import validation |
| **L2** | Application depends only on Domain | Interface injection |
| **L3** | Domain depends only on Infrastructure (interfaces) | Dependency inversion |
| **L4** | Infrastructure depends on external libraries | Direct imports |
| **L5** | No layer skips a level | Go import cycle detection |

## 11.5 Event Architecture

### 11.5.1 Event Bus

DuckOps uses an internal pub/sub event bus for cross-component communication:

```mermaid
graph TB
    subgraph "Event Producers"
        AGENT_EV["Agent Coordinator"]
        SESS_EV["Session Service"]
        LSP_EV["LSP Manager"]
        MCP_EV["MCP Manager"]
        TOOL_EV["Tool Executor"]
    end

    subgraph "Event Bus (pubsub.Broker)"
        BUS["Event Bus"]
    end

    subgraph "Event Consumers"
        UI_EV["TUI (Message Display)"]
        LOG_EV["Logger"]
        METRIC_EV["Metrics Collector"]
    end

    AGENT_EV -->|MessageEvent| BUS
    SESS_EV -->|SessionEvent| BUS
    LSP_EV -->|DiagnosticEvent| BUS
    MCP_EV -->|MCPStateEvent| BUS
    TOOL_EV -->|ToolResultEvent| BUS
    BUS -->|Route| UI_EV
    BUS -->|Route| LOG_EV
    BUS -->|Route| METRIC_EV
```

### 11.5.2 Event Types

| Event Type | Producer | Consumers | Payload |
|------------|----------|-----------|---------|
| MessageEvent | Agent | TUI, Logger | Message struct with content parts |
| SessionEvent | Session | TUI | Session status change |
| ToolResultEvent | Tool Executor | TUI | Tool name, status, output |
| LSPDiagnosticEvent | LSP Manager | TUI | File path, diagnostics list |
| MCPStateEvent | MCP Manager | TUI, Logger | Server ID, connection state |
| ErrorEvent | Any | TUI, Logger | Error message, stack trace |

## 11.6 Data Flow Architecture

### 11.6.1 Request Processing Pipeline

```mermaid
sequenceDiagram
    participant U as User
    participant T as TUI
    participant A as App
    participant C as Coordinator
    participant SA as SessionAgent
    participant BB as ContextBudget
    participant P as Provider
    participant T2 as ToolEngine

    U->>T: Types message
    T->>A: SubmitMessage(text, sessionID)
    A->>C: Run(Request{sessionID, text})
    C->>SA: Start processing
    SA->>BB: Check budget
    
    alt Over budget
        SA->>P: Summarize(context)
        P-->>SA: Summary
        SA->>SA: Trim context
    end

    SA->>P: Chat(messages, tools)
    activate P
    P-->>SA: Stream(chunk)
    deactivate P
    
    loop Each chunk
        SA-->>A: Forward chunk
        A-->>T: Display chunk
    end

    alt Tool call requested
        SA->>T2: Execute(call)
        activate T2
        T2-->>SA: Result
        deactivate T2
        SA->>P: Submit result
        P-->>SA: Continue response
    end

    SA->>A: Complete
    A->>T: Render final
```

## 11.7 Security Architecture

### 11.7.1 Permission System Architecture

```mermaid
graph TB
    subgraph "Permission Flow"
        REQ["Tool Call Request"] --> CHECK{Has Permission?}
        CHECK -->|Never Asked| ASK[Permission Dialog]
        CHECK -->|Always Allow| EXEC[Execute Tool]
        CHECK -->|Always Deny| DENY[Return Denied]
        CHECK -->|Allow Session| EXEC
        ASK -->|Allow Once| EXEC
        ASK -->|Allow Session| EXEC
        ASK -->|Deny| DENY
    end

    subgraph "Permission Storage"
        DB[("Session Memory")]
        CONFIG[("Config File")]
    end

    ASK --> DB
    CHECK --> DB
    CHECK --> CONFIG
```

### 11.7.2 Security Boundaries

| Boundary | Trust Level | Access |
|----------|-------------|--------|
| **User Environment** | High | Full filesystem, network |
| **DuckOps Process** | Medium | Workspace files, env vars |
| **AI Provider API** | Low | Prompt/response only |
| **Docker Sandbox** | Low | Scoped to container |
| **MCP Server** | Low | Defined by tool permissions |

## 11.8 Deployment Architecture

See also: Chapter 16 - Docker Architecture

### 11.8.1 Deployment Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| **Local Binary** | Direct execution on host | Daily development |
| **Client/Server** | Unix socket API | Remote workspace, team use |
| **Docker Container** | Sandboxed execution | Security assessments, CI/CD |

### 11.8.2 Deployment Architecture Diagram

```mermaid
graph TB
    subgraph "Local Mode"
        DEV[Developer Terminal] --> DUCK["duckops (binary)"]
        DUCK --> DB_LOCAL[("SQLite .duckops/")]
        DUCK --> AI_LOCAL["AI Providers"]
    end

    subgraph "Client/Server Mode"
        DEV2[Developer Terminal] --> CLI_CLIENT["duckops (client mode)"]
        CLI_CLIENT -->|Unix Socket| SRV["duckops server"]
        SRV --> DB_SRV[("SQLite")]
        SRV --> AI_SRV["AI Providers"]
    end

    subgraph "Docker Mode"
        DEV3[Developer Terminal] --> DOCKER_CLI["duckops (host)"]
        DOCKER_CLI -->|Docker API| CONT["duckops container"]
        CONT --> DB_CONT[("SQLite")]
        CONT --> TOOLS_CONT["Security Tools<br/>(Nuclei, Nmap, etc.)"]
        CONT --> AI_CONT["AI Providers"]
    end
```

---

**END OF CHAPTER 11**
