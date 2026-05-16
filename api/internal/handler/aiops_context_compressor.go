package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	compressMaxChars    = 2000
	compressMaxItems    = 20
	compressSummarySize = 10

	// ContextCompressor defaults
	defaultMaxTokens       = 128000
	defaultHeadProtect     = 3
	defaultTailProtect     = 20
	defaultToolResultLimit = 4000
	defaultCharsPerToken   = 4

	// Anti-thrashing: minimum savings percentage to consider compression effective
	minEffectiveSavingsPct = 10.0
	// Maximum consecutive ineffective compressions before skipping
	maxIneffectiveCount = 2
)

// summaryPrefix is prepended to compressed context summaries.
const summaryPrefix = "[CONTEXT COMPACTION] Earlier turns were compacted into the summary below. " +
	"Treat as background reference. Respond ONLY to the latest user message after this summary."

// prunedToolPlaceholder replaces old tool results that have been pruned.
const prunedToolPlaceholder = "[Old tool output cleared to save context space]"

// ---------------------------------------------------------------------------
// Message type for conversation-level compression
// ---------------------------------------------------------------------------

// Message represents a single message in the conversation context.
type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolID    string     `json:"tool_id,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool invocation within an assistant message.
type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Args     string `json:"args"`
}

// estimateTokens returns a rough token count for a message.
func (m *Message) estimateTokens() int {
	tokens := len(m.Content) / defaultCharsPerToken
	tokens += 10 // role/metadata overhead
	for _, tc := range m.ToolCalls {
		tokens += len(tc.Args) / defaultCharsPerToken
	}
	return tokens
}

// ---------------------------------------------------------------------------
// ContextCompressor: full Hermes-style conversation compressor
// ---------------------------------------------------------------------------

// ContextCompressor implements the Hermes context compression algorithm:
// 1. Prune duplicate tool results (keep newest full, replace older with summary)
// 2. Truncate oversized tool results while preserving JSON validity
// 3. If still over budget: protect head + tail, summarize middle
// 4. Anti-thrashing check (skip if last 2 compressions saved <10%)
type ContextCompressor struct {
	MaxTokens       int // total token budget for the context window
	HeadProtect     int // number of messages to protect at start
	TailProtect     int // number of messages to protect at end
	ToolResultLimit int // max chars for a single tool result before summarizing

	mu                          sync.Mutex
	compressionCount            int
	lastCompressionSavingsPct   float64
	ineffectiveCompressionCount int
	previousSummary             string
}

// NewContextCompressor creates a compressor with sensible defaults.
func NewContextCompressor() *ContextCompressor {
	return &ContextCompressor{
		MaxTokens:       defaultMaxTokens,
		HeadProtect:     defaultHeadProtect,
		TailProtect:     defaultTailProtect,
		ToolResultLimit: defaultToolResultLimit,
	}
}

// ShouldCompress returns true if the messages exceed the token budget
// and anti-thrashing conditions allow compression.
func (c *ContextCompressor) ShouldCompress(messages []Message) bool {
	total := c.estimateTotalTokens(messages)
	threshold := c.MaxTokens * 50 / 100 // compress at 50% of max
	if total < threshold {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ineffectiveCompressionCount >= maxIneffectiveCount {
		return false
	}
	return true
}

// Compress applies the full Hermes compression algorithm to the message list.
func (c *ContextCompressor) Compress(messages []Message) []Message {
	if len(messages) <= c.HeadProtect+c.TailProtect+1 {
		return messages
	}

	tokensBefore := c.estimateTotalTokens(messages)

	// Phase 1: Prune duplicate tool results (keep newest full copy)
	messages = c.pruneOldToolResults(messages)

	// Phase 2: Truncate oversized tool results
	messages = c.truncateToolResults(messages)

	// Phase 3: Check if still over budget; if so, summarize middle
	tokensAfterPrune := c.estimateTotalTokens(messages)
	threshold := c.MaxTokens * 50 / 100
	if tokensAfterPrune > threshold {
		messages = c.summarizeMiddle(messages)
	}

	// Phase 4: Anti-thrashing tracking
	tokensAfter := c.estimateTotalTokens(messages)
	c.mu.Lock()
	if tokensBefore > 0 {
		savingsPct := float64(tokensBefore-tokensAfter) / float64(tokensBefore) * 100
		c.lastCompressionSavingsPct = savingsPct
		if savingsPct < minEffectiveSavingsPct {
			c.ineffectiveCompressionCount++
		} else {
			c.ineffectiveCompressionCount = 0
		}
	}
	c.compressionCount++
	c.mu.Unlock()

	return messages
}

// estimateTotalTokens returns a rough token count for the entire message list.
func (c *ContextCompressor) estimateTotalTokens(messages []Message) int {
	total := 0
	for i := range messages {
		total += messages[i].estimateTokens()
	}
	return total
}

// pruneOldToolResults deduplicates tool results and replaces old ones with
// 1-line summaries. Keeps only the newest full copy of duplicate results.
func (c *ContextCompressor) pruneOldToolResults(messages []Message) []Message {
	result := make([]Message, len(messages))
	copy(result, messages)

	// Build tool_call_id -> (name, args) index from assistant messages
	callIndex := make(map[string][2]string) // tool_call_id -> [name, args]
	for i := range result {
		if result[i].Role == "assistant" {
			for _, tc := range result[i].ToolCalls {
				callIndex[tc.ID] = [2]string{tc.Name, tc.Args}
			}
		}
	}

	// Determine prune boundary: everything before (len - TailProtect) is pruneable
	pruneBoundary := len(result) - c.TailProtect
	if pruneBoundary < c.HeadProtect {
		pruneBoundary = c.HeadProtect
	}

	// Pass 1: Deduplicate identical tool results (keep newest full copy)
	contentHashes := make(map[string]int) // hash -> newest index
	for i := len(result) - 1; i >= 0; i-- {
		msg := &result[i]
		if msg.Role != "tool" || msg.ToolID == "" {
			continue
		}
		if len(msg.Content) < 200 {
			continue
		}
		h := md5Short(msg.Content)
		if _, seen := contentHashes[h]; seen {
			// This is an older duplicate — replace with back-reference
			msg.Content = "[Duplicate tool output — same content as a more recent call]"
		} else {
			contentHashes[h] = i
		}
	}

	// Pass 2: Replace old tool results (before prune boundary) with summaries
	for i := c.HeadProtect; i < pruneBoundary; i++ {
		msg := &result[i]
		if msg.Role != "tool" || msg.ToolID == "" {
			continue
		}
		if msg.Content == prunedToolPlaceholder {
			continue
		}
		if strings.HasPrefix(msg.Content, "[Duplicate tool output") {
			continue
		}
		if len(msg.Content) <= 200 {
			continue
		}
		// Generate informative 1-line summary
		info := callIndex[msg.ToolID]
		msg.Content = summarizeToolResult(info[0], info[1], msg.Content)
	}

	// Pass 3: Truncate large tool_call arguments in assistant messages
	// outside the protected tail, preserving JSON validity
	for i := c.HeadProtect; i < pruneBoundary; i++ {
		msg := &result[i]
		if msg.Role != "assistant" || len(msg.ToolCalls) == 0 {
			continue
		}
		for j := range msg.ToolCalls {
			if len(msg.ToolCalls[j].Args) > 500 {
				msg.ToolCalls[j].Args = truncateToolCallArgsJSON(msg.ToolCalls[j].Args, 200)
			}
		}
	}

	return result
}

// truncateToolResults shortens tool result content that exceeds ToolResultLimit
// while preserving key information at head and tail.
func (c *ContextCompressor) truncateToolResults(messages []Message) []Message {
	for i := range messages {
		if messages[i].Role != "tool" {
			continue
		}
		content := messages[i].Content
		if len(content) <= c.ToolResultLimit {
			continue
		}
		// Already pruned
		if strings.HasPrefix(content, "[") && len(content) < 200 {
			continue
		}
		headSize := c.ToolResultLimit * 75 / 100
		tailSize := c.ToolResultLimit * 20 / 100
		omitted := len(content) - headSize - tailSize
		messages[i].Content = content[:headSize] +
			fmt.Sprintf("\n\n... [省略 %d 字符] ...\n\n", omitted) +
			content[len(content)-tailSize:]
	}
	return messages
}

// summarizeMiddle protects head and tail messages, replaces the middle with
// a structured summary message.
func (c *ContextCompressor) summarizeMiddle(messages []Message) []Message {
	n := len(messages)
	headEnd := c.HeadProtect
	if headEnd > n {
		headEnd = n
	}
	tailStart := n - c.TailProtect
	if tailStart < headEnd+1 {
		tailStart = headEnd + 1
	}
	if tailStart >= n {
		return messages
	}

	// Extract middle section to summarize
	middle := messages[headEnd:tailStart]
	if len(middle) == 0 {
		return messages
	}

	// Build structured summary of the middle section
	summary := c.buildMiddleSummary(middle)

	// Assemble: head + summary message + tail
	compressed := make([]Message, 0, headEnd+1+(n-tailStart))
	compressed = append(compressed, messages[:headEnd]...)
	compressed = append(compressed, Message{
		Role:    "assistant",
		Content: summaryPrefix + "\n\n" + summary,
	})
	compressed = append(compressed, messages[tailStart:]...)

	c.mu.Lock()
	c.previousSummary = summary
	c.mu.Unlock()

	return compressed
}

// buildMiddleSummary creates a structured summary of the middle messages.
// This is a local summarization (no LLM call) that extracts key information.
func (c *ContextCompressor) buildMiddleSummary(middle []Message) string {
	var b strings.Builder

	c.mu.Lock()
	prevSummary := c.previousSummary
	c.mu.Unlock()

	if prevSummary != "" {
		b.WriteString("## Previous Context\n")
		b.WriteString(prevSummary)
		b.WriteString("\n\n## New Activity Since Last Compaction\n")
	}

	// Collect user requests, assistant actions, and tool results
	var userRequests []string
	var actions []string
	var toolResults []string

	for _, msg := range middle {
		switch msg.Role {
		case "user":
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			userRequests = append(userRequests, content)
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					argsPreview := tc.Args
					if len(argsPreview) > 100 {
						argsPreview = argsPreview[:100] + "..."
					}
					actions = append(actions, fmt.Sprintf("[%s] %s", tc.Name, argsPreview))
				}
			}
			if msg.Content != "" {
				content := msg.Content
				if len(content) > 150 {
					content = content[:150] + "..."
				}
				actions = append(actions, "Response: "+content)
			}
		case "tool":
			content := msg.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			toolResults = append(toolResults, content)
		}
	}

	if len(userRequests) > 0 {
		b.WriteString("## User Requests\n")
		for _, r := range userRequests {
			b.WriteString("- ")
			b.WriteString(r)
			b.WriteString("\n")
		}
	}
	if len(actions) > 0 {
		b.WriteString("\n## Actions Taken\n")
		for _, a := range actions {
			b.WriteString("- ")
			b.WriteString(a)
			b.WriteString("\n")
		}
	}
	if len(toolResults) > 0 {
		b.WriteString("\n## Tool Results (summarized)\n")
		limit := 10
		if len(toolResults) < limit {
			limit = len(toolResults)
		}
		for _, r := range toolResults[:limit] {
			b.WriteString("- ")
			b.WriteString(r)
			b.WriteString("\n")
		}
		if len(toolResults) > 10 {
			b.WriteString(fmt.Sprintf("- ... and %d more tool results\n", len(toolResults)-10))
		}
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

// md5Short returns a short MD5 hash (12 hex chars) for deduplication.
func md5Short(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:6])
}

// summarizeToolResult creates an informative 1-line summary of a tool call + result.
func summarizeToolResult(toolName, toolArgs, content string) string {
	contentLen := len(content)
	lineCount := strings.Count(content, "\n") + 1

	var args map[string]any
	if toolArgs != "" {
		_ = json.Unmarshal([]byte(toolArgs), &args)
	}
	if args == nil {
		args = map[string]any{}
	}

	switch toolName {
	case "prometheus_query", "prometheus_query_range":
		query, _ := args["query"].(string)
		if len(query) > 60 {
			query = query[:57] + "..."
		}
		return fmt.Sprintf("[%s] query=%q (%d chars result)", toolName, query, contentLen)
	case "catpaw_check", "catpaw_diagnose":
		ip, _ := args["ip"].(string)
		return fmt.Sprintf("[%s] ip=%s (%d lines output)", toolName, ip, lineCount)
	case "alert_query":
		return fmt.Sprintf("[alert_query] (%d chars, %d lines)", contentLen, lineCount)
	case "knowledge_search", "runbook_query":
		kw, _ := args["query"].(string)
		if kw == "" {
			kw, _ = args["keyword"].(string)
		}
		return fmt.Sprintf("[%s] query=%q (%d chars)", toolName, kw, contentLen)
	case "remote_exec":
		cmd, _ := args["command"].(string)
		if len(cmd) > 60 {
			cmd = cmd[:57] + "..."
		}
		return fmt.Sprintf("[remote_exec] `%s` (%d lines output)", cmd, lineCount)
	default:
		// Generic fallback
		var firstArg string
		for k, v := range args {
			sv := fmt.Sprint(v)
			if len(sv) > 40 {
				sv = sv[:40]
			}
			firstArg = fmt.Sprintf(" %s=%s", k, sv)
			break
		}
		return fmt.Sprintf("[%s]%s (%d chars result)", toolName, firstArg, contentLen)
	}
}

// truncateToolCallArgsJSON shrinks long string values inside a tool-call
// arguments JSON blob while preserving JSON validity.
func truncateToolCallArgsJSON(argsJSON string, headChars int) string {
	var parsed any
	if err := json.Unmarshal([]byte(argsJSON), &parsed); err != nil {
		return argsJSON // not valid JSON, return as-is
	}
	shrunk := shrinkJSONValues(parsed, headChars)
	b, err := json.Marshal(shrunk)
	if err != nil {
		return argsJSON
	}
	return string(b)
}

// shrinkJSONValues recursively truncates string values in a parsed JSON structure.
func shrinkJSONValues(obj any, headChars int) any {
	switch v := obj.(type) {
	case string:
		if len(v) > headChars {
			return v[:headChars] + "...[truncated]"
		}
		return v
	case map[string]any:
		result := make(map[string]any, len(v))
		for k, val := range v {
			result[k] = shrinkJSONValues(val, headChars)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = shrinkJSONValues(val, headChars)
		}
		return result
	default:
		return obj
	}
}

// ---------------------------------------------------------------------------
// Original tool output compression (used by AIToolExecute and CompressForLLM)
// ---------------------------------------------------------------------------

// CompressToolOutput compresses large tool outputs into structured summaries
// before sending to LLM. If output > 2000 chars, summarize; if array > 20 items,
// take top 10 + summary. Preserves key fields (error messages, metric values, timestamps).
func CompressToolOutput(output any) any {
	if output == nil {
		return output
	}
	switch v := output.(type) {
	case map[string]any:
		return compressMap(v)
	case []any:
		return compressSlice(v)
	default:
		text := fmt.Sprint(output)
		if len(text) > compressMaxChars {
			return compressText(text)
		}
		return output
	}
}

func compressMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if isKeyField(k) {
			result[k] = v
			continue
		}
		switch typed := v.(type) {
		case []any:
			result[k] = compressSlice(typed)
		case map[string]any:
			result[k] = compressMap(typed)
		case string:
			if len(typed) > compressMaxChars {
				result[k] = compressText(typed)
			} else {
				result[k] = typed
			}
		default:
			result[k] = v
		}
	}
	return result
}

func compressSlice(items []any) any {
	if len(items) <= compressMaxItems {
		return items
	}
	summary := make([]any, 0, compressSummarySize+1)
	for i := 0; i < compressSummarySize && i < len(items); i++ {
		summary = append(summary, items[i])
	}
	summary = append(summary, map[string]any{
		"_compressed": true,
		"_total":      len(items),
		"_shown":      compressSummarySize,
		"_omitted":    len(items) - compressSummarySize,
		"_note":       fmt.Sprintf("显示前 %d 条，共 %d 条结果", compressSummarySize, len(items)),
	})
	return summary
}

func compressText(text string) string {
	runes := []rune(text)
	if len(runes) <= compressMaxChars {
		return text
	}
	head := string(runes[:1500])
	tail := string(runes[len(runes)-300:])
	omitted := len(runes) - 1800
	return head + fmt.Sprintf("\n\n... [省略 %d 字符] ...\n\n", omitted) + tail
}

// isKeyField returns true for fields that should always be preserved in full.
func isKeyField(key string) bool {
	preserveKeys := []string{
		"error", "err", "message", "msg",
		"timestamp", "time", "created_at", "updated_at",
		"value", "metric", "score", "severity",
		"status", "id", "name", "title",
		"ip", "host", "target", "source",
		"total", "count",
	}
	lower := strings.ToLower(key)
	for _, k := range preserveKeys {
		if lower == k {
			return true
		}
	}
	return false
}

// CompressForLLM compresses a full context payload (multiple tool results)
// into a size suitable for LLM context windows.
func CompressForLLM(data any, maxLen int) string {
	if maxLen <= 0 {
		maxLen = compressMaxChars
	}
	compressed := CompressToolOutput(data)
	b, err := json.Marshal(compressed)
	if err != nil {
		return fmt.Sprint(data)
	}
	text := string(b)
	if len(text) <= maxLen {
		return text
	}
	return compressText(text)
}
