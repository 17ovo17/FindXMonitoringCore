package handler

import (
	"net/http"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	resourceGroupMaxNameLength = 120
	resourceGroupMaxTextLength = 500
	hostLabelWorkspaceID       = "workspace_id"
	hostLabelResourceGroupID   = "resource_group_id"
	hostLabelTags              = "tags"
	resourceGroupStorageError  = "resource group storage unavailable"
)

var resourceGroupStatuses = map[string]bool{
	"active":   true,
	"disabled": true,
	"archived": true,
}

func ListResourceGroups(c *gin.Context) { c.JSON(http.StatusOK, store.ListResourceGroups()) }

func CreateResourceGroup(c *gin.Context) {
	group, ok := bindResourceGroupInput(c, nil)
	if !ok {
		return
	}
	saved, err := store.SaveResourceGroup(group)
	if err != nil {
		respondResourceGroupStorageError(c)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func GetResourceGroup(c *gin.Context) {
	group, ok, err := store.GetResourceGroup(c.Param("id"))
	if err != nil {
		respondResourceGroupStorageError(c)
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func UpdateResourceGroup(c *gin.Context) {
	existing, ok, err := store.GetResourceGroup(c.Param("id"))
	if err != nil {
		respondResourceGroupStorageError(c)
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource group not found"})
		return
	}
	group, inputOK := bindResourceGroupInput(c, &existing)
	if !inputOK {
		return
	}
	saved, err := store.SaveResourceGroup(group)
	if err != nil {
		respondResourceGroupStorageError(c)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func DeleteResourceGroup(c *gin.Context) {
	groupID := strings.TrimSpace(c.Param("id"))
	for _, target := range store.ListMonitorTargets() {
		if strings.TrimSpace(target.Labels[hostLabelResourceGroupID]) == groupID {
			c.JSON(http.StatusConflict, gin.H{"error": "resource group is in use"})
			return
		}
	}
	deleted, err := store.DeleteResourceGroup(c.Param("id"))
	if err != nil {
		respondResourceGroupStorageError(c)
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource group not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListHostAssets(c *gin.Context) {
	targets := store.ListMonitorTargets()
	agents := agentsByTarget()
	out := make([]model.HostAsset, 0, len(targets))
	for _, target := range targets {
		asset := hostAssetFromTarget(target, agents[target.ID])
		applyCmdbHostFields(&asset)
		if matchHostAssetQuery(c, asset) {
			out = append(out, asset)
		}
	}
	c.JSON(http.StatusOK, out)
}

func GetHostAsset(c *gin.Context) {
	target, ok := getHostTarget(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "host asset not found"})
		return
	}
	asset := hostAssetFromTarget(target, agentsByTarget()[target.ID])
	applyCmdbHostFields(&asset)
	c.JSON(http.StatusOK, asset)
}

func UpdateHostAssetTags(c *gin.Context) {
	target, ok := getHostTarget(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "host asset not found"})
		return
	}
	var input model.HostTagsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	tags := normalizeAssetStrings(input.Tags)
	updateTargetLabel(target, hostLabelTags, strings.Join(tags, ","))
	respondUpdatedHost(c, target, "cmdb.host.tags.assign", "CMDB host tags assigned from resource list", map[string]any{
		"tags": tags,
	})
}

func UpdateHostAssetResourceGroup(c *gin.Context) {
	target, ok := getHostTarget(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "host asset not found"})
		return
	}
	var input model.HostResourceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	groupID := strings.TrimSpace(input.ResourceGroupID)
	if groupID != "" {
		if _, found, err := store.GetResourceGroup(groupID); err != nil {
			respondResourceGroupStorageError(c)
			return
		} else if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource group not found"})
			return
		}
	}
	updateTargetLabel(target, hostLabelResourceGroupID, groupID)
	respondUpdatedHost(c, target, "cmdb.host.resource_group.assign", "CMDB host resource group assigned from resource list", map[string]any{
		"resource_group_id": groupID,
	})
}

func UpdateHostAssetWorkspace(c *gin.Context) {
	target, ok := getHostTarget(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "host asset not found"})
		return
	}
	var input model.HostWorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	workspaceID := strings.TrimSpace(input.WorkspaceID)
	updateTargetLabel(target, hostLabelWorkspaceID, workspaceID)
	respondUpdatedHost(c, target, "cmdb.host.workspace.assign", "CMDB host business workspace assigned from resource list", map[string]any{
		"workspace_id": workspaceID,
	})
}

func bindResourceGroupInput(c *gin.Context, existing *model.ResourceGroup) (model.ResourceGroup, bool) {
	var input model.ResourceGroupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return model.ResourceGroup{}, false
	}
	group, err := normalizeResourceGroupInput(input, existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return model.ResourceGroup{}, false
	}
	return group, true
}

func normalizeResourceGroupInput(input model.ResourceGroupInput, existing *model.ResourceGroup) (model.ResourceGroup, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.ResourceGroup{}, errResourceGroupValidation("name is required")
	}
	if len([]rune(name)) > resourceGroupMaxNameLength {
		return model.ResourceGroup{}, errResourceGroupValidation("name is too long")
	}
	description := strings.TrimSpace(input.Description)
	if len([]rune(description)) > resourceGroupMaxTextLength {
		return model.ResourceGroup{}, errResourceGroupValidation("description is too long")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	if !resourceGroupStatuses[status] {
		return model.ResourceGroup{}, errResourceGroupValidation("unsupported status")
	}
	group := model.ResourceGroup{
		Name:        name,
		Description: description,
		WorkspaceID: strings.TrimSpace(input.WorkspaceID),
		ParentID:    strings.TrimSpace(input.ParentID),
		Status:      status,
		Tags:        normalizeAssetStrings(input.Tags),
	}
	if existing != nil {
		group.ID = existing.ID
		group.CreatedAt = existing.CreatedAt
	}
	return group, nil
}

func matchHostAssetQuery(c *gin.Context, asset model.HostAsset) bool {
	if workspaceID := strings.TrimSpace(c.Query("workspace_id")); workspaceID != "" && asset.WorkspaceID != workspaceID {
		return false
	}
	groupID := strings.TrimSpace(c.Query("resource_group_id"))
	if groupID != "" && asset.ResourceGroupID != groupID {
		return false
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" && asset.Status != status {
		return false
	}
	if online := strings.TrimSpace(c.Query("online")); online != "" {
		want := strings.EqualFold(online, "true") || online == "1"
		if (asset.AgentStatus == "online" || asset.Status == "online") != want {
			return false
		}
	}
	if tags := append(c.QueryArray("tag"), c.QueryArray("tags")...); len(tags) > 0 && !hostAssetHasTag(asset, tags) {
		return false
	}
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))
	if keyword == "" {
		return true
	}
	values := append([]string{asset.Ident, asset.Hostname, asset.AgentID}, asset.IPList...)
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), keyword) {
			return true
		}
	}
	return false
}

func hostAssetHasTag(asset model.HostAsset, tags []string) bool {
	seen := map[string]bool{}
	for _, tag := range asset.Tags {
		seen[strings.ToLower(tag)] = true
	}
	for _, tag := range tags {
		for _, item := range strings.Split(tag, ",") {
			if seen[strings.ToLower(strings.TrimSpace(item))] {
				return true
			}
		}
	}
	return false
}

func agentsByTarget() map[string]*model.FindXAgent {
	out := map[string]*model.FindXAgent{}
	for _, agent := range store.ListFindXAgents() {
		if agent.TargetID != "" {
			out[agent.TargetID] = agent
		}
	}
	return out
}

func getHostTarget(id string) (*model.MonitorTarget, bool) {
	return store.GetMonitorTarget(strings.TrimSpace(id))
}

func hostAssetFromTarget(target *model.MonitorTarget, agent *model.FindXAgent) model.HostAsset {
	labels := safeHostLabels(target.Labels)
	asset := model.HostAsset{
		HostID:          target.ID,
		Ident:           target.Ident,
		Hostname:        firstAssetValue(target.Hostname, target.Name),
		IPList:          splitAssetValues(target.IP),
		OS:              target.OS,
		Arch:            target.Arch,
		WorkspaceID:     labels[hostLabelWorkspaceID],
		ResourceGroupID: labels[hostLabelResourceGroupID],
		Tags:            splitAssetValues(labels[hostLabelTags]),
		Status:          target.Status,
		Source:          target.Source,
		Labels:          labels,
		UpdatedAt:       target.UpdatedAt,
	}
	if target.LastSeen != nil {
		asset.LastSeenAt = target.LastSeen
	}
	applyAgentSummary(&asset, agent)
	return asset
}

func applyAgentSummary(asset *model.HostAsset, agent *model.FindXAgent) {
	if agent == nil {
		return
	}
	asset.AgentID = agent.ID
	asset.AgentStatus = agent.Status
	asset.AgentVersion = agent.Version
	if asset.LastSeenAt == nil || agent.LastSeen.After(*asset.LastSeenAt) {
		asset.LastSeenAt = &agent.LastSeen
	}
}

func respondUpdatedHost(c *gin.Context, target *model.MonitorTarget, action string, summary string, details map[string]any) {
	saved, err := store.UpsertMonitorTarget(target)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid host asset"})
		return
	}
	if _, err := store.AddMonitorAuditLog(model.MonitorAuditLog{
		Actor:        requestActor(c),
		Action:       action,
		ResourceType: "host_asset",
		ResourceID:   saved.ID,
		Scope:        "cmdb",
		Status:       "ok",
		ClientIP:     c.ClientIP(),
		Summary:      summary,
		Details: map[string]any{
			"host_ident":        saved.Ident,
			"host_name":         firstAssetValue(saved.Hostname, saved.Name),
			"ip_count":          len(splitAssetValues(saved.IP)),
			"assignment":        details,
			"assignment_source": "cmdb_resource_list",
		},
	}); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "host asset audit unavailable"})
		return
	}
	asset := hostAssetFromTarget(saved, agentsByTarget()[saved.ID])
	applyCmdbHostFields(&asset)
	c.JSON(http.StatusOK, asset)
}

func applyCmdbHostFields(asset *model.HostAsset) {
	if asset == nil {
		return
	}
	inst, attrs, ok := resolveHostCmdbInstance(*asset)
	if !ok {
		return
	}
	raw := parseCmdbInstanceData(inst.Data)
	values := make(map[string]any, len(raw))
	for key, value := range raw {
		if isSensitiveCmdbKey(key) {
			continue
		}
		values[key] = value
	}
	columns := make([]model.HostAssetCmdbColumn, 0, len(attrs))
	for _, attr := range attrs {
		sensitive, _ := cmdbAttrMaskPolicy(attr)
		if sensitive {
			values[attr.Attr] = cmdbMaskedValue
		} else if value, ok := raw[attr.Attr]; ok {
			values[attr.Attr] = value
		}
		columns = append(columns, model.HostAssetCmdbColumn{
			Attr:      attr.Attr,
			Label:     attr.Label,
			ValueType: attr.ValueType,
			Tag:       attr.Tag,
			Unit:      attr.Unit,
			Sort:      attr.Sort,
			Visible:   attr.Sort <= 8 || attr.Discovery,
			Sensitive: sensitive,
			Masked:    sensitive,
		})
	}
	asset.CmdbInstance = &model.HostAssetCmdbRef{
		InstanceID: inst.ID,
		ObjectID:   inst.ObjectID,
		ObjectName: cmdbObjectName(inst.ObjectID),
		Source:     "cmdb_instance",
	}
	asset.CmdbColumns = columns
	asset.CmdbValues = values
	if value := firstCmdbString(values, "name", "hostname", "host_name", "instance_name"); value != "" {
		asset.Hostname = value
	}
	if value := firstCmdbString(values, "ip_address", "mgmt_ip", "ip", "host_ip", "OS001"); value != "" && len(asset.IPList) == 0 {
		asset.IPList = splitAssetValues(value)
	}
	if value := firstCmdbString(values, "agent_status"); value != "" {
		asset.AgentStatus = value
	}
}

func resolveHostCmdbInstance(asset model.HostAsset) (model.CmdbInstance, []model.CmdbAttribute, bool) {
	candidates := normalizeAssetStrings(append([]string{asset.HostID, asset.Ident, asset.Hostname}, asset.IPList...))
	for _, objectID := range []string{"obj-os", "OperatingSystem1"} {
		if inst, attrs, ok := findHostCmdbInstanceByObject(objectID, candidates); ok {
			return inst, attrs, true
		}
	}
	if inst, attrs, ok := findHostCmdbInstanceByAnyObject(candidates); ok {
		return inst, attrs, true
	}
	return model.CmdbInstance{}, nil, false
}

func findHostCmdbInstanceByObject(objectID string, candidates []string) (model.CmdbInstance, []model.CmdbAttribute, bool) {
	if objectID == "" {
		return model.CmdbInstance{}, nil, false
	}
	attrs := store.ListCmdbAttributes(objectID)
	items, total := store.ListCmdbInstances(objectID, 1, 500)
	if total == 0 && len(items) == 0 {
		return model.CmdbInstance{}, attrs, false
	}
	for _, inst := range items {
		if cmdbInstanceMatchesHost(inst, candidates) {
			return inst, attrs, true
		}
	}
	return model.CmdbInstance{}, attrs, false
}

func findHostCmdbInstanceByAnyObject(candidates []string) (model.CmdbInstance, []model.CmdbAttribute, bool) {
	for _, obj := range store.ListCmdbObjects("") {
		inst, attrs, ok := findHostCmdbInstanceByObject(obj.ID, candidates)
		if ok {
			return inst, attrs, true
		}
	}
	return model.CmdbInstance{}, nil, false
}

func cmdbInstanceMatchesHost(inst model.CmdbInstance, candidates []string) bool {
	seen := map[string]bool{}
	for _, candidate := range candidates {
		seen[strings.ToLower(strings.TrimSpace(candidate))] = true
	}
	if seen[strings.ToLower(strings.TrimSpace(inst.ID))] {
		return true
	}
	raw := parseCmdbInstanceData(inst.Data)
	for _, key := range []string{"name", "hostname", "host_name", "instance_name", "ip_address", "mgmt_ip", "ip", "host_ip", "OS001", "agent_id"} {
		if seen[strings.ToLower(strings.TrimSpace(anyToString(raw[key])))] {
			return true
		}
	}
	return false
}

func firstCmdbString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(anyToString(values[key])); value != "" && value != cmdbMaskedValue {
			return value
		}
	}
	return ""
}

func respondResourceGroupStorageError(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": resourceGroupStorageError})
}

func updateTargetLabel(target *model.MonitorTarget, key, value string) {
	if target.Labels == nil {
		target.Labels = map[string]string{}
	}
	if strings.TrimSpace(value) == "" {
		delete(target.Labels, key)
		return
	}
	target.Labels[key] = strings.TrimSpace(value)
}

func safeHostLabels(in map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range in {
		if isSensitiveAssetKey(key) {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	return out
}

func isSensitiveAssetKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(k, "api_key") || strings.Contains(k, "apikey") ||
		strings.Contains(k, "auth") || strings.Contains(k, "token") ||
		strings.Contains(k, "secret") || strings.Contains(k, "password") ||
		strings.Contains(k, "cookie") || strings.Contains(k, "private") ||
		strings.Contains(k, "dsn")
}

func normalizeAssetStrings(items []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func splitAssetValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return normalizeAssetStrings(strings.Split(value, ","))
}

func firstAssetValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type errResourceGroupValidation string

func (e errResourceGroupValidation) Error() string { return string(e) }
