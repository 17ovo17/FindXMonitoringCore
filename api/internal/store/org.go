package store

import (
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

var (
	orgUserProfiles    = map[string]*orgUserProfile{}
	orgTeams           = map[string]*model.OrgTeam{}
	orgTeamMembers     = map[string]map[string]bool{}
	orgBusinessGroups  = map[string]*model.OrgBusinessGroup{}
	orgBusinessTeams   = map[string]map[string]string{}
	orgRoles           = map[string]*model.OrgRole{}
	orgUserRoles       = map[string]map[string]bool{}
	orgRoleOperations  = map[string]map[string]bool{}
	errOrgEmptyInput   = errors.New("org input ids cannot be empty")
	errOrgRoleReadonly = errors.New("内置角色不可删除")
)

type orgUserProfile struct {
	Nickname     string            `json:"nickname,omitempty"`
	Email        string            `json:"email,omitempty"`
	Phone        string            `json:"phone,omitempty"`
	Contacts     map[string]string `json:"contacts,omitempty"`
	LastActiveAt int64             `json:"last_active_time,omitempty"`
}

var orgOperationGroups = []model.OrgOperationGroup{
	{Name: "org", CName: "人员组织", Ops: []model.OrgOperation{{Name: "org.user.read", CName: "查看用户"}, {Name: "org.user.write", CName: "管理用户"}, {Name: "org.team.read", CName: "查看团队"}, {Name: "org.team.write", CName: "管理团队"}, {Name: "org.business.read", CName: "查看业务组"}, {Name: "org.business.write", CName: "管理业务组"}, {Name: "org.role.read", CName: "查看角色"}, {Name: "org.role.write", CName: "管理角色"}}},
	{Name: "monitor", CName: "监控中心", Ops: []model.OrgOperation{{Name: "monitor.datasource.read", CName: "查看数据源"}, {Name: "monitor.datasource.write", CName: "管理数据源"}, {Name: "monitor.query.execute", CName: "执行查询"}, {Name: "monitor.dashboard.read", CName: "查看仪表盘"}, {Name: "monitor.dashboard.write", CName: "管理仪表盘"}, {Name: "monitor.alert_rule.read", CName: "查看告警规则"}, {Name: "monitor.alert_rule.write", CName: "管理告警规则"}, {Name: "monitor.alert_event.write", CName: "处置告警事件"}}},
	{Name: "aiops", CName: "AI SRE", Ops: []model.OrgOperation{{Name: "aiops.session.read", CName: "查看会话"}, {Name: "aiops.action.execute", CName: "执行动作"}, {Name: "workflow.read", CName: "查看工作流"}, {Name: "workflow.write", CName: "管理工作流"}, {Name: "knowledge.read", CName: "查看知识库"}, {Name: "knowledge.write", CName: "管理知识库"}, {Name: "findx_agent.read", CName: "查看 Agent"}, {Name: "findx_agent.write", CName: "管理 Agent"}}},
	{Name: "platform", CName: "平台配置", Ops: []model.OrgOperation{{Name: "settings.read", CName: "查看设置"}, {Name: "settings.write", CName: "管理设置"}, {Name: "credential.write", CName: "管理凭据"}, {Name: "audit.read", CName: "查看审计"}}},
}

func SeedOrgDefaults() {
	now := time.Now()
	if mysqlOK {
		_, _ = db.Exec(`INSERT IGNORE INTO org_roles (id,name,note,builtin,created_at,updated_at) VALUES ('admin','admin','系统管理员',1,?,?),('viewer','viewer','只读用户',1,?,?)`, now, now, now, now)
		ops := flattenOrgOperations()
		for _, op := range ops {
			_, _ = db.Exec(`INSERT IGNORE INTO org_role_operations (role_id,operation,created_at) VALUES ('admin',?,?)`, op, now)
		}
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if _, ok := orgRoles["admin"]; !ok {
		orgRoles["admin"] = &model.OrgRole{ID: "admin", Name: "admin", Note: "系统管理员", Builtin: true, CreatedAt: now, UpdatedAt: now}
	}
	if _, ok := orgRoles["viewer"]; !ok {
		orgRoles["viewer"] = &model.OrgRole{ID: "viewer", Name: "viewer", Note: "只读用户", Builtin: true, CreatedAt: now, UpdatedAt: now}
	}
	orgRoleOperations["admin"] = map[string]bool{}
	for _, op := range flattenOrgOperations() {
		orgRoleOperations["admin"][op] = true
	}
}

func ListOrgUsers(q string, page, limit int) model.OrgUserList {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	q = strings.ToLower(strings.TrimSpace(q))
	all := listAllOrgUsers()
	filtered := make([]model.OrgUser, 0, len(all))
	for _, u := range all {
		if q == "" || strings.Contains(strings.ToLower(u.Username+" "+u.Nickname+" "+u.Email+" "+u.Phone), q) {
			filtered = append(filtered, u)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	start := (page - 1) * limit
	if start >= len(filtered) {
		return model.OrgUserList{Total: len(filtered), List: []model.OrgUser{}}
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return model.OrgUserList{Total: len(filtered), List: filtered[start:end]}
}

func GetOrgUser(id string) (model.OrgUser, bool) {
	for _, u := range listAllOrgUsers() {
		if u.ID == id {
			return u, true
		}
	}
	return model.OrgUser{}, false
}

func CreateOrgUser(user *model.User, input model.OrgUserInput) (model.OrgUser, error) {
	if mysqlOK {
		tx, err := db.Begin()
		if err != nil {
			return model.OrgUser{}, err
		}
		if err := createOrgUserTx(tx, user, input); err != nil {
			_ = tx.Rollback()
			return model.OrgUser{}, err
		}
		if err := tx.Commit(); err != nil {
			return model.OrgUser{}, err
		}
		return mustGetOrgUser(user.ID), nil
	}
	if err := CreateUser(user); err != nil {
		return model.OrgUser{}, err
	}
	now := time.Now()
	mu.Lock()
	orgUserProfiles[user.ID] = &orgUserProfile{Nickname: input.Nickname, Email: input.Email, Phone: input.Phone, Contacts: cleanContacts(input.Contacts)}
	orgUserRoles[user.ID] = roleSet(input.Roles, user.Role)
	if orgUserRoles[user.ID]["admin"] {
		users[user.Username].Role = "admin"
	}
	users[user.Username].UpdatedAt = now
	mu.Unlock()
	if err := persistFallbackSnapshot(); err != nil {
		return model.OrgUser{}, err
	}
	return mustGetOrgUser(user.ID), nil
}

func UpdateOrgUser(id string, input model.OrgUserInput) (model.OrgUser, error) {
	if mysqlOK {
		tx, err := db.Begin()
		if err != nil {
			return model.OrgUser{}, err
		}
		if err := mysqlUserExists(tx, id); err != nil {
			_ = tx.Rollback()
			return model.OrgUser{}, err
		}
		now := time.Now()
		if _, err := tx.Exec(`INSERT INTO org_user_profiles (user_id,nickname,email,phone,contacts_json,last_active_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE nickname=VALUES(nickname),email=VALUES(email),phone=VALUES(phone),contacts_json=VALUES(contacts_json),updated_at=VALUES(updated_at)`, id, input.Nickname, input.Email, input.Phone, jsonText(cleanContacts(input.Contacts)), int64(0), now, now); err != nil {
			_ = tx.Rollback()
			return model.OrgUser{}, err
		}
		if err := replaceUserRolesTx(tx, id, input.Roles); err != nil {
			_ = tx.Rollback()
			return model.OrgUser{}, err
		}
		if err := tx.Commit(); err != nil {
			return model.OrgUser{}, err
		}
		return mustGetOrgUser(id), nil
	}
	mu.Lock()
	if !userIDExistsLocked(id) {
		mu.Unlock()
		return model.OrgUser{}, sql.ErrNoRows
	}
	orgUserProfiles[id] = &orgUserProfile{Nickname: input.Nickname, Email: input.Email, Phone: input.Phone, Contacts: cleanContacts(input.Contacts)}
	orgUserRoles[id] = roleSet(input.Roles, "")
	mu.Unlock()
	if err := persistFallbackSnapshot(); err != nil {
		return model.OrgUser{}, err
	}
	return mustGetOrgUser(id), nil
}

func DeleteOrgUser(id string) error {
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM users WHERE id=?`, id)
		if err != nil {
			return err
		}
		db.Exec(`DELETE FROM org_user_profiles WHERE user_id=?`, id)
		db.Exec(`DELETE FROM org_user_roles WHERE user_id=?`, id)
		db.Exec(`DELETE FROM org_team_members WHERE user_id=?`, id)
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	mu.Lock()
	var username string
	for name, u := range users {
		if u.ID == id {
			username = name
			break
		}
	}
	if username == "" {
		mu.Unlock()
		return sql.ErrNoRows
	}
	delete(users, username)
	delete(orgUserProfiles, id)
	delete(orgUserRoles, id)
	for _, members := range orgTeamMembers {
		delete(members, id)
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}

func ListOrgTeams(q string) []model.OrgTeam {
	items := allTeams()
	q = strings.ToLower(strings.TrimSpace(q))
	out := []model.OrgTeam{}
	for _, t := range items {
		if q == "" || strings.Contains(strings.ToLower(t.Name+" "+t.Note), q) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func GetOrgTeam(id string) (model.OrgTeam, bool) {
	for _, t := range allTeams() {
		if t.ID == id {
			t.Members = teamUsers(id)
			return t, true
		}
	}
	return model.OrgTeam{}, false
}

func SaveOrgTeam(id string, in model.OrgTeamInput) (model.OrgTeam, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.OrgTeam{}, errors.New("名称不能为空")
	}
	now := time.Now()
	if id == "" {
		id = NewID()
	}
	if mysqlOK {
		_, err := db.Exec(`REPLACE INTO org_teams (id,name,note,parent_id,created_at,updated_at) VALUES (?,?,?,?,COALESCE((SELECT created_at FROM (SELECT created_at FROM org_teams WHERE id=?) x),?),?)`, id, name, in.Note, in.ParentID, id, now, now)
		if err != nil {
			return model.OrgTeam{}, err
		}
		return mustGetTeam(id), nil
	}
	mu.Lock()
	created := now
	if old := orgTeams[id]; old != nil {
		created = old.CreatedAt
	}
	orgTeams[id] = &model.OrgTeam{ID: id, Name: name, Note: in.Note, ParentID: in.ParentID, CreatedAt: created, UpdatedAt: now}
	mu.Unlock()
	return mustGetTeam(id), persistFallbackSnapshot()
}

func DeleteOrgTeam(id string) error {
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM org_teams WHERE id=?`, id)
		db.Exec(`DELETE FROM org_team_members WHERE team_id=?`, id)
		db.Exec(`DELETE FROM org_business_group_teams WHERE team_id=?`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	mu.Lock()
	if orgTeams[id] == nil {
		mu.Unlock()
		return sql.ErrNoRows
	}
	delete(orgTeams, id)
	delete(orgTeamMembers, id)
	for _, teams := range orgBusinessTeams {
		delete(teams, id)
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}

func AddOrgTeamMembers(teamID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return errOrgEmptyInput
	}
	if mysqlOK {
		if err := mysqlTeamExists(db, teamID); err != nil {
			return err
		}
		for _, uid := range userIDs {
			if strings.TrimSpace(uid) == "" {
				return errOrgEmptyInput
			}
			if err := mysqlUserExists(db, uid); err != nil {
				return err
			}
			if _, err := db.Exec(`INSERT IGNORE INTO org_team_members (team_id,user_id,created_at) VALUES (?,?,?)`, teamID, uid, time.Now()); err != nil {
				return err
			}
		}
		return nil
	}
	mu.Lock()
	if orgTeams[teamID] == nil {
		mu.Unlock()
		return sql.ErrNoRows
	}
	for _, uid := range userIDs {
		if strings.TrimSpace(uid) == "" {
			mu.Unlock()
			return errOrgEmptyInput
		}
		if !userIDExistsLocked(uid) {
			mu.Unlock()
			return sql.ErrNoRows
		}
	}
	if orgTeamMembers[teamID] == nil {
		orgTeamMembers[teamID] = map[string]bool{}
	}
	for _, uid := range userIDs {
		orgTeamMembers[teamID][uid] = true
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}

func RemoveOrgTeamMember(teamID, userID string) error {
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM org_team_members WHERE team_id=? AND user_id=?`, teamID, userID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	mu.Lock()
	if orgTeamMembers[teamID] == nil || !orgTeamMembers[teamID][userID] {
		mu.Unlock()
		return sql.ErrNoRows
	}
	delete(orgTeamMembers[teamID], userID)
	mu.Unlock()
	return persistFallbackSnapshot()
}
