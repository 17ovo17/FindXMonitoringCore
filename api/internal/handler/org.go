package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegisterOrgRoutes(rg *gin.RouterGroup, readRequired, adminRequired gin.HandlerFunc) {
	org := rg.Group("/org")
	org.GET("/users", readRequired, OrgListUsers)
	org.POST("/users", adminRequired, OrgCreateUser)
	org.GET("/users/:id", readRequired, OrgGetUser)
	org.PUT("/users/:id", adminRequired, OrgUpdateUser)
	org.DELETE("/users/:id", adminRequired, OrgDeleteUser)
	org.PUT("/users/:id/password", adminRequired, OrgResetUserPassword)

	org.GET("/teams", readRequired, OrgListTeams)
	org.POST("/teams", adminRequired, OrgCreateTeam)
	org.GET("/teams/:id", readRequired, OrgGetTeam)
	org.PUT("/teams/:id", adminRequired, OrgUpdateTeam)
	org.DELETE("/teams/:id", adminRequired, OrgDeleteTeam)
	org.POST("/teams/:id/members", adminRequired, OrgAddTeamMembers)
	org.DELETE("/teams/:id/members/:user_id", adminRequired, OrgRemoveTeamMember)

	org.GET("/business-groups", readRequired, OrgListBusinessGroups)
	org.POST("/business-groups", adminRequired, OrgCreateBusinessGroup)
	org.GET("/business-groups/:id", readRequired, OrgGetBusinessGroup)
	org.PUT("/business-groups/:id", adminRequired, OrgUpdateBusinessGroup)
	org.DELETE("/business-groups/:id", adminRequired, OrgDeleteBusinessGroup)
	org.POST("/business-groups/:id/teams", adminRequired, OrgAddBusinessTeams)
	org.DELETE("/business-groups/:id/teams/:team_id", adminRequired, OrgRemoveBusinessTeam)

	org.GET("/roles", readRequired, OrgListRoles)
	org.POST("/roles", adminRequired, OrgCreateRole)
	org.GET("/roles/:id", readRequired, OrgGetRole)
	org.PUT("/roles/:id", adminRequired, OrgUpdateRole)
	org.DELETE("/roles/:id", adminRequired, OrgDeleteRole)
	org.GET("/permissions/operations", readRequired, OrgOperations)
	org.GET("/roles/:id/operations", readRequired, OrgGetRoleOperations)
	org.PUT("/roles/:id/operations", adminRequired, OrgSetRoleOperations)
}

func OrgListUsers(c *gin.Context) {
	page, limit := positiveQueryInt(c, "page", 1), positiveQueryInt(c, "limit", 20)
	c.JSON(http.StatusOK, store.ListOrgUsers(c.Query("q"), page, limit))
}

func OrgCreateUser(c *gin.Context) {
	var input model.OrgUserInput
	if !bindJSON(c, &input) {
		return
	}
	username := strings.TrimSpace(input.Username)
	if username == "" || strings.TrimSpace(input.Password) == "" {
		writeError(c, http.StatusBadRequest, "用户名和密码不能为空")
		return
	}
	if input.Confirm != "" && input.Confirm != input.Password {
		writeError(c, http.StatusBadRequest, "两次密码不一致")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "密码处理失败")
		return
	}
	legacyRole := "viewer"
	for _, role := range input.Roles {
		if role == "admin" {
			legacyRole = "admin"
			break
		}
	}
	user := &model.User{ID: store.NewID(), Username: username, PasswordHash: string(hash), Role: legacyRole, MustChangePwd: true}
	out, err := store.CreateOrgUser(user, input)
	if err != nil {
		writeStoreError(c, err, "创建用户失败")
		return
	}
	auditEvent(c, "org.user.create", out.ID, "medium", "allow", "创建人员组织用户", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusCreated, out)
}

func OrgGetUser(c *gin.Context) {
	user, ok := store.GetOrgUser(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "用户不存在")
		return
	}
	c.JSON(http.StatusOK, user)
}

func OrgUpdateUser(c *gin.Context) {
	var input model.OrgUserInput
	if !bindJSON(c, &input) {
		return
	}
	out, err := store.UpdateOrgUser(c.Param("id"), input)
	if err != nil {
		writeStoreError(c, err, "更新用户失败")
		return
	}
	auditEvent(c, "org.user.update", out.ID, "medium", "allow", "更新人员组织用户", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, out)
}

func OrgDeleteUser(c *gin.Context) {
	if err := store.DeleteOrgUser(c.Param("id")); err != nil {
		writeStoreError(c, err, "删除用户失败")
		return
	}
	auditEvent(c, "org.user.delete", c.Param("id"), "high", "allow", "删除人员组织用户", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgResetUserPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
		Confirm  string `json:"confirm"`
	}
	if !bindJSON(c, &req) {
		return
	}
	if len(req.Password) < 6 || (req.Confirm != "" && req.Confirm != req.Password) {
		writeError(c, http.StatusBadRequest, "密码至少 6 位且两次输入一致")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "密码处理失败")
		return
	}
	if err := store.UpdateUserPassword(c.Param("id"), string(hash)); err != nil {
		writeStoreError(c, err, "重置密码失败")
		return
	}
	auditEvent(c, "org.user.password", c.Param("id"), "high", "allow", "重置人员组织用户密码", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true, "must_change_pwd": false})
}

func OrgListTeams(c *gin.Context) { c.JSON(http.StatusOK, store.ListOrgTeams(c.Query("q"))) }

func OrgCreateTeam(c *gin.Context) {
	var input model.OrgTeamInput
	if !bindJSON(c, &input) {
		return
	}
	team, err := store.SaveOrgTeam("", input)
	if err != nil {
		writeStoreError(c, err, "创建团队失败")
		return
	}
	auditEvent(c, "org.team.create", team.ID, "medium", "allow", "创建团队", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusCreated, team)
}

func OrgGetTeam(c *gin.Context) {
	team, ok := store.GetOrgTeam(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "团队不存在")
		return
	}
	c.JSON(http.StatusOK, team)
}

func OrgUpdateTeam(c *gin.Context) {
	var input model.OrgTeamInput
	if !bindJSON(c, &input) {
		return
	}
	team, err := store.SaveOrgTeam(c.Param("id"), input)
	if err != nil {
		writeStoreError(c, err, "更新团队失败")
		return
	}
	auditEvent(c, "org.team.update", team.ID, "medium", "allow", "更新团队", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, team)
}

func OrgDeleteTeam(c *gin.Context) {
	if err := store.DeleteOrgTeam(c.Param("id")); err != nil {
		writeStoreError(c, err, "删除团队失败")
		return
	}
	auditEvent(c, "org.team.delete", c.Param("id"), "high", "allow", "删除团队", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgAddTeamMembers(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if !bindJSON(c, &req) {
		return
	}
	if err := store.AddOrgTeamMembers(c.Param("id"), req.IDs); err != nil {
		writeStoreError(c, err, "添加团队成员失败")
		return
	}
	auditEvent(c, "org.team.members.add", c.Param("id"), "medium", "allow", "添加团队成员", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func OrgRemoveTeamMember(c *gin.Context) {
	if err := store.RemoveOrgTeamMember(c.Param("id"), c.Param("user_id")); err != nil {
		writeStoreError(c, err, "移除团队成员失败")
		return
	}
	auditEvent(c, "org.team.members.delete", c.Param("id"), "medium", "allow", "移除团队成员", c.GetHeader("X-Test-Batch-Id"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func positiveQueryInt(c *gin.Context, name string, fallback int) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func bindJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		writeError(c, http.StatusBadRequest, "请求格式不正确")
		return false
	}
	return true
}

func writeStoreError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeError(c, http.StatusNotFound, "记录不存在")
	case store.IsOrgReadonlyRoleError(err):
		writeError(c, http.StatusConflict, "内置角色不可删除或修改权限")
	default:
		writeError(c, http.StatusInternalServerError, fallback)
	}
}

func writeError(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}
