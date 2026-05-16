package handler

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	cmdbBlockedByContract = "pending"
	cmdbMaskedValue       = "******"
)

type cmdbCompatibleColumn struct {
	ID          string `json:"id"`
	Attr        string `json:"attr"`
	Field       string `json:"field"`
	Label       string `json:"label"`
	Name        string `json:"name"`
	ValueType   string `json:"value_type"`
	Tag         string `json:"tag"`
	Required    bool   `json:"required"`
	Unique      bool   `json:"unique"`
	Visible     bool   `json:"visible"`
	Conversion  bool   `json:"conversion"`
	Sensitive   bool   `json:"sensitive"`
	MaskPolicy  string `json:"mask_policy"`
	Unit        string `json:"unit"`
	SourceType  int    `json:"source_type"`
	Width       int    `json:"width"`
	Disabled    bool   `json:"disabled"`
	Dragable    bool   `json:"dragable"`
	IsDiscovery bool   `json:"discovery"`
}

type cmdbCompatibleAttributeValue struct {
	Value      any    `json:"value"`
	ValueType  string `json:"value_type"`
	Unit       string `json:"unit"`
	Conversion bool   `json:"conversion"`
	Type       string `json:"type"`
	Color      string `json:"color"`
	Sensitive  bool   `json:"sensitive"`
	Masked     bool   `json:"masked"`
}

type cmdbCompatibleInstance struct {
	ObjectID     string                                  `json:"object_id"`
	InstanceID   string                                  `json:"instance_id"`
	SortValue    int64                                   `json:"_sort_value"`
	InstanceName string                                  `json:"instance_name"`
	Attribute    map[string]cmdbCompatibleAttributeValue `json:"attribute"`
	Relation     map[string]cmdbCompatibleRelationValue  `json:"relation"`
	RawData      map[string]any                          `json:"raw_data,omitempty"`
}

type cmdbCompatibleRelationValue struct {
	Value       []any  `json:"value"`
	ValueType   string `json:"value_type"`
	Type        string `json:"type"`
	Color       string `json:"color"`
	RelationID  string `json:"relation_id"`
	PRelationID string `json:"p_relation_id"`
}

type cmdbCompatibleInstancesEnvelope struct {
	Total int64                    `json:"total"`
	List  []cmdbCompatibleInstance `json:"list"`
}

type cmdbPersistenceMeta struct {
	Status string `json:"status"`
	Driver string `json:"driver"`
}

type cmdbCompatibleMeta struct {
	Persistence cmdbPersistenceMeta `json:"persistence"`
}

type cmdbDetailGroup struct {
	Tag   string           `json:"tag"`
	Sort  int              `json:"sort"`
	Infos []cmdbDetailInfo `json:"infos"`
}

type cmdbDetailInfo struct {
	ID         string `json:"_id"`
	Label      string `json:"label"`
	Attr       string `json:"attr"`
	Value      any    `json:"value"`
	ValueType  string `json:"value_type"`
	Unit       string `json:"unit"`
	Conversion bool   `json:"conversion"`
	Required   bool   `json:"required"`
	Unique     bool   `json:"unique"`
	Readonly   bool   `json:"readonly"`
	Sort       int    `json:"sort"`
	Sensitive  bool   `json:"sensitive"`
	Masked     bool   `json:"masked"`
	MaskPolicy string `json:"mask_policy"`
	Discovery  bool   `json:"discovery"`
	IsCustom   bool   `json:"is_custom"`
}

// ListCmdbInstancesCompatible 返回接近 AutoOps 抓包结构的 CMDB 实例列表契约。
func ListCmdbInstancesCompatible(c *gin.Context) {
	objectID := c.Param("id")
	page, limit := cmdbPageAndLimit(c)
	instances, total := store.ListCmdbInstances(objectID, page, limit)
	attrs := store.ListCmdbAttributes(objectID)
	columns := buildCmdbCompatibleColumns(attrs, false)

	list := make([]cmdbCompatibleInstance, 0, len(instances))
	legacyItems := make([]model.CmdbInstance, 0, len(instances))
	for _, inst := range instances {
		list = append(list, buildCmdbCompatibleInstance(inst, attrs))
		legacyItems = append(legacyItems, buildCmdbCompatibleLegacyItem(inst, attrs))
	}

	c.JSON(http.StatusOK, gin.H{
		"tree":            buildCmdbCompatibleTree(),
		"instances":       cmdbCompatibleInstancesEnvelope{Total: total, List: list},
		"columns":         columns,
		"default_columns": buildCmdbDefaultColumns(columns),
		"table_id":        []string{"app_modules_cmdb_controllers_instance_list_object_" + objectID},
		"form_id":         "app_modules_cmdb_controllers_instance_list_object_" + objectID,
		"items":           legacyItems,
		"total":           total,
		"page":            page,
		"limit":           limit,
		"meta":            cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	})
}

// GetCmdbInstanceDetailCompatible 返回 base[].tag + infos[] 分组字段详情契约。
func GetCmdbInstanceDetailCompatible(c *gin.Context) {
	inst, ok := store.GetCmdbInstance(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "实例不存在"})
		return
	}
	attrs := store.ListCmdbAttributes(inst.ObjectID)
	base := buildCmdbDetailGroups(*inst, attrs)
	display := cmdbDefaultDisplay(*inst, attrs)

	c.JSON(http.StatusOK, gin.H{
		"instance_id":     inst.ID,
		"_creator":        inst.Creator,
		"_updater":        inst.Updater,
		"create_time":     inst.CreatedAt.Format("2006-01-02 15:04:05"),
		"update_time":     inst.UpdatedAt.Format("2006-01-02 15:04:05"),
		"base":            base,
		"default_display": display,
		"object": gin.H{
			"object_id": inst.ObjectID,
			"name":      cmdbObjectName(inst.ObjectID),
		},
		"meta": cmdbCompatibleMeta{Persistence: cmdbPersistenceStatus()},
	})
}

func cmdbPageAndLimit(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

func buildCmdbCompatibleLegacyItem(inst model.CmdbInstance, attrs []model.CmdbAttribute) model.CmdbInstance {
	raw := parseCmdbInstanceData(inst.Data)
	for _, attr := range attrs {
		sensitive, _ := cmdbAttrMaskPolicy(attr)
		if sensitive {
			raw[attr.Attr] = cmdbMaskedValue
		}
	}
	for key := range raw {
		if isSensitiveCmdbKey(key) {
			raw[key] = cmdbMaskedValue
		}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		inst.Data = "{}"
		return inst
	}
	inst.Data = string(data)
	return inst
}

func buildCmdbCompatibleTree() []gin.H {
	categories := store.ListCmdbCategories()
	objects := store.ListCmdbObjects("")
	counts := store.CountCmdbInstancesByObject()
	childrenByCategory := make(map[string][]gin.H)
	for _, obj := range objects {
		childrenByCategory[obj.CategoryID] = append(childrenByCategory[obj.CategoryID], gin.H{
			"label":       obj.Name,
			"id":          obj.ID,
			"count":       counts[obj.ID],
			"object_type": obj.ObjectType,
			"type":        "model",
			"pid":         obj.CategoryID,
		})
	}

	tree := make([]gin.H, 0, len(categories))
	for _, cat := range categories {
		tree = append(tree, gin.H{
			"label":    cat.Label,
			"id":       cat.ID,
			"type":     "category_top",
			"pid":      cat.ParentID,
			"count":    countCmdbTreeChildren(childrenByCategory[cat.ID]),
			"children": childrenByCategory[cat.ID],
		})
	}
	return tree
}

func countCmdbTreeChildren(children []gin.H) int64 {
	var total int64
	for _, child := range children {
		if count, ok := child["count"].(int64); ok {
			total += count
		}
	}
	return total
}

func buildCmdbCompatibleColumns(attrs []model.CmdbAttribute, detail bool) []cmdbCompatibleColumn {
	columns := make([]cmdbCompatibleColumn, 0, len(attrs))
	for _, attr := range attrs {
		sensitive, policy := cmdbAttrMaskPolicy(attr)
		visible := attr.Sort <= 8
		if detail {
			visible = true
		}
		columns = append(columns, cmdbCompatibleColumn{
			ID:          attr.Attr,
			Attr:        attr.Attr,
			Field:       "attribute",
			Label:       attr.Label,
			Name:        attr.Label,
			ValueType:   attr.ValueType,
			Tag:         attr.Tag,
			Required:    attr.Required,
			Unique:      attr.Unique,
			Visible:     visible,
			Conversion:  true,
			Sensitive:   sensitive,
			MaskPolicy:  policy,
			Unit:        attr.Unit,
			SourceType:  1,
			Width:       110,
			Disabled:    attr.Attr == "name",
			Dragable:    true,
			IsDiscovery: attr.Discovery,
		})
	}
	return columns
}

func buildCmdbDefaultColumns(columns []cmdbCompatibleColumn) []cmdbCompatibleColumn {
	defaultColumns := make([]cmdbCompatibleColumn, 0, len(columns))
	for _, column := range columns {
		if column.Visible {
			defaultColumns = append(defaultColumns, column)
		}
	}
	return defaultColumns
}

func buildCmdbCompatibleInstance(inst model.CmdbInstance, attrs []model.CmdbAttribute) cmdbCompatibleInstance {
	raw := parseCmdbInstanceData(inst.Data)
	attrValues := make(map[string]cmdbCompatibleAttributeValue, len(attrs))
	for _, attr := range attrs {
		value := raw[attr.Attr]
		sensitive, _ := cmdbAttrMaskPolicy(attr)
		masked := false
		if sensitive {
			value = cmdbMaskedValue
			masked = true
		}
		attrValues[attr.Attr] = cmdbCompatibleAttributeValue{
			Value:      value,
			ValueType:  attr.ValueType,
			Unit:       attr.Unit,
			Conversion: true,
			Type:       "attribute",
			Color:      "",
			Sensitive:  sensitive,
			Masked:     masked,
		}
	}
	for key, value := range raw {
		if _, ok := attrValues[key]; ok {
			continue
		}
		sensitive := isSensitiveCmdbKey(key)
		masked := false
		if sensitive {
			value = cmdbMaskedValue
			masked = true
		}
		attrValues[key] = cmdbCompatibleAttributeValue{
			Value:      value,
			ValueType:  inferCmdbValueType(value),
			Unit:       "",
			Conversion: true,
			Type:       "attribute",
			Color:      "",
			Sensitive:  sensitive,
			Masked:     masked,
		}
	}

	return cmdbCompatibleInstance{
		ObjectID:     inst.ObjectID,
		InstanceID:   inst.ID,
		SortValue:    inst.UpdatedAt.UnixMilli(),
		InstanceName: cmdbInstanceName(raw, inst.ID),
		Attribute:    attrValues,
		Relation:     map[string]cmdbCompatibleRelationValue{},
	}
}

func buildCmdbDetailGroups(inst model.CmdbInstance, attrs []model.CmdbAttribute) []cmdbDetailGroup {
	raw := parseCmdbInstanceData(inst.Data)
	groupsByTag := make(map[string][]cmdbDetailInfo)
	groupSort := make(map[string]int)
	for _, attr := range attrs {
		tag := strings.TrimSpace(attr.Tag)
		if tag == "" {
			tag = "基本信息"
		}
		if _, ok := groupSort[tag]; !ok {
			groupSort[tag] = attr.Sort
		}
		sensitive, policy := cmdbAttrMaskPolicy(attr)
		value := raw[attr.Attr]
		masked := false
		if sensitive && !isEmptyCmdbValue(value) {
			value = cmdbMaskedValue
			masked = true
		}
		groupsByTag[tag] = append(groupsByTag[tag], cmdbDetailInfo{
			ID:         attr.ID,
			Label:      attr.Label,
			Attr:       attr.Attr,
			Value:      value,
			ValueType:  attr.ValueType,
			Unit:       attr.Unit,
			Conversion: true,
			Required:   attr.Required,
			Unique:     attr.Unique,
			Readonly:   false,
			Sort:       attr.Sort,
			Sensitive:  sensitive,
			Masked:     masked,
			MaskPolicy: policy,
			Discovery:  attr.Discovery,
			IsCustom:   true,
		})
	}

	groups := make([]cmdbDetailGroup, 0, len(groupsByTag))
	for tag, infos := range groupsByTag {
		sort.Slice(infos, func(i, j int) bool { return infos[i].Sort < infos[j].Sort })
		groups = append(groups, cmdbDetailGroup{Tag: tag, Sort: groupSort[tag], Infos: infos})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Sort < groups[j].Sort })
	return groups
}

func parseCmdbInstanceData(data string) map[string]any {
	values := make(map[string]any)
	if strings.TrimSpace(data) == "" {
		return values
	}
	if err := json.Unmarshal([]byte(data), &values); err != nil {
		return values
	}
	return values
}

func cmdbAttrMaskPolicy(attr model.CmdbAttribute) (bool, string) {
	text := strings.ToLower(attr.Attr + " " + attr.Label)
	switch {
	case isSensitiveCmdbKey(attr.Attr) || isSensitiveCmdbKey(attr.Label):
		return true, "sensitive"
	case strings.Contains(text, "phone") || strings.Contains(attr.Label, "电话") || strings.Contains(attr.Label, "联系电话"):
		return true, "phone"
	case strings.Contains(text, "owner") || strings.Contains(attr.Label, "负责人"):
		return true, "person_name"
	default:
		return false, "none"
	}
}

func cmdbInstanceName(raw map[string]any, fallback string) string {
	for _, key := range []string{"name", "instance_name"} {
		if value, ok := raw[key]; ok {
			if text := strings.TrimSpace(anyToString(value)); text != "" {
				return text
			}
		}
	}
	return fallback
}

func cmdbDefaultDisplay(inst model.CmdbInstance, attrs []model.CmdbAttribute) gin.H {
	raw := parseCmdbInstanceData(inst.Data)
	label := "名称"
	attrCode := "name"
	for _, attr := range attrs {
		if attr.Attr == "name" {
			label = attr.Label
			attrCode = attr.Attr
			break
		}
	}
	return gin.H{"label": label, "attr": attrCode, "value": cmdbInstanceName(raw, inst.ID)}
}

func cmdbObjectName(objectID string) string {
	if obj, ok := store.GetCmdbObject(objectID); ok {
		return obj.Name
	}
	return objectID
}

func cmdbPersistenceStatus() cmdbPersistenceMeta {
	if store.GormOK() {
		return cmdbPersistenceMeta{Status: "ok", Driver: "gorm"}
	}
	return cmdbPersistenceMeta{Status: "blocked_by_persistence", Driver: "memory_fallback"}
}

func inferCmdbValueType(value any) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case float64, int, int64:
		return "float"
	case []any:
		return "array"
	case map[string]any:
		return "struct"
	default:
		return "char"
	}
}

func isEmptyCmdbValue(value any) bool {
	if value == nil {
		return true
	}
	return strings.TrimSpace(anyToString(value)) == ""
}

func isSensitiveCmdbKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, token := range []string{
		"password", "passwd", "secret", "token", "access_key", "secret_key", "private_key", "cookie", "session",
		"phone", "mobile", "owner", "user", "username", "account", "credential", "credentials", "cert", "certificate", "dsn", "connection_string", "connstr",
		"联系电话", "手机号", "资产负责人", "负责人", "运行用户", "连接串", "账号", "密码", "令牌", "凭据", "密钥", "私钥", "证书",
	} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}
