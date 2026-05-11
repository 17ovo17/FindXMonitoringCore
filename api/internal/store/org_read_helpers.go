package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func listAllOrgUsers() []model.OrgUser {
	if mysqlOK {
		rows, err := db.Query(`SELECT u.id,u.username,u.role,u.must_change_pwd,u.created_at,u.updated_at,COALESCE(p.nickname,''),COALESCE(p.email,''),COALESCE(p.phone,''),COALESCE(p.contacts_json,'{}'),COALESCE(p.last_active_at,0) FROM users u LEFT JOIN org_user_profiles p ON p.user_id=u.id`)
		if err == nil {
			defer rows.Close()
			out := []model.OrgUser{}
			for rows.Next() {
				var u model.OrgUser
				var legacyRole, contacts string
				if rows.Scan(&u.ID, &u.Username, &legacyRole, &u.MustChangePwd, &u.CreatedAt, &u.UpdatedAt, &u.Nickname, &u.Email, &u.Phone, &contacts, &u.LastActiveAt) == nil {
					u.Contacts = parseContacts(contacts)
					u.Roles = mysqlUserRoles(u.ID, legacyRole)
					u.UserGroups = userTeamRefs(u.ID)
					u.BusiGroups = userBusinessRefs(u.ID)
					out = append(out, u)
				}
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.OrgUser{}
	for _, u := range users {
		p := orgUserProfiles[u.ID]
		ou := model.OrgUser{ID: u.ID, Username: u.Username, Roles: rolesForUserLocked(u.ID, u.Role), MustChangePwd: u.MustChangePwd, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, UserGroups: teamRefsForUserLocked(u.ID), BusiGroups: businessRefsForUserLocked(u.ID)}
		if p != nil {
			ou.Nickname, ou.Email, ou.Phone, ou.Contacts, ou.LastActiveAt = p.Nickname, p.Email, p.Phone, cleanContacts(p.Contacts), p.LastActiveAt
		}
		out = append(out, ou)
	}
	return out
}

func createOrgUserTx(tx *sql.Tx, user *model.User, input model.OrgUserInput) error {
	now := time.Now()
	user.CreatedAt, user.UpdatedAt = now, now
	if _, err := tx.Exec(`INSERT INTO users (id,username,password_hash,role,must_change_pwd,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, user.ID, user.Username, user.PasswordHash, user.Role, user.MustChangePwd, now, now); err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO org_user_profiles (user_id,nickname,email,phone,contacts_json,last_active_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?,?)`, user.ID, input.Nickname, input.Email, input.Phone, jsonText(cleanContacts(input.Contacts)), int64(0), now, now); err != nil {
		return err
	}
	return replaceUserRolesTx(tx, user.ID, input.Roles)
}

func replaceUserRolesTx(tx *sql.Tx, userID string, roles []string) error {
	if _, err := tx.Exec(`DELETE FROM org_user_roles WHERE user_id=?`, userID); err != nil {
		return err
	}
	for role := range roleSet(roles, "") {
		if _, err := tx.Exec(`INSERT INTO org_user_roles (user_id,role_id,created_at) VALUES (?,?,?)`, userID, role, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

func allTeams() []model.OrgTeam {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,name,note,parent_id,created_at,updated_at FROM org_teams`)
		if err == nil {
			defer rows.Close()
			out := []model.OrgTeam{}
			for rows.Next() {
				var t model.OrgTeam
				if rows.Scan(&t.ID, &t.Name, &t.Note, &t.ParentID, &t.CreatedAt, &t.UpdatedAt) == nil {
					out = append(out, t)
				}
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.OrgTeam{}
	for _, t := range orgTeams {
		cp := *t
		out = append(out, cp)
	}
	return out
}

func allBusinessGroups() []model.OrgBusinessGroup {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,name,note,parent_id,created_at,updated_at FROM org_business_groups`)
		if err == nil {
			defer rows.Close()
			out := []model.OrgBusinessGroup{}
			for rows.Next() {
				var b model.OrgBusinessGroup
				if rows.Scan(&b.ID, &b.Name, &b.Note, &b.ParentID, &b.CreatedAt, &b.UpdatedAt) == nil {
					out = append(out, b)
				}
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.OrgBusinessGroup{}
	for _, b := range orgBusinessGroups {
		cp := *b
		out = append(out, cp)
	}
	return out
}

func teamUsers(teamID string) []model.OrgUser {
	all := listAllOrgUsers()
	out := []model.OrgUser{}
	for _, u := range all {
		for _, ref := range u.UserGroups {
			if ref.ID == teamID {
				out = append(out, u)
			}
		}
	}
	return out
}

func businessTeams(groupID string) []model.OrgTeamLink {
	if mysqlOK {
		rows, err := db.Query(`SELECT t.id,t.name,bgt.perm_flag FROM org_business_group_teams bgt JOIN org_teams t ON t.id=bgt.team_id WHERE bgt.business_group_id=?`, groupID)
		if err == nil {
			defer rows.Close()
			out := []model.OrgTeamLink{}
			for rows.Next() {
				var link model.OrgTeamLink
				if rows.Scan(&link.ID, &link.Name, &link.PermFlag) == nil {
					out = append(out, link)
				}
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := []model.OrgTeamLink{}
	for teamID, perm := range orgBusinessTeams[groupID] {
		if team := orgTeams[teamID]; team != nil {
			out = append(out, model.OrgTeamLink{ID: team.ID, Name: team.Name, PermFlag: perm})
		}
	}
	return out
}

func userTeamRefs(userID string) []model.OrgRef {
	rows, err := db.Query(`SELECT t.id,t.name FROM org_team_members tm JOIN org_teams t ON t.id=tm.team_id WHERE tm.user_id=?`, userID)
	if err != nil {
		return []model.OrgRef{}
	}
	defer rows.Close()
	out := []model.OrgRef{}
	for rows.Next() {
		var r model.OrgRef
		if rows.Scan(&r.ID, &r.Name) == nil {
			out = append(out, r)
		}
	}
	return out
}

func userBusinessRefs(userID string) []model.OrgRef {
	rows, err := db.Query(`SELECT DISTINCT bg.id,bg.name FROM org_team_members tm JOIN org_business_group_teams bgt ON bgt.team_id=tm.team_id JOIN org_business_groups bg ON bg.id=bgt.business_group_id WHERE tm.user_id=?`, userID)
	if err != nil {
		return []model.OrgRef{}
	}
	defer rows.Close()
	out := []model.OrgRef{}
	for rows.Next() {
		var r model.OrgRef
		if rows.Scan(&r.ID, &r.Name) == nil {
			out = append(out, r)
		}
	}
	return out
}

func mysqlUserRoles(userID, legacyRole string) []string {
	rows, err := db.Query(`SELECT role_id FROM org_user_roles WHERE user_id=?`, userID)
	if err != nil {
		return []string{legacyRole}
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var role string
		if rows.Scan(&role) == nil {
			out = append(out, role)
		}
	}
	if len(out) == 0 && legacyRole != "" {
		out = append(out, legacyRole)
	}
	sort.Strings(out)
	return out
}

func mustGetOrgUser(id string) model.OrgUser {
	u, _ := GetOrgUser(id)
	return u
}

func mustGetTeam(id string) model.OrgTeam {
	t, _ := GetOrgTeam(id)
	return t
}

func mustGetBusiness(id string) model.OrgBusinessGroup {
	b, _ := GetOrgBusinessGroup(id)
	return b
}

func getRoleByID(id string) (model.OrgRole, bool) {
	for _, r := range ListOrgRoles() {
		if r.ID == id {
			return r, true
		}
	}
	return model.OrgRole{}, false
}

func roleSet(roles []string, fallback string) map[string]bool {
	out := map[string]bool{}
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			out[role] = true
		}
	}
	if len(out) == 0 {
		if fallback == "" {
			fallback = "viewer"
		}
		out[fallback] = true
	}
	return out
}

func rolesForUserLocked(userID, legacyRole string) []string {
	roles := orgUserRoles[userID]
	if len(roles) == 0 && legacyRole != "" {
		return []string{legacyRole}
	}
	out := []string{}
	for role := range roles {
		out = append(out, role)
	}
	sort.Strings(out)
	return out
}

func teamRefsForUserLocked(userID string) []model.OrgRef {
	out := []model.OrgRef{}
	for teamID, members := range orgTeamMembers {
		if members[userID] && orgTeams[teamID] != nil {
			out = append(out, model.OrgRef{ID: teamID, Name: orgTeams[teamID].Name})
		}
	}
	return out
}

func businessRefsForUserLocked(userID string) []model.OrgRef {
	seen := map[string]bool{}
	out := []model.OrgRef{}
	for teamID, members := range orgTeamMembers {
		if !members[userID] {
			continue
		}
		for groupID, teams := range orgBusinessTeams {
			if teams[teamID] != "" && orgBusinessGroups[groupID] != nil && !seen[groupID] {
				seen[groupID] = true
				out = append(out, model.OrgRef{ID: groupID, Name: orgBusinessGroups[groupID].Name})
			}
		}
	}
	return out
}

func userIDExistsLocked(id string) bool {
	for _, u := range users {
		if u.ID == id {
			return true
		}
	}
	return false
}

type orgExistenceQueryer interface {
	QueryRow(query string, args ...any) *sql.Row
}

func mysqlUserExists(q orgExistenceQueryer, id string) error {
	return mysqlExists(q, `SELECT 1 FROM users WHERE id=?`, id)
}

func mysqlTeamExists(q orgExistenceQueryer, id string) error {
	return mysqlExists(q, `SELECT 1 FROM org_teams WHERE id=?`, id)
}

func mysqlBusinessGroupExists(q orgExistenceQueryer, id string) error {
	return mysqlExists(q, `SELECT 1 FROM org_business_groups WHERE id=?`, id)
}

func mysqlExists(q orgExistenceQueryer, query, id string) error {
	if strings.TrimSpace(id) == "" {
		return errOrgEmptyInput
	}
	var exists int
	return q.QueryRow(query, id).Scan(&exists)
}

func cleanContacts(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	for k, v := range in {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out
}

func parseContacts(raw string) map[string]string {
	out := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func jsonText(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func normalizePerm(value string) string {
	if value == "ro" {
		return "ro"
	}
	return "rw"
}

func flattenOrgOperations() []string {
	var out []string
	for _, group := range orgOperationGroups {
		for _, op := range group.Ops {
			out = append(out, op.Name)
		}
	}
	return out
}

func IsOrgReadonlyRoleError(err error) bool {
	return errors.Is(err, errOrgRoleReadonly)
}
