# File: docs/23-faq.md

# Chapter 23: Frequently Asked Questions

## General

### What is DuckOps?

DuckOps is an AI-powered terminal application for software engineering tasks. It combines LLM capabilities with DevOps tools (security scanners, Docker, git, etc.) in a single, extensible TUI.

### How is DuckOps different from aider or goose?

DuckOps differentiates on these fronts:

| Feature | DuckOps | aider | goose |
|---------|---------|-------|-------|
| Provider support | 25+ providers (catwalk) | OpenAI/Anthropic only | Built-in only |
| Tool system | 35+ built-in + MCP + LSP | Limited | Limited |
| TUI | Full Bubble Tea TUI | CLI only | CLI only |
| Security tools | 15+ integrated scanners | None | None |
| Architecture | Hexagonal + Go | Monolith in Python | Monolith in Go |
| Database | SQLite with full history | None | None |

### Is DuckOps free and open source?

Yes. DuckOps is open source under the MIT license. You pay only for AI provider API usage (e.g., OpenAI, Anthropic).

### What platforms are supported?

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (via WSL2)

## Configuration

### Where is the config file?

```
~/.duckops/duckops.json      # Global user config
.duckops/duckops.json        # Project config
./duckops.json               # Workspace config
```

### How do I add a new AI provider?

```json
{
  "providers": [
    {
      "name": "my-provider",
      "type": "openai",
      "base_url": "https://api.myprovider.com/v1",
      "api_key": "${MY_PROVIDER_KEY}",
      "models": [
        { "name": "my-model", "max_input": 128000 }
      ]
    }
  ]
}
```

### Can I use local models?

Yes. DuckOps supports Ollama, LocalAI, and any OpenAI-compatible local endpoint:

```json
{
  "providers": [
    {
      "name": "ollama",
      "type": "openai",
      "base_url": "http://localhost:11434/v1",
      "models": [
        { "name": "llama3", "max_input": 8000 }
      ]
    }
  ],
  "models": {
    "local": { "provider": "ollama", "model": "llama3" }
  }
}
```

## Usage

### How do I use the TUI?

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Tab` | Focus chat / session list |
| `Ctrl+N` | New session |
| `Ctrl+D` | Delete session |
| `Ctrl+S` | Summarize session |
| `Ctrl+L` | Toggle autocomplete |
| `Ctrl+P` | Profile menu |
| `Ctrl+Q` | Quit |
| `/` | Command mode |

### What commands are available in chat mode?

| Command | Description |
|---------|-------------|
| `/summarize` | Summarize session |
| `/clear` | Clear chat |
| `/export` | Export session as Markdown |
| `/config` | Edit configuration |
| `/help` | Show help |
| `/model` | Switch model |
| `/tokens` | Show token usage |

### Can DuckOps access my file system?

Yes, but with permission controls. By default, DuckOps asks before:

- Reading files outside the workspace
- Editing files
- Executing shell commands
- Making network requests

Configure permissions in `duckops.json`:

```json
{
  "permissions": {
    "default": "ask",
    "tools": {
      "read": "allow",
      "edit": "allow",
      "bash": "ask",
      "web_fetch": "ask"
    }
  }
}
```

## Security

### Are my API keys safe?

- Keys are stored in `~/.duckops/duckops.json` with `chmod 600`
- Keys can be referenced via environment variables: `${OPENAI_API_KEY}`
- Keys are never sent to third parties (only to the AI provider directly)
- Keys are redacted in logs and session exports

### Can DuckOps be used in CI/CD?

Yes. DuckOps supports non-interactive mode:

```bash
# In GitHub Actions
duckops "Run security scan on src/" --no-interactive
```

And server mode for team access:

```bash
duckops server --api-key "sk-xxx" --bind 0.0.0.0:8080
```

### Is DuckOps safe in a shared environment?

- Each user has a separate SQLite database
- MCP servers are isolated per user
- Tool execution has configurable resource limits
- Audit logging tracks all tool executions

## Technical

### How does DuckOps handle context windows?

DuckOps tracks token usage per session. When usage exceeds 80% of the model's context window, it automatically summarizes old messages and replaces them with a concise summary. All original messages are preserved in the database.

### What database does DuckOps use?

SQLite via `modernc.org/sqlite` (pure Go, no CGO dependency for the driver). Schema management via `goose` migrations.

### Can I extend DuckOps with custom tools?

Yes, through MCP (Model Context Protocol). Any MCP-compatible server can be integrated by adding it to `duckops.json`:

```json
{
  "mcps": {
    "my-tool": {
      "command": "node",
      "args": ["my-mcp-server.js"]
    }
  }
}
```

### Does DuckOps support concurrent sessions?

Yes. Multiple TUI sessions and client connections are supported simultaneously. The daemon manages concurrency with mutex-based session isolation.

## Troubleshooting

### Why is my API key not working?

1. Check that the key is set: `echo $OPENAI_API_KEY`
2. Verify it's referenced correctly in config: `"api_key": "${OPENAI_API_KEY}"`
3. Test connectivity: `duckops diagnose providers`
4. Check provider status page for outages

### Why is DuckOps slow?

- Check network latency to your AI provider
- Reduce context window usage: `duckops summarize`
- Use a smaller/faster model for routine tasks
- Check system resource usage: `duckops health`

### Why am I getting rate limited?

- Wait and retry (DuckOps does this automatically)
- Switch to a fallback provider
- Upgrade your API tier for higher rate limits
- Configure a `flat_rate` provider to skip token counting

---

*Next: Chapter 24 - Troubleshooting*
