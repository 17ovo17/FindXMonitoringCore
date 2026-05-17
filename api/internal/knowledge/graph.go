package knowledge

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"ai-workbench-api/internal/embedding"
)

// EntityType 实体类型
type EntityType string

const (
	EntityService  EntityType = "service"
	EntityHost     EntityType = "host"
	EntityMetric   EntityType = "metric"
	EntityAlert    EntityType = "alert"
	EntityRunbook  EntityType = "runbook"
	EntityIncident EntityType = "incident"
	EntityConfig   EntityType = "config"
)

// RelationType 关系类型
type RelationType string

const (
	RelDependsOn RelationType = "depends_on"  // 服务依赖
	RelMonitors  RelationType = "monitors"    // 指标监控
	RelTriggers  RelationType = "triggers"    // 告警触发
	RelResolves  RelationType = "resolves"    // Runbook 解决
	RelAffects   RelationType = "affects"     // 影响范围
	RelRunsOn    RelationType = "runs_on"     // 运行在
	RelCauses    RelationType = "causes"      // 导致
)

// Entity 知识图谱实体
type Entity struct {
	ID         string            `json:"id"`
	Type       EntityType        `json:"type"`
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
}

// Relation 知识图谱关系
type Relation struct {
	Source   string       `json:"source"`
	Target   string       `json:"target"`
	Type     RelationType `json:"type"`
	Weight   float64      `json:"weight"`   // 关系强度
	Evidence string       `json:"evidence"` // 来源证据
}

// GraphStats 图谱统计
type GraphStats struct {
	EntityCount   int            `json:"entity_count"`
	RelationCount int            `json:"relation_count"`
	TypeCounts    map[string]int `json:"type_counts"`
}

// PathResult 路径查询结果
type PathResult struct {
	Entities  []Entity `json:"entities"`
	Relations []Relation `json:"relations"`
	Length    int      `json:"length"`
}

// KnowledgeGraph 运维知识图谱
type KnowledgeGraph struct {
	entities  map[string]*Entity
	relations []Relation
	// 邻接表：entity_id -> []Relation
	adjacency map[string][]Relation
	mu        sync.RWMutex
}

// globalGraph 全局知识图谱实例
var (
	globalGraph *KnowledgeGraph
	graphOnce   sync.Once
)

// GetGraph 获取全局知识图谱实例
func GetGraph() *KnowledgeGraph {
	graphOnce.Do(func() {
		globalGraph = NewKnowledgeGraph()
	})
	return globalGraph
}

// NewKnowledgeGraph 创建知识图谱
func NewKnowledgeGraph() *KnowledgeGraph {
	return &KnowledgeGraph{
		entities:  make(map[string]*Entity),
		relations: make([]Relation, 0),
		adjacency: make(map[string][]Relation),
	}
}

// AddEntity 添加实体
func (g *KnowledgeGraph) AddEntity(entity Entity) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if entity.Properties == nil {
		entity.Properties = make(map[string]string)
	}
	g.entities[entity.ID] = &entity
}

// GetEntity 获取实体
func (g *KnowledgeGraph) GetEntity(id string) (*Entity, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	e, ok := g.entities[id]
	return e, ok
}

// AddRelation 添加关系
func (g *KnowledgeGraph) AddRelation(rel Relation) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if rel.Weight == 0 {
		rel.Weight = 1.0
	}
	g.relations = append(g.relations, rel)
	g.adjacency[rel.Source] = append(g.adjacency[rel.Source], rel)
	// 双向索引（方便反向查询）
	reverseRel := Relation{
		Source:   rel.Target,
		Target:   rel.Source,
		Type:     rel.Type,
		Weight:   rel.Weight,
		Evidence: rel.Evidence,
	}
	g.adjacency[rel.Target] = append(g.adjacency[rel.Target], reverseRel)
}

// QueryNeighbors 查询实体的邻居（1-2 跳）
func (g *KnowledgeGraph) QueryNeighbors(entityID string, maxHops int) ([]Entity, []Relation) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if maxHops <= 0 {
		maxHops = 2
	}

	visited := make(map[string]bool)
	visited[entityID] = true
	var resultEntities []Entity
	var resultRelations []Relation

	// BFS
	queue := []string{entityID}
	for hop := 0; hop < maxHops && len(queue) > 0; hop++ {
		var nextQueue []string
		for _, nodeID := range queue {
			rels := g.adjacency[nodeID]
			for _, rel := range rels {
				targetID := rel.Target
				resultRelations = append(resultRelations, rel)
				if !visited[targetID] {
					visited[targetID] = true
					if e, ok := g.entities[targetID]; ok {
						resultEntities = append(resultEntities, *e)
					}
					nextQueue = append(nextQueue, targetID)
				}
			}
		}
		queue = nextQueue
	}
	return resultEntities, resultRelations
}

// FindPath 查找两个实体之间的最短路径（BFS）
func (g *KnowledgeGraph) FindPath(sourceID, targetID string) *PathResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if sourceID == targetID {
		if e, ok := g.entities[sourceID]; ok {
			return &PathResult{Entities: []Entity{*e}, Length: 0}
		}
		return nil
	}

	// BFS 找最短路径
	type pathNode struct {
		id   string
		path []string
		rels []Relation
	}

	visited := map[string]bool{sourceID: true}
	queue := []pathNode{{id: sourceID, path: []string{sourceID}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// 限制搜索深度
		if len(current.path) > 6 {
			continue
		}

		for _, rel := range g.adjacency[current.id] {
			nextID := rel.Target
			if nextID == targetID {
				// 找到路径
				finalPath := append(current.path, nextID)
				finalRels := append(current.rels, rel)
				entities := make([]Entity, 0, len(finalPath))
				for _, eid := range finalPath {
					if e, ok := g.entities[eid]; ok {
						entities = append(entities, *e)
					}
				}
				return &PathResult{
					Entities:  entities,
					Relations: finalRels,
					Length:    len(finalRels),
				}
			}
			if !visited[nextID] {
				visited[nextID] = true
				newPath := make([]string, len(current.path)+1)
				copy(newPath, current.path)
				newPath[len(current.path)] = nextID
				newRels := make([]Relation, len(current.rels)+1)
				copy(newRels, current.rels)
				newRels[len(current.rels)] = rel
				queue = append(queue, pathNode{id: nextID, path: newPath, rels: newRels})
			}
		}
	}
	return nil // 无路径
}

// ListEntities 列出所有实体（可按类型过滤）
func (g *KnowledgeGraph) ListEntities(entityType EntityType) []Entity {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []Entity
	for _, e := range g.entities {
		if entityType == "" || e.Type == entityType {
			result = append(result, *e)
		}
	}
	return result
}

// ListRelations 列出所有关系
func (g *KnowledgeGraph) ListRelations() []Relation {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]Relation, len(g.relations))
	copy(out, g.relations)
	return out
}

// Stats 获取图谱统计
func (g *KnowledgeGraph) Stats() GraphStats {
	g.mu.RLock()
	defer g.mu.RUnlock()

	typeCounts := make(map[string]int)
	for _, e := range g.entities {
		typeCounts[string(e.Type)]++
	}
	return GraphStats{
		EntityCount:   len(g.entities),
		RelationCount: len(g.relations),
		TypeCounts:    typeCounts,
	}
}

// ExtractFromDocument 从文档中自动提取实体和关系
// 使用规则匹配提取运维领域实体（服务名、主机名、指标名等）
func (g *KnowledgeGraph) ExtractFromDocument(docID, title, content string) ([]Entity, []Relation) {
	var entities []Entity
	var relations []Relation

	// 提取服务名（常见模式：xxxService, xxx-service, xxx_service）
	serviceNames := extractServiceNames(content)
	for _, name := range serviceNames {
		eid := fmt.Sprintf("svc_%s", strings.ToLower(name))
		e := Entity{ID: eid, Type: EntityService, Name: name, Properties: map[string]string{"source_doc": docID}}
		entities = append(entities, e)
		g.AddEntity(e)
	}

	// 提取主机/IP
	hosts := extractHosts(content)
	for _, host := range hosts {
		eid := fmt.Sprintf("host_%s", strings.ReplaceAll(host, ".", "_"))
		e := Entity{ID: eid, Type: EntityHost, Name: host, Properties: map[string]string{"source_doc": docID}}
		entities = append(entities, e)
		g.AddEntity(e)
	}

	// 提取指标名（prometheus 风格：xxx_xxx_total, xxx_xxx_seconds）
	metrics := extractMetricNames(content)
	for _, m := range metrics {
		eid := fmt.Sprintf("metric_%s", m)
		e := Entity{ID: eid, Type: EntityMetric, Name: m, Properties: map[string]string{"source_doc": docID}}
		entities = append(entities, e)
		g.AddEntity(e)
	}

	// 建立文档内实体间的关系
	for i := 0; i < len(serviceNames); i++ {
		svcID := fmt.Sprintf("svc_%s", strings.ToLower(serviceNames[i]))
		// 服务 → 指标
		for _, m := range metrics {
			if strings.Contains(strings.ToLower(m), strings.ToLower(serviceNames[i])) {
				rel := Relation{Source: svcID, Target: fmt.Sprintf("metric_%s", m), Type: RelMonitors, Weight: 0.8, Evidence: docID}
				relations = append(relations, rel)
				g.AddRelation(rel)
			}
		}
		// 服务 → 主机
		for _, h := range hosts {
			rel := Relation{Source: svcID, Target: fmt.Sprintf("host_%s", strings.ReplaceAll(h, ".", "_")), Type: RelRunsOn, Weight: 0.5, Evidence: docID}
			relations = append(relations, rel)
			g.AddRelation(rel)
		}
	}

	return entities, relations
}

// GraphRAG 图增强检索：先找到相关实体，再检索关联文档
func (g *KnowledgeGraph) GraphRAG(ctx context.Context, query string, topK int) []embedding.SearchResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 1. 从 query 中提取可能的实体名
	queryLower := strings.ToLower(query)
	var matchedEntities []string

	for id, e := range g.entities {
		nameLower := strings.ToLower(e.Name)
		if strings.Contains(queryLower, nameLower) || strings.Contains(nameLower, queryLower) {
			matchedEntities = append(matchedEntities, id)
		}
	}

	if len(matchedEntities) == 0 {
		return nil
	}

	// 2. 收集相关实体的文档来源
	docIDs := make(map[string]float64)
	for _, eid := range matchedEntities {
		// 直接实体的文档
		if e, ok := g.entities[eid]; ok {
			if src, exists := e.Properties["source_doc"]; exists {
				docIDs[src] += 1.0
			}
		}
		// 邻居实体的文档（1 跳）
		for _, rel := range g.adjacency[eid] {
			if neighbor, ok := g.entities[rel.Target]; ok {
				if src, exists := neighbor.Properties["source_doc"]; exists {
					docIDs[src] += rel.Weight * 0.5
				}
			}
		}
	}

	// 3. 转换为搜索结果
	var results []embedding.SearchResult
	for docID, score := range docIDs {
		results = append(results, embedding.SearchResult{
			DocID: docID,
			Score: score,
		})
	}

	// 按分数排序
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if len(results) > topK {
		results = results[:topK]
	}
	return results
}

// RemoveEntity 删除实体及其关系
func (g *KnowledgeGraph) RemoveEntity(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.entities, id)
	// 清理关系
	var remaining []Relation
	for _, rel := range g.relations {
		if rel.Source != id && rel.Target != id {
			remaining = append(remaining, rel)
		}
	}
	g.relations = remaining
	// 重建邻接表
	g.adjacency = make(map[string][]Relation)
	for _, rel := range g.relations {
		g.adjacency[rel.Source] = append(g.adjacency[rel.Source], rel)
		reverseRel := Relation{Source: rel.Target, Target: rel.Source, Type: rel.Type, Weight: rel.Weight, Evidence: rel.Evidence}
		g.adjacency[rel.Target] = append(g.adjacency[rel.Target], reverseRel)
	}
}

// --- 实体提取辅助函数 ---

// extractServiceNames 从文本中提取服务名
func extractServiceNames(content string) []string {
	var names []string
	seen := make(map[string]bool)

	// 匹配常见服务命名模式
	words := strings.Fields(content)
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'`")
		lower := strings.ToLower(w)
		if (strings.HasSuffix(lower, "service") || strings.HasSuffix(lower, "-svc") ||
			strings.Contains(lower, "-service") || strings.Contains(lower, "_service")) &&
			len(w) > 4 && len(w) < 50 {
			if !seen[lower] {
				seen[lower] = true
				names = append(names, w)
			}
		}
	}
	return names
}

// extractHosts 从文本中提取主机名/IP
func extractHosts(content string) []string {
	var hosts []string
	seen := make(map[string]bool)

	words := strings.Fields(content)
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'`")
		if isIPv4(w) && !seen[w] {
			seen[w] = true
			hosts = append(hosts, w)
		}
	}
	return hosts
}

// extractMetricNames 从文本中提取 Prometheus 风格指标名
func extractMetricNames(content string) []string {
	var metrics []string
	seen := make(map[string]bool)

	words := strings.Fields(content)
	for _, w := range words {
		w = strings.Trim(w, ".,;:!?()[]{}\"'`")
		if isPrometheusMetric(w) && !seen[w] {
			seen[w] = true
			metrics = append(metrics, w)
		}
	}
	return metrics
}

// isIPv4 简单判断是否为 IPv4 地址
func isIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}
	for _, p := range parts {
		if len(p) == 0 || len(p) > 3 {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// isPrometheusMetric 判断是否为 Prometheus 风格指标名
func isPrometheusMetric(s string) bool {
	if len(s) < 5 || len(s) > 100 {
		return false
	}
	underscoreCount := 0
	for _, c := range s {
		if c == '_' {
			underscoreCount++
		} else if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	// 至少有一个下划线，且以常见后缀结尾
	if underscoreCount == 0 {
		return false
	}
	lower := strings.ToLower(s)
	suffixes := []string{"_total", "_seconds", "_bytes", "_count", "_sum", "_bucket", "_gauge", "_ratio", "_percent", "_info"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return underscoreCount >= 2 // 至少两个下划线也认为可能是指标
}
