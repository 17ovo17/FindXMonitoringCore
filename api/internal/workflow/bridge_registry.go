package workflow

import (
	"fmt"

	"ai-workbench-api/internal/store"
	"ai-workbench-api/internal/workflow/engine"
)

// loadWorkflowGraph tries builtin first, then custom from store.
func loadWorkflowGraph(name string) (*engine.Graph, *engine.EngineConfig, error) {
	// Try builtin
	graph, cfg, err := engine.LoadBuiltinWorkflow(name)
	if err == nil {
		return graph, cfg, nil
	}

	// Try custom from store
	w, ok := store.GetWorkflow(name)
	if !ok {
		return nil, nil, fmt.Errorf("workflow %q not found (builtin err: %v)", name, err)
	}
	graph, cfg, parseErr := engine.ParseDSL([]byte(w.DSL))
	if parseErr != nil {
		return nil, nil, fmt.Errorf("parse custom workflow %q: %w", name, parseErr)
	}
	return graph, cfg, nil
}

// ListBuiltinWorkflowNames returns the names of all builtin workflows.
func ListBuiltinWorkflowNames() []string {
	return []string{
		"smart_diagnosis",
		"domain_diagnosis",
		"health_inspection",
		"metrics_insight",
		"security_compliance",
		"incident_review",
		"network_check",
		"runbook_execute",
		"knowledge_enrich",
	}
}
