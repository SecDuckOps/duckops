# DuckOps

<p align="center">
  <img src="https://img.shields.io/badge/version-1.0.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License">
  <img src="https://img.shields.io/github/stars/SecDuckOps/duckops" alt="Stars">
</p>

> A terminal-first AI DevSecOps AI for software development and security assessments.

![alt text](1781360224593278849.png)

## Overview

DuckOps is an AI-powered terminal assistant that combines intelligent conversation with powerful security assessment capabilities. Built in Go with a polished TUI, it provides seamless integration with multiple AI providers and a comprehensive suite of embedded security tools.

## Key Features

### AI Assistant

- **Multi-Provider Support**: OpenAI, Anthropic, Google, AWS Bedrock, Vercel AI, OpenRouter
- **Session Management**: Persistent conversations with automatic context summarization
- **Tool Integration**: 35+ built-in tools for file operations, shell execution, web search, and diagnostics
- **Loop Detection**: Prevents infinite agent loops

### Terminal UI

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss)
- Syntax highlighting via [Chroma](https://github.com/alecthomas/chroma)
- Markdown rendering with [Glamour](https://github.com/charmbracelet/glamour)

### Security Capabilities

| Category         | Coverage                                     |
| ---------------- | -------------------------------------------- |
| **Injection**    | SQL Injection, Command Injection, RCE, XXE   |
| **Web**          | XSS, CSRF, IDOR, API Security, Open Redirect |
| **System**       | Path Traversal, SSRF, File Inclusion         |
| **Logic**        | Business Logic, Race Conditions              |
| **Auth**         | JWT/OIDC, Authentication Bypass, BFLA        |
| **Crypto**       | Weak Algorithms, IV Reuse, Padding Oracle    |
| **Supply Chain** | Dependency Confusion, Typosquatting          |

**Embedded Tools**: Nuclei, SQLMap, Nmap, Subfinder, FFUF, Httpx, Katana, Naabu, Semgrep

### Integration

- MCP (Model Context Protocol) client/server support
- LSP Integration (Go, with extensible architecture)
- Hooks System for pre/post execution customization
- Skills Framework for custom capabilities
- **DuckOps Platform**: Agent registration, heartbeat, scan result sync, offline queue (`~/.duckops/queue/`)

## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/SecDuckOps/duckops/main/install.sh | bash

# Run interactive mode
duckops

# Run a task non-interactively
duckops run "explain this function"
```

## Installation

### Binary

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/SecDuckOps/duckops/main/install.sh | bash

# Or manually
mv duckops /usr/local/bin/
chmod +x /usr/local/bin/duckops
```

### From Source

```bash
go install github.com/SecDuckOps/duckops@latest
```

### Docker

```bash
docker build -t duckops .

docker run --rm -it \
  -v "$(pwd):/workspace" \
  -v "$HOME/.duckops:/root/.duckops" \
  -e OPENAI_API_KEY="${OPENAI_API_KEY}" \
  duckops
```

### Docker Security Sandbox + MCP

Build the scanner image and expose the HTTP tool server:

```bash
docker build -f containers/Dockerfile -t duckops:latest .

docker run --rm --name duckops-security \
  -p 48081:48081 \
  -e TOOL_SERVER_TOKEN=duckops-local-token \
  -v "$(pwd):/workspace/$(basename "$PWD")" \
  -w "/workspace/$(basename "$PWD")" \
  duckops:latest
```

Call the sandbox directly:

```bash
curl -s http://127.0.0.1:48081/scan \
  -H 'Authorization: Bearer duckops-local-token' \
  -H 'Content-Type: application/json' \
  -d "{\"agent_id\":\"manual\",\"path\":\"/workspace/$(basename "$PWD")\"}"
```

The project `duckops.json` also registers `duckops-security` as a stdio MCP server. DuckOps agents can call `duckops_security_scan` for the default scan preset or `duckops_execute_tool` for one bundled scanner.

## Configuration

### Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export GOOGLE_API_KEY="..."
export DUCKOPS_PROFILE=1   # Enable pprof on port 6060
```

### Config File

Edit `~/.duckops/duckops.json`:

```json
{
  "lsp": {
    "gopls": {
      "commands": ["gopls", "serve"],
      "settings": {
        "staticcheck": true
      }
    }
  }
}
```

## DuckOps Platform

The CLI integrates with the DuckOps Platform server for centralized scan result management.

```bash
# Authenticate with the platform
duckops login

# Sync scan results to the platform
duckops sync --results --tool semgrep --watch

# Or start the daemon for automatic sync
duckops daemon
```

The TUI requires platform login — your email is shown in the header bar when authenticated.

## Usage

```bash
# Interactive mode
duckops

# Non-interactive
duckops run "find security vulnerabilities"
duckops run "write tests for auth.go"

# Session management
duckops --continue              # Continue previous session
duckops --session abc123        # Specify session
duckops session list            # List sessions
duckops session resume <id>    # Resume session
```

### CLI Options

```
  --debug              Enable debug logging
  --cwd <path>        Set working directory
  --data-dir <path>   Custom data directory
  --duck              Skip permission prompts
  --no-color          Disable colored output
  --help, -h          Show help
  --version, -v       Print version
```

## Project Structure

```
duckops/
├── main.go                    # Entry point
├── internal/
│   ├── agent/                 # AI agent core
│   ├── cmd/                   # CLI commands
│   ├── config/                # Configuration
│   ├── db/                    # SQLite persistence
│   ├── session/               # Session management
│   ├── shell/                 # Shell integration
│   ├── ui/                    # Terminal UI
│   └── skills/                # Built-in capabilities
└── .agents/                   # Agent configuration
```

## Development

```bash
# Clone and build
git clone https://github.com/SecDuckOps/duckops.git
cd duckops
go mod download
go build -o duckops .

# Run tests
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) for details.

---

<p align="center">
  Built with <a href="https://github.com/SecDuckOps/duckops">DuckOps</a> · Powered by 🦆
</p>
