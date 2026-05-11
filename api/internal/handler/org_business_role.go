package handler

import (
	"net/http"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func OrgListBusinessGroups(c *gin.Context) {
	c.JSON(http.StatusOK, store.ListOrgBusinessGroups(c.Query("q")))
}

func OrgCreateBusinessGroup(c *gin.Context) {
	var input model.OrgBusinessGroupInput
	if !bindJSON(c, &input) {
		return
	}
	out, err := store.SaveOrgBusinessGroup("", input)
	if err != nil {
		writeStoreError(c, err, "创建业务组失败")
		return
	}
	auditEvent(c, "org.business.create", out.ID, "medium", "allow", "创建业务组", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusCreated, out)
}

func OrgGetBusinessGroup(c *gin.Context) {
	out, ok := store.GetOrgBusinessGroup(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "业务组不存在")
		return
	}
	c.JSON(http.StatusOK, out)
}

func OrgUpdateBusinessGroup(c *gin.Context) {
	var input model.OrgBusinessGroupInput
	if !bindJSON(c, &input) {
		return
	}
	out, err := store.SaveOrgBusinessGroup(c.Param("id"), input)
	if err != nil {
		writeStoreError(c, err, "更新业务组失败")
		return
	}
	auditEvent(c, "org.business.update", out.ID, "medium", "allow", "更新业务组", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, out)
}

func OrgDeleteBusinessGroup(c *gin.Context) {
	if err := store.DeleteOrgBusinessGroup(c.Param("id")); err != nil {
		writeStoreError(c, err, "删除业务组失败")
		return
	}
	auditEvent(c, "org.business.delete", c.Param("id"), "high", "allow", "删除业务组", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgAddBusinessTeams(c *gin.Context) {
	var req struct {
		Teams  []model.OrgTeamLink `json:"teams"`
		ID     string              `json:"team_id"`
		Perm   string              `json:"perm_flag"`
		Legacy []struct {
			UserGroupID string `json:"user_group_id"`
			PermFlag    string `json:"perm_flag"`
		} `json:"members"`
	}
	if !bindJSON(c, &req) {
		return
	}
	links := req.Teams
	if req.ID != "" {
		links = append(links, model.OrgTeamLink{ID: req.ID, PermFlag: req.Perm})
	}
	for _, item := range req.Legacy {
		links = append(links, model.OrgTeamLink{ID: item.UserGroupID, PermFlag: item.PermFlag})
	}
	if err := store.AddOrgBusinessTeams(c.Param("id"), links); err != nil {
		writeStoreError(c, err, "添加业务组团队失败")
		return
	}
	auditEvent(c, "org.business.teams.add", c.Param("id"), "medium", "allow", "添加业务组团队", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgRemoveBusinessTeam(c *gin.Context) {
	if err := store.RemoveOrgBusinessTeam(c.Param("id"), c.Param("team_id")); err != nil {
		writeStoreError(c, err, "移除业务组团队失败")
		return
	}
	auditEvent(c, "org.business.teams.delete", c.Param("id"), "medium", "allow", "移除业务组团队", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgListRoles(c *gin.Context) { c.JSON(http.StatusOK, store.ListOrgRoles()) }

func OrgCreateRole(c *gin.Context) {
	var input model.OrgRoleInput
	if !bindJSON(c, &input) {
		return
	}
	role, err := store.SaveOrgRole("", input)
	if err != nil {
		writeStoreError(c, err, "创建角色失败")
		return
	}
	auditEvent(c, "org.role.create", role.ID, "medium", "allow", "创建角色", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusCreated, role)
}

func OrgGetRole(c *gin.Context) {
	for _, role := range store.ListOrgRoles() {
		if role.ID == c.Param("id") {
			c.JSON(http.StatusOK, role)
			return
		}
	}
	writeError(c, http.StatusNotFound, "角色不存在")
}

func OrgUpdateRole(c *gin.Context) {
	var input model.OrgRoleInput
	if !bindJSON(c, &input) {
		return
	}
	role, err := store.SaveOrgRole(c.Param("id"), input)
	if err != nil {
		writeStoreError(c, err, "更新角色失败")
		return
	}
	auditEvent(c, "org.role.update", role.ID, "medium", "allow", "更新角色", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, role)
}

func OrgDeleteRole(c *gin.Context) {
	if err := store.DeleteOrgRole(c.Param("id")); err != nil {
		writeStoreError(c, err, "删除角色失败")
		return
	}
	auditEvent(c, "org.role.delete", c.Param("id"), "high", "allow", "删除角色", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgOperations(c *gin.Context) { c.JSON(http.StatusOK, store.OrgOperations()) }

func OrgGetRoleOperations(c *gin.Context) {
	c.JSON(http.StatusOK, store.GetOrgRoleOperations(c.Param("id")))
}

func OrgSetRoleOperations(c *gin.Context) {
	var ops []string
	if !bindJSON(c, &ops) {
		return
	}
	if err := store.SetOrgRoleOperations(c.Param("id"), ops); err != nil {
		writeStoreError(c, err, "更新角色权限失败")
		return
	}
	auditEvent(c, "org.role.operations.update", c.Param("id"), "high", "allow", "更新角色权限", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
