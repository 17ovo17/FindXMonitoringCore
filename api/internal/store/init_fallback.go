package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type fallbackSnapshot struct {
	ChatSessions       map[string]*model.ChatSession      `json:"chat_sessions"`
	ChatMessages       map[string][]model.ChatMessage     `json:"chat_messages"`
	TopologyBusinesses map[string]*model.TopologyBusiness `json:"topology_businesses"`
	Users              map[string]*fallbackSnapshotUser   `json:"users,omitempty"`
	OrgUserProfiles    map[string]*orgUserProfile         `json:"org_user_profiles,omitempty"`
	OrgTeams           map[string]*model.OrgTeam          `json:"org_teams,omitempty"`
	OrgTeamMembers     map[string]map[string]bool         `json:"org_team_members,omitempty"`
	OrgBusinessGroups  map[string]*model.OrgBusinessGroup `json:"org_business_groups,omitempty"`
	OrgBusinessTeams   map[string]map[string]string       `json:"org_business_teams,omitempty"`
	OrgRoles           map[string]*model.OrgRole          `json:"org_roles,omitempty"`
	OrgUserRoles       map[string]map[string]bool         `json:"org_user_roles,omitempty"`
	OrgRoleOperations  map[string]map[string]bool         `json:"org_role_operations,omitempty"`
}

type fallbackSnapshotUser struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	PasswordHash  string    `json:"password_hash"`
	Role          string    `json:"role"`
	MustChangePwd bool      `json:"must_change_pwd"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func fallbackSnapshotPath() string {
	if value := strings.TrimSpace(viper.GetString("storage.fallback_file")); value != "" {
		return value
	}
	return filepath.Join("data", "memory-store.json")
}

func loadFallbackSnapshot() {
	if mysqlOK {
		return
	}
	path := fallbackSnapshotPath()
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return
	}
	var snap fallbackSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		logrus.Warnf("memory fallback snapshot ignored: %v", err)
		return
	}
	mu.Lock()
	if snap.ChatSessions != nil {
		chatSessions = snap.ChatSessions
	}
	if snap.ChatMessages != nil {
		chatMessages = snap.ChatMessages
	}
	if snap.TopologyBusinesses != nil {
		topologyBusinesses = snap.TopologyBusinesses
	}
	if snap.Users != nil {
		users = map[string]*model.User{}
		for username, user := range snap.Users {
			users[username] = &model.User{
				ID:            user.ID,
				Username:      user.Username,
				PasswordHash:  user.PasswordHash,
				Role:          user.Role,
				MustChangePwd: user.MustChangePwd,
				CreatedAt:     user.CreatedAt,
				UpdatedAt:     user.UpdatedAt,
			}
		}
	}
	if snap.OrgUserProfiles != nil {
		orgUserProfiles = snap.OrgUserProfiles
	}
	if snap.OrgTeams != nil {
		orgTeams = snap.OrgTeams
	}
	if snap.OrgTeamMembers != nil {
		orgTeamMembers = snap.OrgTeamMembers
	}
	if snap.OrgBusinessGroups != nil {
		orgBusinessGroups = snap.OrgBusinessGroups
	}
	if snap.OrgBusinessTeams != nil {
		orgBusinessTeams = snap.OrgBusinessTeams
	}
	if snap.OrgRoles != nil {
		orgRoles = snap.OrgRoles
	}
	if snap.OrgUserRoles != nil {
		orgUserRoles = snap.OrgUserRoles
	}
	if snap.OrgRoleOperations != nil {
		orgRoleOperations = snap.OrgRoleOperations
	}
	mu.Unlock()
	logrus.Infof("loaded memory fallback snapshot from %s", path)
}

func persistFallbackSnapshot() error {
	if mysqlOK {
		return nil
	}
	mu.RLock()
	snap := fallbackSnapshot{
		ChatSessions:       map[string]*model.ChatSession{},
		ChatMessages:       map[string][]model.ChatMessage{},
		TopologyBusinesses: map[string]*model.TopologyBusiness{},
		Users:              map[string]*fallbackSnapshotUser{},
		OrgUserProfiles:    map[string]*orgUserProfile{},
		OrgTeams:           map[string]*model.OrgTeam{},
		OrgTeamMembers:     map[string]map[string]bool{},
		OrgBusinessGroups:  map[string]*model.OrgBusinessGroup{},
		OrgBusinessTeams:   map[string]map[string]string{},
		OrgRoles:           map[string]*model.OrgRole{},
		OrgUserRoles:       map[string]map[string]bool{},
		OrgRoleOperations:  map[string]map[string]bool{},
	}
	for id, session := range chatSessions {
		cp := *session
		snap.ChatSessions[id] = &cp
	}
	for id, messages := range chatMessages {
		snap.ChatMessages[id] = append([]model.ChatMessage{}, messages...)
	}
	for id, business := range topologyBusinesses {
		cp := *business
		snap.TopologyBusinesses[id] = &cp
	}
	for username, user := range users {
		snap.Users[username] = &fallbackSnapshotUser{
			ID:            user.ID,
			Username:      user.Username,
			PasswordHash:  user.PasswordHash,
			Role:          user.Role,
			MustChangePwd: user.MustChangePwd,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}
	}
	for id, profile := range orgUserProfiles {
		cp := *profile
		cp.Contacts = cleanContacts(profile.Contacts)
		snap.OrgUserProfiles[id] = &cp
	}
	for id, team := range orgTeams {
		cp := *team
		snap.OrgTeams[id] = &cp
	}
	for teamID, members := range orgTeamMembers {
		snap.OrgTeamMembers[teamID] = map[string]bool{}
		for userID, ok := range members {
			snap.OrgTeamMembers[teamID][userID] = ok
		}
	}
	for id, group := range orgBusinessGroups {
		cp := *group
		snap.OrgBusinessGroups[id] = &cp
	}
	for groupID, teams := range orgBusinessTeams {
		snap.OrgBusinessTeams[groupID] = map[string]string{}
		for teamID, perm := range teams {
			snap.OrgBusinessTeams[groupID][teamID] = perm
		}
	}
	for id, role := range orgRoles {
		cp := *role
		snap.OrgRoles[id] = &cp
	}
	for userID, roles := range orgUserRoles {
		snap.OrgUserRoles[userID] = map[string]bool{}
		for roleID, ok := range roles {
			snap.OrgUserRoles[userID][roleID] = ok
		}
	}
	for roleID, ops := range orgRoleOperations {
		snap.OrgRoleOperations[roleID] = map[string]bool{}
		for op, ok := range ops {
			snap.OrgRoleOperations[roleID][op] = ok
		}
	}
	mu.RUnlock()
	path := fallbackSnapshotPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		logrus.Warnf("memory fallback snapshot mkdir failed: %v", err)
		return err
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		logrus.Warnf("memory fallback snapshot marshal failed: %v", err)
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		logrus.Warnf("memory fallback snapshot write failed: %v", err)
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		logrus.Warnf("memory fallback snapshot replace failed: %v", err)
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
