# File: docs/27-references.md

# Chapter 27: References

## AI Model Providers

| Provider | URL | Documentation |
|----------|-----|---------------|
| OpenAI | https://platform.openai.com | https://platform.openai.com/docs |
| Anthropic | https://anthropic.com | https://docs.anthropic.com |
| Google Gemini | https://ai.google.dev | https://ai.google.dev/docs |
| OpenRouter | https://openrouter.ai | https://openrouter.ai/docs |
| Ollama | https://ollama.ai | https://github.com/ollama/ollama |
| Groq | https://groq.com | https://console.groq.com/docs |
| Together AI | https://together.ai | https://docs.together.ai |
| Mistral | https://mistral.ai | https://docs.mistral.ai |
| DeepSeek | https://deepseek.com | https://platform.deepseek.com |

## Core Libraries

| Library | Version | URL | Purpose |
|---------|---------|-----|---------|
| Go | 1.26+ | https://go.dev | Runtime |
| Bubble Tea | v2 | https://github.com/charmbracelet/bubbletea | TUI framework |
| Lipgloss | latest | https://github.com/charmbracelet/lipgloss | TUI styling |
| Glamour | latest | https://github.com/charmbracelet/glamour | Markdown rendering |
| BubbleZone | latest | https://github.com/lrstanley/bubblezone | Mouse support |
| SQLite (modernc) | latest | https://modernc.org/sqlite | Pure Go SQLite driver |
| sqlc | 1.27+ | https://sqlc.dev | Type-safe SQL codegen |
| goose | latest | https://github.com/pressly/goose | Database migrations |
| Catwalk | latest | https://github.com/anthropics/catwalk | Provider registry |
| Fantasy | latest | https://github.com/anthropics/fantasy | LLM SDK |
| mvdan.cc/sh | v3 | https://mvdan.cc/sh | Shell command parsing |
| gorilla/websocket | latest | https://github.com/gorilla/websocket | WebSocket client |
| rs/cors | latest | https://github.com/rs/cors | CORS middleware |
| go-isatty | latest | https://github.com/mattn/go-isatty | TTY detection |
| chroma | latest | https://github.com/alecthomas/chroma | Syntax highlighting |

## Security Standards & References

| Standard | URL | Description |
|----------|-----|-------------|
| OWASP Top 10 | https://owasp.org/Top10 | Web application security risks |
| OWASP API Top 10 | https://owasp.org/API-Security | API security risks |
| OWASP ASVS | https://owasp.org/ASVS | Application Security Verification Standard |
| STRIDE | https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats | Microsoft threat modeling |
| CWE | https://cwe.mitre.org | Common Weakness Enumeration |
| CVE | https://cve.mitre.org | Common Vulnerabilities and Exposures |
| NIST CSF | https://www.nist.gov/cyberframework | Cybersecurity Framework |
| MITRE ATT&CK | https://attack.mitre.org | Adversarial tactics and techniques |
| CIS Benchmarks | https://www.cisecurity.org/cis-benchmarks | Security configuration benchmarks |

## Protocol Documentation

| Protocol | URL | Description |
|----------|-----|-------------|
| MCP Spec | https://modelcontextprotocol.io | Model Context Protocol specification |
| LSP Spec | https://microsoft.github.io/language-server-protocol | Language Server Protocol specification |
| SSE (W3C) | https://html.spec.whatwg.org/multipage/server-sent-events.html | Server-Sent Events specification |
| JSON-RPC 2.0 | https://www.jsonrpc.org/specification | JSON-RPC specification |
| JSON Schema | https://json-schema.org | JSON Schema specification |

## DevOps Tools

| Tool | URL | Purpose |
|------|-----|---------|
| Docker | https://docker.com | Containerization |
| Docker Compose | https://docs.docker.com/compose | Multi-container orchestration |
| Kubernetes | https://kubernetes.io | Container orchestration |
| GitHub Actions | https://docs.github.com/actions | CI/CD |
| Nmap | https://nmap.org | Network scanning |
| Nuclei | https://nuclei.projectdiscovery.io | Vulnerability scanning |
| SQLMap | https://sqlmap.org | SQL injection testing |
| ffuf | https://github.com/ffuf/ffuf | Web fuzzing |
| Semgrep | https://semgrep.dev | Static analysis |
| Playwright | https://playwright.dev | Browser automation |

## Development Resources

| Resource | URL | Description |
|----------|-----|-------------|
| Go Documentation | https://go.dev/doc | Go language documentation |
| Go by Example | https://gobyexample.com | Go code examples |
| Effective Go | https://go.dev/doc/effective_go | Go best practices |
| Go Modules | https://go.dev/doc/modules | Module system reference |
| SQLite Documentation | https://sqlite.org/docs.html | SQLite reference |
| Bubble Tea Tutorial | https://github.com/charmbracelet/bubbletea/tree/master/tutorials | TUI tutorials |
| sqlc Documentation | https://docs.sqlc.dev | SQL codegen reference |

## Community

| Resource | URL |
|----------|-----|
| DuckOps GitHub | https://github.com/SecDuckOps/duckops |
| DuckOps Issues | https://github.com/SecDuckOps/duckops/internalissues |
| DuckOps Discord | https://discord.gg/duckops |
| DuckOps Documentation | https://docs.duckops.dev |

## Tools Referenced

| Tool | Package Path |
|------|-------------|
| Read | `internal/tools/read.go` |
| Edit | `internal/tools/edit.go` |
| Write | `internal/tools/write.go` |
| Create | `internal/tools/create.go` |
| Glob | `internal/tools/glob.go` |
| Grep | `internal/tools/grep.go` |
| Bash | `internal/tools/bash.go` |
| ListDir | `internal/tools/list_dir.go` |
| WebSearch | `internal/tools/web_search.go` |
| WebFetch | `internal/tools/web_fetch.go` |
| FileTracker | `internal/tools/file_tracker.go` |
| TodoWrite | `internal/tools/todowrite.go` |
| Nmap | `internal/tools/nmap.go` |
| Nuclei | `internal/tools/nuclei.go` |
| Subfinder | `internal/tools/subfinder.go` |
| HTTPx | `internal/tools/httpx.go` |
| FFUF | `internal/tools/ffuf.go` |
| Katana | `internal/tools/katana.go` |
| Naabu | `internal/tools/naabu.go` |
| SQLMap | `internal/tools/sqlmap.go` |
| Semgrep | `internal/tools/semgrep.go` |

## Services

| Service | Package Path |
|---------|-------------|
| SessionService | `internal/service/session.go` |
| MessageService | `internal/service/message.go` |
| HistoryService | `internal/service/history.go` |
| PermissionService | `internal/permission/permission.go` |
| FileTrackerService | `internal/service/file_tracker.go` |

## Agents & Coordination

| Component | Package Path |
|-----------|-------------|
| Coordinator | `internal/agent/coordinator.go` |
| SessionAgent | `internal/agent/session.go` |
| ContextBudget | `internal/agent/budget.go` |
| LoopDetector | `internal/agent/loop.go` |
| ToolRepair | `internal/agent/repair.go` |

---

*This concludes the DuckOps technical documentation. For the latest updates, visit https://github.com/SecDuckOps/duckops*
