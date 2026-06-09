# File: docs/24-troubleshooting.md

# Chapter 24: Troubleshooting Guide

## 24.1 Common Errors

### Config Errors

```
Error: failed to load config: open ~/.duckops/duckops.json: no such file
```

**Solution**: Run `duckops init` to create the default config, or manually create `~/.duckops/duckops.json`.

---

```
Error: invalid config: provider "openai" not found in registry
```

**Solution**: Check that the provider type is valid. Supported types: `openai`, `anthropic`, `gemini`, `openrouter`, `ollama`, `groq`, `together`, `mistral`, `deepseek`.

---

```
Error: config validation failed: "models.large" requires both "provider" and "model"
```

**Solution**: Ensure each model entry has both `provider` and `model` fields:

```json
"models": {
  "large": { "provider": "openai", "model": "gpt-4o" }
}
```

### Connection Errors

```
Error: connect: connection refused
```

- Is the DuckOps daemon running? `duckops daemon`
- Is the socket path correct? Default: `/tmp/duckops.sock`
- Check with: `ls -la /tmp/duckops.sock`

---

```
Error: dial unix /tmp/duckops.sock: connect: no such file or directory
```

The daemon is not running or the socket was cleaned up. Start the daemon:

```bash
duckops daemon
```

---

```
Error: provider "openai" returned 401 Unauthorized
```

Your API key is invalid or missing. Check:

```bash
echo $OPENAI_API_KEY
duckops config get openai.api_key
```

### AI Provider Errors

```
Error: provider "openai" returned 429 Too Many Requests
```

You have exceeded your rate limit. DuckOps automatically retries with exponential backoff (2s, 4s, 8s). If it persists:

- Check your OpenAI usage dashboard
- Configure a fallback provider in `duckops.json`
- Add `"flat_rate": true` to the provider config

---

```
Error: provider "anthropic" returned 529 Overloaded
```

Anthropic is experiencing high demand. DuckOps will retry with backoff. Configure a fallback:

```json
{
  "models": {
    "large": { "provider": "openai", "model": "gpt-4o" }
  }
}
```

---

```
Error: context deadline exceeded (timeout)
```

The AI provider took too long to respond. Increase the timeout in provider config:

```json
{
  "providers": [{
    "name": "openai",
    "timeout": 120
  }]
}
```

### Database Errors

```
Error: database is locked
```

SQLite is busy with another connection. DuckOps auto-retries with a 5s timeout. If persistent:

```bash
# Check for stuck processes
lsof ~/.duckops/data/duckops.db

# Recover
duckops db check
duckops db vacuum
```

---

```
Error: no such table: messages
```

Database is corrupted or missing migrations. Run:

```bash
duckops db migrate
```

---

```
Error: migration failed: duplicate column name
```

Migration state is inconsistent. Reset and re-run:

```bash
duckops db reset  # Warning: deletes all data
duckops db migrate
```

### Tool Execution Errors

```
Error: tool "edit" failed: old_string not found
```

The string to replace was not found in the file. Use the `read` tool first to verify exact content.

---

```
Error: tool "bash" execution timed out (120s)
```

The command took too long. Either:
- Wait for it to complete by increasing timeout
- Cancel and try a more specific command

---

```
Error: permission denied: tool "bash" requires user approval
```

The permission policy blocked the tool. Either:
- Approve the tool in the TUI permission dialog
- Pre-approve in config: `"tools": {"bash": "allow"}`

### MCP Errors

```
Error: MCP server "my-server" not found
```

The command or path is wrong. Test the MCP server independently:

```bash
npx @modelcontextprotocol/github --help
```

---

```
Error: MCP server "my-server" crashed: signal: killed
```

The server ran out of memory or was forcibly terminated. Increase resource limits or reduce tool call complexity.

## 24.2 Diagnostic Commands

```bash
# Quick health check
duckops health

# Full diagnostic
duckops diagnose

# Test provider connectivity
duckops diagnose providers

# Check database integrity
duckops db check

# View debug logs
duckops --debug

# Tool call tracing
duckops --trace-tools

# MCP debugging
duckops --debug-mcp
```

## 24.3 Debug Mode

Enable debug logging to trace issues:

```bash
# Environment variable
export DUCKOPS_DEBUG=true
duckops

# Command flag
duckops --debug

# Increase log verbosity
duckops --log-level trace
```

### What Debug Mode Shows

```
[DEBUG] 2026-01-15T10:30:00Z config: loading config from /home/user/.duckops/duckops.json
[DEBUG] 2026-01-15T10:30:00Z config: provider "openai" enabled with model "gpt-4o"
[DEBUG] 2026-01-15T10:30:01Z db: connected to /home/user/.duckops/data/duckops.db
[DEBUG] 2026-01-15T10:30:01Z db: running migrations (current: 5, target: 5)
[DEBUG] 2026-01-15T10:30:01Z agent: creating session "sess_abc123"
[DEBUG] 2026-01-15T10:30:02Z provider: sending request to openai (model: gpt-4o, tokens: 1520)
[DEBUG] 2026-01-15T10:30:05Z provider: received response (tokens: 234, finish_reason: stop)
[DEBUG] 2026-01-15T10:30:05Z tools: executing tool "read" (file: main.go)
```

## 24.4 Log File Locations

| Platform | Log Location |
|----------|-------------|
| Linux | `~/.duckops/logs/` or `/var/log/duckops/` |
| macOS | `~/Library/Logs/duckops/` |
| Docker | stdout/stderr (configurable via Docker logging driver) |

## 24.5 Crash Recovery

### DuckOps Won't Start

```bash
# 1. Check config validity
duckops config validate

# 2. Reset config
mv ~/.duckops/duckops.json ~/.duckops/duckops.json.bak
duckops init

# 3. Check database
mv ~/.duckops/data/duckops.db ~/.duckops/data/duckops.db.bak
duckops
```

### Lost Session Data

```bash
# 1. Check if database exists
ls -la ~/.duckops/data/duckops.db

# 2. Verify backups
ls -la ~/.duckops/backups/

# 3. Restore from backup
cp ~/.duckops/backups/duckops-20260101.db ~/.duckops/data/duckops.db
```

### Freeze or Hang

```bash
# Graceful shutdown
pkill -INT duckops

# Force kill
pkill -9 duckops

# Remove stale socket
rm -f /tmp/duckops.sock

# Restart
duckops daemon
```

## 24.6 Known Issues

| Issue | Workaround | Status |
|-------|-----------|--------|
| TUI flickers on WSL2 | Use Windows Terminal, not cmd.exe | Open |
| SQLite busy on NFS | Store DB locally, not on NFS | By design |
| Large paste (100K+ chars) | Paste in chunks or use file input | Open |
| MCP server zombie processes | DuckOps cleans up on exit; kill manually if needed | Open |
| High memory with long sessions | Use `/summarize` periodically | By design |

## 24.7 Getting Help

```bash
# Built-in help
duckops --help
duckops help [command]

# Version info
duckops version

# GitHub Issues
# https://github.com/SecDuckOps/duckops/issues

# Community Discord
# https://discord.gg/duckops
```

---

*Next: Chapter 25 - Appendix*
