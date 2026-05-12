#!/bin/bash
# DuckOps Agent Entrypoint - Enhanced
# Provides robust startup with logging, health checks, and graceful shutdown

set -euo pipefail

# ─── Configuration ─────────────────────────────────────────────────
CAIDO_PORT="${CAIDO_PORT:-48080}"
CAIDO_LOG="${CAIDO_LOG:-/var/log/duckops/caido_startup.log}"
RUNTIME_HOME="${DUCKOPS_RUNTIME_HOME:-/var/lib/duckops/home}"
TOOL_SERVER_PORT="${TOOL_SERVER_PORT:-48081}"
TOOL_SERVER_TIMEOUT="${DUCKOPS_SANDBOX_EXECUTION_TIMEOUT:-120}"
TOOL_SERVER_HEALTH_ENDPOINT="http://localhost:${TOOL_SERVER_PORT}/health"
CAIDO_PID=""
CAIDO_STATUS="disabled"
LOG_DIR="/var/log/duckops"

# ─── Logging Setup ────────────────────────────────────────────────
mkdir -p "$LOG_DIR" "$RUNTIME_HOME" "${XDG_CACHE_HOME:-/var/lib/duckops/xdg-cache}" "${XDG_CONFIG_HOME:-/var/lib/duckops/xdg-config}" 2>/dev/null || true

# Check if log directory is writable
LOG_DIR_WRITABLE=true
if [ ! -w "$LOG_DIR" ]; then
    LOG_DIR_WRITABLE=false
fi

log() {
    local msg="[$(date +'%Y-%m-%d %H:%M:%S')] $*"
    echo "$msg" >&2
    if [ "$LOG_DIR_WRITABLE" = true ]; then
        echo "$msg" >> "$LOG_DIR/entrypoint.log" 2>/dev/null || true
    fi
}

# ─── Cleanup Handler ──────────────────────────────────────────────
cleanup() {
    local exit_code=$?
    log "Cleanup initiated (exit code: $exit_code)"
    
    if [ -n "$CAIDO_PID" ] && kill -0 "$CAIDO_PID" 2>/dev/null; then
        log "Stopping Caido (PID: $CAIDO_PID)"
        kill -TERM "$CAIDO_PID" 2>/dev/null || true
        wait "$CAIDO_PID" 2>/dev/null || true
    fi
    
    log "Cleanup complete"
    exit $exit_code
}

trap cleanup EXIT INT TERM

# ─── Caido Proxy Setup ─────────────────────────────────────────────
if command -v caido-cli >/dev/null 2>&1 && [ -f /app/certs/ca.p12 ]; then
    log "Starting Caido proxy on port ${CAIDO_PORT}..."
    
    mkdir -p "$(dirname "$CAIDO_LOG")" 2>/dev/null || true
    
    # Try to log to file; fall back to /dev/null if not writable
    if [ -w "$(dirname "$CAIDO_LOG")" ]; then
        caido-cli --listen 0.0.0.0:${CAIDO_PORT} \
                  --allow-guests --no-logging --no-open \
                  --import-ca-cert /app/certs/ca.p12 \
                  --import-ca-cert-pass "" > "$CAIDO_LOG" 2>&1 &
    else
        caido-cli --listen 0.0.0.0:${CAIDO_PORT} \
                  --allow-guests --no-logging --no-open \
                  --import-ca-cert /app/certs/ca.p12 \
                  --import-ca-cert-pass "" > /dev/null 2>&1 &
    fi
    CAIDO_PID=$!

    # Wait for Caido to be ready
    CAIDO_READY=false
    for i in $(seq 1 30); do
        if ! kill -0 "$CAIDO_PID" 2>/dev/null; then
            log "WARN: Caido process died during startup; continuing without proxy."
            if [ -f "$CAIDO_LOG" ]; then
                cat "$CAIDO_LOG" >&2 2>/dev/null || true
            fi
            CAIDO_PID=""
            break
        fi
        
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:${CAIDO_PORT}/graphql/ 2>/dev/null || echo "000")
        if [[ "$HTTP_CODE" =~ ^(200|400)$ ]]; then
            CAIDO_READY=true
            break
        fi
        
        log "Waiting for Caido... (attempt $i/30)"
        sleep 1
    done

    if [ "$CAIDO_READY" = true ]; then
        sleep 2

        # Get API token
        TOKEN=""
        for attempt in 1 2 3 4 5; do
            RESPONSE=$(curl -sL -X POST -H "Content-Type: application/json" \
                -d '{"query":"mutation LoginAsGuest { loginAsGuest { token { accessToken } } }"}' \
                http://localhost:${CAIDO_PORT}/graphql 2>/dev/null || echo "{}")
            TOKEN=$(echo "$RESPONSE" | jq -r '.data.loginAsGuest.token.accessToken // empty' 2>/dev/null || echo "")
            
            if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
                log "Successfully obtained Caido API token"
                break
            fi
            
            log "Token attempt $attempt failed, retrying..."
            sleep $((attempt * 2))
        done

        if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
            export CAIDO_API_TOKEN=$TOKEN

            # Create and select project
            CREATE_RESPONSE=$(curl -sL -X POST -H "Content-Type: application/json" \
                -H "Authorization: Bearer $TOKEN" \
                -d '{"query":"mutation CreateProject { createProject(input: {name: "duckops", temporary: true}) { project { id } } }"}' \
                http://localhost:${CAIDO_PORT}/graphql 2>/dev/null || echo "{}")
            PROJECT_ID=$(echo "$CREATE_RESPONSE" | jq -r '.data.createProject.project.id // empty' 2>/dev/null || echo "")

            if [ -n "$PROJECT_ID" ] && [ "$PROJECT_ID" != "null" ] && [ "$PROJECT_ID" != "" ]; then
                curl -sL -X POST -H "Content-Type: application/json" \
                    -H "Authorization: Bearer $TOKEN" \
                    -d "{\"query\":\"mutation SelectProject { selectProject(id: \\\"'$PROJECT_ID'\\\") { currentProject { project { id } } } \"}" \
                    http://localhost:${CAIDO_PORT}/graphql > /dev/null 2>&1 || true
            fi

            # Set proxy environment variables
            export http_proxy=http://127.0.0.1:${CAIDO_PORT}
            export https_proxy=http://127.0.0.1:${CAIDO_PORT}
            export HTTP_PROXY=http://127.0.0.1:${CAIDO_PORT}
            export HTTPS_PROXY=http://127.0.0.1:${CAIDO_PORT}
            export ALL_PROXY=http://127.0.0.1:${CAIDO_PORT}
            CAIDO_STATUS="ready on :${CAIDO_PORT}"
            log "Caido proxy configured and ready"
        else
            log "WARN: Failed to get Caido API token; continuing without proxy env."
        fi
    else
        log "WARN: Caido API did not become ready; continuing without proxy."
    fi
else
    log "INFO: caido-cli or CA bundle not found; starting sandbox without Caido proxy."
fi

# ─── SSL/TLS Configuration ────────────────────────────────────────
export REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt
export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# ─── Browser Trust Store (Commented - certutil not installed) ─────
# if command -v certutil >/dev/null 2>&1; then
#     mkdir -p "${RUNTIME_HOME}/.pki/nssdb"
#     certutil -N -d "sql:${RUNTIME_HOME}/.pki/nssdb" --empty-password 2>/dev/null || true
#     certutil -A -n "DuckOps Testing Root CA" -t "C,," -i /app/certs/ca.crt -d "sql:${RUNTIME_HOME}/.pki/nssdb" 2>/dev/null || true
# else
#     log "INFO: certutil is unavailable; browser trust-store bootstrap skipped."
# fi

# ─── Startup Complete ─────────────────────────────────────────────
log "DuckOps sandbox ready (Caido: ${CAIDO_STATUS}, tool server on :${TOOL_SERVER_PORT})"

# ─── Generate Tool Server Token ───────────────────────────────────
TOOL_SERVER_TOKEN_FILE="${RUNTIME_HOME}/.tool_server_token"
if [ -f "$TOOL_SERVER_TOKEN_FILE" ]; then
    TOOL_SERVER_TOKEN=$(cat "$TOOL_SERVER_TOKEN_FILE" 2>/dev/null || echo "")
else
    TOOL_SERVER_TOKEN=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
    mkdir -p "$RUNTIME_HOME"
    echo "$TOOL_SERVER_TOKEN" > "$TOOL_SERVER_TOKEN_FILE" 2>/dev/null || true
    chmod 600 "$TOOL_SERVER_TOKEN_FILE"
fi

# ─── Launch Tool Server ────────────────────────────────────────────
# Use stderr for debug logs - only suppress on explicit request
TOOL_SERVER_DEBUG="${TOOL_SERVER_DEBUG:-false}"
if [ "$TOOL_SERVER_DEBUG" != "true" ]; then
    trap - EXIT INT TERM
    exec 2>/dev/null
fi

exec /usr/local/bin/duckops tool-server \
    --port "$TOOL_SERVER_PORT" \
    --timeout "$TOOL_SERVER_TIMEOUT" \
    --token "$TOOL_SERVER_TOKEN"
