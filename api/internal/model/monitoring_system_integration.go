package model

const (
	MonitoringSystemIntegrationStatusActive            = "active"
	MonitoringSystemIntegrationStatusBlockedByContract = "PENDING"
)

type MonitoringSystemIntegration struct {
	ID             string                                     `json:"id"`
	Weight         int                                        `json:"weight"`
	Name           string                                     `json:"name"`
	URL            string                                     `json:"url"`
	ConfigPreview  string                                     `json:"config_preview"`
	IsPrivate      bool                                       `json:"is_private"`
	TeamIDs        []int                                      `json:"team_ids"`
	Hide           bool                                       `json:"hide"`
	ShowInMenu     bool                                       `json:"show_in_menu"`
	Status         string                                     `json:"status"`
	CreateAt       int64                                      `json:"create_at"`
	UpdateAt       int64                                      `json:"update_at"`
	CreateBy       string                                     `json:"create_by"`
	UpdateBy       string                                     `json:"update_by"`
	Builtin        bool                                       `json:"builtin"`
	Capabilities   MonitoringSystemIntegrationCapabilities    `json:"capabilities"`
	BlockedActions []MonitoringSystemIntegrationBlockedAction `json:"blocked_actions"`
}

type MonitoringSystemIntegrationCapabilities struct {
	List          bool   `json:"list"`
	Detail        bool   `json:"detail"`
	Read          bool   `json:"read"`
	Write         bool   `json:"write"`
	Sort          bool   `json:"sort"`
	MenuEmbedding bool   `json:"menu_embedding"`
	OpenEmbedded  bool   `json:"open_embedded"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
}

type MonitoringSystemIntegrationBlockedAction struct {
	Action string `json:"action"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type MonitoringSystemIntegrationFilter struct {
	Query      string
	Status     string
	Visibility string
}

type MonitoringSystemIntegrationInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	ConfigPreview string `json:"config_preview"`
	IsPrivate     bool   `json:"is_private"`
	TeamIDs       []int  `json:"team_ids"`
	Weight        int    `json:"weight"`
	Hide          bool   `json:"hide"`
}

type MonitoringSystemIntegrationWeightInput struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

type MonitoringSystemIntegrationHideInput struct {
	Hide bool `json:"hide"`
}
