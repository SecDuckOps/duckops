# Agent ↔ Server Integration Architecture

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     CUSTOMER MACHINE (Agent)                     │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────┐   │
│  │  Config   │  │   Auth   │  │    API   │  │   Local       │   │
│  │  Manager  │  │  Manager │  │   Client │  │   Storage     │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  │  (SQLite/     │   │
│       │              │              │        │   BadgerDB)  │   │
│       ▼              ▼              ▼        │               │   │
│  ┌─────────────────────────────────────┐    └───────────────┘   │
│  │           Agent Core                │                        │
│  │  ┌────────┐ ┌────────┐ ┌─────────┐  │                        │
│  │  │Heartbeat│ │  Scan  │ │ Result  │  │     INTERNET          │
│  │  │Service │ │ Runner │ │Uploader │  │        │              │
│  │  └────────┘ └────────┘ └─────────┘  │        │              │
│  │  ┌────────┐ ┌────────┐              │   TLS/ │ HTTPS        │
│  │  │  Job   │ │  Event │              │    mTLS │              │
│  │  │Scheduler│ │ Logger │              │        │              │
│  │  └────────┘ └────────┘              │        │              │
│  └─────────────────────────────────────┘        │              │
└──────────────────────────────────────────────────┘              │
                                                                    │
═══════════════════════════════════════════════════════════════════════
                                                                    │
┌──────────────────────────────────────────────────────────────────┐│
│                     DUCKOPS SERVER                                ││
│                                                                  ││
│  ┌─────────────────────────────────────────────────────────┐    ││
│  │               API Gateway (TLS/mTLS)                    │    ││
│  │  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐ │    ││
│  │  │ Rate Limiter │  │ Auth (PAT)  │  │ Request Log   │ │    ││
│  │  └─────────────┘  └──────────────┘  └───────────────┘ │    ││
│  └─────────────────────────────────────────────────────────┘    ││
│                                                                  ││
│  ┌─────────────────────────────────────────────────────────┐    ││
│  │                    Route Handlers                        │    ││
│  │  /agents/*    /scans/*    /findings/*    /vulns/*       │    ││
│  │  /heartbeats/* /projects/* /pipelines/*  /insights/*   │    ││
│  └─────────────────────────────────────────────────────────┘    ││
│                                                                  ││
│  ┌─────────────────────────────────────────────────────────┐    ││
│  │                    Service Layer                         │    ││
│  │  AgentMgmt  ScanService  FindingService  VulnService    │    ││
│  │  Heartbeat  JobService   InsightService  ProjectSvc     │    ││
│  └─────────────────────────────────────────────────────────┘    ││
│                                                                  ││
│  ┌─────────────────────────────────────────────────────────┐    ││
│  │                    Database (PostgreSQL)                 │    ││
│  │  agents | sessions | heartbeats | scan_jobs |           │    ││
│  │  scan_results | findings | vulnerabilities | insights   │    ││
│  └─────────────────────────────────────────────────────────┘    ││
│                                                                  ││
│  ┌──────────────────────┐  ┌──────────────────────────────┐    ││
│  │    Background Jobs    │  │      AI Insight Engine       │    ││
│  │  (stale agent cleanup,│  │  (pattern analysis, trend    │    ││
│  │   heartbeat monitor,  │  │   detection, prioritization) │    ││
│  │   job scheduler)      │  │                              │    ││
│  └──────────────────────┘  └──────────────────────────────┘    ││
└──────────────────────────────────────────────────────────────────┘
```

---

## 2. Agent Lifecycle

```
                         ┌──────────────┐
                         │ Agent Startup │
                         └──────┬───────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   Load Config         │
                    │   Read PAT, AgentID,  │
                    │   Server URL, Scan    │
                    │   Definitions, etc.   │
                    └───────────┬───────────┘
                                │
                                ▼
                    ┌───────────────────────┐
               ┌───│   Persisted Agent ID   │◄────┐
               │   │   Exists?              │     │
               │   └───────────┬───────────┘     │
               │          Yes │              No  │
               │               ▼                 │
               │   ┌───────────────────────┐     │
               │   │  Authenticate via     │     │
               │   │  Existing PAT + ID    │     │
               │   └───────────┬───────────┘     │
               │               │                 │
               │               ▼                 │
               │   ┌───────────────────────┐     │
               │   │ POST /agents/connect  │     │
               │   │ Server validates PAT  │     │
               │   │ Returns session token │     │
               │   └───────────┬───────────┘     │
               │               │                 │
               │        ┌──────┴──────┐         │
               │        ▼             ▼          │
               │   Success        401/403        │
               │        │             │          │
               │        ▼             ▼          │
               │   ┌────────┐  ┌───────────┐    │
               │   │Online  │  │ Re-register│────┘
               │   │State   │  │ /register  │
               │   └────┬───┘  └───────────┘
               │        │
               │        ▼
               │   ┌───────────────────────┐
               │   │   Heartbeat Loop      │
               │   │   Every N seconds     │
               │   │   POST /heartbeat     │
               │   │   Includes: status,   │
               │   │   load, scan progress │
               │   └───────────┬───────────┘
               │               │
               │               ▼
               │   ┌───────────────────────┐
               │   │   Job Poll / Push     │
               │   │   GET /jobs (poll)    │
               │   │   or Server-Sent      │
               │   │   Events (push)       │
               │   └───────────┬───────────┘
               │               │
               │               ▼
               │   ┌───────────────────────┐
               │   │   Execute Scan        │
               │   │   Run tools: semgrep, │
               │   │   nuclei, nmap, etc.  │
               │   │   Collect findings    │
               │   └───────────┬───────────┘
               │               │
               │               ▼
               │   ┌───────────────────────┐
               │   │   Upload Results      │
               │   │   POST /scans/{id}/   │
               │   │     results            │
               │   │   POST /findings       │
               │   │   POST /vulnerabilities│
               │   └───────────┬───────────┘
               │               │
               │               ▼
               │   ┌───────────────────────┐
               │   │   Error Recovery      │
               │   │   Exponential backoff │
               │   │   Queue for retry     │
               │   │   Max retries → alert │
               │   └───────────────────────┘
               │
               │   ┌───────────────────────┐
               │   │   Disconnect/Sleep    │
               │   │   Graceful shutdown:  │
               │   │   Final heartbeat     │
               │   │   POST /disconnect    │
               │   │   Persist state       │
               │   └───────────────────────┘
               │
               ▼
          Agent Stopped
```

---

## 3. Sequence Diagrams

### 3.1 First Registration

```
Agent                         DuckOps Server                    Database
  │                               │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ On first start:         │  │                               │
  │  │ - No stored AgentID     │  │                               │
  │  │ - No stored PAT         │  │                               │
  │  │ - Read PAT from env/    │  │                               │
  │  │   config file           │  │                               │
  │  └──────────┬──────────────┘  │                               │
  │             │                  │                               │
  │  POST /api/v1/agents/register │                               │
  │  ┌──────────────────────────►│                               │
  │  │ {                         │                               │
  │  │   hostname: "cust-01",    │                               │
  │  │   platform: "linux",      │                               │
  │  │   version: "1.0.0",       │                               │
  │  │   capabilities: [         │                               │
  │  │     "semgrep","nuclei",   │                               │
  │  │     "nmap","trivy"        │                               │
  │  │   ],                      │                               │
  │  │   public_key: "..."       │                               │
  │  │ }                         │                               │
  │  │                           │  INSERT INTO agents (...)     │
  │  │                           │  ───────────────────────────► │
  │  │                           │  ◄── agent record ───────────│
  │  │  ◄── 201 Created ────────│                               │
  │  │  {                       │                               │
  │  │   data: {                │                               │
  │  │    agent: {              │                               │
  │  │     uuid: "ag_abc123",   │                               │
  │  │     token: "pat_xyz...", │                               │
  │  │     expires_at: "..."    │                               │
  │  │    }                     │                               │
  │  │   }                      │                               │
  │  │  }                       │                               │
  │  │                           │                               │
  │  │  ┌─────────────────────┐  │                               │
  │  │  │ Persist to local    │  │                               │
  │  │  │ storage:            │  │                               │
  │  │  │ - AgentID           │  │                               │
  │  │  │ - PAT (encrypted)   │  │                               │
  │  │  │ - Server URL        │  │                               │
  │  │  └─────────────────────┘  │                               │
  │  │                           │                               │
  │  │  POST /api/v1/agents/connect                              │
  │  │  ┌──────────────────────────►│                            │
  │  │  │ { agent_id: "ag_abc123", }│   INSERT INTO sessions    │
  │  │  │   Bearer: pat_xyz...      │  ───────────────────────► │
  │  │  │                           │  ◄── session record ─────│
  │  │  ◄── 200 OK ────────────────│                            │
  │  │  { "session": {             │                            │
  │  │    "token": "sess_jkl...",  │                            │
  │  │    "expires_at": "...",     │                            │
  │  │    "heartbeat_interval": 30 │                            │
  │  │  }}                         │                            │
  │  │                           │                               │
  │  │  ┌─────────────────────┐  │                               │
  │  │  │ Store session token │  │                               │
  │  │  │ Start heartbeat     │  │                               │
  │  │  └─────────────────────┘  │                               │
  │  │                           │                               │
```

### 3.2 Normal Operation

```
Agent                         DuckOps Server                    Database
  │                               │                               │
  │  ┌──────────┬──────────┐     │                               │
  │  │ Periodically:       │     │                               │
  │  │ Every 30s: Heartbeat│     │                               │
  │  │ Every 60s: Job Poll │     │                               │
  │  └─────────────────────┘     │                               │
  │                               │                               │
  │  POST /api/v1/agents/heartbeat                                │
  │  ┌────────────────────────────►│                              │
  │  │ {                          │   UPDATE agents               │
  │  │   session_token: "...",    │     SET last_seen=NOW(),      │
  │  │   status: "idle",          │     status='online'          │
  │  │   cpu_load: 0.45,          │   INSERT INTO heartbeats     │
  │  │   mem_usage: 2048,         │  ───────────────────────────►│
  │  │   scan_in_progress: null,  │  ◄── done ───────────────────│
  │  │   scanned_projects: [{...}]│                               │
  │  │ }                          │                               │
  │  ◄── 200 OK ─────────────────│                               │
  │  { jobs: [], config: {...} }  │                               │
  │                               │                               │
  │  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─                                  │
  │                               │                               │
  │  GET /api/v1/jobs             │                               │
  │  ┌────────────────────────────►│                              │
  │  │ Authorization: Bearer ... │  SELECT * FROM scan_jobs      │
  │  │                           │    WHERE agent_id=?           │
  │  │                           │    AND status='pending'       │
  │  │                           │  ───────────────────────────► │
  │  │                           │  ◄── jobs ───────────────────│
  │  ◄── 200 OK ─────────────────│                               │
  │  { jobs: [{                  │                               │
  │    uuid: "job_001",          │                               │
  │    type: "scan",             │                               │
  │    target: { workspace: "..", project: ".." },               │
  │    scan_type: "semgrep",     │                               │
  │    priority: "high",         │                               │
  │    created_at: "..."         │                               │
  │  }]}                         │                               │
  │                               │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Execute Scan:           │  │                               │
  │  │ 1. Update status:       │  │                               │
  │  │    "running"            │  │                               │
  │  │ 2. Run security tool    │  │                               │
  │  │ 3. Collect findings     │  │                               │
  │  │ 4. Upload results       │  │                               │
  │  └─────────────────────────┘  │                               │
  │                               │                               │
  │  PATCH /api/v1/scans/{uuid}                                  │
  │  ┌────────────────────────────►│   UPDATE scan_jobs           │
  │  │ { status: "running",       │     SET status='running'     │
  │  │   current_stage:           │  ───────────────────────────►│
  │  │   "dependency_scan" }      │                               │
  │  ◄── 200 OK ─────────────────│                               │
  │                               │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Run semgrep/nuclei/etc  │  │                               │
  │  └────────┬────────────────┘  │                               │
  │           ▼                   │                               │
  │      Scan Completed           │                               │
  │           │                   │                               │
  │  POST /api/v1/scans/{uuid}/results                            │
  │  ┌────────────────────────────►│                              │
  │  │ { findings: [...],         │  INSERT INTO findings (batch)│
  │  │   vulnerabilities: [...],  │  INSERT INTO vulns (batch)   │
  │  │   summary: {               │  UPDATE scan_jobs            │
  │  │     total_findings: 12,    │    SET status='completed'    │
  │  │     critical: 2,           │  ───────────────────────────►│
  │  │     high: 4, medium: 3,   │                               │
  │  │     low: 3, duration: 45s │                               │
  │  │   }                        │                               │
  │  }                            │                               │
  │  ◄── 201 Created ────────────│                               │
  │                               │                               │
```

### 3.3 Lost Connection Recovery

```
Agent                         DuckOps Server                    Database
  │                               │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Normal heartbeat fails  │  │                               │
  │  │ - Timeout               │  │                               │
  │  │ - Network partition     │  │                               │
  │  │ - Server unreachable    │  │                               │
  │  └────────┬────────────────┘  │                               │
  │           ▼                   │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Retry with backoff:     │  │                               │
  │  │ Attempt 1: +1s          │  │                               │
  │  │ Attempt 2: +2s          │  │                               │
  │  │ Attempt 3: +4s          │  │                               │
  │  │ Attempt 4: +8s          │  │                               │
  │  │ Attempt 5: +15s          │  │                               │
  │  │ ... up to 5min max      │  │                               │
  │  │ Max retries: 10         │  │                               │
  │  └────────┬────────────────┘  │                               │
  │           │                   │                               │
  │           ▼                   │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Enter offline mode:     │  │                               │
  │  │ - Queue scans locally   │  │                               │
  │  │ - Store results in      │  │                               │
  │  │   local SQLite/         │  │                               │
  │  │   BadgerDB              │  │                               │
  │  │ - Continue scanning     │  │                               │
  │  │   cached jobs           │  │                               │
  │  │ - Mark all as pending   │  │                               │
  │  │   sync                  │  │                               │
  │  └────────┬────────────────┘  │                               │
  │           │                   │                               │
  │    Connection Restored        │                               │
  │           │                   │                               │
  │  POST /api/v1/agents/heartbeat                                │
  │  ┌────────────────────────────►│                              │
  │  │ Authorization: Bearer ...  │  Check session validity      │
  │  │                             │  ───────────────────────────►│
  │  ◄── 401 Unauthorized ────────│                              │
  │       (session expired)       │                               │
  │           │                   │                               │
  │  POST /api/v1/agents/connect                                  │
  │  ┌────────────────────────────►│                              │
  │  │ { agent_id: "ag_abc123",  }│  Validate PAT                │
  │  │   Bearer: pat_xyz...      }│  ───────────────────────────►│
  │  ◄── 200 OK ─────────────────│                              │
  │  { session: { ... } }        │                               │
  │           │                   │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Sync queued results:    │  │                               │
  │  │ For each offline scan:  │  │                               │
  │  │ POST /scans (create)    │  │                               │
  │  │ POST /scans/{id}/results│  │                               │
  │  └────────┬────────────────┘  │                               │
  │           │                   │                               │
  │           ▼                   │                               │
  │  ┌─────────────────────────┐  │                               │
  │  │ Resume normal heartbeat │  │                               │
  │  │ Resume job polling      │  │                               │
  │  └─────────────────────────┘  │                               │
```

---

## 4. Local Agent Components

### 4.1 Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                  Agent Core Process                          │
│                                                              │
│  ┌─────────────┐   ┌─────────────┐   ┌──────────────────┐  │
│  │ Config      │   │ Auth        │   │ API Client       │  │
│  │ Manager     │   │ Manager     │   │ (HTTP/REST)      │  │
│  │             │   │             │   │                  │  │
│  │ - Load YAML │   │ - PAT load  │   │ - Retry logic    │  │
│  │ - Validate  │   │ - Encrypt   │   │ - Backoff        │  │
│  │ - Watch     │   │ - Rotate    │   │ - Auth header    │  │
│  │ - Reload    │   │ - Validate  │   │ - Timeout        │  │
│  └──────┬──────┘   └──────┬──────┘   └────────┬─────────┘  │
│         │                 │                    │            │
│         └────────┬────────┴────────┬───────────┘            │
│                  │                 │                         │
│         ┌────────▼─────────────────▼──────────┐             │
│         │          Agent Controller            │             │
│         │  ┌──────────────────────────────┐   │             │
│         │  │ Lifecycle Manager            │   │             │
│         │  │ - Startup sequence           │   │             │
│         │  │ - State machine              │   │             │
│         │  │ - Shutdown                   │   │             │
│         │  └──────────────────────────────┘   │             │
│         │  ┌──────────────────────────────┐   │             │
│         │  │ Connection Manager           │   │             │
│         │  │ - Online/Offline             │   │             │
│         │  │ - Reconnection               │   │             │
│         │  │ - Heartbeat                  │   │             │
│         │  └──────────────────────────────┘   │             │
│         │  ┌──────────────────────────────┐   │             │
│         │  │ Job Dispatcher               │   │             │
│         │  │ - Poll/push receive          │   │             │
│         │  │ - Priority queue             │   │             │
│         │  │ - Concurrency control        │   │             │
│         │  └──────────────────────────────┘   │             │
│         │  ┌──────────────────────────────┐   │             │
│         │  │ Sync Manager                 │   │             │
│         │  │ - Offline queue              │   │             │
│         │  │ - Batch upload               │   │             │
│         │  │ - Conflict resolution        │   │             │
│         │  └──────────────────────────────┘   │             │
│         └─────────────────────────────────────┘             │
│                                                              │
│  ┌────────────┐  ┌────────────┐  ┌───────────────────────┐  │
│  │ Heartbeat  │  │ Scan       │  │ Result Uploader       │  │
│  │ Service    │  │ Runner     │  │                       │  │
│  │            │  │            │  │ - Findings batch      │  │
│  │ - Timer    │  │ - Semgrep  │  │ - Vulns batch         │  │
│  │ - Payload  │  │ - Nuclei   │  │ - Retry on fail       │  │
│  │ - Metrics  │  │ - Trivy    │  │ - Compression         │  │
│  │ - Retry    │  │ - Nmap     │  │ - Chunking (large)    │  │
│  └─────┬──────┘  │ - Katana   │  └──────────┬────────────┘  │
│        │         │ - Custom   │             │               │
│        │         └────────────┘             │               │
│        │                                      │               │
│  ┌─────┴──────────────────────────────────────┴──────┐       │
│  │              Job Scheduler                         │       │
│  │  ┌──────────┐ ┌────────┐ ┌────────┐ ┌─────────┐  │       │
│  │  │Schedule  │ │Priority│ │Timeout │ │Resource │  │       │
│  │  │Management│ │Queue   │ │Control │ │Limiter  │  │       │
│  │  └──────────┘ └────────┘ └────────┘ └─────────┘  │       │
│  └────────────────────────────────────────────────────┘       │
│                                                              │
│  ┌────────────────────────────────────────────────────┐      │
│  │              Local Storage (SQLite/Badger)         │      │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────┐ │      │
│  │  │ Agent    │ │ Scan     │ │ Offline  │ │Event │ │      │
│  │  │ Identity │ │ Results  │ │ Queue    │ │Log   │ │      │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────┘ │      │
│  └────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 Component Definitions

| Component | Responsibility | Key Methods |
|-----------|----------------|-------------|
| **ConfigManager** | Load, validate, watch, and reload configuration from file/env | `Load(path)`, `Get(key)`, `Watch()`, `Reload()` |
| **AuthManager** | Store PAT encrypted, manage rotation, validate tokens | `Init(pat)`, `GetPAT()`, `Rotate()`, `IsValid()`, `Encrypt()`, `Decrypt()` |
| **APIClient** | Low-level HTTP client with retry, backoff, auth injection | `Do(method, path, body)`, `RetryPolicy()`, `SetToken()` |
| **HeartbeatService** | Send periodic heartbeats, measure system metrics | `Start(interval)`, `Stop()`, `Send()`, `collectMetrics()` |
| **ScanRunner** | Execute security scans via configured tools | `Run(spec)`, `Cancel()`, `parseOutput()`, `validateResults()` |
| **ResultUploader** | Upload findings/vulns to server with batching & retry | `UploadResults(scanID, results)`, `UploadFindings(batch)`, `UploadVulns(batch)` |
| **JobScheduler** | Poll/push job queue, manage concurrency, timeouts | `Poll()`, `Dispatch(job)`, `ack(jobID)`, `nack(jobID, reason)` |
| **LocalStorage** | Persist agent identity, offline queue, scan cache | `StoreAgentID(id)`, `GetAgentID()`, `QueueOffline(job)`, `FlushQueue()` |
| **AgentController** | Orchestrate all components, manage state machine | `Start()`, `Shutdown()`, `transition(state)`, `handleError(err)` |

---

## 5. Server-Side Components

| Component | Responsibility | Implementation |
|-----------|----------------|----------------|
| **AuthMiddleware** | Validate Bearer PAT on every request | Extract token, lookup `agent_sessions`, validate expiry |
| **AgentHandler** | `POST /register`, `POST /connect`, `POST /disconnect`, `GET /{id}` | Route handlers |
| **HeartbeatHandler** | `POST /heartbeat` - receive metrics, update `last_seen` | Route handler |
| **AgentRegistry** | CRUD for agents, generate PATs, manage agent lifecycle | Service layer |
| **SessionManager** | Create/validate/invalidate agent sessions, rotate tokens | Service layer |
| **ScanService** | Create/update/query scan jobs, manage lifecycle | Service layer |
| **FindingService** | Batch insert findings, deduplicate, query, export | Service layer |
| **VulnerabilityService** | Batch insert vulns, CVE correlation, query | Service layer |
| **JobService** | Job assignment, priority queues, scheduling | Service layer + background goroutine |
| **ProjectService** | Project CRUD for scan targets | Service layer |
| **PipelineService** | CI/CD pipeline integration | Service layer |
| **InsightService** | AI-driven pattern analysis, trend detection | Async engine (LLM-based) |
| **NotificationService** | Webhooks, alerts for critical findings | Background worker |
| **BackgroundJanitor** | Periodic cleanup: stale agents, expired sessions, orphaned scans | Cron-like goroutine |

---

## 6. API Contracts

### 6.1 Agent Registration

```http
POST /api/v1/agents/register
Content-Type: application/json

Request:
{
  "hostname": "customer-dev-01",
  "platform": "linux/amd64",
  "version": "1.0.0",
  "capabilities": ["semgrep", "nuclei", "trivy", "nmap", "katana"],
  "public_key": "ssh-ed25519 AAAAC3...",
  "labels": {
    "env": "production",
    "team": "security",
    "region": "us-east-1"
  },
  "max_concurrent_scans": 2,
  "local_timezone": "America/New_York"
}

Response: 201 Created
{
  "data": {
    "agent": {
      "uuid": "ag_01J2XYZ...",
      "token": "duckops_pat_prod_abc123def456...",
      "expires_at": "2026-12-31T23:59:59Z",
      "session": {
        "token": "sess_01J2ABC...",
        "expires_at": "2026-06-17T00:00:00Z",
        "heartbeat_interval_seconds": 30
      },
      "config": {
        "heartbeat_interval": 30,
        "scan_timeout_seconds": 600,
        "max_retries": 5,
        "offline_queue_capacity": 1000,
        "upload_chunk_size": 100
      }
    }
  }
}

Errors:
  400 - Invalid request body
  409 - Agent already registered (use /connect)
  429 - Rate limited
```

### 6.2 Agent Connect

```http
POST /api/v1/agents/connect
Content-Type: application/json
Authorization: Bearer <pat>

Request:
{
  "agent_id": "ag_01J2XYZ...",
  "session_info": {
    "client_version": "1.0.0",
    "uptime_seconds": 3600,
    "last_known_ip": "192.168.1.100"
  }
}

Response: 200 OK
{
  "data": {
    "session": {
      "token": "sess_01J2ABC...",
      "expires_at": "2026-06-18T00:00:00Z",
      "heartbeat_interval_seconds": 30
    },
    "agent": {
      "uuid": "ag_01J2XYZ...",
      "status": "online",
      "last_seen": "2026-06-16T12:00:00Z"
    }
  }
}

Errors:
  401 - Invalid or expired PAT
  404 - Agent not found (requires registration)
```

### 6.3 Agent Heartbeat

```http
POST /api/v1/agents/heartbeat
Content-Type: application/json
Authorization: Bearer <session_token>

Request:
{
  "status": "idle",
  "metrics": {
    "cpu_load_1m": 0.45,
    "memory_used_mb": 2048,
    "memory_total_mb": 8192,
    "disk_used_percent": 62.5,
    "network_rx_bytes": 1048576,
    "network_tx_bytes": 524288
  },
  "scan_progress": {
    "running": null,
    "completed_today": 5,
    "failed_today": 1
  },
  "uptime_seconds": 7200
}

Response: 200 OK
{
  "data": {
    "received_at": "2026-06-16T12:00:30Z",
    "next_heartbeat_in": 30,
    "pending_jobs_count": 2,
    "config_update": null,
    "actions": [
      { "type": "sync_config", "payload": { ... } },
      { "type": "update_agent", "payload": { ... } }
    ]
  }
}

Errors:
  401 - Session invalid/expired (trigger re-auth)
  408 - Heartbeat interval too large (retry with smaller interval)
```

### 6.4 Get Agent

```http
GET /api/v1/agents/{agent_id}
Authorization: Bearer <pat>

Response: 200 OK
{
  "data": {
    "agent": {
      "uuid": "ag_01J2XYZ...",
      "hostname": "customer-dev-01",
      "platform": "linux/amd64",
      "version": "1.0.0",
      "status": "online",
      "capabilities": ["semgrep", "nuclei", "trivy", "nmap"],
      "labels": { "env": "production" },
      "last_seen": "2026-06-16T12:00:30Z",
      "registered_at": "2026-06-01T00:00:00Z",
      "connected_since": "2026-06-16T08:00:00Z",
      "current_session": "sess_01J2ABC...",
      "scan_summary": {
        "total": 120,
        "completed": 115,
        "failed": 3,
        "in_progress": 2
      }
    }
  }
```

### 6.5 Agent Disconnect

```http
POST /api/v1/agents/disconnect
Authorization: Bearer <session_token>

Request:
{
  "reason": "shutdown",
  "uptime_seconds": 86400,
  "pending_results_count": 0
}

Response: 200 OK
{
  "data": {
    "disconnected_at": "2026-06-16T23:59:59Z",
    "invalidate_session": true,
    "offline_queue_accepted": true
  }
}
```

### 6.6 Create Scan

```http
POST /api/v1/scans
Content-Type: application/json
Authorization: Bearer <session_token>

Request:
{
  "agent_id": "ag_01J2XYZ...",
  "project_id": "proj_01J2...",
  "workspace_id": "ws_01J2...",
  "type": "semgrep",
  "target": {
    "path": "/workspace/my-app",
    "branch": "main",
    "commit": "a1b2c3d4...",
    "exclude_patterns": ["vendor/**", "node_modules/**"],
    "include_patterns": ["**/*.go", "**/*.py", "**/*.js"]
  },
  "parameters": {
    "rules": ["p/default", "p/owasp-top-ten"],
    "severity_threshold": "medium",
    "timeout_seconds": 300
  },
  "priority": "high",
  "metadata": {
    "triggered_by": "schedule",
    "scan_group_id": "sg_01J2..."
  }
}

Response: 201 Created
{
  "data": {
    "scan": {
      "uuid": "scan_01J2XYZ...",
      "status": "pending",
      "created_at": "2026-06-16T12:00:00Z"
    }
  }
}
```

### 6.7 Update Scan Status

```http
PATCH /api/v1/scans/{scan_uuid}
Authorization: Bearer <session_token>

Request:
{
  "status": "running",
  "current_stage": "dependency_sast",
  "progress_percent": 45,
  "estimated_remaining_seconds": 120,
  "artifacts": {
    "log_file": "/tmp/scan_abc123.log"
  }
}

Response: 200 OK
```

### 6.8 Upload Scan Results

```http
POST /api/v1/scans/{scan_uuid}/results
Content-Type: application/json
Authorization: Bearer <session_token>

Request:
{
  "status": "completed",
  "duration_seconds": 245,
  "summary": {
    "total_findings": 12,
    "by_severity": { "critical": 2, "high": 4, "medium": 3, "low": 3 },
    "by_type": { "sast": 8, "sca": 3, "secret": 1 },
    "new_findings": 5,
    "suppressed": 0
  },
  "findings": [
    {
      "tool": "semgrep",
      "rule_id": "python.lang.correctness.audit.assert-eval",
      "title": "Use of eval() detected",
      "severity": "high",
      "cvss_score": 7.5,
      "cwe_id": "CWE-95",
      "description": "Dynamic execution via eval() can lead to code injection.",
      "file_path": "src/utils.py",
      "line_start": 42,
      "line_end": 42,
      "code_snippet": "result = eval(user_input)",
      "confidence": "high",
      "remediation": "Avoid eval(). Use ast.literal_eval() if necessary.",
      "status": "open",
      "scanner_metadata": {
        "rule_url": "https://semgrep.dev/r/python.lang.correctness.audit.assert-eval"
      }
    }
  ],
  "vulnerabilities": [
    {
      "cve_id": "CVE-2024-21626",
      "title": "runc container breakout",
      "severity": "critical",
      "cvss_score": 9.8,
      "package_name": "runc",
      "package_version": "1.1.11",
      "fixed_version": "1.1.12",
      "source": "trivy",
      "type": "os",
      "status": "open",
      "advisory_url": "https://nvd.nist.gov/vuln/detail/CVE-2024-21626",
      "detected_at": "/usr/bin/docker-runc"
    }
  ],
  "errors": []
}

Response: 201 Created
{
  "data": {
    "scan_id": "scan_01J2XYZ...",
    "findings_created": 12,
    "vulnerabilities_created": 3,
    "duplicates_skipped": 0
  }
}
```

### 6.9 Poll Jobs

```http
GET /api/v1/jobs?agent_id={agent_id}&limit=5&types=semgrep,trivy
Authorization: Bearer <session_token>

Response: 200 OK
{
  "data": {
    "jobs": [
      {
        "uuid": "job_01J2...",
        "type": "scan",
        "scan_type": "semgrep",
        "payload": { ... },
        "priority": "high",
        "created_at": "2026-06-16T11:00:00Z",
        "ttl_seconds": 600
      }
    ],
    "pagination": {
      "total": 1,
      "limit": 5,
      "offset": 0
    }
  }
}
```

### 6.10 Acknowledge Job

```http
POST /api/v1/jobs/{job_uuid}/ack
Authorization: Bearer <session_token>

Request:
{
  "action": "accept | reject",
  "reason": ""  // required if reject
}

Response: 200 OK
```

---

## 7. Retry Policies, Timeouts, Backoff, Offline Mode

### 7.1 Default Parameters

```
┌────────────────────┬────────────┬──────────────┬──────────────────┐
│      Context       │ Max Retries│ Initial      │ Max Backoff      │
│                    │            │ Backoff      │                  │
├────────────────────┼────────────┼──────────────┼──────────────────┤
│ Agent Registration │ 3          │ 1s           │ 10s              │
│ Agent Connect      │ 3          │ 1s           │ 10s              │
│ Heartbeat          │ 10         │ 1s           │ 5min             │
│ Scan Create        │ 3          │ 2s           │ 30s              │
│ Scan Result Upload │ 5          │ 5s           │ 5min             │
│ Finding Batch      │ 5          │ 5s           │ 5min             │
│ Job Poll           │ 3          │ 500ms        │ 10s              │
│ Job Ack            │ 3          │ 1s           │ 10s              │
└────────────────────┴────────────┴──────────────┴──────────────────┘
```

### 7.2 Backoff Strategy

```
Exponential Backoff with Jitter:

  delay_n = min(max_delay, initial_delay * 2^n + random(0, initial_delay))

  Formula applied uniformly:

  n=0: 1.0s + jitter(0-1s)
  n=1: 2.0s + jitter(0-1s)
  n=2: 4.0s + jitter(0-1s)
  n=3: 8.0s + jitter(0-1s)
  n=4: 15.0s + jitter(0-1s)
  n=5: 30.0s + jitter(0-1s)
  n=6: 60.0s + jitter(0-1s)
  n=7: 120.0s + jitter(0-1s)
  n=8: 240.0s + jitter(0-1s)
  n=9: 300.0s + jitter(0-1s)
```

### 7.3 Timeouts

| Operation | Client Timeout | Server Timeout |
|-----------|---------------|----------------|
| Registration | 30s | 10s |
| Connect | 15s | 10s |
| Heartbeat | 30s | 15s |
| Scan Create | 15s | 10s |
| Scan Results Upload | 120s | 60s |
| Finding Batch Upload | 120s | 60s |
| Job Poll | 30s | 15s |
| Scan Execution (local) | N/A | 600s (configurable) |

### 7.4 Offline Mode

```
States:
  online:     Normal operation, all communication active
  degraded:   Heartbeat fails, retrying with backoff
  offline:    Max retries exhausted, entering offline mode
  reconnecting: Connection restored, syncing queued data

Offline Behavior:
  1. Queue all scan results locally in SQLite
  2. Continue executing cached jobs (if available)
  3. Persist scan results with status='pending_sync'
  4. Periodically attempt reconnection (every 5min)
  5. On reconnection:
     a. Re-authenticate (POST /connect)
     b. Query all pending_sync results
     c. Upload in chronological order
     d. Mark as synced or failed
  6. Offline queue capacity: 1000 results (LRU eviction)

Reconnection Backoff (once offline):
  Attempt 1: 30s
  Attempt 2: 60s
  Attempt 3: 120s
  Attempt 4: 300s (5 min, stays here)
  After 24h offline: switch to every 15min
```

---

## 8. Security Requirements

### 8.1 PAT Storage

```
Agent Machine:
  ┌──────────────────────────────────┐
  │   $HOME/.duckops/agent/         │
  │   ├── config.yaml               │ ← plaintext (non-secret config)
  │   ├── identity.json             │ ← encrypted (AgentID, PAT, keys)
  │   ├── offline_queue.db          │ ← SQLite (scan cache pending sync)
  │   ├── agent.log                 │ ← local audit log
  │   └── socket/duckops.sock       │ ← local API socket (existing)
  │
  │   identity.json format:
  │   {
  │     "agent_id": "ag_...",
  │     "pat": "AES256-GCM-BASE64==",
  │     "server_url": "https://api.duckops.io",
  │     "created_at": "...",
  │     "public_key": "...",
  │     "private_key": "AES256-GCM-BASE64=="
  │   }
  │
  │   AES-256-GCM key derived from:
  │   machine_id + user_id + salt
  │   (via PBKDF2, 100k iterations)
  └──────────────────────────────────┘
```

### 8.2 Token Rotation

| Token | Lifetime | Rotation Mechanism |
|-------|----------|-------------------|
| PAT (static) | 90 days | Server-initiated (check on heartbeat), client renewal via `/connect` |
| Session Token | 24 hours | Renewed on heartbeat if < 1h remaining |
| Agent ID | Permanent | Tied to machine identity, regenerated on factory reset |

### 8.3 TLS Requirements

```
Agent → Server:
  - TLS 1.3 minimum (1.2 with strong ciphers allowed)
  - Certificate pinning (SHA-256 of server cert)
  - mTLS optional (if agent has client cert)
  - Cipher suites: TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384
  - HSTS: require Strict-Transport-Security header

Certificate Validation:
  - Full chain verification
  - CRL/OCSP checking
  - Pinned public key for first connection (TOFU)
  - Re-pin on admin-triggered rotation only
```

### 8.4 Agent Identity Verification

```
Agent proves identity via:
  1. Registration: server-issued AgentID + PAT
  2. Connection: PAT (provided in Authorization header)
  3. Session: short-lived session token signed by server
  4. Message-level: HMAC-SHA256 over request body using session secret

Anti-Tampering:
  - All critical responses include HMAC signature in X-Signature header
  - Agent validates signature before processing "actions" from heartbeat
  - Server verifies agent hasn't been impersonated via session token binding

Hardware Binding (optional/advanced):
  - Agent generates Ed25519 keypair on registration
  - Sends public key to server
  - Signs heartbeat with private key
  - Server verifies signature each heartbeat
  - Private key stored encrypted with machine-bound key
```

### 8.5 Agent Authorization Model

```
┌──────────────┬────────────┬───────────────┬─────────────────┐
│  Resource    │  Create    │   Read        │   Update/Delete │
├──────────────┼────────────┼───────────────┼─────────────────┤
│ Agent (self) │ -          │ Always        │ With PAT       │
│ Other Agents │ Never      │ Never         │ Never           │
│ Scans (own)  │ Always     │ Always        │ Always          │
│ Scans (other)│ Never      │ Never         │ Never           │
│ Findings     │ Always     │ Always (own)  │ Only via scan   │
│ Projects     │ On assign  │ Assigned only │ -               │
│ Users        │ Never      │ Never         │ Never           │
│ Config       │ -          │ -             │ Admin API only  │
└──────────────┴────────────┴───────────────┴─────────────────┘
```

---

## 9. Database Schema

### 9.1 Agents Table

```sql
CREATE TABLE agents (
    uuid            TEXT PRIMARY KEY,          -- ag_01J2XYZ...
    hostname        TEXT NOT NULL,
    platform        TEXT NOT NULL,             -- linux/amd64, darwin/arm64
    version         TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'offline'
                    CHECK (status IN ('offline','online','degraded','disabled')),
    pat_hash        TEXT NOT NULL,             -- bcrypt hash of PAT
    public_key      TEXT,                      -- Ed25519 public key (optional)
    capabilities    TEXT NOT NULL DEFAULT '[]', -- JSON array
    labels          TEXT DEFAULT '{}',          -- JSON object
    max_concurrent_scans INTEGER NOT NULL DEFAULT 2,
    last_seen       TIMESTAMP WITH TIME ZONE,
    last_ip         TEXT,
    registered_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata        TEXT DEFAULT '{}'           -- JSON blob for extensible fields
);

CREATE INDEX idx_agents_status ON agents (status);
CREATE INDEX idx_agents_last_seen ON agents (last_seen);
CREATE INDEX idx_agents_hostname ON agents (hostname);
```

### 9.2 Agent Sessions Table

```sql
CREATE TABLE agent_sessions (
    uuid            TEXT PRIMARY KEY,           -- sess_01J2ABC...
    agent_uuid      TEXT NOT NULL REFERENCES agents(uuid) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,              -- SHA-256 of session token
    token_prefix    TEXT NOT NULL,              -- First 8 chars for lookup (sess_...)
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active','expired','revoked')),
    issued_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at      TIMESTAMP WITH TIME ZONE,
    client_version  TEXT,
    client_ip       TEXT,
    metadata        TEXT DEFAULT '{}'
);

CREATE INDEX idx_sessions_agent ON agent_sessions (agent_uuid);
CREATE INDEX idx_sessions_expires ON agent_sessions (expires_at);
CREATE INDEX idx_sessions_token_prefix ON agent_sessions (token_prefix);
CREATE INDEX idx_sessions_status ON agent_sessions (status);
```

### 9.3 Heartbeats Table

```sql
CREATE TABLE heartbeats (
    uuid            TEXT PRIMARY KEY,
    agent_uuid      TEXT NOT NULL REFERENCES agents(uuid) ON DELETE CASCADE,
    session_uuid    TEXT NOT NULL REFERENCES agent_sessions(uuid) ON DELETE CASCADE,
    status          TEXT NOT NULL,              -- idle, scanning, busy
    cpu_load_1m     REAL,
    memory_used_mb  INTEGER,
    memory_total_mb INTEGER,
    disk_used_percent REAL,
    network_rx_bytes BIGINT,
    network_tx_bytes BIGINT,
    scan_running    TEXT,                       -- scan UUID or null
    completed_today INTEGER DEFAULT 0,
    failed_today    INTEGER DEFAULT 0,
    uptime_seconds  INTEGER,
    received_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_heartbeats_agent ON heartbeats (agent_uuid, received_at DESC);
CREATE INDEX idx_heartbeats_session ON heartbeats (session_uuid);
CREATE INDEX idx_heartbeats_received ON heartbeats (received_at);
-- Partition by month for retention (e.g., 90 days)
```

### 9.4 Scan Jobs Table

```sql
CREATE TABLE scan_jobs (
    uuid            TEXT PRIMARY KEY,           -- job_01J2...
    agent_uuid      TEXT REFERENCES agents(uuid),
    project_id      TEXT,
    workspace_id    TEXT,
    scan_type       TEXT NOT NULL,              -- semgrep, nuclei, trivy, nmap, custom
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','assigned','accepted',
                           'running','completed','failed','cancelled','timed_out')),
    priority        TEXT NOT NULL DEFAULT 'medium'
                    CHECK (priority IN ('critical','high','medium','low')),
    target          TEXT NOT NULL,               -- JSON: path, branch, commit, patterns
    parameters      TEXT DEFAULT '{}',           -- JSON: tool-specific params
    assigned_at     TIMESTAMP WITH TIME ZONE,
    accepted_at     TIMESTAMP WITH TIME ZONE,
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    ttl_seconds     INTEGER NOT NULL DEFAULT 600,
    retry_count     INTEGER NOT NULL DEFAULT 0,
    max_retries     INTEGER NOT NULL DEFAULT 3,
    error_message   TEXT,
    metadata        TEXT DEFAULT '{}',           -- JSON
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status ON scan_jobs (status, priority DESC, created_at);
CREATE INDEX idx_jobs_agent ON scan_jobs (agent_uuid, status);
CREATE INDEX idx_jobs_type ON scan_jobs (scan_type);
CREATE INDEX idx_jobs_created ON scan_jobs (created_at);
```

### 9.5 Scan Results / Findings Table

```sql
CREATE TABLE findings (
    uuid                TEXT PRIMARY KEY,        -- fnd_01J2...
    scan_job_uuid       TEXT NOT NULL REFERENCES scan_jobs(uuid) ON DELETE CASCADE,
    agent_uuid          TEXT REFERENCES agents(uuid),
    -- Source
    tool                TEXT NOT NULL,           -- semgrep, nuclei, trivy, etc.
    rule_id             TEXT,
    -- Classification
    title               TEXT NOT NULL,
    severity            TEXT NOT NULL
                        CHECK (severity IN ('critical','high','medium','low','info')),
    cvss_score          REAL,
    cwe_id              TEXT,
    cve_id              TEXT,
    confidence          TEXT,                    -- high, medium, low
    -- Location
    file_path           TEXT,
    line_start          INTEGER,
    line_end            INTEGER,
    code_snippet        TEXT,
    -- Details
    description         TEXT,
    remediation         TEXT,
    -- Package (for SCA)
    package_name        TEXT,
    package_version     TEXT,
    fixed_version       TEXT,
    -- Status
    status              TEXT NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open','in_progress','fixed',
                               'accepted','false_positive','suppressed')),
    suppressed_by       TEXT,
    suppressed_reason   TEXT,
    suppressed_at       TIMESTAMP WITH TIME ZONE,
    -- Fingerprinting
    fingerprint         TEXT,                   -- SHA-256 of (tool, rule_id, file, line, context)
    -- Metadata
    scanner_metadata    TEXT DEFAULT '{}',
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_findings_scan ON findings (scan_job_uuid);
CREATE INDEX idx_findings_agent ON findings (agent_uuid);
CREATE INDEX idx_findings_severity ON findings (severity, created_at DESC);
CREATE INDEX idx_findings_cwe ON findings (cwe_id);
CREATE INDEX idx_findings_cve ON findings (cve_id);
CREATE INDEX idx_findings_status ON findings (status);
CREATE UNIQUE INDEX idx_findings_fingerprint ON findings (fingerprint);
```

### 9.6 Vulnerabilities Table

```sql
CREATE TABLE vulnerabilities (
    uuid                TEXT PRIMARY KEY,
    scan_job_uuid       TEXT NOT NULL REFERENCES scan_jobs(uuid) ON DELETE CASCADE,
    agent_uuid          TEXT REFERENCES agents(uuid),
    cve_id              TEXT NOT NULL,
    title               TEXT NOT NULL,
    severity            TEXT NOT NULL
                        CHECK (severity IN ('critical','high','medium','low','info')),
    cvss_score          REAL,
    epss_score          REAL,                   -- Exploit Prediction Scoring System
    package_name        TEXT,
    package_version     TEXT,
    fixed_version       TEXT,
    source              TEXT,                    -- trivy, grype, osv.dev, nvd
    type                TEXT,                    -- os, library, language
    status              TEXT NOT NULL DEFAULT 'open'
                        CHECK (status IN ('open','in_progress','fixed',
                               'accepted','false_positive')),
    advisory_url        TEXT,
    exploit_available   BOOLEAN DEFAULT FALSE,
    detected_paths      TEXT,                    -- JSON array of paths
    vuln_metadata       TEXT DEFAULT '{}',
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vulns_scan ON vulnerabilities (scan_job_uuid);
CREATE INDEX idx_vulns_cve ON vulnerabilities (cve_id);
CREATE INDEX idx_vulns_severity ON vulnerabilities (severity, cvss_score DESC);
CREATE INDEX idx_vulns_status ON vulnerabilities (status);
CREATE INDEX idx_vulns_package ON vulnerabilities (package_name, package_version);
```

### 9.7 AI Insights Table

```sql
CREATE TABLE insights (
    uuid                TEXT PRIMARY KEY,
    agent_uuid          TEXT REFERENCES agents(uuid),
    project_id          TEXT,
    type                TEXT NOT NULL,           -- trend, anomaly, recommendation, summary
    title               TEXT NOT NULL,
    severity            TEXT,
    description         TEXT NOT NULL,
    data                TEXT,                    -- JSON: supporting data
    source_scan_uuids   TEXT,                    -- JSON array of scan UUIDs
    model_version       TEXT,
    confidence_score    REAL,
    actionable          BOOLEAN DEFAULT TRUE,
    status              TEXT DEFAULT 'active'
                        CHECK (status IN ('active','dismissed','actioned')),
    expires_at          TIMESTAMP WITH TIME ZONE,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_insights_agent ON insights (agent_uuid, created_at DESC);
CREATE INDEX idx_insights_type ON insights (type, created_at DESC);
CREATE INDEX idx_insights_severity ON insights (severity);
```

### 9.8 Projects Table

```sql
CREATE TABLE projects (
    uuid                TEXT PRIMARY KEY,
    name                TEXT NOT NULL,
    description         TEXT,
    repository_url      TEXT,
    default_branch      TEXT DEFAULT 'main',
    language            TEXT,
    framework           TEXT,
    agent_uuids         TEXT DEFAULT '[]',        -- JSON array of assigned agent UUIDs
    labels              TEXT DEFAULT '{}',
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_name ON projects (name);
CREATE INDEX idx_projects_language ON projects (language);
```

### 9.9 Entity Relationships

```
agents 1───* agent_sessions
agents 1───* heartbeats
agents 1───* scan_jobs
agents 1───* findings
agents 1───* vulnerabilities
agents 1───* insights

scan_jobs 1───* findings
scan_jobs 1───* vulnerabilities

projects 1───* scan_jobs
projects 1───* insights
```

---

## 10. Folder Structure

### 10.1 Server-Side (DuckOps Server)

```
internal/
  agentapi/                # NEW: Agent API package
    handler.go             # Route registration
    register.go            # POST /agents/register
    connect.go             # POST /agents/connect
    heartbeat.go           # POST /agents/heartbeat
    disconnect.go          # POST /agents/disconnect
    get.go                 # GET /agents/{id}
    middleware.go          # PAT/session auth middleware

  agent/                   # EXTEND: Agent service layer
    registry.go            # Agent CRUD, PAT generation
    session.go             # Session management, token rotation
    heartbeat.go           # Heartbeat processing
    metrics.go             # Agent metrics tracking

  scanapi/                 # NEW: Scan API package
    handler.go             # Route registration
    create.go              # POST /scans
    update.go              # PATCH /scans/{uuid}
    results.go             # POST /scans/{uuid}/results
    get.go                 # GET /scans/{uuid}
    list.go                # GET /scans

  scan/                    # NEW: Scan service layer
    service.go             # Scan lifecycle management
    findings.go            # Finding batch insert, dedup
    vulns.go               # Vulnerability batch insert
    scheduler.go           # Job assignment to agents

  jobapi/                  # NEW: Job API package
    handler.go
    poll.go                # GET /jobs
    ack.go                 # POST /jobs/{uuid}/ack

  job/                     # NEW: Job service layer
    queue.go               # Priority queue implementation
    dispatcher.go          # Job assignment logic
    timeout.go             # TTL enforcement

  projectapi/              # NEW: Project API package
    handler.go
    crud.go                # CRUD for projects

  project/                 # NEW: Project service layer
    service.go

  insight/                 # NEW: AI insight engine
    engine.go              # Analysis engine
    trends.go              # Trend detection
    recommendations.go     # Recommendation generation

  notification/            # NEW: Notification service
    webhook.go             # Webhook dispatch
    alert.go               # Alert escalation

  db/
    migrations/
      20260101000000_add_agent_tables.sql
      20260102000000_add_scan_tables.sql
      20260103000000_add_finding_tables.sql
      20260104000000_add_insight_tables.sql
    sql/
      agents.sql           # sqlc queries
      sessions.sql
      heartbeats.sql
      scans.sql
      findings.sql
      vulns.sql
      insights.sql
      projects.sql
```

### 10.2 Agent-Side (Standalone Go Binary or Library)

```
agent/                     # NEW: Agent standalone binary/library
  cmd/
    agent/
      main.go              # Entry point

  internal/
    config/
      manager.go           # Config load/watch/reload
      defaults.go          # Default configuration

    auth/
      manager.go           # PAT load/encrypt/decrypt
      rotation.go          # Token rotation logic
      identity.go          # Agent identity management

    client/
      client.go            # HTTP client with retry/backoff
      middleware.go        # Auth header injection
      errors.go            # Error classification

    heartbeat/
      service.go           # Heartbeat loop
      metrics.go           # System metrics collection
      payload.go           # Heartbeat payload builder

    scan/
      runner.go            # Scan execution orchestrator
      semgrep.go           # Semgrep integration
      nuclei.go            # Nuclei integration
      trivy.go             # Trivy integration
      nmap.go              # Nmap integration
      katana.go            # Katana integration
      custom.go            # Custom script execution

    upload/
      uploader.go          # Result upload orchestrator
      findings.go          # Finding batch upload
      vulns.go             # Vulnerability batch upload
      compressor.go        # Payload compression

    scheduler/
      scheduler.go         # Job scheduling & dispatch
      queue.go             # Priority queue
      timeout.go           # Execution timeout
      limiter.go           # Concurrency limiter

    storage/
      storage.go           # Interface
      sqlite.go            # SQLite implementation
      identity.go          # Agent identity persistence
      queue.go             # Offline queue persistence

    lifecycle/
      controller.go        # Agent state machine
      startup.go           # Startup sequence
      shutdown.go          # Graceful shutdown
      reconnect.go         # Reconnection logic

    log/
      logger.go            # Structured logging

  config.yaml.example      # Example configuration
```

---

## 11. Implementation Roadmap

### Phase 1: Agent Registration (Week 1-2)

```
Goals:
  - Agent can register with the server
  - Server issues PAT and AgentID
  - Agent persists identity locally
  - Agent can reconnect with existing credentials

Deliverables:
  [Server] Create agents table, agent_sessions table (migration)
  [Server] POST /api/v1/agents/register endpoint
  [Server] POST /api/v1/agents/connect endpoint
  [Server] POST /api/v1/agents/disconnect endpoint
  [Server] GET /api/v1/agents/{id} endpoint
  [Server] PAT generation (crypto/rand, bcrypt hash)
  [Server] Session token generation (SHA-256 HMAC)
  [Server] Auth middleware (Bearer token validation)
  [Server] Background janitor: expire stale sessions

  [Agent] Config manager (load YAML, env override)
  [Agent] Auth manager (PAT encryption, identity persistence)
  [Agent] API client base (HTTP, TLS, retry)
  [Agent] Registration flow (register → persist → connect)
  [Agent] Reconnection flow (connect with existing PAT)

Tests:
  - Agent register + connect end-to-end
  - Token expiry and reconnection
  - Invalid PAT rejection
  - Concurrent registration prevention (idempotency key)

Verification:
  curl -X POST https://api.duckops.io/api/v1/agents/register \
    -H 'Content-Type: application/json' \
    -d '{"hostname":"test-01","platform":"linux/amd64","version":"1.0.0"}'
  → 201 + agent_uuid + pat
```

### Phase 2: Heartbeats (Week 3-4)

```
Goals:
  - Agent sends periodic heartbeats
  - Server tracks agent online/offline status
  - Heartbeat includes system metrics
  - Server can push config updates via heartbeat response

Deliverables:
  [Server] POST /api/v1/agents/heartbeat endpoint
  [Server] Create heartbeats table (migration)
  [Server] Heartbeat processing: update last_seen, status
  [Server] Server-issued actions in heartbeat response
  [Server] Stale agent detection (no heartbeat in N intervals)
  [Server] Agent status dashboard query API

  [Agent] Heartbeat service (ticker, payload, retry)
  [Agent] System metrics collection (CPU, memory, disk, network)
  [Agent] Heartbeat response processing (actions, config updates)
  [Agent] Degraded mode detection (missing heartbeats)

Tests:
  - Heartbeat loop timing (30s interval ± 5s)
  - Server detects agent offline after 3 missed heartbeats
  - Server actions in response (e.g., update scan interval)
  - Graceful degradation when server unreachable

Verification:
  # Agent running, heartbeats every 30s
  curl -s https://api.duckops.io/api/v1/agents/ag_01J2... \
    -H 'Authorization: Bearer pat_...' | jq '.data.agent.status'
  → "online"
```

### Phase 3: Scan Uploads (Week 5-7)

```
Goals:
  - Agent can upload scan results (findings + vulnerabilities)
  - Server stores, deduplicates, and serves results
  - Batch upload with progress
  - Offline queue for disconnected operation

Deliverables:
  [Server] Create scan_jobs, findings, vulnerabilities tables
  [Server] POST /api/v1/scans endpoint
  [Server] PATCH /api/v1/scans/{uuid} endpoint
  [Server] POST /api/v1/scans/{uuid}/results endpoint
  [Server] GET /api/v1/scans/{uuid} endpoint
  [Server] GET /api/v1/scans endpoint (list, filter)
  [Server] Finding deduplication via fingerprint
  [Server] CVE correlation (NVD lookup optional)

  [Agent] Scan result uploader (JSON batch, compression)
  [Agent] Finding uploading with retry
  [Agent] Vulnerability uploading with retry
  [Agent] Offline queue (SQLite, pending_sync status)
  [Agent] Offline sync on reconnection
  [Agent] Result validation before upload

Tests:
  - Upload 10k findings in single batch
  - Fingerprint-based deduplication
  - Offline queue capacity (1000 entries)
  - Offline → online sync integrity
  - Compression ratio (target: 5:1 for large payloads)

Verification:
  curl -X POST https://api.duckops.io/api/v1/scans/scan_01J2.../results \
    -H 'Authorization: Bearer sess_...' \
    -H 'Content-Type: application/json' \
    -d '{"status":"completed","findings":[...],"vulnerabilities":[...]}'
  → 201 + finding count
```

### Phase 4: Job Execution (Week 8-10)

```
Goals:
  - Server assigns scan jobs to agents
  - Agent polls or receives jobs via SSE
  - Agent executes scans based on job spec
  - Results uploaded with job reference
  - Full scan lifecycle management

Deliverables:
  [Server] Scan jobs priority queue
  [Server] GET /api/v1/jobs endpoint (agent poll)
  [Server] POST /api/v1/jobs/{uuid}/ack endpoint
  [Server] Job timeout enforcement (background)
  [Server] Job retry logic (max 3 retries)
  [Server] Server-sent events for push (optional SSE)

  [Agent] Job scheduler (poll, priority queue, dispatch)
  [Agent] Scan runner (tool execution, output parsing)
  [Agent] Semgrep integration module
  [Agent] Nuclei integration module
  [Agent] Trivy integration module
  [Agent] Nmap integration module
  [Agent] Custom scan module (script-based)
  [Agent] Execution timeout and cancellation
  [Agent] Concurrency limiter (per agent config)

Tests:
  - Job creation → assignment → acceptance → execution → results
  - Multiple agents polling same queue (no double-assignment)
  - Job timeout and reassignment
  - Concurrent scan execution (max 2 per agent)
  - Large project scan (100k+ files, truncate before timeout)

Verification:
  # Create job via API
  curl -X POST https://api.duckops.io/api/v1/scans \
    -H 'Authorization: Bearer sess_...' \
    -d '{...}'
  → 201 + job appears in agent poll
```

### Phase 5: AI Insights (Week 11-12)

```
Goals:
  - Server analyzes scan results for patterns
  - Generate recommendations and trend reports
  - Provide actionable security insights
  - Support for cross-project and cross-agent analysis

Deliverables:
  [Server] Create insights table
  [Server] AI insight engine (LLM integration)
  [Server] Trend detection: finding frequency over time
  [Server] Pattern recognition: recurring vulnerability types
  [Server] Severity escalation detection
  [Server] Remediation recommendation generation
  [Server] Cross-agent comparison (benchmarking)
  [Server] Insight query API with filters
  [Server] Email/webhook notifications for critical insights

Tests:
  - AI insight generation with sample findings
  - Trend detection accuracy (10+ scans)
  - Cross-project pattern matching
  - Insight deduplication
  - Insight expiration and archiving

Verification:
  curl -s https://api.duckops.io/api/v1/insights?agent_id=ag_01J2... \
    -H 'Authorization: Bearer pat_...' | jq '.data.insights[0]'
  → "title": "Critical: 3 new RCE vulnerabilities detected in src/"
```

---

## 12. Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|------------|------------|
| **Agent PAT leaked** | Unauthorized access to scan results, impersonation | Medium | Encrypt PAT at rest (AES-256-GCM + machine-bound key), short PAT lifetime (90d), rotation on heartbeat, audit log all API access, immediate revocation endpoint |
| **Offline queue overflow** | Loss of scan results | Low | LRU eviction, configurable capacity (default 1000), alert when >80% full, oldest entries get lowest priority |
| **Server DDoS from agents** | Service degradation for all customers | Medium | Rate limiting per agent (burst + sustained), exponential backoff enforcement on client, CAPTCHA on registration if suspicious |
| **Agent identity theft** | False scan results, data pollution | Medium | Session token binding, heartbeat signature verification (Ed25519), PAT rotation, IP reputation scoring |
| **Network partition** | Agent unable to upload results | High | Offline mode with local queue, exponential reconnection backoff, sync integrity checks on reconnect, eventual consistency model |
| **Race condition on job assignment** | Two agents get same job | Medium | Atomic job assignment with FOR UPDATE SKIP LOCKED, status check before execution, idempotent result upload |
| **Critical finding delayed** | Security exposure window | Medium | Real-time push via SSE (when connected), priority heartbeat with urgent flag, mobile push/webhook integration |
| **Database migration backward compat** | Server deployment rollback issues | Low | All migrations must be additive (no destructive DDL), phased rollouts, canary testing, read replicas for zero-downtime migration |
| **Agent binary tampering** | Compromised scan results, data exfiltration | Medium | Code signing, integrity hash on startup, server-side scan parameter validation, immutable infrastructure for agent deployment |
| **Scanner tool vulnerability** | RCE on agent machine | Medium | Run scanners in sandboxed containers, strict resource limits, no-root execution, tool binary hash verification, timeout enforcement |
| **Large scan result payload** | HTTP timeout, memory OOM | Medium | Chunked upload (100 findings/batch), server-side streaming validation, compression (gzip), size limits with clear error codes |
| **Clock skew between agent/server** | Session token validation failures, timestamps incorrect | Low | NTP synchronization, clock skew tolerance (±5min), use monotonic clock for internal timing, relative timestamps for critical operations |
