# File: docs/01-project-overview.md

# Chapter 1: Project Overview

## 1.1 Project Title

### Multiple Title Suggestions

| # | Title | Rationale |
|---|-------|-----------|
| 1 | **DuckOps: AI-Powered DevSecOps Terminal Assistant** | Emphasizes both development and security capabilities |
| 2 | **DuckOps: The Terminal-Native AI Coding & Security Platform** | Captures the dual coding + security assessment nature |
| 3 | **DuckOps: Multi-Provider AI Agent for Software Engineering** | Focuses on the AI agent orchestration aspect |
| 4 | **DuckOps: Context-Aware AI DevSecOps Framework** | Highlights the context management and compression engine |
| 5 | **Project Duck: Intelligent Terminal Agent for Modern DevOps** | Short, memorable brand name |

### Final Selected Title

**DuckOps: An AI-Powered DevSecOps Terminal Assistant with Multi-Provider Orchestration, Context-Aware Compression, and Embedded Security Assessment**

### Project Tagline

> *Your terminal-native AI teammate for development, security assessment, and everything in between.*

### Project Vision

To create a unified, terminal-first AI assistant that seamlessly integrates software development workflows with comprehensive security assessment capabilities — all within the developer's native environment. DuckOps envisions a world where developers never need to leave the terminal to get AI assistance, run security scans, manage context, or orchestrate complex multi-agent workflows. It aims to be the single pane of glass for AI-assisted software engineering, combining the power of multiple language models, embedded security tooling, and rich terminal UI into one cohesive experience.

## 1.2 Project Overview Diagram

```mermaid
graph TB
    subgraph "User Interface Layer"
        TUI["Terminal UI (Bubble Tea)"]
        CLI["CLI (Cobra Commands)"]
        API["REST API (Unix Socket)"]
    end

    subgraph "Application Layer"
        APP["App Controller"]
        COORD["Agent Coordinator"]
        BACKEND["Backend Service"]
    end

    subgraph "Core Services"
        SESSION["Session Service"]
        MESSAGE["Message Service"]
        HISTORY["File History Service"]
        CONFIG["Config Store"]
        PERM["Permission Service"]
        FILETRACK["File Tracker"]
    end

    subgraph "AI Provider Layer"
        PROVIDERS["Provider Orchestrator"]
        CATWALK["Catwalk (Provider Registry)"]
        FANTASY["Fantasy (LLM Abstraction)"]
    end

    subgraph "Tool Layer"
        TOOLS["35+ Built-in Tools"]
        MCP_TOOLS["MCP Tools"]
        SHELL["Shell Executor"]
    end

    subgraph "Infrastructure"
        DB[("SQLite Database")]
        DOCKER["Docker Sandbox"]
        LSP["LSP Client"]
        SEC["Security Tools<br/>(Nuclei, Nmap, SQLMap, etc.)"]
    end

    TUI --> APP
    CLI --> APP
    API --> BACKEND --> APP
    APP --> COORD
    COORD --> PROVIDERS
    COORD --> TOOLS
    COORD --> MCP_TOOLS
    COORD --> SHELL
    APP --> SESSION
    APP --> MESSAGE
    APP --> HISTORY
    APP --> CONFIG
    APP --> PERM
    APP --> FILETRACK
    PROVIDERS --> CATWALK
    PROVIDERS --> FANTASY
    FANTASY --> AI_PROVIDERS[("OpenAI / Anthropic /<br/>Google / Azure /<br/>OpenRouter / etc.")]
    SESSION --> DB
    MESSAGE --> DB
    HISTORY --> DB
    TOOLS --> SEC
    TOOLS --> DOCKER
    LSP --> APP

    style TUI fill:#e1f5fe
    style CLI fill:#e1f5fe
    style API fill:#e1f5fe
    style APP fill:#c8e6c9
    style COORD fill:#c8e6c9
    style BACKEND fill:#c8e6c9
    style SESSION fill:#fff9c4
    style MESSAGE fill:#fff9c4
    style HISTORY fill:#fff9c4
    style CONFIG fill:#fff9c4
    style PROVIDERS fill:#f3e5f5
    style CATWALK fill:#f3e5f5
    style FANTASY fill:#f3e5f5
    style TOOLS fill:#ffccbc
    style MCP_TOOLS fill:#ffccbc
    style SHELL fill:#ffccbc
    style DB fill:#e8eaf6
    style DOCKER fill:#e8eaf6
    style LSP fill:#e8eaf6
    style SEC fill:#e8eaf6
```

## 1.3 Key Differentiators

| Feature | DuckOps | GitHub Copilot | Claude Code | ChatGPT | Shell-GPT |
|---------|---------|---------------|-------------|---------|-----------|
| **Multi-Provider** | 25+ providers | OpenAI only | Anthropic only | OpenAI only | OpenAI/Anthropic |
| **Terminal Native** | Full TUI | IDE plugin | CLI only | Web/Desktop | CLI only |
| **Security Tools** | 15+ embedded | None | None | None | None |
| **Session Persistence** | SQLite-backed | Per-file | Per-session | Web history | None |
| **Context Compression** | Auto-summarization | None | Limited | Manual | None |
| **MCP Protocol** | Client + Server | None | None | None | None |
| **LSP Integration** | Full (gopls) | Limited | None | None | None |
| **Skills Framework** | Extensible | None | None | None (GPTs) | None |
| **Docker Sandbox** | Full support | None | None | None | None |
| **Cost Tracking** | Per-session/model | Subscription | Per-token | Subscription | Per-token |
| **Open Source** | ✅ MIT | ❌ Proprietary | ❌ Proprietary | ❌ Proprietary | ✅ |

## 1.4 Technical Stack Overview

```mermaid
mindmap
  root((DuckOps))
    Language
      Go 1.26+
      sqlc (codegen)
    UI Framework
      Bubble Tea v2
      Lipgloss v2
      Glamour v2
      Ultraviolet
      Chroma
    AI Providers
      catwalk (provider registry)
      fantasy (LLM abstraction)
      openai-go
      anthropic-sdk-go
      google-genai
    Database
      SQLite (modernc.org)
      ncruces/go-sqlite3
      goose (migrations)
    Shell
      mvdan.cc/sh v3
      mvdan.cc/sh/moreinterp
    Network
      modelcontextprotocol/go-sdk
      sourcegraph/jsonrpc2
      go-git/go-git v5
    Security
      nuclei
      nmap
      sqlmap
      ffuf
      subfinder
      httpx
      katana
      naabu
      semgrep
```

## 1.5 Module Count by Layer

| Layer | Packages | Files | Purpose |
|-------|----------|-------|---------|
| CLI | 1 (`cmd`) | 22 | Cobra command definitions |
| Application | 3 (`app`, `backend`, `workspace`) | 12 | Service wiring and lifecycle |
| UI | 15 subpackages (`ui/*`) | 60+ | Terminal user interface |
| Agent | 2 (`agent`, `agent/tools`) | 80+ | AI orchestration and tool implementations |
| Core Services | 8 (`session`, `message`, `config`, etc.) | 30+ | Business logic |
| Infrastructure | 6 (`db`, `server`, `client`, `lsp`, etc.) | 25+ | Database, networking, protocol support |
| Security | 2 (`security`, `graphx`) | 40+ | Security analysis engine |
| **Total** | **43 packages** | **~270 files** | |

## 1.6 Project Metadata

| Attribute | Value |
|-----------|-------|
| **Module Path** | `github.com/SecDuckOps/duckops` |
| **Go Version** | 1.26.3 |
| **Direct Dependencies** | 77 |
| **Indirect Dependencies** | ~100+ |
| **Build Time** | ~30s (clean build) |
| **Binary Size** | ~50MB (static) |
| **License** | MIT |
| **API Version** | v1 (Unix socket) |

---

**END OF CHAPTER 1**
