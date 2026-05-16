package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func cmdbRelationSchema(rel model.CmdbInstanceRelation, relationType model.CmdbRelationType, from model.CmdbInstance, to model.CmdbInstance) gin.H {
	return gin.H{
		"_id":                  rel.RelationTypeID,
		"instance_relation_id": rel.ID,
		"relation_id":          rel.ID,
		"left_object_id":       from.ObjectID,
		"left_id":              to.ObjectID,
		"left_name":            cmdbObjectName(to.ObjectID),
		"left_groups":          []string{},
		"right_object_id":      to.ObjectID,
		"right_id":             from.ObjectID,
		"right_name":           cmdbObjectName(from.ObjectID),
		"right_groups":         []string{},
		"asst_id":              cmdbRelationAssocID(relationType),
		"asst_name":            cmdbRelationDisplayName(relationType),
		"left_asst_name":       firstNonEmptyRelationString(relationType.LeftAsstName, cmdbRelationDisplayName(relationType)),
		"right_asst_name":      firstNonEmptyRelationString(relationType.RightAsstName, cmdbRelationDisplayName(relationType)),
		"mapping":              firstNonEmptyRelationString(relationType.Mapping, "n:1"),
		"visible":              cmdbRelationVisible(relationType),
		"description":          relationType.Description,
		"rule_logic":           firstNonEmptyRelationString(relationType.RuleLogic, "and"),
		"rule_expression":      relationType.RuleExpression,
		"rules":                cmdbRelationRules(relationType.RulesJSON),
		"right_min":            relationType.RightMin,
		"left_min":             relationType.LeftMin,
		"left_max":             relationType.LeftMax,
		"right_max":            relationType.RightMax,
		"left_object_type":     cmdbObjectType(from.ObjectID),
		"right_object_type":    cmdbObjectType(to.ObjectID),
		"source":               relationType.Source,
		"left_object_name":     cmdbObjectName(from.ObjectID),
		"right_object_name":    cmdbObjectName(to.ObjectID),
	}
}

func cmdbRelationSchemaFromAction(item model.CmdbRelationActionRequest) (gin.H, bool) {
	for _, rel := range store.ListCmdbInstanceRelations(item.InstanceID) {
		if rel.ID != item.RelationID {
			continue
		}
		relationType, ok := cmdbRelationType(rel.RelationTypeID)
		if !ok {
			return nil, false
		}
		return cmdbRelationSchemaForRelation(rel, relationType)
	}
	return nil, false
}

func cmdbRelationSchemaForRelation(rel model.CmdbInstanceRelation, relationType model.CmdbRelationType) (gin.H, bool) {
	source, sourceOK := store.GetCmdbInstance(rel.SourceInstanceID)
	target, targetOK := store.GetCmdbInstance(rel.TargetInstanceID)
	if !sourceOK || !targetOK {
		return nil, false
	}
	return cmdbRelationSchema(rel, relationType, *source, *target), true
}

func cmdbRelationPathLabel(nodes []string, relationName string) string {
	if len(nodes) == 0 {
		return ""
	}
	if relationName == "" {
		relationName = "relation"
	}
	return strings.Join(nodes, "-("+relationName+")")
}

func cmdbRelationSchemaMetadata(schema gin.H) map[string]string {
	meta := map[string]string{}
	for _, key := range []string{
		"asst_id", "asst_name", "mapping", "rule_logic", "rule_expression",
		"left_object_id", "right_object_id", "left_id", "right_id",
	} {
		if value := strings.TrimSpace(anyToString(schema[key])); value != "" {
			meta["cmdb_relation_"+key] = value
		}
	}
	meta["cmdb_relation_visible"] = strconv.FormatBool(schema["visible"] == true)
	return meta
}

func cmdbRelationVisible(relationType model.CmdbRelationType) bool {
	if relationType.Visible == nil {
		return true
	}
	return *relationType.Visible
}

func cmdbRelationRules(raw string) []any {
	if strings.TrimSpace(raw) == "" {
		return []any{}
	}
	var rules []any
	if err := json.Unmarshal([]byte(raw), &rules); err != nil {
		return []any{}
	}
	return rules
}

func firstNonEmptyRelationString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
