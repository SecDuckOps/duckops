# File: docs/diagrams/flowcharts.md

# Process Flowcharts

## Application Startup Flow

```mermaid
flowchart TD
    A([duckops command]) --> B{Parse flags}
    B --> C{Mode?}
    C -->|Local| D[Resolve working directory]
    C -->|Client/Server| E[Connect to server]
    D --> F[Load config files]
    F --> G[Resolve env vars]
    G --> H{Merge configs}
    H --> I[Connect SQLite]
    I --> J[Run migrations]
    J --> K[Initialize services]
    K --> L[Discover skills]
    L --> M[Start MCP connections]
    M --> N[Start LSP manager]
    N --> O{Interactive?}
    O -->|Yes| P[Start Bubble Tea TUI]
    O -->|No| Q[Run non-interactive]
    P --> R[Render prompt]
    Q --> S[Execute prompt]
    S --> T[Stream response]
    T --> U[Exit]
    R --> V([Ready for input])
```

## Message Processing Flow

```mermaid
flowchart TD
    A([User message]) --> B[Save to messages table]
    B --> C[Load session history]
    C --> D[Calculate token budget]
    D --> E{Under limit?}
    E -->|No| F[Generate summary]
    F --> G[Save summary message]
    G --> H[Trim context]
    E -->|Yes| H
    H --> I[Build provider request]
    I --> J[Select provider]
    J --> K[Send to AI provider]
    K --> L{Response chunk?}
    L -->|Text token| M[Forward to UI]
    M --> L
    L -->|Tool call| N[Validate tool]
    N --> O{Valid?}
    O -->|No| P[Auto-repair]
    P --> Q{Repaired?}
    Q -->|Yes| R[Execute tool]
    Q -->|No| S[Return error to AI]
    O -->|Yes| R
    R --> T[Format result]
    T --> U[Send result to AI]
    U --> L
    L -->|Finish reason| V[Save messages]
    V --> W[Update session stats]
    W --> X([Done])
```

## Security Scan Flow

```mermaid
flowchart TD
    A([Security scan request]) --> B[Parse target]
    B --> C{Scan type?}
    C -->|Network| D[Nmap scan]
    C -->|Vulnerability| E[Nuclei scan]
    C -->|Web| F[FFUF + Katana]
    C -->|SQLi| G[SQLMap]
    C -->|All| H[Orchestrate all]
    
    H --> D
    H --> E
    H --> F
    H --> G
    
    D --> I[Parse XML output]
    E --> J[Parse JSON output]
    F --> K[Parse JSON output]
    G --> L[Parse text output]
    
    I --> M{Merge results}
    J --> M
    K --> M
    L --> M
    
    M --> N[AI analyze findings]
    N --> O[Assign severities]
    O --> P[Generate report]
    P --> Q{Report format?}
    Q -->|Summary| R[Brief findings list]
    Q -->|Detailed| S[Full report with remediation]
    Q -->|JSON| T[Structured data]
    
    R --> U([Return to user])
    S --> U
    T --> U
```

## Tool Execution Flow

```mermaid
flowchart TD
    A([Tool call from AI]) --> B{Name recognized?}
    B -->|Unknown| C[Return error: unknown tool]
    B -->|Known| D{Has permission?}
    D -->|Default (Ask)| E[Show permission dialog]
    E --> F{User choice?}
    F -->|Allow Once| H[Execute]
    F -->|Allow Session| G[Cache permission]
    G --> H
    F -->|Deny| I[Return error: denied]
    D -->|Always Allow| H
    D -->|Always Deny| I
    
    H --> J{Input valid?}
    J -->|Schema error| K[Auto-repair input]
    K --> L{Repair succeeded?}
    L -->|Yes| H
    L -->|No| M[Return error: invalid]
    J -->|Valid| N[Execute with timeout]
    
    N --> O{Result?}
    O -->|Success| P[Format output]
    O -->|Timeout| Q[Kill process]
    Q --> R[Return partial output]
    O -->|Error| S{Retry?}
    S -->|Yes| T[Wait + retry]
    T --> N
    S -->|No| U[Return error]
    
    P --> V[Track file version]
    V --> W([Return to agent])
    R --> W
    U --> W
```

## Configuration Loading Flow

```mermaid
flowchart TD
    A([Start config load]) --> B[Search paths]
    B --> C{Found duckops.json?}
    
    C -->|~/.duckops/| D[Load global config]
    C -->|.duckops/| E[Load project config]
    C -->|CWD duckops.json| F[Load workspace config]
    C -->|Legacy XDG| G[Load legacy if exists]
    
    D --> H{More paths?}
    E --> H
    F --> H
    G --> H
    
    H -->|Yes| B
    H -->|No| I{Merge configs}
    I --> J[Apply defaults]
    J --> K[Resolve env vars]
    K --> L{Valid?}
    L -->|Schema errors| M[Log warnings]
    M --> N[Load providers]
    L -->|Valid| N
    
    N --> O{Has providers?}
    O -->|No| P[Search embedded providers]
    P --> Q[Load catwalk providers]
    Q --> R[Setup default models]
    O -->|Yes| R
    
    R --> S[Apply CLI overrides]
    S --> T[Store in ConfigStore]
    T --> U([Config ready])
```

## Error Recovery Flow

```mermaid
flowchart TD
    A([Error occurred]) --> B{Error type?}
    
    B -->|Provider 429| C[Calculate backoff]
    C --> D[Log rate limit]
    D --> E{Retries < 3?}
    E -->|Yes| F[Wait exp backoff]
    F --> G[Retry request]
    G --> H{Success?}
    H -->|Yes| I([Continue])
    H -->|No| E
    
    B -->|Provider 5xx| J[Log server error]
    J --> E
    
    B -->|Tool timeout| K[Kill process]
    K --> L[Return partial]
    L --> I
    
    B -->|Network error| M[Log connection failure]
    M --> N{Has fallback provider?}
    N -->|Yes| O[Switch provider]
    O --> G
    N -->|No| E
    
    E -->|Exhausted| P[Notify user]
    P --> Q([Graceful degradation])
```

---

*Referenced from Chapters 6, 8, 9, 10*
