# File: docs/diagrams/erd.md

# Entity-Relationship Diagrams

## Main ERD

```mermaid
erDiagram
    sessions ||--o{ messages : "contains"
    sessions ||--o{ files : "tracks"
    sessions ||--o{ read_files : "reads"
    sessions ||--o| sessions : "branches_from"

    sessions {
        text id PK "UUID v4"
        text parent_session_id FK "Parent session for branching"
        text title "Auto-generated title"
        integer message_count "Total message count"
        integer prompt_tokens "Cumulative input tokens"
        integer completion_tokens "Cumulative output tokens"
        real cost "Total cost in USD"
        text summary_message_id "Latest summary"
        text todos "JSON todo list"
        integer updated_at "Unix timestamp"
        integer created_at "Unix timestamp"
    }

    messages {
        text id PK "UUID v4"
        text session_id FK "Parent session"
        text role "user|assistant|system|tool"
        text parts "JSON ContentPart array"
        text model "Model that generated this"
        text provider "Provider used"
        integer is_summary_message "Auto-summary flag"
        integer created_at "Unix timestamp"
        integer updated_at "Unix timestamp"
        integer finished_at "When streaming completed"
    }

    files {
        text id PK "UUID v4"
        text session_id FK "Creating session"
        text path "File path (relative)"
        text content "File contents"
        integer version "Auto-incrementing"
        integer created_at "Unix timestamp"
        integer updated_at "Unix timestamp"
    }

    read_files {
        text session_id PK,FK "Session that read"
        text path PK "File path"
        integer read_at "When it was read"
    }
```

## Physical Data Model

```mermaid
erDiagram
    sessions {
        text id
        text parent_session_id
        text title
        integer message_count
        integer prompt_tokens
        integer completion_tokens
        real cost
        text summary_message_id
        text todos
        integer updated_at
        integer created_at
        text idx_sessions_updated_at "INDEX DESC"
        text idx_sessions_created_at "INDEX DESC"
        text idx_sessions_parent "INDEX"
    }

    messages {
        text id
        text session_id
        text role
        text parts
        text model
        text provider
        integer is_summary_message
        integer created_at
        integer updated_at
        integer finished_at
        text idx_messages_session "INDEX (session_id,created_at)"
        text idx_messages_created_at "INDEX DESC"
        text idx_messages_model "INDEX"
    }

    files {
        text id
        text session_id
        text path
        text content
        integer version
        integer created_at
        integer updated_at
        text idx_files_session "INDEX"
        text idx_files_path "INDEX"
        text idx_files_session_path_version "UNIQUE INDEX"
    }

    read_files {
        text session_id
        text path
        integer read_at
        text pk_read_files "PRIMARY KEY (session_id, path)"
    }
```

## Data Flow Diagram

```mermaid
graph TB
    subgraph "Data Sources"
        USER_IN["User Input"]
        AI_RESP["AI Response"]
        TOOL_OUT["Tool Output"]
        CONFIG_FILES["Config Files"]
    end

    subgraph "Data Storage"
        DB_SESS["sessions table"]
        DB_MSG["messages table"]
        DB_FILES["files table"]
        DB_READ["read_files table"]
    end

    subgraph "Data Processing"
        MSG_MARSHAL["Marshal Content Parts<br/>Go structs -> JSON"]
        MSG_UNMARSHAL["Unmarshal Content Parts<br/>JSON -> Go structs"]
        CONTEXT_BUILD["Context Builder<br/>messages -> LLM format"]
        SUMMARY_GEN["Summary Generator<br/>Long context -> short"]
    end

    USER_IN --> MSG_MARSHAL --> DB_MSG
    AI_RESP --> MSG_MARSHAL --> DB_MSG
    TOOL_OUT --> MSG_MARSHAL --> DB_MSG
    DB_MSG --> MSG_UNMARSHAL --> CONTEXT_BUILD
    CONTEXT_BUILD --> SUMMARY_GEN
    SUMMARY_GEN --> DB_MSG
    CONFIG_FILES --> DB_SESS
    DB_SESS --> DB_MSG
    DB_SESS --> DB_FILES
    DB_SESS --> DB_READ
```

## Migration Timeline

```mermaid
gantt
    title Database Migration History
    dateFormat  YYYY-MM-DD
    axisFormat %Y-%m-%d

    section Schema
    Initial schema (sessions, messages, files, read_files)     :done, m1, 2024-09-01, 1d
    Add summary_message_id column                              :done, m2, after m1, 1d
    Add created_at indexes                                     :done, m3, after m2, 1d
    Add provider column to messages                            :done, m4, after m3, 1d
    Add is_summary_message column                              :done, m5, after m4, 1d
    Add todos column to sessions                               :done, m6, after m5, 1d
    Add read_files table                                       :done, m7, after m6, 1d
```

## Index Strategy Visualization

```mermaid
graph TB
    subgraph "Query Patterns"
        Q1["SELECT * FROM sessions<br/>ORDER BY updated_at DESC<br/>LIMIT 50"]
        Q2["SELECT * FROM messages<br/>WHERE session_id = ?<br/>ORDER BY created_at ASC"]
        Q3["SELECT * FROM files<br/>WHERE session_id = ?<br/>ORDER BY version"]
        Q4["SELECT SUM(prompt_tokens)<br/>FROM sessions"]
    end

    subgraph "Index Usage"
        I1["idx_sessions_updated_at<br/>(sessions.updated_at DESC)"]
        I2["idx_messages_session<br/>(messages.session_id, messages.created_at)"]
        I3["idx_files_session<br/>(files.session_id)"]
        I4["idx_sessions_all<br/>(full table scan OK for aggregates)"]
    end

    Q1 --> I1
    Q2 --> I2
    Q3 --> I3
    Q4 --> I4
```

---

*Referenced from Chapter 12: Database Design*
