# File: docs/diagrams/sequence-diagrams.md

# Sequence Diagrams

## Full Conversation Flow

```mermaid
sequenceDiagram
    participant User as Developer
    participant TUI as Bubble Tea TUI
    participant App as App Controller
    participant Coord as Coordinator
    participant Agent as SessionAgent
    participant Budget as ContextBudget
    participant Prov as AI Provider
    participant Tools as ToolEngine
    participant Db as SQLite

    User->>TUI: Types message
    TUI->>TUI: Render in chat view
    
    TUI->>App: SubmitMessage(text)
    App->>Db: Save user message
    App->>Coord: Run(Request{sessionID, text})
    
    Coord->>Agent: Process(request)
    
    Agent->>Budget: Check(messages, model.MaxTokens)
    
    alt Budget > threshold
        Agent->>Prov: Summarize(context)
        Prov-->>Agent: Summary text
        Agent->>Db: Save summary message
        Agent->>Agent: Trim messages to budget
    end
    
    Agent->>Agent: Build provider request
    
    par Send to Provider
        Agent->>Prov: ChatCompletion(messages, tools)
        Prov-->>Agent: Stream response chunk
        
        loop Response Stream
            alt Text Token
                Prov-->>Agent: {type: "text", text: "..."}
                Agent-->>App: Forward token
                App-->>TUI: Append to message
                TUI-->>User: Display token
                
            else Tool Call
                Prov-->>Agent: {type: "tool_call", name: "...", input: {...}}
                Agent->>Tools: Execute(name, input)
                
                alt Success
                    Tools-->>Agent: Result content
                else Error
                    Tools-->>Agent: Error message
                end
                
                Agent->>Prov: SubmitToolResult(callID, result)
                
            else Finish
                Prov-->>Agent: {type: "finish", reason: "stop"}
            end
        end
    end
    
    Agent->>Db: Save assistant message
    Agent->>Db: Update session token counts
    Agent-->>Coord: Complete
    Coord-->>App: Done
    
    App-->>TUI: Signal complete
    TUI-->>User: Show ready prompt
```

## Tool Call with Auto-Repair

```mermaid
sequenceDiagram
    participant AI as AI Provider
    participant Agent as SessionAgent
    participant Detector as LoopDetector
    participant Repair as ToolRepair
    participant Tool as Tool Executor

    AI->>Agent: tool_call(name="edit", input={file: "x.go", old: "foo", new: "bar"})
    
    Agent->>Detector: Record(call)
    
    alt Is looping?
        Detector-->>Agent: Loop detected (>50 calls)
        Agent-->>AI: Return error: loop broken
        AI-->>Agent: Okay, stopping
    else Normal
        Agent->>Repair: Validate(name, input)
        
        alt Valid input
            Repair-->>Agent: OK
            Agent->>Tool: Execute(name, input)
            
            alt Success
                Tool-->>Agent: Result: success
                Agent-->>AI: tool_result(content="File updated")
            else Error
                Tool-->>Agent: Result: error
                Agent-->>Repair: Can repair?
                
                alt Repairable
                    Repair-->>Agent: Fixed input
                    Agent->>Tool: Execute(name, fixedInput)
                    Tool-->>Agent: Result: success
                    Agent-->>AI: tool_result(content="File updated")
                else Not repairable
                    Agent-->>AI: tool_result(isError=true, content="...")
                end
            end
            
        else Invalid input
            Repair-->>Agent: Input validation failed
            Repair->>Repair: Auto-fix input
            
            alt Fixed
                Repair-->>Agent: Fixed input
                Agent->>Tool: Execute(name, fixedInput)
                Tool-->>Agent: Result
                Agent-->>AI: tool_result(content=result)
            else Cannot fix
                Agent-->>AI: tool_result(isError=true, content="Invalid params")
            end
        end
    end
```

## Provider Failover Flow

```mermaid
sequenceDiagram
    participant Agent as SessionAgent
    participant Router as Provider Router
    participant OAI as OpenAI
    participant ANTH as Anthropic
    participant OAI2 as OpenRouter (Fallback)

    Agent->>Router: Select provider for request
    
    Note over Router: Primary: OpenAI
    
    Router->>OAI: ChatCompletion(messages)
    OAI-->>Router: HTTP 429 Rate Limited
    
    Router->>Router: Log rate limit
    Router->>Router: Wait 2s (backoff)
    Router->>OAI: Retry (attempt 2/3)
    OAI-->>Router: HTTP 503 Service Unavailable
    
    Router->>Router: Wait 4s (exponential backoff)
    Router->>OAI: Retry (attempt 3/3)
    OAI-->>Router: HTTP 503 Service Unavailable
    
    Note over Router: Primary exhausted, try fallback
    
    Router->>Router: Select fallback provider (Anthropic)
    
    Router->>ANTH: ChatCompletion(messages)
    ANTH-->>Router: Stream response
    
    Router-->>Agent: Forward response stream
    
    Agent->>Agent: Continue processing
```

## Session Lifecycle

```mermaid
sequenceDiagram
    participant User as Developer
    participant TUI as TUI
    participant App as App
    participant Session as SessionService
    participant Agent as Coordinator
    participant Db as SQLite

    Note over User,Db: Session Creation
    User->>TUI: Open application
    TUI->>App: Start
    App->>Session: Create("New Session")
    Session->>Db: INSERT sessions
    Db-->>Session: Session{ID: "sess_123"}
    Session-->>App: Session
    App-->>TUI: Ready

    Note over User,Db: Active Conversation
    User->>TUI: Type message (multiple turns)
    TUI->>App: SubmitMessage
    App->>Agent: Run
    Agent->>Db: Save messages
    
    Note over User,Db: Session Update
    User->>TUI: Rename to "Bug Fix"
    TUI->>App: RenameSession("sess_123", "Bug Fix")
    App->>Session: Rename("sess_123", "Bug Fix")
    Session->>Db: UPDATE sessions SET title = ?
    Db-->>Session: OK
    Session-->>App: Done
    App-->>TUI: Updated

    Note over User,Db: Session Resumption
    User->>TUI: Close and restart
    TUI->>App: Start
    App->>Session: Get("sess_123")
    Session->>Db: SELECT * FROM sessions WHERE id = ?
    Db-->>Session: Session data
    Session->>Db: SELECT * FROM messages WHERE session_id = ? ORDER BY created_at
    Db-->>Session: [50 messages]
    Session-->>App: Full session history
    App-->>TUI: Ready with history

    Note over User,Db: Session Deletion
    User->>TUI: Delete session
    TUI->>App: DeleteSession("sess_123")
    App->>Session: Delete("sess_123")
    Session->>Db: DELETE FROM messages WHERE session_id = ?
    Session->>Db: DELETE FROM files WHERE session_id = ?
    Session->>Db: DELETE sessions WHERE id = ?
    Db-->>Session: OK (cascade)
    Session-->>App: Done
    App-->>TUI: Removed
```

## MCP Tool Execution

```mermaid
sequenceDiagram
    participant Agent as Agent
    participant MCPC as MCP Client
    participant SRV as MCP Server
    participant EXT as External Tool

    Note over Agent,EXT: Initialization
    MCPC->>SRV: initialize(protocolVersion, capabilities)
    SRV-->>MCPC: serverInfo, capabilities
    
    MCPC->>SRV: tools/list
    SRV-->>MCPC: [{name, description, inputSchema}]
    MCPC->>MCPC: Register tools
    MCPC-->>Agent: Tools available

    Note over Agent,EXT: Tool Execution
    Agent->>MCPC: Call tool "github_create_pr"
    MCPC->>SRV: tools/call(name="github_create_pr", arguments={...})
    SRV->>EXT: Create PR via GitHub API
    EXT-->>SRV: PR URL, status
    
    alt Progress
        SRV-->>MCPC: progress notification
        MCPC-->>Agent: Progress update
    end
    
    SRV-->>MCPC: tool result: {content: [{type: "text", text: "PR created: ..."}]}
    MCPC-->>Agent: Formatted result

    Note over Agent,EXT: Resource Access
    Agent->>MCPC: Read resource "file:///config"
    MCPC->>SRV: resources/read(uri="file:///config")
    SRV-->>MCPC: resource contents
    MCPC-->>Agent: Resource data
```

## LSP Integration Flow

```mermaid
sequenceDiagram
    participant App as App Controller
    participant LSPM as LSP Manager
    participant LSPC as LSP Client
    participant GPLS as gopls
    participant TUI as TUI

    Note over App,TUI: LSP Connection
    App->>LSPM: Start(gopls)
    LSPM->>LSPC: Connect
    LSPC->>GPLS: initialize(rootUri, capabilities)
    GPLS-->>LSPC: serverCapabilities
    LSPC->>GPLS: initialized
    GPLS-->>LSPC: OK
    
    LSPC->>GPLS: textDocument/didOpen(fileUri, text)
    GPLS-->>LSPC: diagnostics

    Note over App,TUI: Diagnostics
    LSPC-->>LSPM: PublishDiagnostics(uri, diagnostics)
    LSPM-->>App: LSPEvent{diagnostics}
    App-->>TUI: Show diagnostics in sidebar
    TUI-->>User: Error highlighting

    Note over App,TUI: Code Actions
    User->>TUI: Request code action
    TUI->>App: GetCodeActions(file, line)
    App->>LSPM: CodeAction(file, range)
    LSPM->>LSPC: textDocument/codeAction
    LSPC->>GPLS: codeAction(params)
    GPLS-->>LSPC: [{title, kind, edit}]
    LSPC-->>LSPM: Code actions
    LSPM-->>App: Actions list
    App-->>TUI: Show available actions
    TUI-->>User: Pick action

    Note over App,TUI: Hover/Completion
    User->>TUI: Hover on symbol
    TUI->>App: Hover(file, position)
    App->>LSPM: HoverRequest(file, pos)
    LSPM->>LSPC: textDocument/hover
    LSPC->>GPLS: hover(params)
    GPLS-->>LSPC: Hover contents
    LSPC-->>LSPM: Hover result
    LSPM-->>App: Documentation
    App-->>TUI: Show tooltip
```

## Configuration Reload Flow

```mermaid
sequenceDiagram
    participant FS as File System
    participant Watch as File Watcher
    participant Store as ConfigStore
    participant App as Application
    participant Coord as Coordinator

    Note over FS,Coord: Initial Load
    FS->>Store: ReadConfig(paths...)
    Store->>Store: Merge configs
    Store->>Store: Resolve env vars
    Store-->>App: Config ready

    Note over FS,Coord: Runtime Change
    User->>FS: Edit duckops.json
    FS->>Watch: File change event
    
    Watch->>Store: ReloadFromDisk()
    Store->>FS: Read updated config
    FS-->>Store: New config
    
    Store->>Store: Merge & validate
    alt Valid
        Store->>Store: Apply new config
        Store-->>App: ConfigUpdated event
        
        App->>App: Check for model changes
        alt Models changed
            App->>Coord: UpdateModels(large, small)
            Coord->>Coord: Rebuild agent models
        end
        
        App->>App: Check for provider changes
        alt Providers changed
            App->>App: Notify agent of provider update
        end
        
        App->>App: Check for MCP changes
        alt MCP config changed
            App->>App: Restart MCP connections
        end
        
    else Invalid
        Store-->>App: ConfigError
        App-->>User: Show error notification
    end
```

---

*Referenced from Chapters 9, 10, 11*
