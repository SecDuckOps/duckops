# File: docs/diagrams/uml.md

# UML Diagrams

## Class Diagram: Core Domain Models

```mermaid
classDiagram
    class Session {
        +string ID
        +string ParentSessionID
        +string Title
        +int64 MessageCount
        +int64 PromptTokens
        +int64 CompletionTokens
        +float64 Cost
        +string SummaryMessageID
        +[]Todo Todos
        +int64 CreatedAt
        +int64 UpdatedAt
        +HashID() string
        +HasIncompleteTodos() bool
    }

    class Todo {
        +string Content
        +TodoStatus Status
        +string ActiveForm
    }

    class Message {
        +string ID
        +string SessionID
        +string Role
        +[]ContentPart Parts
        +string Model
        +string Provider
        +int64 CreatedAt
        +int64 UpdatedAt
        +Content() []ContentPart
        +ToolCalls() []ToolCall
        +IsFinished() bool
        +AppendContent(ContentPart)
        +Clone() Message
        +ToAIMessage() fantasy.Message
    }

    class ContentPart {
        <<interface>>
        +isPart()
    }

    class TextContent {
        +string Text
    }

    class ReasoningContent {
        +string Reasoning
    }

    class ImageURLContent {
        +string ImageURL
        +string Detail
    }

    class BinaryContent {
        +[]byte Data
        +string MimeType
    }

    class ToolCall {
        +string ID
        +string Name
        +json.RawMessage Input
    }

    class ToolResult {
        +string ToolCallID
        +string Content
        +bool IsError
    }

    class Finish {
        +string FinishReason
    }

    ContentPart <|.. TextContent
    ContentPart <|.. ReasoningContent
    ContentPart <|.. ImageURLContent
    ContentPart <|.. BinaryContent
    ContentPart <|.. ToolCall
    ContentPart <|.. ToolResult
    ContentPart <|.. Finish
    Message "1" --> "*" ContentPart : contains
    Session "1" --> "*" Message : has
    Session "1" --> "*" Todo : has
```

## Class Diagram: Service Layer

```mermaid
classDiagram
    class SessionService {
        <<interface>>
        +Create(ctx, title) (Session, error)
        +Get(ctx, id) (Session, error)
        +List(ctx) ([]Session, error)
        +Save(ctx, session) (Session, error)
        +Delete(ctx, id) error
        +UpdateTitleAndUsage(ctx, id, title, tokens, cost) error
    }

    class MessageService {
        <<interface>>
        +Create(ctx, params) (Message, error)
        +Get(ctx, id) (Message, error)
        +List(ctx, sessionID) ([]Message, error)
        +Delete(ctx, id) error
        +DeleteSessionMessages(ctx, sessionID) error
    }

    class HistoryService {
        <<interface>>
        +Create(ctx, path, content, sessionID) (File, error)
        +CreateVersion(ctx, path, content, sessionID) (File, error)
        +Get(ctx, id) (File, error)
        +ListBySession(ctx, sessionID) ([]File, error)
        +Delete(ctx, id) error
    }

    class ConfigStore {
        +Config() *Config
        +SetConfigField(path string, value any) error
        +RemoveConfigField(path string) error
        +UpdatePreferredModel(modelType, provider, model) error
        +SetProviderAPIKey(provider, key) error
        +ReloadFromDisk(ctx) error
    }

    class PermissionService {
        +Check(ctx, toolName, input) (PermissionResult, error)
        +Grant(decision, toolName, sessionID)
        +SetSkip(skip bool)
    }

    class FileTrackerService {
        <<interface>>
        +RecordRead(ctx, sessionID, path) error
        +LastReadTime(ctx, sessionID, path) (int64, error)
        +ListReadFiles(ctx, sessionID) ([]ReadFile, error)
    }

    class Querier {
        <<interface>>
        +CreateSession(ctx, params) (Session, error)
        +GetSessionByID(ctx, id) (Session, error)
        +ListSessions(ctx) ([]Session, error)
        +CreateMessage(ctx, params) (Message, error)
        +ListMessagesBySession(ctx, sessionID) ([]Message, error)
        +CreateFile(ctx, params) (File, error)
        +GetFile(ctx, id) (File, error)
    }

    SessionService ..> Querier : uses
    MessageService ..> Querier : uses
    HistoryService ..> Querier : uses
    FileTrackerService ..> Querier : uses
```

## Class Diagram: Agent System

```mermaid
classDiagram
    class Coordinator {
        <<interface>>
        +Run(ctx, req) error
        +Cancel(sessionID)
        +CancelAll()
        +IsSessionBusy(sessionID) bool
        +Summarize(ctx, sessionID) error
        +Model() *Model
    }

    class SessionAgent {
        <<interface>>
        +Run(ctx, params) error
        +SetModels(large, small, local)
        +SetTools([]Tool)
        +SetSystemPrompt(string)
        +Cancel()
        +Summarize(ctx) error
    }

    class Model {
        +fantasy.LanguageModel LanguageModel
        +catwalk.Model CatwalkCfg
        +config.SelectedModel ModelCfg
        +bool FlatRate
    }

    class Tool {
        +string Name
        +string Description
        +*jsonschema.Schema ParameterSchema
        +Execute(ctx, params) (ToolResult, error)
    }

    class ContextBudget {
        +Check(messages, maxTokens) BudgetResult
        +EstimateTokens(content) int
        +SuggestTruncation(messages, target) []Message
    }

    class LoopDetector {
        +Record(toolCall)
        +IsLooping() bool
        +Reset()
    }

    class ToolRepair {
        +Repair(tool, input, error) (json.RawMessage, error)
        +CanRepair(tool, error) bool
    }

    Coordinator "1" --> "*" SessionAgent : manages
    Coordinator "1" --> "*" Model : holds
    SessionAgent "1" --> "1" Model : uses
    SessionAgent "1" --> "1" ContextBudget : uses
    SessionAgent "1" --> "1" LoopDetector : uses
    SessionAgent "1" --> "1" ToolRepair : uses
    SessionAgent "1" --> "*" Tool : uses
```

## Class Diagram: Configuration

```mermaid
classDiagram
    class Config {
        +string Schema
        +map[SelectedModelType]SelectedModel Models
        +map[SelectedModelType][]SelectedModel RecentModels
        +csync.Map[string, ProviderConfig] Providers
        +MCPs MCP
        +LSPs LSP
        +*Options Options
        +*Permissions Permissions
        +EnabledProviders() []ProviderConfig
        +GetModel(provider, model) *catwalk.Model
        +LargeModel() *catwalk.Model
        +SmallModel() *catwalk.Model
        +SetupAgents()
    }

    class ProviderConfig {
        +string ID
        +string Name
        +string BaseURL
        +catwalk.Type Type
        +string APIKey
        +string APIKeyTemplate
        +*oauth.Token OAuthToken
        +bool Disable
        +string SystemPromptPrefix
        +map[string]string ExtraHeaders
        +map[string]any ExtraBody
        +bool FlatRate
        +[]catwalk.Model Models
        +ToProvider() catwalk.Provider
    }

    class SelectedModel {
        +string Model
        +string Provider
        +string ReasoningEffort
        +bool Think
        +int64 MaxTokens
        +*float64 Temperature
        +*float64 TopP
        +*int64 TopK
        +map[string]any ProviderOptions
    }

    class Options {
        +[]string ContextPaths
        +[]string SkillsPaths
        +*TUIOptions TUI
        +bool Debug
        +bool DebugLSP
        +bool DisableAutoSummarize
        +string DataDirectory
        +[]string DisabledTools
        +bool DisableDefaultProviders
        +*Attribution Attribution
        +bool DisableMetrics
    }

    class MCPConfig {
        +string Command
        +map[string]string Env
        +[]string Args
        +MCPType Type
        +string URL
        +bool Disabled
        +[]string DisabledTools
        +int Timeout
        +map[string]string Headers
        +ResolvedEnv() []string
        +ResolvedArgs() []string
    }

    class LSPConfig {
        +bool Disabled
        +string Command
        +[]string Args
        +map[string]string Env
        +[]string FileTypes
        +[]string RootMarkers
        +map[string]any InitOptions
        +map[string]any Options
        +int Timeout
    }

    Config "1" --> "*" ProviderConfig : has
    Config "1" --> "*" SelectedModel : uses
    Config "1" --> "1" Options : has
    Config "1" --> "*" MCPConfig : has
    Config "1" --> "*" LSPConfig : has
    Config "1" --> "*" Agent : has
```

## Component Diagram: Module Dependencies

```mermaid
classDiagram
    class App {
        +SessionService sessions
        +MessageService messages
        +HistoryService history
        +PermissionService permissions
        +FileTrackerService fileTracker
        +Coordinator agentCoordinator
        +*lsp.Manager lspManager
        +*config.ConfigStore config
        +New() *App
        +RunNonInteractive()
        +Shutdown()
    }

    class Coordinator {
        +Run(ctx, req) error
        +Cancel(sessionID)
    }

    class ConfigStore {
        +Config() *Config
        +SetConfigField()
    }

    class Querier {
        <<interface>>
    }

    App --> Coordinator : creates
    App --> ConfigStore : reads
    App ..> Querier : creates (via sqlite)
    Coordinator --> Querier : uses
    Coordinator --> ConfigStore : reads
```

---

*Referenced from Chapters 10, 11, 12*
