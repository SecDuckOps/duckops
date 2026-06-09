# File: docs/19-context-compression.md

# Chapter 19: Context Compression

## 19.1 Overview

Context compression prevents token overflow in long-running conversations. DuckOps implements a multi-layered strategy that balances conversation coherence against token budget constraints.

### 19.1.1 Why Context Compression?

| Problem | Consequence | Solution |
|---------|-------------|----------|
| Token limit exceeded | Conversation truncated, AI loses context | Auto-summarization |
| High API costs | Expensive long prompts | Selective trimming |
| Degraded response quality | Model loses focus on recent messages | Priority-based retention |
| Session resumption | Full history must fit in context window | Progressive summarization |

## 19.2 Architecture

```
┌──────────────────────────────────────────────┐
│            Context Budget Manager              │
│                                                │
│  1. Estimate total tokens in session           │
│  2. Compare against model's context window     │
│  3. If over threshold (80%):                   │
│     a. Generate summary of oldest messages     │
│     b. Replace them with summary message       │
│     c. Continue with full recent context       │
│                                                │
│  Threshold: model.MaxTokens * 0.80             │
│  Max summarizations per session: unlimited     │
└──────────────────────────────────────────────┘
```

## 19.3 Summary Generation

When the context budget is exceeded, DuckOps uses the **small model** to generate a summary:

```go
func (a *SessionAgent) Summarize(ctx context.Context) error {
    // 1. Select messages to summarize (oldest 50% of history)
    toSummarize := a.messages[:len(a.messages)/2]

    // 2. Build summarization prompt
    prompt := "Summarize the following conversation, preserving:\n"
    prompt += "- Key decisions made\n"
    prompt += "- Files created or modified\n"
    prompt += "- Requirements gathered\n"
    prompt += "- Current task status\n\n"
    prompt += formatMessages(toSummarize)

    // 3. Call small model for summary
    summary, err := a.small.Complete(ctx, prompt)

    // 4. Replace summarized messages with summary message
    summaryMsg := Message{
        Role: "system",
        Parts: []ContentPart{
            TextContent{Text: "[Context Summary] " + summary},
        },
    }

    a.messages = append([]Message{summaryMsg}, a.messages[len(toSummarize):]...)

    // 5. Save to database
    return a.db.SaveSummary(ctx, a.sessionID, summary)
}
```

### Summary Content

Each summary preserves:

| Element | Preserved? | Format |
|---------|-----------|--------|
| Code changes | Yes | File path + diff summary |
| Architecture decisions | Yes | Decision + rationale |
| User preferences | Yes | Full context preserved |
| Error details | Yes | Error + resolution |
| Security requirements | Yes | Full context preserved |
| Tool outputs | Partial | Key results, truncated |
| AI reasoning | Partial | Main conclusions |

## 19.4 Compression Strategies

### Strategy 1: Auto-Summarize (Default)

Triggered automatically at 80% of the model's context window.

```
Before: [250 messages, 120K tokens] → exceeds 128K limit (94%)
After:  [Summary msg + 125 messages, 100K tokens] → within 128K (78%)
```

### Strategy 2: Selective Truncation

Remove low-value content first:

```go
type TruncationPriority int

const (
    PrioritySystem     TruncationPriority = 5  // Keep at all costs
    PriorityDecision   TruncationPriority = 4  // Keep
    PriorityCode       TruncationPriority = 3  // Keep
    PriorityToolOutput TruncationPriority = 2  // Truncate first if long
    PriorityReasoning  TruncationPriority = 1  // Summarize if needed
)
```

1. Start with oldest reasoning blocks
2. Truncate tool outputs to their summaries
3. Remove resolved todo items
4. Summarize code discussions

### Strategy 3: Hierarchical Summarization

For very long sessions, DuckOps maintains a summary tree:

```
Session Root Summary
  ├── Phase 1: Requirements (10 messages) → Summary
  ├── Phase 2: Implementation (50 messages) → Summary
  │     ├── Day 1: Auth module (15 messages) → Summary
  │     ├── Day 2: API routes (20 messages) → Summary
  │     └── Day 3: Testing (15 messages) → Summary
  └── Phase 3: Review (5 messages) → Full context
```

Each phase is summarized when archived. The root summary provides a high-level overview, while deeper levels retain more detail.

## 19.5 Token Estimation

DuckOps estimates token counts without an external tokenizer:

```go
func EstimateTokens(text string) int {
    // ~4 characters per token for English text
    // ~1.5 characters per token for code
    return int(float64(len(text)) / 3.5)
}
```

For models with known tokenizers (like `gpt-4o` and `claude-opus-4`), DuckOps uses the Fantasy SDK's built-in token estimation, which is model-aware.

### Model Context Windows

| Model | Context Window | Safety Threshold (80%) |
|-------|---------------|----------------------|
| GPT-4o | 128,000 | 102,400 |
| GPT-4o-mini | 128,000 | 102,400 |
| Claude Opus 4 | 200,000 | 160,000 |
| Claude Sonnet 4 | 200,000 | 160,000 |
| Gemini 2.5 Pro | 1,000,000+ | 800,000 |
| Llama 3 (local) | 8,000 | 6,400 |

## 19.6 Compression Events

When compression occurs, the UI displays a notification:

```
[System] Context compressed: 45 old messages summarized.
128,000 → 95,000 tokens (26% reduction)
```

The summary message is collapsible in the TUI:

```
▸ [Context Summary] (click to expand)
  User asked to implement JWT authentication...
```

## 19.7 Manual Compression

Users can trigger compression manually:

```bash
duckops session summarize <session-id>
```

Or from the TUI:

- `Ctrl+S` — Summarize current session
- `/summarize` — Chat command to trigger compression

## 19.8 Configuration

```json
{
  "options": {
    "disable-auto-summarize": false,
    "context": {
      "threshold": 0.80,
      "strategy": "auto",
      "min-messages-before-summary": 20
    }
  }
}
```

| Option | Default | Description |
|--------|---------|-------------|
| `disable-auto-summarize` | `false` | Disable automatic compression |
| `threshold` | `0.80` | Fraction of context window before compression |
| `strategy` | `"auto"` | Compression strategy (`auto`, `selective`, `hierarchical`) |
| `min-messages-before-summary` | `20` | Minimum messages before first compression |

## 19.9 Data Preservation

All original messages remain in the SQLite database after compression. The compression only affects what is sent to the AI provider. Users can:

- Expand summary messages to see original content
- Export full session history to JSON/Markdown
- Reconstruct uncompressed context for review

The compression is purely a **context window optimization** — no data loss occurs.

## 19.10 Performance Impact

| Session Length | Before Compression | After Compression | Reduction |
|---------------|-------------------|-------------------|-----------|
| 20 messages | 8K tokens | 8K tokens | 0% |
| 50 messages | 25K tokens | 25K tokens | 0% (below threshold) |
| 100 messages | 60K tokens | 45K tokens | 25% |
| 250 messages | 128K tokens | 95K tokens | 26% |
| 500 messages | 200K tokens | 120K tokens | 40% |

Compression triggers approximately once every 50-100 messages in typical usage, depending on message length and context window size.

---

*Next: Chapter 20 - Implementation Plan*
