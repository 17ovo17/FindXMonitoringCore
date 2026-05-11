package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"ai-workbench-api/internal/model"
	"ai-workbench-api/internal/store"

	"github.com/gin-gonic/gin"
)

func ListNotificationTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": store.ListNotificationTemplates(c.Query("notify_channel_ident"))})
}

func GetNotificationTemplate(c *gin.Context) {
	tpl, ok := store.GetNotificationTemplate(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification template not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

func SaveNotificationTemplates(c *gin.Context) {
	var items []model.NotificationTemplate
	if err := bindArrayOrSingle(c, &items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification template payload"})
		return
	}
	saved, err := store.SaveNotificationTemplates(items, notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification template requires name and channel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": saved})
}

func UpdateNotificationTemplate(c *gin.Context) {
	var tpl model.NotificationTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification template payload"})
		return
	}
	tpl.ID = c.Param("id")
	saved, err := store.SaveNotificationTemplate(&tpl, notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification template requires name and channel"})
		return
	}
	c.JSON(http.StatusOK, saved)
}

func DeleteNotificationTemplates(c *gin.Context) {
	ids := parseIDsPayload(c)
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": store.DeleteNotificationTemplates(ids)})
}

func CloneNotificationTemplate(c *gin.Context) {
	tpl, ok, err := store.CloneNotificationTemplate(c.Param("id"), notificationActor(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "notification template cannot be cloned"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification template not found"})
		return
	}
	c.JSON(http.StatusOK, tpl)
}

func PreviewNotificationTemplate(c *gin.Context) {
	var req notificationPreviewRequest
	if err := bindNotificationPreview(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preview payload"})
		return
	}
	if len(req.TPL.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template content required"})
		return
	}
	renderTemplatePreview(c, req.TPL.Content, req.EventIDs, req.Event)
}

func PreviewNotificationTemplateByID(c *gin.Context) {
	tpl, ok := store.GetNotificationTemplate(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification template not found"})
		return
	}
	var req notificationPreviewRequest
	if err := bindNotificationPreview(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preview payload"})
		return
	}
	renderTemplatePreview(c, tpl.Content, req.EventIDs, req.Event)
}

type notificationPreviewRequest struct {
	EventIDs []string `json:"event_ids"`
	TPL      struct {
		Content map[string]string `json:"content"`
	} `json:"tpl"`
	Event map[string]any `json:"event"`
}

func bindNotificationPreview(c *gin.Context, req *notificationPreviewRequest) error {
	if c.Request.ContentLength == 0 {
		return nil
	}
	return c.ShouldBindJSON(req)
}

func renderTemplatePreview(c *gin.Context, content map[string]string, eventIDs []string, event map[string]any) {
	if event == nil {
		event = firstNotificationPreviewEvent(eventIDs)
	}
	rendered, missing := renderNotificationTemplate(content, event)
	c.JSON(http.StatusOK, model.NotificationRenderResult{Content: rendered, Event: event, Missing: missing})
}

var notificationTemplateVarRE = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

func renderNotificationTemplate(content map[string]string, event map[string]any) (map[string]string, []string) {
	out := map[string]string{}
	missingSet := map[string]bool{}
	for key, text := range content {
		out[key] = notificationTemplateVarRE.ReplaceAllStringFunc(text, func(token string) string {
			match := notificationTemplateVarRE.FindStringSubmatch(token)
			if len(match) != 2 {
				return token
			}
			value, ok := lookupNotificationValue(event, match[1])
			if !ok {
				missingSet[match[1]] = true
				return token
			}
			return fmt.Sprint(value)
		})
	}
	missing := make([]string, 0, len(missingSet))
	for key := range missingSet {
		missing = append(missing, key)
	}
	return out, missing
}

func lookupNotificationValue(root map[string]any, path string) (any, bool) {
	var cur any = root
	for _, part := range strings.Split(path, ".") {
		switch typed := cur.(type) {
		case map[string]any:
			next, ok := typed[part]
			if !ok {
				return nil, false
			}
			cur = next
		case map[string]string:
			value, ok := typed[part]
			if !ok {
				return nil, false
			}
			cur = value
		default:
			return nil, false
		}
	}
	return cur, true
}
