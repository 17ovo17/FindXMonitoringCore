package store

import (
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
)

func ListOrgBusinessGroups(q string) []model.OrgBusinessGroup {
	items := allBusinessGroups()
	q = strings.ToLower(strings.TrimSpace(q))
	out := []model.OrgBusinessGroup{}
	for _, b := range items {
		if q == "" || strings.Contains(strings.ToLower(b.Name+" "+b.Note), q) {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func GetOrgBusinessGroup(id string) (model.OrgBusinessGroup, bool) {
	for _, b := range allBusinessGroups() {
		if b.ID == id {
			b.Teams = businessTeams(id)
			return b, true
		}
	}
	return model.OrgBusinessGroup{}, false
}

func SaveOrgBusinessGroup(id string, in model.OrgBusinessGroupInput) (model.OrgBusinessGroup, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.OrgBusinessGroup{}, errors.New("名称不能为空")
	}
	now := time.Now()
	if id == "" {
		id = NewID()
	}
	if mysqlOK {
		_, err := db.Exec(`REPLACE INTO org_business_groups (id,name,note,parent_id,created_at,updated_at) VALUES (?,?,?,?,COALESCE((SELECT created_at FROM (SELECT created_at FROM org_business_groups WHERE id=?) x),?),?)`, id, name, in.Note, in.ParentID, id, now, now)
		if err != nil {
			return model.OrgBusinessGroup{}, err
		}
		return mustGetBusiness(id), nil
	}
	mu.Lock()
	created := now
	if old := orgBusinessGroups[id]; old != nil {
		created = old.CreatedAt
	}
	orgBusinessGroups[id] = &model.OrgBusinessGroup{ID: id, Name: name, Note: in.Note, ParentID: in.ParentID, CreatedAt: created, UpdatedAt: now}
	mu.Unlock()
	return mustGetBusiness(id), persistFallbackSnapshot()
}

func DeleteOrgBusinessGroup(id string) error {
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM org_business_groups WHERE id=?`, id)
		db.Exec(`DELETE FROM org_business_group_teams WHERE business_group_id=?`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	mu.Lock()
	if orgBusinessGroups[id] == nil {
		mu.Unlock()
		return sql.ErrNoRows
	}
	delete(orgBusinessGroups, id)
	delete(orgBusinessTeams, id)
	mu.Unlock()
	return persistFallbackSnapshot()
}

func AddOrgBusinessTeams(groupID string, links []model.OrgTeamLink) error {
	if len(links) == 0 {
		return errOrgEmptyInput
	}
	if mysqlOK {
		if err := mysqlBusinessGroupExists(db, groupID); err != nil {
			return err
		}
		for _, link := range links {
			if strings.TrimSpace(link.ID) == "" {
				return errOrgEmptyInput
			}
			if err := mysqlTeamExists(db, link.ID); err != nil {
				return err
			}
			perm := normalizePerm(link.PermFlag)
			if _, err := db.Exec(`INSERT INTO org_business_group_teams (business_group_id,team_id,perm_flag,created_at) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE perm_flag=VALUES(perm_flag)`, groupID, link.ID, perm, time.Now()); err != nil {
				return err
			}
		}
		return nil
	}
	mu.Lock()
	if orgBusinessGroups[groupID] == nil {
		mu.Unlock()
		return sql.ErrNoRows
	}
	for _, link := range links {
		if strings.TrimSpace(link.ID) == "" {
			mu.Unlock()
			return errOrgEmptyInput
		}
		if orgTeams[link.ID] == nil {
			mu.Unlock()
			return sql.ErrNoRows
		}
	}
	if orgBusinessTeams[groupID] == nil {
		orgBusinessTeams[groupID] = map[string]string{}
	}
	for _, link := range links {
		orgBusinessTeams[groupID][link.ID] = normalizePerm(link.PermFlag)
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}

func RemoveOrgBusinessTeam(groupID, teamID string) error {
	if mysqlOK {
		res, err := db.Exec(`DELETE FROM org_business_group_teams WHERE business_group_id=? AND team_id=?`, groupID, teamID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return sql.ErrNoRows
		}
		return nil
	}
	mu.Lock()
	if orgBusinessTeams[groupID] == nil || orgBusinessTeams[groupID][teamID] == "" {
		mu.Unlock()
		return sql.ErrNoRows
	}
	delete(orgBusinessTeams[groupID], teamID)
	mu.Unlock()
	return persistFallbackSnapshot()
}

func ListOrgRoles() []model.OrgRole {
	if mysqlOK {
		rows, err := db.Query(`SELECT id,name,note,builtin,created_at,updated_at FROM org_roles ORDER BY builtin DESC,name ASC`)
		if err == nil {
			defer rows.Close()
			out := []model.OrgRole{}
			for rows.Next() {
				var r model.OrgRole
				if rows.Scan(&r.ID, &r.Name, &r.Note, &r.Builtin, &r.CreatedAt, &r.UpdatedAt) == nil {
					out = append(out, r)
				}
			}
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	out := make([]model.OrgRole, 0, len(orgRoles))
	for _, r := range orgRoles {
		cp := *r
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func SaveOrgRole(id string, in model.OrgRoleInput) (model.OrgRole, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return model.OrgRole{}, errors.New("名称不能为空")
	}
	now := time.Now()
	if id == "" {
		id = NewID()
	}
	if mysqlOK {
		_, err := db.Exec(`REPLACE INTO org_roles (id,name,note,builtin,created_at,updated_at) VALUES (?,?,?,COALESCE((SELECT builtin FROM (SELECT builtin FROM org_roles WHERE id=?) x),0),COALESCE((SELECT created_at FROM (SELECT created_at FROM org_roles WHERE id=?) y),?),?)`, id, name, in.Note, id, id, now, now)
		if err != nil {
			return model.OrgRole{}, err
		}
		role, ok := getRoleByID(id)
		if !ok {
			return model.OrgRole{}, sql.ErrNoRows
		}
		return role, nil
	}
	mu.Lock()
	created, builtin := now, false
	if old := orgRoles[id]; old != nil {
		created, builtin = old.CreatedAt, old.Builtin
	}
	orgRoles[id] = &model.OrgRole{ID: id, Name: name, Note: in.Note, Builtin: builtin, CreatedAt: created, UpdatedAt: now}
	mu.Unlock()
	role, ok := getRoleByID(id)
	if !ok {
		return model.OrgRole{}, sql.ErrNoRows
	}
	return role, nil
}

func DeleteOrgRole(id string) error {
	role, ok := getRoleByID(id)
	if !ok {
		return sql.ErrNoRows
	}
	if role.Builtin {
		return errOrgRoleReadonly
	}
	if mysqlOK {
		db.Exec(`DELETE FROM org_role_operations WHERE role_id=?`, id)
		db.Exec(`DELETE FROM org_user_roles WHERE role_id=?`, id)
		_, err := db.Exec(`DELETE FROM org_roles WHERE id=?`, id)
		return err
	}
	mu.Lock()
	delete(orgRoles, id)
	delete(orgRoleOperations, id)
	for _, roles := range orgUserRoles {
		delete(roles, id)
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}

func OrgOperations() []model.OrgOperationGroup { return orgOperationGroups }

func GetOrgRoleOperations(roleID string) []string {
	if mysqlOK {
		rows, err := db.Query(`SELECT operation FROM org_role_operations WHERE role_id=?`, roleID)
		if err == nil {
			defer rows.Close()
			var out []string
			for rows.Next() {
				var op string
				if rows.Scan(&op) == nil {
					out = append(out, op)
				}
			}
			sort.Strings(out)
			return out
		}
	}
	mu.RLock()
	defer mu.RUnlock()
	var out []string
	for op := range orgRoleOperations[roleID] {
		out = append(out, op)
	}
	sort.Strings(out)
	return out
}

func SetOrgRoleOperations(roleID string, ops []string) error {
	if role, ok := getRoleByID(roleID); ok && role.Builtin {
		return errOrgRoleReadonly
	}
	allowed := map[string]bool{}
	for _, op := range flattenOrgOperations() {
		allowed[op] = true
	}
	if mysqlOK {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		tx.Exec(`DELETE FROM org_role_operations WHERE role_id=?`, roleID)
		for _, op := range ops {
			if allowed[op] {
				if _, err := tx.Exec(`INSERT INTO org_role_operations (role_id,operation,created_at) VALUES (?,?,?)`, roleID, op, time.Now()); err != nil {
					_ = tx.Rollback()
					return err
				}
			}
		}
		return tx.Commit()
	}
	mu.Lock()
	orgRoleOperations[roleID] = map[string]bool{}
	for _, op := range ops {
		if allowed[op] {
			orgRoleOperations[roleID][op] = true
		}
	}
	mu.Unlock()
	return persistFallbackSnapshot()
}
