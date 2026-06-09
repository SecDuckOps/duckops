# File: docs/26-glossary.md

# Chapter 26: Glossary

| Term | Definition |
|------|-----------|
| **Agent** | An AI-powered session handler that manages conversation context, tool execution, and provider communication for a single session. |
| **BFLA** | Broken Function Level Authorization — a security vulnerability where users can access unauthorized functions. |
| **BOLA** | Broken Object Level Authorization — a security vulnerability where users can access unauthorized objects. |
| **Bubble Tea** | A Go framework for building terminal user interfaces based on The Elm Architecture. |
| **Catwalk** | A provider registry library that supports 25+ AI model providers with a unified interface. |
| **CDP** | Chrome DevTools Protocol — protocol for controlling Chromium-based browsers. |
| **CLI** | Command-Line Interface — the terminal-based interface for DuckOps. |
| **Coordinator** | The central orchestrator that manages multiple session agents and models. |
| **CSP** | Content Security Policy — a browser security standard for preventing XSS attacks. |
| **CSRF** | Cross-Site Request Forgery — a web security vulnerability. |
| **Daemon** | Background server process that DuckOps clients connect to via Unix socket. |
| **Dockerization** | The process of packaging an application into a Docker container following best practices. |
| **DuckOps** | AI-powered development toolkit combining LLM capabilities with DevOps tools. |
| **Fantasy** | An LLM SDK providing a unified interface to multiple AI model providers. |
| **Flat Rate** | A provider configuration flag that skips token counting (useful for OpenRouter flat-rate models). |
| **Glamour** | A Go library for rendering Markdown in terminal applications. |
| **IDOR** | Insecure Direct Object Reference — a security vulnerability where users can access objects by ID. |
| **IPC** | Inter-Process Communication — DuckOps uses Unix domain sockets for communication between client and daemon. |
| **JSON Schema** | A vocabulary for annotating and validating JSON documents, used for tool parameter validation. |
| **JSON-RPC** | A remote procedure call protocol encoded in JSON, used by MCP. |
| **Large Model** | The primary AI model used for complex reasoning and code generation tasks. |
| **Lipgloss** | A Go library for styling terminal output with colors, padding, and alignment. |
| **Local Model** | An AI model running on the local machine (e.g., via Ollama), used for offline or sensitive tasks. |
| **Loop Detection** | A mechanism to detect and break infinite tool call loops. |
| **LSP** | Language Server Protocol — a protocol for providing language intelligence features (code completion, diagnostics, hover). |
| **MCP** | Model Context Protocol — an open protocol by Anthropic for connecting AI applications to external tools and data sources. |
| **Mermaid** | A Markdown-inspired diagramming and charting tool, used for DuckOps documentation. |
| **Migration** | A versioned database schema change managed by the `goose` library. |
| **Non-Interactive Mode** | DuckOps mode that runs a single prompt and exits, suitable for CI/CD pipelines. |
| **OAuth Token** | An access token used for OAuth-based authentication with AI providers. |
| **Permission Service** | The component that enforces access control for tool execution (allow, deny, or ask). |
| **Provider** | An AI model service provider such as OpenAI, Anthropic, Google, or Ollama. |
| **Provider Router** | The component that selects and fails over between AI providers. |
| **RLS** | Row-Level Security — a database security feature for restricting row access. |
| **SessionAgent** | The per-session AI conversation handler. |
| **Small Model** | A lightweight AI model used for summarization and quick tasks. |
| **SQLite** | Embedded SQL database engine used by DuckOps for persistent storage. |
| **sqlc** | A Go code generation tool for type-safe SQL queries. |
| **SSE** | Server-Sent Events — a server push protocol used by DuckOps for streaming AI responses. |
| **SSH** | Secure Shell — protocol for secure remote access, used by DuckOps for git operations. |
| **SSRF** | Server-Side Request Forgery — a web security vulnerability. |
| **STRIDE** | A threat modeling methodology (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege). |
| **TUI** | Terminal User Interface — the interactive Bubble Tea interface for DuckOps. |
| **Todo** | An inline task item extracted from conversation and tracked in the session. |
| **Tool** | A capability exposed to the AI agent (read file, edit file, bash, web search, etc.). |
| **Tool Repair** | Automatic correction of malformed tool call inputs using a secondary AI model. |
| **Unix Socket** | A local inter-process communication endpoint (`/tmp/duckops.sock`). |
| **WAL Mode** | Write-Ahead Logging — a SQLite journal mode that improves concurrent read performance. |
| **XDG** | Cross-Desktop Group standards for config and data directory locations. |
| **XSS** | Cross-Site Scripting — a web security vulnerability. |
| **XXE** | XML External Entity — a security vulnerability in XML parsers. |

---

*Next: Chapter 27 - References*
