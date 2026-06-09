# File: docs/22-maintenance.md

# Chapter 22: Maintenance Guide

## 22.1 Overview

This chapter covers day-to-day maintenance tasks for DuckOps: database optimization, log management, configuration updates, and health monitoring.

## 22.2 Database Maintenance

### SQLite Optimization

```bash
# Analyze query planner
duckops db analyze

# Vacuum (reclaim disk space)
duckops db vacuum

# Integrity check
duckops db check

# Manual maintenance
sqlite3 ~/.duckops/data/duckops.db "
  PRAGMA analysis_limit=1000;
  PRAGMA optimize;
  VACUUM;
"
```

### Automatic Maintenance

DuckOps runs automated maintenance on startup:

```go
func (s *DatabaseService) AutoMaintenance(ctx context.Context) {
    // Run every 24 hours
    ticker := time.NewTicker(24 * time.Hour)
    for {
        select {
        case <-ticker.C:
            s.db.Exec("PRAGMA optimize")
            s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
        case <-ctx.Done():
            return
        }
    }
}
```

### Data Retention

| Data Type | Retention | Cleanup Policy |
|-----------|-----------|---------------|
| Messages | 90 days | Auto-delete after 90 days |
| Sessions | 90 days | Auto-delete with messages |
| File versions | 30 days | Keep only 10 versions per file |
| Audit logs | 1 year | Archive to JSON, then delete |
| Error logs | 30 days | Rotate daily |

### Manual Cleanup

```bash
# Delete sessions older than N days
duckops db cleanup --older-than 30d

# Delete specific session
duckops session delete <session-id>

# Clear all data
duckops db reset
```

## 22.3 Log Management

### Log Rotation

For file-based logging, use `logrotate`:

```bash
# /etc/logrotate.d/duckops
/var/log/duckops/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0640 duckops duckops
    sharedscripts
    postrotate
        kill -USR1 $(cat /var/run/duckops.pid) 2>/dev/null || true
    endscript
}
```

### Log Levels

| Level | Use Case | Output |
|-------|----------|--------|
| `error` | Production issues | Production |
| `warn` | Non-critical issues | Production |
| `info` | Normal operations | Production |
| `debug` | Development debugging | Development |
| `trace` | Function-level tracing | Development |

## 22.4 Configuration Updates

### Safe Reload

DuckOps supports hot-reloading configuration without restart:

```bash
# Edit config
vim ~/.duckops/duckops.json

# Trigger reload
duckops config reload
```

Or send `SIGHUP`:

```bash
kill -HUP $(pgrep duckops)
```

### Validating Config

```bash
duckops config validate
# Config valid: 3 providers, 2 MCP servers, 15 tools
```

## 22.5 Health Monitoring

### Health Check Endpoint (Server Mode)

```bash
curl http://localhost:8080/health
```

Response:

```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime_seconds": 9000,
  "active_sessions": 3,
  "total_sessions": 150,
  "total_messages": 5000,
  "database_size_mb": 45,
  "memory_usage_mb": 65,
  "cpu_usage_percent": 12
}
```

### Prometheus Alerts (Kubernetes)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: duckops-alerts
spec:
  groups:
  - name: duckops
    rules:
    - alert: DuckOpsDown
      expr: up{job="duckops"} == 0
      for: 5m
      annotations:
        summary: DuckOps is down

    - alert: DuckOpsHighMemory
      expr: process_resident_memory_bytes{job="duckops"} > 500e6
      for: 10m
      annotations:
        summary: DuckOps memory > 500MB

    - alert: DuckOpsHighErrorRate
      expr: rate(duckops_errors_total[5m]) > 0.1
      for: 5m
      annotations:
        summary: DuckOps error rate > 0.1/s
```

## 22.6 Performance Tuning

### SQLite Performance

| Setting | Default | Tuned | Description |
|---------|---------|-------|-------------|
| `journal_mode` | delete | `WAL` | Write-ahead logging for concurrent reads |
| `synchronous` | full | `NORMAL` | Balance safety vs. write speed |
| `cache_size` | -2000 | -64000 | 64MB page cache |
| `busy_timeout` | 0 | 5000 | Wait 5s before SQLITE_BUSY |
| `foreign_keys` | off | on | Enforce referential integrity |

### Memory

```bash
# Limit Go memory
export GOGC=100
export GOMEMLIMIT=2GiB

# Profile memory usage
duckops pprof heap
```

### Concurrency

| Setting | Default | Description |
|---------|---------|-------------|
| `max_concurrent_sessions` | 10 | Maximum TUI sessions |
| `max_concurrent_tools` | 5 | Parallel tool executions |
| `mcp_connection_limit` | 5 | MCP servers per user |

## 22.7 Troubleshooting Tools

### Diagnostic Commands

```bash
# Full system diagnostics
duckops diagnose

# Connectivity test
duckops diagnose providers

# Database integrity
duckops db check

# Configuration dump (redacted)
duckops config dump --redact-keys
```

### Debug Mode

```bash
# Run with debug logging
duckops --debug

# Enable tool call tracing
duckops --trace-tools

# Dump raw JSON-RPC for MCP
duckops --debug-mcp
```

## 22.8 Routine Schedule

| Frequency | Task | Command |
|-----------|------|---------|
| Daily | Health check | `duckops health` |
| Weekly | DB optimization | `duckops db vacuum` |
| Weekly | Log review | Check `/var/log/duckops/` |
| Monthly | DB integrity | `duckops db check` |
| Monthly | Backup | `duckops db backup` |
| Quarterly | Cleanup old sessions | `duckops db cleanup --older-than 90d` |
| Per-release | Upgrade | See upgrade procedure |

---

*Next: Chapter 23 - FAQ*
