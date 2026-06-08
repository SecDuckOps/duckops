package compression

type Tokenizer struct {
	charsPerToken float64
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		charsPerToken: 4.0,
	}
}

func (t *Tokenizer) EstimateText(text string) int {
	if text == "" {
		return 0
	}
	return int(float64(len(text)) / t.charsPerToken) + 1
}

func (t *Tokenizer) EstimateMessages(messages []Message) int {
	total := 0
	for _, msg := range messages {
		total += t.EstimateText(msg.Content)
		// role and metadata overhead
		total += 2
	}
	return total
}

func (t *Tokenizer) EstimateToolOutputs(outputs []ToolOutput) int {
	total := 0
	for _, o := range outputs {
		total += t.EstimateText(o.Input)
		total += t.EstimateText(o.Output)
		total += 2
	}
	return total
}

func (t *Tokenizer) EstimateCodeSnippets(snippets []CodeSnippet) int {
	total := 0
	for _, s := range snippets {
		total += t.EstimateText(s.Content)
		// file path and language overhead
		total += 4
	}
	return total
}

func (t *Tokenizer) EstimateLogs(logs []string) int {
	total := 0
	for _, l := range logs {
		total += t.EstimateText(l)
	}
	return total
}

func (t *Tokenizer) EstimateContext(ctx Context) int {
	total := 0
	total += t.EstimateText(ctx.SystemPrompt)
	total += t.EstimateText(ctx.AgentMemory)
	total += t.EstimateText(ctx.Config)
	total += t.EstimateText(ctx.SecurityRules)
	total += t.EstimateMessages(ctx.Messages)
	total += t.EstimateToolOutputs(ctx.ToolOutputs)
	total += t.EstimateCodeSnippets(ctx.CodeSnippets)
	total += t.EstimateLogs(ctx.Logs)
	return total
}

func (t *Tokenizer) EstimateCompressed(cc CompressedContext) int {
	total := 0
	total += t.EstimateText(cc.SystemPrompt)
	total += t.EstimateMessages(cc.Messages)
	total += t.EstimateToolOutputs(cc.ToolOutputs)
	total += t.EstimateText(cc.Summary)
	total += t.EstimateCodeSnippets(cc.CodeSnippets)
	return total
}

func (t *Tokenizer) EstimateString(s string) int {
	return t.EstimateText(s)
}
