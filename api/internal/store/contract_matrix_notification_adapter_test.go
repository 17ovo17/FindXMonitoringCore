package store

import (
	"strings"
	"testing"

	"ai-workbench-api/internal/model"
)

func TestMonitoringContractMatrixNotificationAdapterSplit(t *testing.T) {
	ResetContractMatrixForTest()

	aggregate, ok, err := GetContractMatrixEntry("FX-CONTRACT-N9E-NOTIFICATION-FINDX-ADAPTER")
	if err != nil {
		t.Fatalf("get notification adapter aggregate: %v", err)
	}
	if !ok {
		t.Fatal("notification adapter aggregate seed missing")
	}
	if aggregate.Status != model.ContractStatusBlocked || aggregate.SafeToRetry {
		t.Fatalf("notification adapter aggregate must stay blocked and non-ready: %#v", aggregate)
	}
	if aggregate.Metadata["findx_route"] != "/notifications?section=rules" {
		t.Fatalf("aggregate findx_route = %q, want /notifications?section=rules", aggregate.Metadata["findx_route"])
	}
	if aggregate.Metadata["gap_type"] != "notification_adapter_aggregate" {
		t.Fatalf("aggregate gap_type = %q, want notification_adapter_aggregate", aggregate.Metadata["gap_type"])
	}
	if aggregate.Metadata["upstream_ref"] != "notification-adapter-aggregate" {
		t.Fatalf("aggregate upstream_ref = %q, want notification-adapter-aggregate", aggregate.Metadata["upstream_ref"])
	}
	for _, want := range []string{"rules", "channel configs", "templates", "contacts", "manage lookup", "test", "statistics"} {
		if !strings.Contains(aggregate.Metadata["upstream_scope"], want) {
			t.Fatalf("aggregate upstream_scope missing %q: %#v", want, aggregate.Metadata)
		}
	}
	if aggregate.Handler != "" || aggregate.Backend != "" || aggregate.Datasource != "" || aggregate.Executor != "" || len(aggregate.EvidenceRefs) != 0 {
		t.Fatalf("aggregate must not expose executable evidence: %#v", aggregate)
	}

	counts := map[string]int{}
	for _, tt := range notificationAdapterGapExpectations() {
		t.Run(tt.id, func(t *testing.T) {
			item, ok, err := GetContractMatrixEntry(tt.id)
			if err != nil {
				t.Fatalf("get %s: %v", tt.id, err)
			}
			if !ok {
				t.Fatalf("%s seed missing", tt.id)
			}
			if item.Status != tt.status {
				t.Fatalf("%s status = %s, want %s", tt.id, item.Status, tt.status)
			}
			if item.SafeToRetry || item.Handler != "" || item.Backend != "" || item.Datasource != "" || item.Executor != "" || len(item.EvidenceRefs) != 0 {
				t.Fatalf("%s should remain non-ready without executable evidence: %#v", tt.id, item)
			}
			if item.Metadata["findx_route"] != "/notifications?section=rules" {
				t.Fatalf("%s findx_route = %q, want /notifications?section=rules", tt.id, item.Metadata["findx_route"])
			}
			if item.Metadata["gap_type"] != tt.gapType {
				t.Fatalf("%s gap_type = %q, want %q", tt.id, item.Metadata["gap_type"], tt.gapType)
			}
			assertMonitoringN9eGapMetadata(t, item, tt.upstreamRef)
			for _, want := range notificationAdapterSourceRefs() {
				if !contractListContains(item.SourceRefs, want) {
					t.Fatalf("%s source_refs missing %s: %#v", tt.id, want, item.SourceRefs)
				}
			}
			assertNotificationAdapterTextIsSafe(t, item)
			counts[item.Status]++
		})
	}
	if counts[model.ContractStatusMissingBackend] != 26 ||
		counts[model.ContractStatusMissingDatasource] != 4 ||
		counts[model.ContractStatusMissingExecutor] != 5 {
		t.Fatalf("notification adapter status distribution mismatch: %#v", counts)
	}
}

func TestMonitoringContractMatrixNotificationAdapterSourceRefs(t *testing.T) {
	got := notificationAdapterSourceRefs()
	want := []string{
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\services.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\List.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\Form\TestButton.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\Detail\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\Detail\Events.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\Detail\AlertRules.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationRules\pages\Detail\SubscribeRules.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationChannels\services.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationChannels\pages\ListNG\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationChannels\pages\Form\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationTemplates\services.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\notificationTemplates\pages\List\index.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\contacts\services.ts`,
		`D:\项目迁移文件\平台源码\fe-main\src\pages\contacts\pages\List.tsx`,
		`D:\项目迁移文件\平台源码\fe-main\src\services\manage.ts`,
	}
	if len(got) != len(want) {
		t.Fatalf("notification adapter source refs count = %d, want %d: %#v", len(got), len(want), got)
	}
	for _, ref := range want {
		if !contractListContains(got, ref) {
			t.Fatalf("notification adapter source refs missing %q: %#v", ref, got)
		}
	}
}

func notificationAdapterGapExpectations() []struct {
	id          string
	status      string
	gapType     string
	upstreamRef string
} {
	return []struct {
		id          string
		status      string
		gapType     string
		upstreamRef string
	}{
		{"FX-CONTRACT-N9E-NOTIFY-RULES-LIST", model.ContractStatusMissingBackend, "notification_rules_list", "GET /api/n9e/notify-rules"},
		{"FX-CONTRACT-N9E-NOTIFY-RULE-DETAIL", model.ContractStatusMissingBackend, "notification_rule_detail", "GET /api/n9e/notify-rule/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-RULES-CREATE", model.ContractStatusMissingBackend, "notification_rules_create", "POST /api/n9e/notify-rules"},
		{"FX-CONTRACT-N9E-NOTIFY-RULE-UPDATE", model.ContractStatusMissingBackend, "notification_rule_update", "PUT /api/n9e/notify-rule/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-RULES-DELETE", model.ContractStatusMissingBackend, "notification_rules_delete", "DELETE /api/n9e/notify-rules"},
		{"FX-CONTRACT-N9E-NOTIFY-RULE-CUSTOM-PARAMS", model.ContractStatusMissingBackend, "notification_rule_custom_params", "GET /api/n9e/notify-rule/custom-params"},
		{"FX-CONTRACT-N9E-NOTIFY-RULE-TEST", model.ContractStatusMissingExecutor, "notification_rule_test", "POST /api/n9e/notify-rule/test"},
		{"FX-CONTRACT-N9E-NOTIFY-STATISTICS", model.ContractStatusMissingDatasource, "notification_statistics", "GET /api/n9e-plus/notify/{id}/statistics"},
		{"FX-CONTRACT-N9E-NOTIFY-EVENTS", model.ContractStatusMissingDatasource, "notification_events", "GET /api/n9e-plus/notify/{id}/alert-cur-events"},
		{"FX-CONTRACT-N9E-NOTIFY-ALERT-RULES", model.ContractStatusMissingBackend, "notification_alert_rules", "GET /api/n9e-plus/notify/{id}/alert-rules"},
		{"FX-CONTRACT-N9E-NOTIFY-SUBSCRIBE-RULES", model.ContractStatusMissingBackend, "notification_subscribe_rules", "GET /api/n9e-plus/notify/{id}/sub-alert-rules"},
		{"FX-CONTRACT-N9E-NOTIFY-EVENT-TAGKEYS", model.ContractStatusMissingDatasource, "notification_event_tagkeys", "GET /api/n9e/event-tagkeys"},
		{"FX-CONTRACT-N9E-NOTIFY-FEISHU-GROUPS", model.ContractStatusMissingExecutor, "notification_feishu_groups", "GET /api/n9e/feishu-visible-chats/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-FLASHDUTY-CHANNELS", model.ContractStatusMissingExecutor, "notification_flashduty_channels", "GET /api/n9e/flashduty-channel-list/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-SERVICES", model.ContractStatusMissingExecutor, "notification_pagerduty_services", "pagerduty-services-lookup"},
		{"FX-CONTRACT-N9E-NOTIFY-PAGERDUTY-CONNECTOR-LOOKUP", model.ContractStatusMissingExecutor, "notification_pagerduty_connector_lookup", "pagerduty-connector-lookup"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-LIST", model.ContractStatusMissingBackend, "notification_channel_configs_list", "GET /api/n9e/notify-channel-configs"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-SIMPLIFIED", model.ContractStatusMissingBackend, "notification_channel_configs_simplified", "GET /api/n9e/simplified-notify-channel-configs"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-CREATE", model.ContractStatusMissingBackend, "notification_channel_configs_create", "POST /api/n9e/notify-channel-configs"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-UPDATE", model.ContractStatusMissingBackend, "notification_channel_config_update", "PUT /api/n9e/notify-channel-config/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-DETAIL", model.ContractStatusMissingBackend, "notification_channel_config_detail", "GET /api/n9e/notify-channel-config/{id}"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-BY-IDENT", model.ContractStatusMissingBackend, "notification_channel_config_by_ident", "GET /api/n9e/notify-channel-config?ident={ident}"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIGS-DELETE", model.ContractStatusMissingBackend, "notification_channel_configs_delete", "DELETE /api/n9e/notify-channel-configs"},
		{"FX-CONTRACT-N9E-NOTIFY-CHANNEL-CONFIG-IDENTS", model.ContractStatusMissingBackend, "notification_channel_config_idents", "GET /api/n9e/notify-channel-config/idents"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATES-LIST", model.ContractStatusMissingBackend, "message_templates_list", "GET /api/n9e/message-templates"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATE-DETAIL", model.ContractStatusMissingBackend, "message_template_detail", "GET /api/n9e/message-template/{id}"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATES-CREATE", model.ContractStatusMissingBackend, "message_templates_create", "POST /api/n9e/message-templates"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATE-UPDATE", model.ContractStatusMissingBackend, "message_template_update", "PUT /api/n9e/message-template/{id}"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATES-DELETE", model.ContractStatusMissingBackend, "message_templates_delete", "DELETE /api/n9e/message-templates"},
		{"FX-CONTRACT-N9E-MESSAGE-TEMPLATE-PREVIEW", model.ContractStatusMissingDatasource, "message_template_preview", "POST /api/n9e/events-message"},
		{"FX-CONTRACT-N9E-NOTIFY-CONTACTS-LIST", model.ContractStatusMissingBackend, "notification_contacts_list", "GET /api/n9e/notify-contact"},
		{"FX-CONTRACT-N9E-NOTIFY-CONTACTS-UPDATE", model.ContractStatusMissingBackend, "notification_contacts_update", "PUT /api/n9e/notify-contact"},
		{"FX-CONTRACT-N9E-MANAGE-NOTIFY-CHANNELS", model.ContractStatusMissingBackend, "manage_notify_channels", "GET /api/n9e/notify-channels"},
		{"FX-CONTRACT-N9E-MANAGE-CONTACT-CHANNELS", model.ContractStatusMissingBackend, "manage_contact_channels", "GET /api/n9e/contact-channels"},
		{"FX-CONTRACT-N9E-MANAGE-CONTACT-KEYS", model.ContractStatusMissingBackend, "manage_contact_keys", "GET /api/n9e/contact-keys"},
	}
}

func assertNotificationAdapterTextIsSafe(t *testing.T, item model.ContractMatrixEntry) {
	t.Helper()
	values := []string{item.ID, item.Capability, item.BlockedReason}
	values = append(values, item.SourceRefs...)
	for key, value := range item.Metadata {
		values = append(values, key, value)
	}
	for _, value := range values {
		if containsContractSensitiveText(value) {
			t.Fatalf("%s contains sensitive contract text in %q: %#v", item.ID, value, item)
		}
		if containsContractFakeSuccessText(value) {
			t.Fatalf("%s contains fake runtime state in %q: %#v", item.ID, value, item)
		}
		for _, mojibake := range monitoringMojibakeDenylist() {
			if strings.Contains(value, mojibake) {
				t.Fatalf("%s contains mojibake %q in value %q: %#v", item.ID, mojibake, value, item)
			}
		}
	}
}
