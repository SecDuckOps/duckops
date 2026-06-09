# File: docs/09-use-cases.md

# Chapter 9: Use Cases

## 9.1 Use Case Diagram

### 9.1.1 High-Level Use Case Diagram

```mermaid
graph TB
    subgraph "DuckOps System"
        UC1["UC1: AI Conversation"]
        UC2["UC2: File Operations"]
        UC3["UC3: Shell Execution"]
        UC4["UC4: Security Scan"]
        UC5["UC5: Session Management"]
        UC6["UC6: Configuration"]
        UC7["UC7: MCP Integration"]
        UC8["UC8: Web Search/Fetch"]
        UC9["UC9: Code Analysis (LSP)"]
        UC10["UC10: Model Management"]
    end

    subgraph "Actors"
        DEV["Developer"]
        ADMIN["Administrator"]
        AI_PROV["AI Provider"]
        MCP_SRV["MCP Server"]
        LSP_SRV["LSP Server"]
    end

    DEV --> UC1
    DEV --> UC2
    DEV --> UC3
    DEV --> UC4
    DEV --> UC5
    DEV --> UC6
    DEV --> UC7
    DEV --> UC8
    DEV --> UC9
    DEV --> UC10
    ADMIN --> UC6
    AI_PROV <--> UC1
    MCP_SRV <--> UC7
    LSP_SRV <--> UC9
```

## 9.2 Detailed Use Cases

### 9.2.1 Use Case UC1: AI Conversation

| Field | Value |
|-------|-------|
| **ID** | UC1 |
| **Name** | AI Conversation |
| **Actor** | Developer |
| **Description** | User sends a message to the AI and receives a response with optional tool execution |
| **Preconditions** | At least one AI provider configured with valid API key |
| **Postconditions** | Message and response saved to session history |
| **Priority** | Critical |
| **Frequency** | Continuous |

**Main Flow:**

```mermaid
sequenceDiagram
    participant User as Developer
    participant TUI as Terminal UI
    participant Agent as Agent Coordinator
    participant CB as Context Budget
    participant Prov as AI Provider
    participant Tool as Tool Executor
    participant DB as Database

    User->>TUI: Type message
    TUI->>Agent: SubmitMessage(text)
    Agent->>DB: Save user message
    Agent->>CB: Check context budget

    alt Budget exceeded
        Agent->>Prov: Summarize session
        Prov-->>Agent: Summary text
        Agent->>DB: Save summary message
    end

    Agent->>Prov: Send messages (with tool schemas)
    Prov-->>Agent: Stream response tokens
    TUI-->>User: Display streaming tokens

    loop Tool Calls
        Prov->>Agent: Tool call request
        Agent->>Tool: Execute tool
        Tool-->>Agent: Tool result
        Agent->>Prov: Submit tool result
        Prov-->>Agent: Continue response
    end

    Agent->>DB: Save assistant message
    TUI-->>User: Display complete response
```

**Alternative Flows:**

| Flow | Trigger | Action |
|------|---------|--------|
| **AF1.1: Provider Error** | Provider API returns error | Agent retries with exponential backoff (3 attempts), then notifies user |
| **AF1.2: Rate Limited** | Provider rate limit hit | Agent waits and retries with backoff |
| **AF1.3: Session Busy** | Agent is processing another request | Message queued, user notified of position |
| **AF1.4: Context Full** | Context window exhausted | Auto-summarization before sending |
| **AF1.5: Tool Loop** | Tool calls exceed 50 iterations | Agent breaks loop and notifies user |

### 9.2.2 Use Case UC2: File Operations

| Field | Value |
|-------|-------|
| **ID** | UC2 |
| **Name** | File Operations |
| **Actor** | Developer (via AI agent) |
| **Description** | Agent reads, writes, edits, and searches files in the workspace |
| **Preconditions** | Active session, workspace directory accessible |
| **Postconditions** | File changes tracked with version history |
| **Priority** | Critical |
| **Frequency** | High |

**Activity Diagram:**

```mermaid
stateDiagram-v2
    [*] --> FileOpRequest
    FileOpRequest --> ValidatePath
    ValidatePath --> DetermineType
    DetermineType --> ReadFile: View/Read
    DetermineType --> WriteFile: Create/Write
    DetermineType --> EditFile: Edit
    DetermineType --> SearchFiles: Glob/Grep
    ReadFile --> CheckPerms
    WriteFile --> CheckPerms
    EditFile --> CheckPerms
    SearchFiles --> CheckPerms
    CheckPerms --> Deny: No permission
    CheckPerms --> Execute: Permission granted
    Deny --> [*]
    Execute --> ReadResult: Read
    Execute --> WriteResult: Write
    Execute --> EditResult: Edit
    Execute --> SearchResult: Search
    ReadResult --> TrackVersion
    WriteResult --> TrackVersion
    EditResult --> TrackVersion
    SearchResult --> ReturnResults
    TrackVersion --> ReturnResults
    ReturnResults --> [*]
```

**Scenarios:**

| Operation | Input | Output |
|-----------|-------|--------|
| **View** | File path | File content with syntax highlighting |
| **Write** | File path + content | New file created, version 1 |
| **Edit** | File path + line range + new content | File updated, version incremented |
| **Glob** | Pattern (e.g., `**/*.go`) | Matching file paths |
| **Grep** | Pattern (regex) + optional path | Matching lines with context |
| **Multi-Edit** | List of {file, edits} | All files updated atomically |

### 9.2.3 Use Case UC3: Shell Execution

| Field | Value |
|-------|-------|
| **ID** | UC3 |
| **Name** | Shell Execution |
| **Actor** | Developer (via AI agent) |
| **Description** | Agent executes shell commands with output capture |
| **Preconditions** | Active session |
| **Postconditions** | Command output saved in message history |
| **Priority** | High |
| **Frequency** | High |

**Flowchart:**

```mermaid
graph TB
    START([Shell Command Request]) --> VALIDATE{Validate Command}
    VALIDATE -->|Safe| TIMEOUT{Has Timeout?}
    VALIDATE -->|Blocked| REJECT[Reject Command]
    TIMEOUT -->|Yes| EXEC_WITH[Execute with timeout]
    TIMEOUT -->|No| EXEC_DEF[Execute with default timeout 120s]
    EXEC_WITH --> MONITOR[Monitor Execution]
    EXEC_DEF --> MONITOR
    MONITOR --> SPLIT{Output Size?}
    SPLIT -->|Small| RETURN[Return full output]
    SPLIT -->|Large| TRUNCATE[Return truncated + notice]
    REJECT --> ERR([Error: Command not allowed])
    RETURN --> DONE([Done])
    TRUNCATE --> DONE
```

**Error Handling:**

| Error | Cause | Handling |
|-------|-------|----------|
| **Timeout** | Command exceeds time limit | Kill process, return partial output |
| **Permission** | Command not in allowed list | Return permission denied message |
| **Not Found** | Command does not exist | Suggest installation |
| **Non-Zero Exit** | Command failed | Return exit code + output |
| **Output Too Large** | Exceeds 100MB limit | Truncate and notify |

### 9.2.4 Use Case UC4: Security Scan

| Field | Value |
|-------|-------|
| **ID** | UC4 |
| **Name** | Security Scan |
| **Actor** | Developer (via AI agent) |
| **Description** | Agent orchestrates security scanning tools and analyzes results |
| **Preconditions** | Security tools installed (natively or via Docker) |
| **Postconditions** | Scan results analyzed and report generated |
| **Priority** | Medium |
| **Frequency** | Low-Medium |

**Sequence Diagram:**

```mermaid
sequenceDiagram
    participant User as Developer
    participant Agent as Agent
    participant TP as Tool Proxy
    participant Nmap as Nmap
    participant Nuclei as Nuclei
    participant Parser as Result Parser
    participant AI as AI Provider

    User->>Agent: "Scan example.com for vulnerabilities"
    Agent->>TP: Execute("nmap -sV -sC example.com")
    TP->>Nmap: Run scan
    Nmap-->>TP: XML output
    TP->>Parser: Parse Nmap results
    Parser-->>TP: {open_ports, services, versions}

    Agent->>TP: Execute("nuclei -u https://example.com")
    TP->>Nuclei: Run scan
    Nuclei-->>TP: JSON output
    TP->>Parser: Parse Nuclei results
    Parser-->>TP: {vulnerabilities, severity}

    Agent->>AI: Analyze combined findings
    AI-->>Agent: Severity assessment, remediation
    Agent-->>User: Formatted security report
```

**Supported Scan Types:**

| Tool | Input | Output Format | Use Case |
|------|-------|---------------|----------|
| Nmap | Target host/IP | XML | Port discovery, service detection |
| Naabu | Target host | JSON | Fast port scanning |
| Nuclei | Target URL | JSON | Vulnerability scanning |
| SQLMap | Target URL + params | Text | SQL injection detection |
| FFUF | Target URL + wordlist | JSON | Directory fuzzing |
| Subfinder | Domain | Text | Subdomain enumeration |
| Httpx | URL/probe | JSON | HTTP probing |
| Katana | Target URL | JSON | Web crawling |
| Semgrep | Source directory | SARIF | Static code analysis |

### 9.2.5 Use Case UC5: Session Management

| Field | Value |
|-------|-------|
| **ID** | UC5 |
| **Name** | Session Management |
| **Actor** | Developer |
| **Description** | User manages chat sessions (create, list, resume, delete) |
| **Preconditions** | Application initialized |
| **Postconditions** | Session state changes persisted |
| **Priority** | High |
| **Frequency** | Medium |

**Use Case Flow:**

```mermaid
stateDiagram-v2
    [*] --> AppStart
    AppStart --> SessionSelect: List sessions
    SessionSelect --> NewSession: Create new
    SessionSelect --> ResumeSession: Select existing
    NewSession --> ActiveSession
    ResumeSession --> ActiveSession
    ActiveSession --> SendMessage: Type message
    SendMessage --> ActiveSession
    ActiveSession --> RenameSession: Rename
    ActiveSession --> DeleteSession: Delete
    DeleteSession --> SessionSelect
    ActiveSession --> CloseSession: Quit
    CloseSession --> [*]
```

**Session Operations:**

| Operation | Description | CLI Command |
|-----------|-------------|-------------|
| **Create** | Start new conversation | `duckops` (new session) |
| **List** | Show all sessions | `duckops session list` |
| **Resume** | Continue existing session | `duckops --session <id>` |
| **Rename** | Change session title | Via TUI or API |
| **Delete** | Remove session | `duckops session delete <id>` |
| **Archive** | Hide from active list | Future feature |

### 9.2.6 Use Case UC6: Configuration Management

| Field | Value |
|-------|-------|
| **ID** | UC6 |
| **Name** | Configuration Management |
| **Actor** | Developer, Administrator |
| **Description** | User configures providers, models, tools, and preferences |
| **Preconditions** | Configuration file accessible |
| **Postconditions** | Configuration changes applied |
| **Priority** | High |
| **Frequency** | Low |

**Configuration Sources:**

```mermaid
graph TB
    subgraph "Config Sources (Priority Order)"
        CLI[CLI Flags: highest priority]
        LOCAL[.duckops/duckops.json: project level]
        GLOBAL[~/.duckops/duckops.json: user global]
        XDG[~/.config/duckops/duckops.json: legacy XDG]
        DEFAULT[Embedded defaults: lowest priority]
    end

    CLI --> MERGE((Merge))
    LOCAL --> MERGE
    GLOBAL --> MERGE
    XDG --> MERGE
    DEFAULT --> MERGE
    MERGE --> VALIDATE{Validate}
    VALIDATE -->|Valid| APPLY[Apply Configuration]
    VALIDATE -->|Invalid| ERROR[Report Errors]
```

### 9.2.7 Use Case UC7: MCP Integration

| Field | Value |
|-------|-------|
| **ID** | UC7 |
| **Name** | MCP Integration |
| **Actor** | Developer, MCP Server |
| **Description** | System connects to MCP servers for additional tools and resources |
| **Preconditions** | MCP server configured with valid transport |
| **Postconditions** | MCP tools available to agent |
| **Priority** | Medium |
| **Frequency** | Low |

```mermaid
sequenceDiagram
    participant Agent as Agent
    participant MCPC as MCP Client
    participant MCPS as MCP Server
    participant Tool as External Tool

    Note over Agent,MCPC: Startup
    MCPC->>MCPS: Initialize session
    MCPS-->>MCPC: Server capabilities
    MCPC->>MCPS: List tools
    MCPS-->>MCPC: Tool definitions
    MCPC->>MCPS: List resources
    MCPS-->>MCPC: Resource URIs
    MCPC-->>Agent: Register tools

    Note over Agent,Tool: Runtime
    Agent->>MCPC: Call tool(params)
    MCPC->>MCPS: tools/call
    MCPS->>Tool: Execute
    Tool-->>MCPS: Result
    MCPS-->>MCPC: Tool result
    MCPC-->>Agent: Formatted result
```

### 9.2.8 Use Case UC8: Web Search/Fetch

| Field | Value |
|-------|-------|
| **ID** | UC8 |
| **Name** | Web Search and Fetch |
| **Actor** | Developer (via AI agent) |
| **Description** | Agent fetches web pages or performs web searches |
| **Preconditions** | Network connectivity |
| **Postconditions** | Content saved in message history |
| **Priority** | Medium |
| **Frequency** | Medium |

**Fetch Flow:**

```mermaid
graph TB
    START([Web Request]) --> DETERMINE{Type?}
    DETERMINE -->|Fetch URL| FETCH[HTTP GET URL]
    DETERMINE -->|Search| SEARCH[Web Search API]
    FETCH --> CHECK_STATUS{Status OK?}
    CHECK_STATUS -->|Yes| CONVERT[HTML to Markdown]
    CHECK_STATUS -->|No| ERR[Return Error]
    SEARCH --> RESULTS[Return Search Results]
    CONVERT --> TRUNCATE{Content Size?}
    TRUNCATE -->|Under Limit| RETURN[Return Full Content]
    TRUNCATE -->|Over Limit| SUMMARY[Return Summary + Link]
    RESULTS --> DONE([Done])
    RETURN --> DONE
    SUMMARY --> DONE
    ERR --> DONE
```

## 9.3 User Stories

### 9.3.1 Epic: AI Assistance

| Story ID | As a... | I want to... | So that... | Priority | Effort |
|----------|---------|--------------|------------|----------|--------|
| US-001 | Developer | Chat with multiple AI providers in one session | I can use the best model for each task | P0 | L |
| US-002 | Developer | Resume interrupted sessions | I don't lose context when closing the app | P0 | M |
| US-003 | Developer | Have the AI read my project files | I don't need to copy-paste code | P0 | M |
| US-004 | Developer | Execute shell commands through AI | I stay in the terminal context | P1 | M |
| US-005 | Developer | Have the AI edit multiple files | I can refactor across the codebase | P1 | H |
| US-006 | Developer | Search my codebase through AI | I can find relevant code quickly | P0 | M |
| US-007 | Developer | Get web search results in chat | I don't switch to browser | P2 | M |

### 9.3.2 Epic: Security Assessment

| Story ID | As a... | I want to... | So that... | Priority | Effort |
|----------|---------|--------------|------------|----------|--------|
| US-008 | Developer | Run vulnerability scans from chat | I find security issues early | P1 | H |
| US-009 | Security Researcher | Orchestrate multiple security tools | I do comprehensive assessments | P2 | H |
| US-010 | Developer | Get AI analysis of scan results | I understand the findings | P1 | M |
| US-011 | Developer | Scan dependencies for vulnerabilities | I avoid supply chain attacks | P2 | M |

### 9.3.3 Epic: Configuration & Control

| Story ID | As a... | I want to... | So that... | Priority | Effort |
|----------|---------|--------------|------------|----------|--------|
| US-012 | Developer | Configure AI providers easily | I can start using the tool quickly | P0 | M |
| US-013 | Administrator | Set default models for my team | everyone uses approved models | P2 | L |
| US-014 | Developer | Control which tools the AI can use | I prevent dangerous operations | P1 | M |
| US-015 | Developer | See my token usage and costs | I manage my API budget | P1 | L |

### 9.3.4 Epic: Integration

| Story ID | As a... | I want to... | So that... | Priority | Effort |
|----------|---------|--------------|------------|----------|--------|
| US-016 | Developer | Connect MCP servers for custom tools | I extend DuckOps capabilities | P2 | M |
| US-017 | Developer | See LSP diagnostics in chat | I fix code issues faster | P2 | M |
| US-018 | Developer | Use Docker sandbox for secure execution | I safely run untrusted tools | P3 | H |
| US-019 | Developer | Create custom skills | I share workflows with my team | P3 | M |

## 9.4 Use Case Prioritization Matrix

| Use Case | Business Value | Technical Risk | User Impact | Priority |
|----------|---------------|----------------|-------------|----------|
| UC1: AI Conversation | 95 | 40 | 95 | P0 |
| UC5: Session Management | 90 | 25 | 85 | P0 |
| UC2: File Operations | 85 | 35 | 80 | P0 |
| UC6: Configuration | 80 | 20 | 70 | P0 |
| UC3: Shell Execution | 75 | 40 | 75 | P1 |
| UC8: Web Search/Fetch | 70 | 25 | 65 | P1 |
| UC4: Security Scan | 65 | 55 | 60 | P2 |
| UC7: MCP Integration | 55 | 45 | 50 | P2 |
| UC9: Code Analysis (LSP) | 50 | 40 | 45 | P2 |
| UC10: Model Management | 45 | 30 | 40 | P2 |

## 9.5 System Workflows

### 9.5.1 Startup Workflow

```mermaid
sequenceDiagram
    participant User as User
    participant CLI as CLI
    participant CFG as Config
    participant DB as Database
    participant APP as App
    participant UI as TUI

    User->>CLI: duckops
    CLI->>CFG: Load configuration
    CFG->>CFG: Search config paths
    CFG->>CFG: Merge configs
    CFG-->>CLI: Resolved config

    CLI->>DB: Connect SQLite
    DB->>DB: Run pending migrations
    DB-->>CLI: Ready

    CLI->>APP: Initialize services
    APP->>APP: Create session service
    APP->>APP: Create message service
    APP->>APP: Create agent coordinator
    APP->>APP: Discover skills
    APP->>APP: Start MCP connections
    APP-->>CLI: Services ready

    CLI->>UI: Start Bubble Tea
    UI-->>User: Interactive prompt
```

### 9.5.2 Non-Interactive Run Workflow

```mermaid
sequenceDiagram
    participant User as User
    participant CLI as CLI
    participant APP as App
    participant SESS as Session
    participant AGENT as Agent
    participant PROV as Provider

    User->>CLI: duckops run "Fix this bug" --session last
    CLI->>APP: Initialize (no TUI)
    APP->>SESS: Resolve session (last or new)
    SESS-->>APP: Session ID
    APP->>AGENT: Prepare agent with config
    APP->>USER: Show spinner

    par Stream Response
        AGENT->>PROV: Send prompt
        PROV-->>AGENT: Stream tokens
        AGENT-->>APP: Forward tokens
        APP-->>User: Print to stdout
    end

    APP-->>User: Done
    APP->>APP: Cleanup
```

## 9.6 Process Flowcharts

### 9.6.1 Tool Execution Flowchart

```mermaid
flowchart TD
    A([Tool Call Request]) --> B{Valid Tool?}
    B -->|No| C[Return Error: Unknown Tool]
    B -->|Yes| D{Has Permission?}
    D -->|Ask| E[Show Permission Dialog]
    E -->|Allow| F[Proceed]
    E -->|Allow Session| F
    E -->|Deny| G[Return Error: Permission Denied]
    D -->|Allow| F
    D -->|Deny| G
    F --> H{Input Valid?}
    H -->|No| I[Auto-Repair Input]
    I -->|Success| J[Execute Tool]
    I -->|Failed| K[Return Error: Invalid Input]
    H -->|Yes| J
    J --> L{Execution OK?}
    L -->|Yes| M[Format Result]
    L -->|Timeout| N[Kill + Return Partial]
    L -->|Error| O{Retry?}
    O -->|Yes| P[Wait + Retry]
    P --> J
    O -->|No| Q[Return Error]
    M --> R[Track Version / Log]
    N --> R
    Q --> R
    R --> S([Return to Agent])
```

### 9.6.2 Message Processing Flowchart

```mermaid
flowchart TD
    A([User Message]) --> B[Save to DB]
    B --> C[Check Context Budget]
    C --> D{Budget OK?}
    D -->|Yes| E[Build Message Array]
    D -->|No| F[Auto-Summarize]
    F --> G[Save Summary Message]
    G --> E
    E --> H[Select Provider]
    H --> I[Send to Provider API]
    I --> J{Response Type?}
    J -->|Token Stream| K[Forward Tokens to UI]
    K --> J
    J -->|Tool Call| L[Execute Tool]
    L --> M[Send Result to Provider]
    M --> J
    J -->|Stop/Finish| N[Save Messages]
    N --> O[Update Session Stats]
    O --> P([Done])
```

---

**END OF CHAPTER 9**
