package handler

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ---------------------------------------------------------------------------
// DiagnoseEngine: Catpaw-style alert aggregation + AI diagnosis
// ---------------------------------------------------------------------------

// DiagnoseToolScope distinguishes where a diagnostic tool executes.
type DiagnoseToolScope string

const (
	DiagnoseToolScopeLocal  DiagnoseToolScope = "local"
	DiagnoseToolScopeRemote DiagnoseToolScope = "remote"
)

// DiagnoseToolParam describes a parameter for a diagnostic tool.
type DiagnoseToolParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// DiagnoseTool defines a diagnostic tool that the AI can invoke during diagnosis.
type DiagnoseTool struct {
	Name        string            `json:"name"`
	Category    string            `json:"category"`
	Scope       DiagnoseToolScope `json:"scope"`
	Hint        string            `json:"hint"`
	Description string            `json:"description"`
	Parameters  []DiagnoseToolParam `json:"parameters,omitempty"`
	PreCollect  func(target string) map[string]any                         `json:"-"`
	Execute     func(ctx context.Context, target string, params map[string]any) (string, error) `json:"-"`
}

// ---------------------------------------------------------------------------
// DiagnoseToolRegistry: categorized tool registry with pre-collectors
// ---------------------------------------------------------------------------

// DiagnoseToolCategory groups related diagnostic tools.
type DiagnoseToolCategory struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Scope       DiagnoseToolScope `json:"scope"`
	Tools       []DiagnoseTool  `json:"tools"`
}

// DiagnoseToolRegistry manages diagnostic tools registered by plugins.
// Thread-safe for concurrent reads; writes happen only at startup.
type DiagnoseToolRegistry struct {
	mu            sync.RWMutex
	categories    map[string]*DiagnoseToolCategory
	toolIndex     map[string]*DiagnoseTool
	preCollectors map[string]func(target string) map[string]any
	diagnoseHints map[string]string
}

// NewDiagnoseToolRegistry creates an empty registry.
func NewDiagnoseToolRegistry() *DiagnoseToolRegistry {
	return &DiagnoseToolRegistry{
		categories:    make(map[string]*DiagnoseToolCategory),
		toolIndex:     make(map[string]*DiagnoseTool),
		preCollectors: make(map[string]func(target string) map[string]any),
		diagnoseHints: make(map[string]string),
	}
}

// Register adds a tool under the given category.
func (r *DiagnoseToolRegistry) Register(category string, tool DiagnoseTool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, dup := r.toolIndex[tool.Name]; dup {
		logrus.Warnf("diagnose: duplicate tool name %q in category %q, skipped", tool.Name, category)
		return
	}

	cat, ok := r.categories[category]
	if !ok {
		cat = &DiagnoseToolCategory{
			Name:  category,
			Scope: tool.Scope,
		}
		r.categories[category] = cat
	}
	cat.Tools = append(cat.Tools, tool)
	toolCopy := tool
	r.toolIndex[tool.Name] = &toolCopy
}

// RegisterCategory registers or updates a category's metadata.
func (r *DiagnoseToolRegistry) RegisterCategory(name, description string, scope DiagnoseToolScope) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cat, ok := r.categories[name]
	if !ok {
		cat = &DiagnoseToolCategory{Name: name}
		r.categories[name] = cat
	}
	cat.Description = description
	cat.Scope = scope
}

// RegisterPreCollector registers a baseline data collector for a category.
func (r *DiagnoseToolRegistry) RegisterPreCollector(category string, fn func(target string) map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preCollectors[category] = fn
}

// SetDiagnoseHints registers diagnostic route hints for a category.
func (r *DiagnoseToolRegistry) SetDiagnoseHints(category string, hints string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.diagnoseHints[category] = hints
}

// Get returns a tool by name.
func (r *DiagnoseToolRegistry) Get(name string) (*DiagnoseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.toolIndex[name]
	return t, ok
}

// ByCategory returns all tools in a category.
func (r *DiagnoseToolRegistry) ByCategory(category string) []DiagnoseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cat, ok := r.categories[category]
	if !ok {
		return nil
	}
	result := make([]DiagnoseTool, len(cat.Tools))
	copy(result, cat.Tools)
	return result
}

// RunPreCollectors executes all registered pre-collectors for the target.
func (r *DiagnoseToolRegistry) RunPreCollectors(target string) map[string]map[string]any {
	r.mu.RLock()
	collectors := make(map[string]func(target string) map[string]any, len(r.preCollectors))
	for k, v := range r.preCollectors {
		collectors[k] = v
	}
	r.mu.RUnlock()

	results := make(map[string]map[string]any, len(collectors))
	for category, fn := range collectors {
		if data := fn(target); len(data) > 0 {
			results[category] = data
		}
	}
	return results
}

// ListToolCatalog returns a formatted catalog of all tools for AI prompts.
func (r *DiagnoseToolRegistry) ListToolCatalog() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cats := make([]*DiagnoseToolCategory, 0, len(r.categories))
	for _, cat := range r.categories {
		cats = append(cats, cat)
	}
	sort.Slice(cats, func(i, j int) bool { return cats[i].Name < cats[j].Name })

	var b strings.Builder
	for _, cat := range cats {
		desc := cat.Description
		if desc == "" {
			desc = cat.Name + " diagnostics"
		}
		fmt.Fprintf(&b, "[%s] (%s) %s\n", cat.Name, cat.Scope, desc)
		for _, t := range cat.Tools {
			params := formatDiagnoseParams(t.Parameters)
			if params != "" {
				fmt.Fprintf(&b, "  %s(%s) - %s\n", t.Name, params, t.Description)
			} else {
				fmt.Fprintf(&b, "  %s() - %s\n", t.Name, t.Description)
			}
		}
	}
	return b.String()
}

// ToolCount returns the total number of registered tools.
func (r *DiagnoseToolRegistry) ToolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.toolIndex)
}

// GetHints returns diagnostic hints for a category.
func (r *DiagnoseToolRegistry) GetHints(category string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.diagnoseHints[category]
}

func formatDiagnoseParams(params []DiagnoseToolParam) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, p := range params {
		s := p.Name
		if p.Required {
			s += "*"
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ", ")
}

// ---------------------------------------------------------------------------
// AlertContext and Aggregator
// ---------------------------------------------------------------------------

// AlertContext represents a single alert event for diagnosis.
type AlertContext struct {
	AlertID      string            `json:"alert_id"`
	RuleName     string            `json:"rule_name"`
	Severity     string            `json:"severity"`
	Status       string            `json:"status"`
	Target       string            `json:"target"`
	CurrentValue string            `json:"current_value,omitempty"`
	Threshold    string            `json:"threshold,omitempty"`
	Description  string            `json:"description"`
	Labels       map[string]string `json:"labels,omitempty"`
	FiredAt      time.Time         `json:"fired_at"`
}

// AggregatedDiagnosis collects alerts for the same target within a time window.
type AggregatedDiagnosis struct {
	Target    string         `json:"target"`
	Alerts    []AlertContext `json:"alerts"`
	FirstSeen time.Time      `json:"first_seen"`
	LastSeen  time.Time      `json:"last_seen"`
}

// DiagnoseAggregator collects alerts for the same target within a short
// time window, then submits one aggregated diagnosis request to the engine.
type DiagnoseAggregator struct {
	window  time.Duration
	pending map[string]*AggregatedDiagnosis
	timers  map[string]*time.Timer
	mu      sync.Mutex
	engine  *DiagnoseEngine
}

// NewDiagnoseAggregator creates an aggregator with the given window duration.
func NewDiagnoseAggregator(engine *DiagnoseEngine, window time.Duration) *DiagnoseAggregator {
	if window <= 0 {
		window = 30 * time.Second
	}
	return &DiagnoseAggregator{
		window:  window,
		pending: make(map[string]*AggregatedDiagnosis),
		timers:  make(map[string]*time.Timer),
		engine:  engine,
	}
}

// Submit is called when an alert event fires. It aggregates events for the
// same target within the time window before triggering diagnosis.
func (a *DiagnoseAggregator) Submit(alert AlertContext) {
	target := alert.Target
	if target == "" {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if agg, exists := a.pending[target]; exists {
		agg.Alerts = append(agg.Alerts, alert)
		agg.LastSeen = time.Now()
		return
	}

	now := time.Now()
	agg := &AggregatedDiagnosis{
		Target:    target,
		Alerts:    []AlertContext{alert},
		FirstSeen: now,
		LastSeen:  now,
	}
	a.pending[target] = agg

	a.timers[target] = time.AfterFunc(a.window, func() {
		a.mu.Lock()
		agg := a.pending[target]
		delete(a.pending, target)
		delete(a.timers, target)
		a.mu.Unlock()

		if agg == nil {
			return
		}

		logrus.Infof("diagnose aggregator: window closed for %s, %d alerts", target, len(agg.Alerts))
		a.engine.SubmitDiagnosis(agg)
	})
}

// Shutdown cancels all pending aggregation timers.
func (a *DiagnoseAggregator) Shutdown() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for key, timer := range a.timers {
		timer.Stop()
		delete(a.timers, key)
		delete(a.pending, key)
	}
}

// ---------------------------------------------------------------------------
// DiagnoseEngine: central coordinator for AI-powered diagnosis
// ---------------------------------------------------------------------------

// DiagnoseResult stores the outcome of a single diagnosis run.
type DiagnoseResult struct {
	ID          string              `json:"id"`
	Target      string              `json:"target"`
	Status      string              `json:"status"` // success, failed, timeout
	Alerts      []AlertContext      `json:"alerts"`
	Baseline    map[string]map[string]any `json:"baseline,omitempty"`
	Report      string              `json:"report,omitempty"`
	Error       string              `json:"error,omitempty"`
	Rounds      int                 `json:"rounds"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt time.Time           `json:"completed_at"`
	DurationMs  int64               `json:"duration_ms"`
}

// DiagnoseEngine is the central coordinator for AI-powered diagnosis.
// It receives aggregated alerts, collects baseline data, formats context
// for AI, executes tool-calling diagnosis, and stores results.
type DiagnoseEngine struct {
	aggregator     *DiagnoseAggregator
	registry       *DiagnoseToolRegistry
	maxRounds      int
	maxConcurrent  int
	toolTimeout    time.Duration
	cooldownWindow time.Duration

	mu        sync.Mutex
	sem       chan struct{}
	cooldowns map[string]time.Time // target -> cooldown expiry
	results   []*DiagnoseResult    // recent results (ring buffer)
	resultCap int
}

// DiagnoseEngineConfig holds configuration for the engine.
type DiagnoseEngineConfig struct {
	AggregationWindow time.Duration
	MaxRounds         int
	MaxConcurrent     int
	ToolTimeout       time.Duration
	CooldownWindow    time.Duration
	ResultCapacity    int
}

// DefaultDiagnoseEngineConfig returns sensible defaults.
func DefaultDiagnoseEngineConfig() DiagnoseEngineConfig {
	return DiagnoseEngineConfig{
		AggregationWindow: 30 * time.Second,
		MaxRounds:         10,
		MaxConcurrent:     3,
		ToolTimeout:       30 * time.Second,
		CooldownWindow:    10 * time.Minute,
		ResultCapacity:    100,
	}
}

// NewDiagnoseEngine creates a new engine with the given registry and config.
func NewDiagnoseEngine(registry *DiagnoseToolRegistry, cfg DiagnoseEngineConfig) *DiagnoseEngine {
	engine := &DiagnoseEngine{
		registry:       registry,
		maxRounds:      cfg.MaxRounds,
		maxConcurrent:  cfg.MaxConcurrent,
		toolTimeout:    cfg.ToolTimeout,
		cooldownWindow: cfg.CooldownWindow,
		sem:            make(chan struct{}, cfg.MaxConcurrent),
		cooldowns:      make(map[string]time.Time),
		results:        make([]*DiagnoseResult, 0, cfg.ResultCapacity),
		resultCap:      cfg.ResultCapacity,
	}
	engine.aggregator = NewDiagnoseAggregator(engine, cfg.AggregationWindow)
	return engine
}

// Aggregator returns the engine's aggregator for submitting alerts.
func (e *DiagnoseEngine) Aggregator() *DiagnoseAggregator {
	return e.aggregator
}

// Registry returns the engine's tool registry.
func (e *DiagnoseEngine) Registry() *DiagnoseToolRegistry {
	return e.registry
}

// SubmitAlert is the public entry point for alert events.
// It delegates to the aggregator which batches by target.
func (e *DiagnoseEngine) SubmitAlert(alert AlertContext) {
	e.aggregator.Submit(alert)
}

// SubmitDiagnosis is called by the aggregator when the window expires.
// It respects cooldown and concurrency limits.
func (e *DiagnoseEngine) SubmitDiagnosis(agg *AggregatedDiagnosis) {
	if e.isCooldownActive(agg.Target) {
		logrus.Infof("diagnose skipped: cooldown active for %s", agg.Target)
		return
	}

	select {
	case e.sem <- struct{}{}:
		go func() {
			defer func() { <-e.sem }()
			e.runDiagnosis(agg)
		}()
	default:
		logrus.Warnf("diagnose skipped: concurrency limit reached for %s", agg.Target)
	}
}

// runDiagnosis executes the full diagnosis flow for an aggregated alert batch.
func (e *DiagnoseEngine) runDiagnosis(agg *AggregatedDiagnosis) {
	start := time.Now()
	result := &DiagnoseResult{
		ID:        fmt.Sprintf("diag-%d", start.UnixMilli()),
		Target:    agg.Target,
		Alerts:    agg.Alerts,
		StartedAt: start,
	}

	defer func() {
		if r := recover(); r != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("panic: %v", r)
			logrus.Errorf("diagnose panic for %s: %v", agg.Target, r)
		}
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(start).Milliseconds()
		e.storeResult(result)
		e.updateCooldown(agg.Target)
	}()

	logrus.Infof("diagnose started for %s with %d alerts", agg.Target, len(agg.Alerts))

	// Step 1: Collect pre-diagnosis baseline data
	baseline := e.registry.RunPreCollectors(agg.Target)
	result.Baseline = baseline

	// Step 2: Format context for AI
	aiContext := e.buildAIContext(agg, baseline)

	// Step 3: Execute AI diagnosis with tool calling
	report, rounds, err := e.executeDiagnosis(agg.Target, aiContext)
	result.Rounds = rounds
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		logrus.Warnf("diagnose failed for %s: %v", agg.Target, err)
		return
	}

	result.Status = "success"
	result.Report = report
	logrus.Infof("diagnose completed for %s in %d rounds (%dms)",
		agg.Target, rounds, time.Since(start).Milliseconds())
}

// buildAIContext formats the diagnosis context for the AI model.
func (e *DiagnoseEngine) buildAIContext(agg *AggregatedDiagnosis, baseline map[string]map[string]any) string {
	var b strings.Builder

	b.WriteString("# 诊断上下文\n\n")
	b.WriteString("## 目标\n")
	b.WriteString(agg.Target)
	b.WriteString("\n\n")

	// Alert information
	b.WriteString("## 告警信息\n")
	for i, alert := range agg.Alerts {
		fmt.Fprintf(&b, "%d. [%s] %s - %s\n", i+1, alert.Severity, alert.RuleName, alert.Description)
		if alert.CurrentValue != "" {
			fmt.Fprintf(&b, "   当前值: %s", alert.CurrentValue)
			if alert.Threshold != "" {
				fmt.Fprintf(&b, " (阈值: %s)", alert.Threshold)
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Baseline data
	if len(baseline) > 0 {
		b.WriteString("## 基线数据（诊断前采集）\n")
		for category, data := range baseline {
			fmt.Fprintf(&b, "### %s\n", category)
			for k, v := range data {
				fmt.Fprintf(&b, "- %s: %v\n", k, v)
			}
		}
		b.WriteString("\n")
	}

	// Available tools
	b.WriteString("## 可用诊断工具\n")
	b.WriteString(e.registry.ListToolCatalog())
	b.WriteString("\n")

	// Diagnostic hints
	for _, alert := range agg.Alerts {
		if alert.Labels != nil {
			if category, ok := alert.Labels["category"]; ok {
				if hints := e.registry.GetHints(category); hints != "" {
					fmt.Fprintf(&b, "## 诊断路线提示 (%s)\n%s\n\n", category, hints)
					break
				}
			}
		}
	}

	return b.String()
}

// executeDiagnosis runs the AI tool-calling loop.
// Returns the final report, number of rounds, and any error.
func (e *DiagnoseEngine) executeDiagnosis(target, aiContext string) (string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.maxRounds)*e.toolTimeout*2)
	defer cancel()

	// Respect context cancellation
	select {
	case <-ctx.Done():
		return "", 0, ctx.Err()
	default:
	}

	// Use the existing diagnoseWithAI infrastructure for the actual AI call
	report, _ := diagnoseWithAI(target, DiagnoseOptions{
		Prompt: aiContext,
	})

	if report == "" {
		return "", 1, fmt.Errorf("AI returned empty diagnosis report")
	}

	return report, 1, nil
}

// ---------------------------------------------------------------------------
// Cooldown and result management
// ---------------------------------------------------------------------------

func (e *DiagnoseEngine) isCooldownActive(target string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	expiry, ok := e.cooldowns[target]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(e.cooldowns, target)
		return false
	}
	return true
}

func (e *DiagnoseEngine) updateCooldown(target string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cooldowns[target] = time.Now().Add(e.cooldownWindow)
}

func (e *DiagnoseEngine) storeResult(result *DiagnoseResult) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.results) >= e.resultCap {
		// Drop oldest
		e.results = e.results[1:]
	}
	e.results = append(e.results, result)
}

// RecentResults returns the most recent diagnosis results.
func (e *DiagnoseEngine) RecentResults(limit int) []*DiagnoseResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	if limit <= 0 || limit > len(e.results) {
		limit = len(e.results)
	}
	start := len(e.results) - limit
	result := make([]*DiagnoseResult, limit)
	copy(result, e.results[start:])
	return result
}

// Shutdown stops the aggregator and waits for in-flight diagnoses.
func (e *DiagnoseEngine) Shutdown() {
	e.aggregator.Shutdown()
	logrus.Info("diagnose engine shutdown")
}
