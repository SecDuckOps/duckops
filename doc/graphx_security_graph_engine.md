# GraphX: Persistent Security Code Graph Engine

## Implementation Plan

**Version**: 1.0
**Created**: 2026-05-12
**Target Duration**: 8-12 weeks
**Architecture**: Local-first, Graph-based, AI-native

---

## 1. Project Overview

GraphX is a persistent AI-native code intelligence and security graph engine that serves as the "Security-Aware Persistent Software Intelligence Graph" for enterprise-scale repositories.

### Core Purpose
- Understand large repositories incrementally
- Trace code relationships and security impacts
- Minimize token usage via graph-first context loading
- Power threat modeling and blast radius analysis

### Key Metrics
- **Token Reduction**: 5x-10x via graph-first querying
- **Update Speed**: Sub-2 second for small diffs
- **Languages**: 15+ (extensible via plugins)
- **Storage**: SQLite with WAL mode (local-first)

---

## 2. Architecture Overview

```
Repository
    ↓
AST Parsing (tree-sitter + parsers)
    ↓
IR Extraction
    ↓
Knowledge Graph (SQLite + embeddings)
    ↓
Incremental Update Engine
    ↓
Blast Radius Engine
    ↓
Threat Modeling Layer
    ↓
AI Reasoning Layer (context engine)
    ↓
Visualization Layer
```

---

## 3. Module Structure

### 3.1 Core Modules

| Module | Path | Purpose |
|--------|------|---------|
| **graph** | `internal/graphx/graph` | SQLite-backed graph database |
| **parser** | `internal/graphx/parser` | AST parsing with tree-sitter |
| **ir** | `internal/graphx/ir` | Intermediate representation |
| **storage** | `internal/graphx/storage` | Persistence layer |
| **config** | `internal/graphx/config` | Configuration management |

### 3.2 Analysis Modules

| Module | Path | Purpose |
|--------|------|---------|
| **blast_radius** | `internal/graphx/analysis/blast` | Impact analysis |
| **threat_model** | `internal/graphx/analysis/threat` | STRIDE analysis |
| **execution_flow** | `internal/graphx/analysis/flow` | Request lifecycle tracing |
| **community** | `internal/graphx/analysis/community` | Graph clustering |
| **architecture** | `internal/graphx/analysis/arch` | Architectural patterns |

### 3.3 Engine Modules

| Module | Path | Purpose |
|--------|------|---------|
| **incremental** | `internal/graphx/engine/incremental` | Change detection |
| **semantic** | `internal/graphx/engine/semantic` | Vector search |
| **context** | `internal/graphx/engine/context` | AI context generation |
| **diff** | `internal/graphx/engine/diff` | Graph comparison |
| **watch** | `internal/graphx/engine/watch` | File monitoring |

### 3.4 Interface Modules

| Module | Path | Purpose |
|--------|------|---------|
| **cli** | `internal/graphx/cli` | CLI commands |
| **visualization** | `internal/graphx/ui` | Graph visualization |
| **export** | `internal/graphx/export` | Format exporters |

---

## 4. Database Schema

### 4.1 Nodes Table
```sql
CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    type TEXT NOT NULL,  -- file, function, class, api, service, infra, db, queue, pipeline, k8s
    name TEXT NOT NULL,
    fully_qualified_name TEXT,
    file_path TEXT,
    language TEXT,
    framework TEXT,
    security_sensitivity REAL DEFAULT 0.0,
    internet_exposed INTEGER DEFAULT 0,
    auth_sensitive INTEGER DEFAULT 0,
    pii_handling INTEGER DEFAULT 0,
    ownership TEXT,
    risk_score REAL DEFAULT 0.0,
    blast_radius_score REAL DEFAULT 0.0,
    content_hash TEXT,
    metadata TEXT,  -- JSON blob
    created_at INTEGER,
    updated_at INTEGER,
    indexed_at INTEGER
);

CREATE INDEX idx_nodes_repo ON nodes(repo_id);
CREATE INDEX idx_nodes_type ON nodes(type);
CREATE INDEX idx_nodes_file ON nodes(file_path);
CREATE INDEX idx_nodes_risk ON nodes(risk_score DESC);
```

### 4.2 Edges Table
```sql
CREATE TABLE edges (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    edge_type TEXT NOT NULL,  -- imports, calls, inherits, reads, writes, api_call, produces, consumes, trust_boundary, auth_dep
    confidence REAL DEFAULT 1.0,  -- 0.0-1.0
    confidence_level TEXT DEFAULT 'EXTRACTED',  -- EXTRACTED, INFERRED, AMBIGUOUS
    extraction_source TEXT,  -- ast, regex, manual, inferred
    metadata TEXT,  -- JSON blob
    created_at INTEGER,
    updated_at INTEGER
);

CREATE INDEX idx_edges_source ON edges(source_id);
CREATE INDEX idx_edges_target ON edges(target_id);
CREATE INDEX idx_edges_type ON edges(edge_type);
CREATE INDEX idx_edges_repo ON edges(repo_id);
```

### 4.3 Embeddings Table
```sql
CREATE TABLE embeddings (
    node_id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    vector BLOB NOT NULL,  -- Stored as blob for efficiency
    model TEXT NOT NULL,
    dimensionality INTEGER,
    created_at INTEGER
);

CREATE INDEX idx_embeddings_repo ON embeddings(repo_id);
```

### 4.4 Repositories Table
```sql
CREATE TABLE repositories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    root_path TEXT,
    language_stats TEXT,  -- JSON
    last_analyzed INTEGER,
    last_hash TEXT,
    config TEXT,  -- JSON config
    created_at INTEGER,
    updated_at INTEGER
);
```

### 4.5 Snapshots Table
```sql
CREATE TABLE snapshots (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    snapshot_type TEXT NOT NULL,  -- full, incremental
    node_count INTEGER,
    edge_count INTEGER,
    content_hash TEXT,
    created_at INTEGER,
    metadata TEXT
);

CREATE INDEX idx_snapshots_repo ON snapshots(repo_id);
CREATE INDEX idx_snapshots_time ON snapshots(created_at DESC);
```

### 4.6 Findings Table
```sql
CREATE TABLE findings (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    node_id TEXT,
    severity TEXT NOT NULL,  -- critical, high, medium, low, info
    category TEXT NOT NULL,  -- security, architecture, quality
    title TEXT NOT NULL,
    description TEXT,
    affected_nodes TEXT,  -- JSON array
    recommendations TEXT,
    created_at INTEGER,
    updated_at INTEGER
);

CREATE INDEX idx_findings_repo ON findings(repo_id);
CREATE INDEX idx_findings_severity ON findings(severity);
```

---

## 5. Implementation Phases

### Phase 1: Core Infrastructure (Weeks 1-3)

#### Week 1: Project Setup & Graph Database
- [ ] Initialize `internal/graphx` module
- [ ] Set up go.mod with dependencies:
  - `modernc.org/sqlite` (already in go.mod)
  - `github.com/tree-sitter/tree-sitter` (for parsing)
  - `github.com/redis/go-redis/v9` (optional, for distributed)
  - `github.com/semi-technologies/weaviate-go-client` (optional, for vector search)
- [ ] Implement SQLite schema migration
- [ ] Create basic graph operations (CRUD nodes/edges)
- [ ] Add WAL mode for performance
- [ ] Test basic persistence

#### Week 2: AST Parsing Engine
- [ ] Implement `parser.Parser` interface
- [ ] Add tree-sitter bindings for:
  - Python
  - TypeScript/JavaScript
  - Go
  - Rust
- [ ] Create `ir.Program` struct
- [ ] Implement basic IR extraction:
  - Function definitions
  - Class definitions
  - Import statements
  - Call relationships
- [ ] Add regex fallback for simple patterns

#### Week 3: Knowledge Graph Layer
- [ ] Connect IR extraction to graph storage
- [ ] Implement node creation from IR
- [ ] Implement edge creation (imports, calls)
- [ ] Add metadata extraction (framework, security tags)
- [ ] Build basic CLI: `graphx graph build`
- [ ] Test on sample repository

### Phase 2: Incremental Updates (Weeks 4-5)

#### Week 4: Change Detection Engine
- [ ] Implement file hashing (xxhash/fastcdc)
- [ ] Create incremental scanner
- [ ] Add git integration for change detection
- [ ] Implement targeted re-parsing
- [ ] Update affected edges only
- [ ] Add embedding updates for changed nodes

#### Week 5: AI Context Engine
- [ ] Implement graph-first query engine
- [ ] Build context window optimization
- [ ] Add security-relevant node prioritization
- [ ] Create minimal context generator
- [ ] Add token estimation
- [ ] Test token reduction (target 5-10x)

### Phase 3: Security Analysis (Weeks 6-7)

#### Week 6: Blast Radius Engine
- [ ] Implement BFS/DFS traversal
- [ ] Add weighted path analysis
- [ ] Create auth impact detection
- [ ] Build RBAC impact analysis
- [ ] Add API exposure tracking
- [ ] Implement privilege escalation detection
- [ ] Add CLI: `graphx blast-radius <path>`

#### Week 7: Threat Modeling
- [ ] Implement STRIDE analysis
- [ ] Create attack path generation
- [ ] Add trust boundary detection
- [ ] Build abuse case generation
- [ ] Implement entry point identification
- [ ] Add lateral movement paths
- [ ] Add CLI: `graphx threat-model`

### Phase 4: Advanced Analysis (Weeks 8-9)

#### Week 8: Execution Flow & Community Detection
- [ ] Implement request lifecycle tracing
- [ ] Add auth flow analysis
- [ ] Build queue flow detection
- [ ] Implement Leiden/Louvain clustering
- [ ] Add service boundary detection
- [ ] Build architectural pattern detection

#### Week 9: Architecture Intelligence
- [ ] Implement centrality metrics (betweenness, degree)
- [ ] Add god service detection
- [ ] Build architectural bottleneck detection
- [ ] Implement cyclic dependency detection
- [ ] Create architectural risk scoring
- [ ] Add surprise/unusual relationship detection

### Phase 5: Integration & Polish (Weeks 10-12)

#### Week 10: Search & Export
- [ ] Implement semantic search (vector + keyword)
- [ ] Add FTS5 for full-text search
- [ ] Create embedding generation (local model)
- [ ] Add export formats:
  - GraphML
  - Neo4j Cypher
  - Mermaid
  - PlantUML
  - JSON

#### Week 11: Watch Mode & Visualization
- [ ] Implement file system watcher
- [ ] Add git hook integration
- [ ] Create auto-rebuild on changes
- [ ] Build web visualization UI
- [ ] Add Cytoscape.js integration
- [ ] Implement attack path explorer

#### Week 12: Graph Diff & Multi-Repo
- [ ] Implement snapshot comparison
- [ ] Add security drift detection
- [ ] Build architecture change summaries
- [ ] Add multi-repo support
- [ ] Create cross-repo dependency analysis
- [ ] Final testing and documentation

---

## 6. Key Interfaces

### 6.1 Parser Interface
```go
type Parser interface {
    Name() string
    Language() string
    Parse(content []byte) (*ir.Program, error)
    ExtractSecurityMetadata(*ir.Program) map[string]interface{}
}
```

### 6.2 Graph Interface
```go
type Graph interface {
    AddNode(node *Node) error
    AddEdge(edge *Edge) error
    GetNode(id string) (*Node, error)
    Query(q *Query) ([]*Node, []*Edge, error)
    Traverse(start string, opts *TraversalOpts) ([]*Node, error)
}
```

### 6.3 Context Engine Interface
```go
type ContextEngine interface {
    GenerateContext(req *ContextRequest) (*ContextResponse, error)
    EstimateTokens(ctx *ContextResponse) int
    OptimizeForWindow(ctx *ContextResponse, maxTokens int) *ContextResponse
}
```

### 6.4 Blast Radius Engine Interface
```go
type BlastRadiusEngine interface {
    Analyze(nodeID string, opts *BlastOpts) (*BlastResult, error)
    FindAffectedAuthFlows(nodeID string) []string
    FindImpactedAPIs(nodeID string) []string
    CalculateRiskScore(nodeID string) float64
}
```

---

## 7. CLI Commands

### Graph Commands
```bash
graphx graph init <repo-path>           # Initialize repository
graphx graph build                      # Build graph
graphx graph watch                      # Watch mode
graphx graph query <query>              # Query graph
graphx graph export <format>            # Export graph
graphx graph diff <commit-a> <commit-b>  # Compare snapshots
```

### Analysis Commands
```bash
graphx blast-radius <path>               # Analyze blast radius
graphx blast-radius <path> --auth       # Focus on auth impact
graphx threat-model                     # Generate threat model
graphx threat-model --stride             # STRIDE analysis
graphx attack-paths                      # Find attack paths
graphx execution-flow <endpoint>        # Trace execution
```

### Architecture Commands
```bash
graphx architecture review               # Architecture analysis
graphx architecture bottlenecks          # Find bottlenecks
graphx architecture clusters             # Community detection
graphx architecture surprise             # Unusual patterns
```

### Utility Commands
```bash
graphx search <query>                    # Semantic search
graphx export <format>                   # Export graph
graphx snapshot create                   # Create snapshot
graphx snapshot list                     # List snapshots
graphx findings                          # List security findings
```

---

## 8. Configuration

### 8.1 graphx Config (graphx.yaml)
```yaml
repository:
  path: /path/to/repo
  ignore:
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/*.test.go"

graph:
  storage: sqlite
  path: .graphx/graph.db
  wal_mode: true

parser:
  languages:
    - python
    - typescript
    - go
  parallel: 4

analysis:
  security:
    enabled: true
    sensitivity_threshold: 0.5
  blast_radius:
    max_depth: 10
    weight_auth: 2.0
    weight_api: 1.5

watch:
  enabled: true
  debounce_ms: 500

export:
  formats:
    - mermaid
    - graphml
    - json
```

---

## 9. Security Metadata Schema

Each node should have security-relevant metadata:

```json
{
  "node": {
    "id": "auth.login",
    "type": "function",
    "security_metadata": {
      "internet_exposed": true,
      "auth_sensitive": true,
      "handles_tokens": true,
      "handles_pii": false,
      "requires_authorization": true,
      "trust_boundary": "public",
      "risk_score": 8.7,
      "blast_radius_score": 9.2
    }
  }
}
```

---

## 10. Edge Confidence Levels

| Level | Description | Score Range |
|-------|-------------|--------------|
| **EXTRACTED** | Directly from AST/parser | 1.0 |
| **INFERRED** | Logically deduced | 0.7-0.9 |
| **AMBIGUOUS** | Uncertain relationship | 0.3-0.6 |

---

## 11. Technology Decisions

### Storage
- **SQLite with WAL mode**: Local-first, no cloud dependency, high performance
- **FTS5**: Full-text search for code search
- **Blobs**: For vector embeddings

### Parsing
- **tree-sitter**: Deterministic AST parsing
- **Regex fallback**: For simple pattern matching
- **Plugin system**: Extensible for additional languages

### Analysis
- **Custom algorithms**: Leiden/Louvain for clustering
- **Graph traversal**: BFS/DFS with weighted paths

### Embeddings
- **Local models**: sentence-transformers (local-first)
- **Optional**: OpenAI-compatible APIs

---

## 12. Testing Strategy

### Unit Tests
- Parser tests for each language
- Graph operation tests
- Algorithm tests (blast radius, clustering)

### Integration Tests
- End-to-end graph build
- Incremental update tests
- CLI command tests

### Performance Tests
- Large repository handling (10k+ files)
- Token reduction measurement
- Update speed benchmarks

---

## 13. Acceptance Criteria

### Core Functionality
- [ ] Graph builds successfully for Python, TypeScript, Go repos
- [ ] Incremental updates complete in <2 seconds for small diffs
- [ ] Storage persists between runs
- [ ] CLI commands work as specified

### Token Efficiency
- [ ] Graph-first querying reduces token usage by 5-10x
- [ ] Context engine prioritizes security-relevant content

### Security Analysis
- [ ] Blast radius correctly identifies affected nodes
- [ ] Threat modeling generates attack paths
- [ ] Execution flow traces request lifecycle

### Visualization
- [ ] Graph exports to Mermaid, GraphML, JSON
- [ ] Watch mode triggers automatic rebuilds
- [ ] Multi-repo support works

---

## 14. Future Considerations

- Distributed graph storage (optional)
- Cloud sync (optional)
- IDE integrations (VS Code, JetBrains)
- CI/CD pipeline integration
- Additional language support (C++, Scala, etc.)