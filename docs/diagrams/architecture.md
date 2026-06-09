# File: docs/diagrams/architecture.md

# Architecture Diagrams

## System Context Diagram

```mermaid
graph TB
    subgraph "DuckOps System"
        CORE["DuckOps Core"]
    end

    DEV("Developer") -->|Uses terminal| CORE
    CORE -->|API calls| AI_PROV("AI Providers")
    CORE -->|HTTP| WEB("Web Services")
    CORE -->|Docker API| DOCKER("Docker Engine")
    CORE -->|stdio/HTTP| MCP_SRV("MCP Servers")
    CORE -->|JSON-RPC| LSP_SRV("LSP Servers")
    CORE -->|syscalls| OS("Operating System")
    OS --> FS[("Filesystem")]
    OS --> DB[("SQLite Database")]
```

## Layered Architecture

```mermaid
graph TB
    subgraph "Presentation Layer"
        TUI["TUI: bubbletea.Model"]
        CLI["CLI: cobra.Command"]
        API["API: http.Handler"]
    end

    subgraph "Application Layer"
        APP["app.App"]
        BACKEND["backend.Backend"]
        WSP["workspace.Workspace"]
    end

    subgraph "Domain Layer"
        SVC_SESSION["session.Service"]
        SVC_MSG["message.Service"]
        SVC_HIST["history.Service"]
        SVC_PERM["permission.Service"]
        SVC_TRACK["filetracker.Service"]
    end

    subgraph "Infrastructure Layer"
        DB["db.Queries<br/>(sqlc)"]
        SHELL["shell.Executor"]
        PROVIDER["fantasy.Provider"]
        TOOLS["agent.Tools"]
        LSP["lsp.Client"]
        MCP["mcp.Client"]
    end

    subgraph "External Interfaces"
        EXT_AI["AI Providers<br/>OpenAI, Anthropic, etc."]
        EXT_WEB["Web"]
        EXT_TOOLS["Security Tools<br/>Nuclei, Nmap, etc."]
        EXT_DOCKER["Docker"]
        EXT_MCP["MCP Ecosystem"]
        EXT_LSP["gopls etc."]
    end

    TUI --> APP
    CLI --> APP
    API --> BACKEND --> APP
    APP --> SVC_SESSION
    APP --> SVC_MSG
    APP --> SVC_HIST
    APP --> SVC_PERM
    APP --> SVC_TRACK
    APP --> AGENT["agent.Coordinator"]
    AGENT --> PROVIDER --> EXT_AI
    AGENT --> TOOLS --> EXT_WEB
    AGENT --> TOOLS --> SHELL
    AGENT --> TOOLS --> EXT_TOOLS
    AGENT --> TOOLS --> MCP --> EXT_MCP
    APP --> LSP --> EXT_LSP
    SVC_SESSION --> DB
    SVC_MSG --> DB
    SVC_HIST --> DB
    AGENT --> DB
```

## Component Architecture

```mermaid
graph TB
    subgraph "Config Layer"
        CONFIG["config.ConfigStore"]
        SCHEMA["JSON Schema<br/>Validation"]
        RESOLVER["Variable<br/>Resolver"]
        CATWALK["Catwalk<br/>Provider Registry"]
    end

    subgraph "Core Domain"
        SESSION["Session Service<br/>CRUD + todos"]
        MESSAGE["Message Service<br/>Content parts"]
        HISTORY["History Service<br/>File versions"]
    end

    subgraph "Agent System"
        COORD["Coordinator<br/>Multi-session"]
        AGENT["SessionAgent<br/>Per-conversation"]
        BB["ContextBudget<br/>Token management"]
        TOOL_ENG["ToolEngine<br/>35+ tools"]
        LOOP["LoopDetector<br/>Safety"]
        REPAIR["ToolRepair<br/>Auto-correction"]
    end

    subgraph "Provider Abstraction"
        FANTASY["fantasy SDK<br/>LLM abstraction"]
        OAI["OpenAI Adapter"]
        ANTH["Anthropic Adapter"]
        GEM["Gemini Adapter"]
        OCOMP["OpenAI-Compatible<br/>Adapter"]
    end

    subgraph "UI System"
        MODEL["ui.Model<br/>Application state"]
        CHAT["ui/chat<br/>Renderers"]
        DIALOG["ui/dialog<br/>Overlays"]
        STYLES["ui/styles<br/>Themes"]
    end

    CONFIG --> COORD
    CONFIG --> SESSION
    CONFIG --> MESSAGE
    COORD --> AGENT
    AGENT --> BB
    AGENT --> TOOL_ENG
    AGENT --> LOOP
    AGENT --> REPAIR
    AGENT --> FANTASY
    FANTASY --> OAI
    FANTASY --> ANTH
    FANTASY --> GEM
    FANTASY --> OCOMP
    MODEL --> CHAT
    MODEL --> DIALOG
    MODEL --> STYLES
    COORD --> MODEL
    SESSION --> DB[(SQLite)]
    MESSAGE --> DB
    HISTORY --> DB
```

## Agent Architecture

```mermaid
graph TB
    subgraph "Agent Runtime"
        MAIN_LOOP["Main Loop<br/>while !finished"]
        STREAM["Stream Handler<br/>process chunks"]
        TOOL_PROC["Tool Processor<br/>execute & retry"]
    end

    subgraph "Context Pipeline"
        COLLECT["Context Collector<br/>gather messages"]
        BUDGET["Budget Checker<br/>token estimation"]
        SUMM["Summarizer<br/>compress if needed"]
        BUILD["Message Builder<br/>construct payload"]
    end

    subgraph "Provider Layer"
        ENCODE["Request Encoder<br/>provider format"]
        SEND["API Sender<br/>HTTP + streaming"]
        DECODE["Response Decoder<br/>chunk parser"]
    end

    subgraph "Tool Engine"
        VALIDATE["Tool Validator<br/>schema check"]
        EXEC["Tool Executor<br/>run tool"]
        FIX["Auto Repair<br/>fix & retry"]
        RESULT["Result Formatter<br/>to content parts"]
    end

    USER_INPUT["User Input"] --> COLLECT
    COLLECT --> BUDGET
    BUDGET -->|Under limit| BUILD
    BUDGET -->|Over limit| SUMM --> BUILD
    BUILD --> ENCODE
    ENCODE --> SEND
    SEND --> DECODE --> STREAM
    STREAM -->|Text token| MAIN_LOOP
    STREAM -->|Tool call| TOOL_PROC
    TOOL_PROC --> VALIDATE
    VALIDATE -->|Valid| EXEC
    VALIDATE -->|Invalid| FIX --> EXEC
    EXEC --> RESULT --> MAIN_LOOP
    MAIN_LOOP -->|Finished| RESP["Final Response"]
```

## Deployment Architecture

```mermaid
graph TB
    subgraph "Local Deployment"
        DEV[Developer Machine]
        DUCK["duckops binary"]
        DB_LOCAL[(".duckops/ duckops.db")]
        CONFIG_LOCAL[("Config Files")]
        DUCK --> DB_LOCAL
        DUCK --> CONFIG_LOCAL
    end

    subgraph "Client/Server Deployment"
        CLIENT["duckops (client)"]
        NET[("Unix Socket")]
        SRV["duckops server"]
        DB_SRV[("duckops.db")]
        CLIENT --> NET --> SRV --> DB_SRV
    end

    subgraph "Docker Deployment"
        HOST["duckops (host)"]
        DOCKER[("Docker Engine")]
        CONT["Security Container"]
        TOOLS_CONT["Nuclei, Nmap,<br/>SQLMap, etc."]
        HOST -->|Docker API| DOCKER --> CONT
        CONT --> TOOLS_CONT
    end

    DUCK -->|HTTPS| AI_PROVIDERS[AI Providers]
    SRV --> AI_PROVIDERS
    CONT --> AI_PROVIDERS
```

## MCP Architecture

```mermaid
graph TB
    subgraph "DuckOps MCP Client"
        MCP_CLIENT["mcp.Client"]
        TOOL_REG["Tool Registry"]
        RESOURCE_REG["Resource Registry"]
        PROMPT_REG["Prompt Registry"]
    end

    subgraph "DuckOps MCP Server"
        MCP_SRV["mcp.Server"]
        EXPORTED_TOOLS["Exported DuckOps Tools"]
    end

    subgraph "External MCP Servers"
        EXT_MCP1["File System MCP"]
        EXT_MCP2["GitHub MCP"]
        EXT_MCP3["Database MCP"]
    end

    AGENT["Agent"] --> MCP_CLIENT
    MCP_CLIENT -->|stdio| EXT_MCP1
    MCP_CLIENT -->|HTTP/SSE| EXT_MCP2
    MCP_CLIENT -->|stdio| EXT_MCP3
    MCP_CLIENT --> TOOL_REG
    MCP_CLIENT --> RESOURCE_REG
    MCP_CLIENT --> PROMPT_REG
    TOOL_REG --> AGENT

    MCP_SRV -->|stdio| EXT_CLIENT["External MCP Clients"]
    EXPORTED_TOOLS --> MCP_SRV
```

---

*These diagrams are referenced from the main documentation chapters.*
