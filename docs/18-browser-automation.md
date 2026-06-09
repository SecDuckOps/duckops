# File: docs/18-browser-automation.md

# Chapter 18: Browser Automation

## 18.1 Overview

DuckOps provides browser automation capabilities through Playwright and the CDP (Chrome DevTools Protocol). The automation system enables the AI agent to interact with web pages — navigating, clicking, typing, extracting data, and capturing screenshots.

### 18.1.1 Architecture

```
+-----------+      JSON-RPC      +-------------+
|  DuckOps  | <----------------> |  Playwright  |
|  Agent    |     over stdio     |   MCP Server |
+-----------+                    +-------------+
                                         |
                                    +----------+
                                    |  Browser  |
                                    | (Chromium)|
                                    +----------+
```

## 18.2 Playwright MCP Integration

DuckOps uses the Playwright MCP server (`@playwright/mcp`) for browser automation. The server is configured as an MCP server in `duckops.json`:

```json
{
  "mcps": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp"],
      "env": {
        "PLAYWRIGHT_BROWSERS_PATH": "0",
        "DISPLAY": ":99"
      }
    }
  }
}
```

### Available Playwright Tools

| Tool | Description |
|------|-------------|
| `browser_navigate` | Navigate to a URL |
| `browser_click` | Click an element by selector |
| `browser_type` | Type text into an input |
| `browser_snapshot` | Get page snapshot (accessibility tree) |
| `browser_screenshot` | Capture a screenshot |
| `browser_evaluate` | Execute JavaScript in page context |
| `browser_set_locale` | Set browser locale |
| `browser_press_key` | Press a keyboard key |

## 18.3 Puppeteer-Based Alternative

For environments where Playwright is not available, DuckOps also supports a Puppeteer-based approach via the `web-scraper` tool:

```go
// internal/tools/web_scraper.go
func (t *WebScraper) Execute(ctx context.Context, params json.RawMessage) (string, error) {
    var req struct {
        URL string `json:"url"`
    }
    json.Unmarshal(params, &req)

    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Launch headless browser
    // Navigate and wait for page load
    // Extract text content
    // Take screenshot
    // Return structured data
}
```

This tool is used for simpler use cases like documentation scraping and content extraction.

## 18.4 Use Cases

### Automated Web Testing

The AI can execute test scenarios against web applications:

```
User: "Test the login flow on staging.example.com"

Agent:
1. browser_navigate("https://staging.example.com/login")
2. browser_type("[name=email]", "test@example.com")
3. browser_type("[name=password]", "password123")
4. browser_click("[type=submit]")
5. browser_snapshot() → verify dashboard loaded
```

### Documentation Scraping

```
User: "Read the FastAPI documentation about dependencies"

Agent:
1. browser_navigate("https://fastapi.tiangolo.com/tutorial/dependencies/")
2. browser_snapshot() → extract page content
3. Return formatted documentation
```

### Visual Regression Spot-Checking

```
User: "Check if the homepage renders correctly"

Agent:
1. browser_navigate("https://example.com")
2. browser_screenshot("fullpage")
3. Return screenshot to user for visual inspection
```

### Form Filling Automation

```
User: "Fill out this license registration form"

Agent:
1. browser_navigate(form_url)
2. browser_type("#name", "John Doe")
3. browser_type("#email", "john@example.com")
4. browser_type("#company", "Acme Corp")
5. browser_click("[type=checkbox]")  // agree to terms
6. browser_click("[type=submit]")
```

## 18.5 Headless Mode

By default, Playwright runs in headless mode. For debugging, set:

```json
{
  "args": ["@playwright/mcp", "--headless=false"]
}
```

For server environments without a display, use `xvfb`:

```yaml
services:
  duckops:
    environment:
      - DISPLAY=:99
    volumes:
      - /tmp/.X11-unix:/tmp/.X11-unix:ro
```

## 18.6 Stealth Mode

For sites that block headless browsers, DuckOps supports stealth configurations:

```go
// Additional browser launch args for stealth
const stealthArgs = []string{
    "--disable-blink-features=AutomationControlled",
    "--no-sandbox",
    "--disable-dev-shm-usage",
    "--disable-web-security",
}
```

These flags reduce the chance of bot detection by masking headless browser fingerprints.

## 18.7 Security Considerations

### URL Allowlist

Administrators can restrict navigation to specific domains:

```json
{
  "options": {
    "browser": {
      "allowed_domains": ["*.example.com", "docs.python.org"],
      "blocked_domains": ["internal.company.com"],
      "max_pages": 10
    }
  }
}
```

### Timeout Controls

Browser operations are subject to strict timeouts to prevent hanging:

| Operation | Default Timeout |
|-----------|----------------|
| Page navigation | 30 seconds |
| Element wait | 10 seconds |
| Screenshot | 15 seconds |
| JavaScript evaluation | 5 seconds |

### No Credential Storage

Browser automation never stores credentials. Forms are filled by the AI based on user-provided information. Credentials are not persisted in session history.

## 18.8 Browser Installation

Playwright browsers must be installed separately, either globally or via `npx`:

```bash
# Install Chromium for Playwright
npx playwright install chromium

# Verify installation
npx playwright install --dry-run
```

For Docker deployments, browsers are installed in the build stage:

```dockerfile
FROM mcr.microsoft.com/playwright:v1.52.0 AS playwright-browsers
# Pre-installed: Chromium, Firefox, WebKit

FROM duckops:latest
COPY --from=playwright-browsers /home/duckops/.cache/ms-playwright /home/duckops/.cache/ms-playwright
```

## 18.9 Limitations

- **Single Page**: Playwright MCP operates on one page at a time
- **No Multi-Tab**: Multiple browser tabs are not supported in the current integration
- **No File Downloads**: File download automation is not implemented
- **JavaScript-Heavy Sites**: Some SPAs may require explicit waitForSelector calls

---

*Next: Chapter 19 - Context Compression*
