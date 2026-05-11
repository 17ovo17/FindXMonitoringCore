package model

import "time"

type OrgRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OrgUser struct {
	ID            string            `json:"id"`
	Username      string            `json:"username"`
	Nickname      string            `json:"nickname,omitempty"`
	Email         string            `json:"email,omitempty"`
	Phone         string            `json:"phone,omitempty"`
	Roles         []string          `json:"roles"`
	Contacts      map[string]string `json:"contacts,omitempty"`
	UserGroups    []OrgRef          `json:"user_groups"`
	BusiGroups    []OrgRef          `json:"busi_groups"`
	MustChangePwd bool              `json:"must_change_pwd"`
	LastActiveAt  int64             `json:"last_active_time,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type OrgUserInput struct {
	Username string            `json:"username"`
	Password string            `json:"password"`
	Confirm  string            `json:"confirm"`
	Nickname string            `json:"nickname"`
	Email    string            `json:"email"`
	Phone    string            `json:"phone"`
	Roles    []string          `json:"roles"`
	Contacts map[string]string `json:"contacts"`
}

type OrgUserList struct {
	Total int       `json:"total"`
	List  []OrgUser `json:"list"`
}

type OrgTeam struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note,omitempty"`
	ParentID  string    `json:"parent_id,omitempty"`
	Members   []OrgUser `json:"members,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrgTeamInput struct {
	Name     string `json:"name"`
	Note     string `json:"note"`
	ParentID string `json:"parent_id"`
}

type OrgBusinessGroup struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Note      string        `json:"note,omitempty"`
	ParentID  string        `json:"parent_id,omitempty"`
	Teams     []OrgTeamLink `json:"teams,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type OrgTeamLink struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	PermFlag string `json:"perm_flag"`
}

type OrgBusinessGroupInput struct {
	Name     string `json:"name"`
	Note     string `json:"note"`
	ParentID string `json:"parent_id"`
}

type OrgRole struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Note      string    `json:"note,omitempty"`
	Builtin   bool      `json:"builtin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrgRoleInput struct {
	Name string `json:"name"`
	Note string `json:"note"`
}

type OrgOperationGroup struct {
	Name  string         `json:"name"`
	CName string         `json:"cname"`
	Ops   []OrgOperation `json:"ops"`
}

type OrgOperation struct {
	Name  string `json:"name"`
	CName string `json:"cname"`
}
