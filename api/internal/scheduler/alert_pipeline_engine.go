package scheduler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/sirupsen/logrus"
)

// pipelineHTTPClient 用于 callback 处理器的异步 HTTP 请求。
var pipelineHTTPClient = &http.Client{Timeout: 10 * time.Second}

// ApplyPipelines 将事件依次通过所有已启用的流水线处理器。
// 返回处理后的事件和是否应丢弃（不发送通知）。
func ApplyPipelines(event *model.MonitorAlertEvent) (*model.MonitorAlertEvent, bool) {
	if event == nil {
		return event, false
	}
	pipelines := store.ListEnabledAlertPipelines()
	for _, pipeline := range pipelines {
		for _, proc := range pipeline.Processors {
			if !proc.Enabled {
				continue
			}
			if !matchPipelineConditions(event, proc.Conditions) {
				continue
			}
			switch proc.Type {
			case "relabel":
				applyRelabel(event, proc.Config)
			case "drop":
				logrus.WithFields(logrus.Fields{
					"pipeline": pipeline.Name,
					"event":    event.ID,
				}).Info("alert event dropped by pipeline")
				return event, true
			case "callback":
				go applyCallback(event, proc.Config)
			case "enrich":
				applyEnrich(event, proc.Config)
			}
		}
	}
	return event, false
}

// matchPipelineConditions 检查事件是否满足所有条件（AND 语义）。
func matchPipelineConditions(event *model.MonitorAlertEvent, conditions []model.LabelMatcher) bool {
	if len(conditions) == 0 {
		return true
	}
	for _, cond := range conditions {
		if !matchSingleCondition(event, cond) {
			return false
		}
	}
	return true
}

func matchSingleCondition(event *model.MonitorAlertEvent, cond model.LabelMatcher) bool {
	value := getEventLabelValue(event, cond.Key)
	switch cond.Operator {
	case "=":
		return value == cond.Value
	case "!=":
		return value != cond.Value
	case "=~":
		matched, err := regexp.MatchString(cond.Value, value)
		if err != nil {
			logrus.WithError(err).WithField("pattern", cond.Value).Warn("pipeline regex match failed")
			return false
		}
		return matched
	case "!~":
		matched, err := regexp.MatchString(cond.Value, value)
		if err != nil {
			logrus.WithError(err).WithField("pattern", cond.Value).Warn("pipeline regex match failed")
			return false
		}
		return !matched
	default:
		return false
	}
}

func getEventLabelValue(event *model.MonitorAlertEvent, key string) string {
	switch key {
	case "__name__":
		return event.Name
	case "__severity__":
		return event.Severity
	case "__status__":
		return event.Status
	case "__rule_id__":
		return event.RuleID
	case "__datasource_id__":
		return event.DatasourceID
	default:
		if event.Labels != nil {
			return event.Labels[key]
		}
		return ""
	}
}

// applyRelabel 修改事件标签。
// 配置支持: "actions" 数组，每个 action 包含 "action"(set/delete), "key", "value"。
func applyRelabel(event *model.MonitorAlertEvent, config map[string]any) {
	if event.Labels == nil {
		event.Labels = map[string]string{}
	}
	actions, ok := config["actions"]
	if !ok {
		return
	}
	actionList, ok := actions.([]any)
	if !ok {
		return
	}
	for _, item := range actionList {
		action, ok := item.(map[string]any)
		if !ok {
			continue
		}
		act, _ := action["action"].(string)
		key, _ := action["key"].(string)
		if key == "" {
			continue
		}
		switch act {
		case "set":
			val, _ := action["value"].(string)
			event.Labels[key] = val
		case "delete":
			delete(event.Labels, key)
		}
	}
}

// applyCallback 异步将事件数据 POST 到配置的 URL。
func applyCallback(event *model.MonitorAlertEvent, config map[string]any) {
	urlStr, _ := config["url"].(string)
	if strings.TrimSpace(urlStr) == "" {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		logrus.WithError(err).Error("pipeline callback marshal failed")
		return
	}
	req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		logrus.WithError(err).WithField("url", urlStr).Error("pipeline callback request build failed")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if headers, ok := config["headers"].(map[string]any); ok {
		for k, v := range headers {
			if sv, ok := v.(string); ok {
				req.Header.Set(k, sv)
			}
		}
	}
	resp, err := pipelineHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("url", urlStr).Warn("pipeline callback request failed")
		return
	}
	resp.Body.Close()
}

// applyEnrich 从配置中添加注解到事件。
// 配置支持: "annotations" map[string]string 直接追加到事件 annotations。
func applyEnrich(event *model.MonitorAlertEvent, config map[string]any) {
	if event.Annotations == nil {
		event.Annotations = map[string]string{}
	}
	annotations, ok := config["annotations"]
	if !ok {
		return
	}
	annotationMap, ok := annotations.(map[string]any)
	if !ok {
		return
	}
	for k, v := range annotationMap {
		if sv, ok := v.(string); ok {
			event.Annotations[k] = sv
		}
	}
}
