package handler

import "testing"

func TestWhiteboxAIOpsV2TopologyTriggerByQuestion(t *testing.T) {
	cases := []string{
		"10.0.1.21 和 10.0.1.22 CPU 都很高",
		"Redis 异常影响哪些链路",
		"生成 clims 拓扑",
	}
	for _, input := range cases {
		graph, ok := aiopsTopologyGraphForQuestion(input)
		if !ok {
			t.Fatalf("expected topology trigger for %q", input)
		}
		if graph.Summary.Planner == "" {
			t.Fatalf("topology summary planner should be set for %q: %+v", input, graph.Summary)
		}
	}
	if _, ok := aiopsTopologyGraphForQuestion("单台机器 CPU 高"); ok {
		t.Fatal("generic single-machine diagnosis should not force topology update")
	}
}
