# File: docs/15-agent-system.md

# Chapter 15: Agent System

## 15.1 Overview

The DuckOps agent system is a multi-layered AI orchestration framework that manages conversational sessions, tool execution, context budgets, provider failover, and model selection. At its core is the `Coordinator` — a lifecycle manager that creates and controls individual `SessionAgent` instances, each responsible for one conversation session.

### 15.1.1 Architecture Summary

```
┌─────────────────────────────────────────────────────┐
│                    Coordinator                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │ SessionAgent │  │ SessionAgent │  │ SessionAgent │  │
│  │   (ID: a1)  │  │   (ID: b2)  │  │   (ID: c3)  │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │          │
│         └─────────────────┼─────────────────┘          │
│                           │                            │
│                    ┌──────┴──────┐                     │
│                    │  Provider   │                     │
│                    │   Router    │                     │
│                    └──────┬──────┘                     │
│                           │                            │
│              ┌────────────┼────────────┐               │
│              │            │            │               │
│          OpenAI      Anthropic    OpenRouter           │
└─────────────────────────────────────────────────────┘
```

## 15.2 The Coordinator

The `Coordinator` interface (`internal/agent/coordinator.go`) is the central orchestrator:

```go
type Coordinator interface {
    Run(ctx context.Context, req *RunRequest) error
    Cancel(sessionID string)
    CancelAll()
    IsSessionBusy(sessionID string) bool
    Summarize(ctx context.Context, sessionID string) error
    Model() *Model
}
```

### Responsibilities

| Responsibility | Description |
|---------------|-------------|
| **Session Management** | Creates, tracks, and manages concurrent session agents |
| **Model Provision** | Holds the active large/small/local model references |
| **Cancellation** | Gracefully cancels in-flight requests per session |
| **Summarization** | Triggers context compression when budgets are exceeded |
| **Lifecycle** | Starts agents on demand, cleans up on shutdown |

### Implementation Details

The concrete `coordinator` struct stores a map of active agents:

```go
type coordinator struct {
    mu       sync.RWMutex
    sessions map[string]*sessionAgent
    large    *Model
    small    *Model
    local    *Model
    db       *sql.DB
    q        *db.Queries
}
```

When `Run` is called, it either reuses an existing agent (if one exists for the session ID) or creates a new one. Each agent runs in its own goroutine, processing the conversation stream.

## 15.3 SessionAgent

The `SessionAgent` interface (`internal/agent/session.go`) handles a single conversation:

```go
type SessionAgent interface {
    Run(ctx context.Context, params *RunParams) error
    SetModels(large, small, local *Model)
    SetTools(tools []Tool)
    SetSystemPrompt(prompt string)
    Cancel()
    Summarize(ctx context.Context) error
}
```

### Conversation Loop

The agent's `Run` method implements the core chat loop:

```
┌────────────────────────────────────────────┐
│           Conversation Loop                 │
│                                            │
│  1. Load messages from DB                  │
│  2. Check context budget (summarize if  )│
│  3. Build provider request (system+msgs)   │
│  4. Stream response from provider          │
│  5. For each chunk:                        │
│     a. Text → forward to UI                │
│     b. Tool call → execute, repair, loop   │
│     c. Finish → save, update session         │
│  6. Repeat on next user message            │
└────────────────────────────────────────────┘
```

### Tool Call Handling

When the AI invokes a tool, the agent follows this sequence:

1. **Validation**: Check the tool exists and the input matches its JSON schema
2. **Permission Check**: Consult the `PermissionService` — always allow, always deny, or ask the user
3. **Execution**: Run the tool with a configurable timeout (default 120s)
4. **Auto-Repair**: If the tool errors, attempt automatic input repair (up to 3 retries)
5. **Loop Detection**: Monitor for repetitive tool calls (same tool, same input >50 times) to break infinite loops
6. **Result Injection**: Return the formatted result to the AI provider

```go
type RunParams struct {
    SessionID string
    UserMessage string
    Images []string
    Tools []Tool
    SystemPrompt string
}
```

## 15.4 Models

The `Model` struct wraps three layers of model configuration:

```go
type Model struct {
    LanguageModel fantasy.LanguageModel  // Fantasy SDK runtime
    CatwalkCfg    catwalk.Model          // Catwalk provider config
    ModelCfg      config.SelectedModel   // DuckOps user config
    FlatRate      bool
}
```

### Fantasy SDK Integration

`fantasy.LanguageModel` is the core interface from the Fantasy LLM SDK:

```go
type LanguageModel interface {
    ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (<-chan ChatEvent, error)
}
```

DuckOps converts its internal message format to Fantasy's `ChatCompletionRequest`, which includes messages, tools, and provider-specific options. The response is a channel of events (text tokens, tool calls, finish reasons, errors).

### Model Selection

Models are configured in `duckops.json` under three types:

```json
{
  "large": { "provider": "openai", "model": "gpt-4o" },
  "small": { "provider": "openai", "model": "gpt-4o-mini" },
  "local": { "provider": "ollama", "model": "llama3" }
}
```

| Type | Purpose | Size/Cost |
|------|---------|-----------|
| `large` | Complex reasoning, code generation | Full-featured, higher cost |
| `small` | Quick tasks, summarization | Faster, cheaper |
| `local` | Offline usage, sensitive data | Runs on-device |

The coordinator selects which model to use based on the task cost estimate and configuration.

## 15.5 Context Budget

The `ContextBudget` system (`internal/agent/budget.go`) prevents token overflow by monitoring conversation length:

```go
type BudgetResult struct {
    IsOverBudget  bool
    TotalTokens   int
    MaxTokens     int
    SuggestedAction string  // "summarize", "truncate", or "none"
}
```

### Token Estimation

Token counts are estimated using the model's context window. When a session exceeds the budget, the agent:

1. Generates a summary of earlier messages (using the small model)
2. Inserts the summary as a message with `role: system`
3. Removes the original messages from context (but keeps them in the database)
4. Continues the conversation with the truncated context

```
Before:
[User] How do I implement auth?
[Assistant] Here's how...
[User] Now add rate limiting
[Assistant] Sure...
[User] What about caching?   ← over budget

After:
[System] Summary: User asked about auth, rate limiting. Provided solutions.
[User] What about caching?   ← within budget
```

### Auto-Summarization

When `DisableAutoSummarize` is `false` (default), summarization happens automatically. The small model generates a concise summary, and the conversation continues seamlessly. The summary message is clearly labeled so the user can expand it to see full context.

## 15.6 Loop Detection

The `LoopDetector` (`internal/agent/loop.go`) prevents infinite tool execution cycles:

```go
type LoopDetector struct {
    recentCalls []ToolCallRecord
    threshold   int  // default: 50
}

func (d *LoopDetector) IsLooping() bool
func (d *LoopDetector) Record(call ToolCall)
func (d *LoopDetector) Reset()
```

It tracks the last N tool calls. If the same tool is called with the same input more than the threshold (50), it signals the agent to break the loop and return an error. After breaking, the agent sends a message to the AI explaining the loop.

## 15.7 Tool Repair

The `ToolRepair` system (`internal/agent/repair.go`) attempts to fix malformed tool calls:

```go
type ToolRepair struct {
    maxAttempts int  // default: 3
}

func (r *ToolRepair) Repair(tool Tool, input json.RawMessage, execErr string) (json.RawMessage, error)
```

When a tool execution fails, the repair system uses the small model to analyze the error and suggest corrected input. It follows a strict protocol:

1. **Analyze**: Determine why the tool call failed (wrong field name, missing parameter, type mismatch)
2. **Fix**: Generate corrected JSON input
3. **Validate**: Check the corrected input against the tool's JSON schema
4. **Execute**: Run the tool with corrected input (up to `maxAttempts`)

```json
{
  "original": { "file": "src/foo.go", "old_str": "abc", "new_str": "def" },
  "error": "old_str not found in file",
  "analysis": "The string 'abc' appears after leading whitespace",
  "fixed": { "file": "src/foo.go", "old": "abc", "new": "def" }
}
```

## 15.8 Provider Router

The provider routing system selects the best AI provider for each request:

```go
func selectProvider(req *ChatRequest, config *config.Config) (Provider, error)
```

### Selection Logic

1. **Model Configuration**: Start with the user's configured model (large/small/local)
2. **Provider Availability**: Check if the selected provider is enabled and has a valid API key
3. **Fallback Chain**: If the primary provider fails (rate limit, server error), try the next available provider
4. **Exponential Backoff**: On rate limits (HTTP 429), wait with exponential backoff (2s, 4s, 8s) up to 3 retries

### Provider Priority

Providers are tried in the order they appear in `duckops.json`. The first enabled provider with the requested model wins. Common configurations:

```
Large Model Provider Chain:
  1. OpenAI (primary)
  2. Anthropic (fallback 1)
  3. OpenRouter (fallback 2)

Small Model Provider Chain:
  1. OpenAI (primary)
  2. OpenRouter (fallback)
```

## 15.9 Session Persistence

The agent system persists all session data to SQLite:

- **Messages**: Every user and assistant message is saved with timestamps
- **Token Usage**: Running count of prompt and completion tokens per session
- **Cost Tracking**: Accumulated cost for API usage
- **Todos**: Inline task items extracted from the conversation
- **File Versions**: Read/write operations tracked per session

### Session Resume

When the application restarts, all sessions are loaded from SQLite. The user sees their full history and can continue any session. The agent loads the last N messages (within the token budget) and reconstructs the context.

## 15.10 Security Context

The agent system respects DuckOps' security model:

- **Permission Checks**: Every tool call is validated against the permission policy (always allow, always deny, or ask)
- **Tool Isolation**: Tools run in subprocesses with configurable resource limits
- **Provider Keys**: API keys are stored in the config file and never exposed to the agent's context
- **Data Isolation**: Each session's data is isolated; agents cannot access other sessions' data

## 15.11 Configuration Example

```json
{
  "models": {
    "large": { "provider": "openai", "model": "gpt-4o" },
    "small": { "provider": "openai", "model": "gpt-4o-mini" },
    "local": { "provider": "ollama", "model": "llama3" }
  },
  "options": {
    "disable-auto-summarize": false
  },
  "permissions": {
    "default": "ask",
    "tools": {
      "edit": "allow",
      "read": "allow",
      "bash": "ask",
      "web_fetch": "ask"
    }
  }
}
```

---

*Next: Chapter 16 — Docker Deployment*
