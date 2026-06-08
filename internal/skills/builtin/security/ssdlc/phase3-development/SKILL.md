---
name: phase3-development
description: Phase 3 - Secure coding, static application security testing (SAST), vulnerability skills loading, and framework-specific risks
---

# Phase 3: Development & Coding

Write secure source code, run parallel static analysis, and apply framework-specific security patterns.

## Objectives

- Write secure code according to language-specific standards
- Run parallel SAST container scanners (gosec, bandit, semgrep)
- Audit injection vectors, authentication flaws, cryptography misuse, and XXE
- Load relevant vulnerability skills before reviewing code
- Apply framework-specific security guidelines

## Phase Activities

### 3.1 SAST Scanning Pipeline

Run the appropriate static analysis tools inside isolated container runtimes:

```bash
# Parallel execution of appropriate SAST container scanners:
container_scanner --tool sast__semgrep --path .
container_scanner --tool sast__gosec --path .     # For Go codebases
container_scanner --tool sast__bandit --path .    # For Python codebases
```

### 3.2 Vulnerability Focus Areas

When reviewing source code, prioritize finding and validating:

*   **Injection Vectors**: SQL injection, command injection, OS execution, log injection.
*   **Authentication & Session Flaws**: Weak JWT verification, insecure cookie attributes, credentials in code.
*   **Cryptography Misuse**: MD5/SHA1 usage for passwords, hardcoded keys, static IVs, lack of integrity checks.
*   **Insecure Deserialization**: Use of `pickle` (Python), `unserialize` (PHP), or unsafe YAML loaders.
*   **XML External Entity (XXE)**: Unsafe parsing configurations permitting external entity resolution.
*   **Path Traversal**: Arbitrary file reads, LFI, RFI, lack of path sanitization.

### 3.3 Dynamic Skill Loading

**Always load relevant vulnerability skills BEFORE reviewing or validating code:**
- If checking JWT/OIDC implementation → load `vulnerabilities/authentication-jwt` skill.
- If checking IDOR/Access controls → load `vulnerabilities/idor` skill.
- If checking API endpoints → load `vulnerabilities/api-security` skill.

**Load framework-specific security rules based on target stack:**
- FastAPI: load `frameworks/fastapi` skill.
- Next.js: load `frameworks/nextjs` skill.
- Go Frameworks: load `frameworks/go-frameworks` skill.

### 3.4 Secure Coding Refactoring Patterns

#### Go
```go
// ❌ Insecure SQL Query
db.Query("SELECT * FROM users WHERE username = '" + username + "'")

// ✅ Secure Parameterized Query
db.Query("SELECT * FROM users WHERE username = ?", username)
```

#### Python (FastAPI)
```python
# ❌ Insecure path traversal
@app.get("/files/{filename}")
def read_file(filename: str):
    return open(f"/var/www/files/{filename}").read()

# ✅ Secure path resolution
@app.get("/files/{filename}")
def read_file(filename: str):
    safe_path = Path("/var/www/files").resolve()
    target_path = safe_path.joinpath(filename).resolve()
    if not target_path.is_relative_to(safe_path):
        raise HTTPException(status_code=400, detail="Invalid path")
    return target_path.read_text()
```

## GraphX Integration

**Continuous Code Tracking**

```bash
# Query recently added/modified code structures
GraphX: Query type="function" AND created_at > last_review

# Track data flows from public controllers to database sinks
GraphX: Query path_from="controller" AND path_to="db"
```

## Deliverables

- **Refactored Codebase**: Validated code changes.
- **Vulnerability Checklists**: Loaded skills checklists.
- **Initial SAST Log**: Preliminary scan findings.

## Exit Criteria

- [ ] Secure code standards verified.
- [ ] Language-appropriate SAST run.
- [ ] No critical/high static issues present in code.
- [ ] Framework security guidelines applied.
- [ ] GraphX updated with code symbols.