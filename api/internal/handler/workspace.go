package handler

import (
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

const (
	workspaceMaxNameLength = 120
	workspaceMaxTextLength = 500
)

var workspaceStatuses = map[string]bool{
	"active":   true,
	"disabled": true,
	"archived": true,
}

func ListWorkspaces(c *gin.Context) {
	businesses := store.ListTopologyBusinesses()
	out := make([]model.Workspace, 0, len(businesses))
	for _, business := range businesses {
		out = append(out, workspaceFromBusiness(business))
	}
	c.JSON(http.StatusOK, out)
}

func CreateWorkspace(c *gin.Context) {
	var input model.WorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	business, err := workspaceInputToBusiness(input, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved := store.SaveTopologyBusiness(business)
	c.JSON(http.StatusOK, workspaceFromBusiness(saved))
}

func GetWorkspace(c *gin.Context) {
	business, ok := store.GetTopologyBusiness(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	c.JSON(http.StatusOK, workspaceFromBusiness(business))
}

func UpdateWorkspace(c *gin.Context) {
	existing, ok := store.GetTopologyBusiness(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	var input model.WorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	business, err := workspaceInputToBusiness(input, &existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved := store.SaveTopologyBusiness(business)
	c.JSON(http.StatusOK, workspaceFromBusiness(saved))
}

func DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	if _, ok := store.GetTopologyBusiness(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	store.DeleteTopologyBusiness(id)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func workspaceInputToBusiness(input model.WorkspaceInput, existing *model.TopologyBusiness) (model.TopologyBusiness, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.TopologyBusiness{}, errWorkspaceValidation("name is required")
	}
	if len([]rune(name)) > workspaceMaxNameLength {
		return model.TopologyBusiness{}, errWorkspaceValidation("name is too long")
	}
	description := strings.TrimSpace(input.Description)
	if len([]rune(description)) > workspaceMaxTextLength {
		return model.TopologyBusiness{}, errWorkspaceValidation("description is too long")
	}
	owner := strings.TrimSpace(input.Owner)
	if len([]rune(owner)) > workspaceMaxNameLength {
		return model.TopologyBusiness{}, errWorkspaceValidation("owner is too long")
	}
	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = "active"
	}
	if !workspaceStatuses[status] {
		return model.TopologyBusiness{}, errWorkspaceValidation("unsupported status")
	}
	hosts := normalizeWorkspaceStrings(input.Hosts)
	endpoints, err := normalizeWorkspaceEndpoints(input.Endpoints)
	if err != nil {
		return model.TopologyBusiness{}, err
	}
	tags := normalizeWorkspaceStrings(input.Tags)
	attrs := map[string]string{
		"description": description,
		"owner":       owner,
		"status":      status,
		"tags":        strings.Join(tags, ","),
	}
	business := model.TopologyBusiness{
		Name:       name,
		Hosts:      hosts,
		Endpoints:  endpoints,
		Attributes: attrs,
	}
	if existing != nil {
		business.ID = existing.ID
		business.CreatedAt = existing.CreatedAt
		business.Graph = existing.Graph
		for key, value := range existing.Attributes {
			if _, managed := attrs[key]; !managed {
				business.Attributes[key] = value
			}
		}
	}
	return business, nil
}

func workspaceFromBusiness(business model.TopologyBusiness) model.Workspace {
	attrs := business.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	status := strings.TrimSpace(attrs["status"])
	if status == "" || !workspaceStatuses[status] {
		status = "active"
	}
	return model.Workspace{
		ID:            business.ID,
		Name:          business.Name,
		Description:   attrs["description"],
		Owner:         attrs["owner"],
		Status:        status,
		Tags:          splitWorkspaceTags(attrs["tags"]),
		Hosts:         normalizeWorkspaceStrings(business.Hosts),
		Endpoints:     business.Endpoints,
		ResourceCount: workspaceResourceCount(business),
		CreatedAt:     business.CreatedAt,
		UpdatedAt:     business.UpdatedAt,
	}
}

func normalizeWorkspaceEndpoints(items []model.TopologyEndpoint) ([]model.TopologyEndpoint, error) {
	out := make([]model.TopologyEndpoint, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		ip := strings.TrimSpace(item.IP)
		if ip == "" {
			continue
		}
		if net.ParseIP(ip) == nil {
			return nil, errWorkspaceValidation("endpoint ip is invalid")
		}
		if item.Port < 0 || item.Port > 65535 {
			return nil, errWorkspaceValidation("endpoint port is invalid")
		}
		item.IP = ip
		item.ServiceName = strings.TrimSpace(item.ServiceName)
		item.Protocol = strings.TrimSpace(item.Protocol)
		key := ip + ":" + strconv.Itoa(item.Port) + ":" + item.ServiceName + ":" + item.Protocol
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out, nil
}

func normalizeWorkspaceStrings(items []string) []string {
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

func splitWorkspaceTags(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return normalizeWorkspaceStrings(strings.Split(value, ","))
}

func workspaceResourceCount(business model.TopologyBusiness) int {
	hostSet := map[string]bool{}
	for _, host := range business.Hosts {
		host = strings.TrimSpace(host)
		if host != "" {
			hostSet[host] = true
		}
	}
	for _, endpoint := range business.Endpoints {
		if strings.TrimSpace(endpoint.IP) != "" {
			hostSet[strings.TrimSpace(endpoint.IP)] = true
		}
	}
	return len(hostSet) + len(business.Endpoints)
}

type errWorkspaceValidation string

func (e errWorkspaceValidation) Error() string {
	return string(e)
}
