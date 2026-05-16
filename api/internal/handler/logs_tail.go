package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var logsTailUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

// LogsTailWebSocket upgrades to WebSocket and streams real-time logs from Loki tail.
// GET /api/v1/logs/tail/ws
func LogsTailWebSocket(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "loki datasource is not configured"})
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	// Upgrade client connection to WebSocket
	clientWS, err := logsTailUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Warn("logs_tail: client websocket upgrade failed")
		return
	}
	defer clientWS.Close()

	// Build Loki tail WebSocket URL
	lokiWsURL := strings.Replace(lokiURL, "http://", "ws://", 1)
	lokiWsURL = strings.Replace(lokiWsURL, "https://", "wss://", 1)

	params := url.Values{}
	params.Set("query", query)
	if limit := strings.TrimSpace(c.Query("limit")); limit != "" {
		params.Set("limit", limit)
	}
	if delayFor := strings.TrimSpace(c.Query("delay_for")); delayFor != "" {
		params.Set("delay_for", delayFor)
	}
	if start := strings.TrimSpace(c.Query("start")); start != "" {
		params.Set("start", start)
	}

	tailURL := lokiWsURL + "/loki/api/v1/tail?" + params.Encode()

	// Connect to Loki tail WebSocket
	header := http.Header{}
	if auth := c.GetHeader("X-Loki-Auth"); auth != "" {
		header.Set("Authorization", auth)
	}
	if orgID := c.GetHeader("X-Scope-OrgID"); orgID != "" {
		header.Set("X-Scope-OrgID", orgID)
	}

	lokiWS, _, err := websocket.DefaultDialer.Dial(tailURL, header)
	if err != nil {
		logrus.WithError(err).Warn("logs_tail: loki websocket connection failed")
		writeWSError(clientWS, "failed to connect to loki tail endpoint")
		return
	}
	defer lokiWS.Close()

	// Channel to signal shutdown
	done := make(chan struct{})

	// Read from client (detect disconnect)
	go func() {
		defer close(done)
		for {
			_, _, err := clientWS.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Forward messages from Loki to client
	for {
		select {
		case <-done:
			return
		default:
		}

		lokiWS.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, message, err := lokiWS.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logrus.WithError(err).Warn("logs_tail: loki websocket read error")
			}
			return
		}

		// Transform Loki tail message to FindX format
		records := parseLokiTailMessage(message)
		if len(records) == 0 {
			continue
		}

		payload, err := json.Marshal(gin.H{
			"type":  "logs",
			"items": records,
			"count": len(records),
		})
		if err != nil {
			continue
		}

		if err := clientWS.WriteMessage(websocket.TextMessage, payload); err != nil {
			return
		}
	}
}

func writeWSError(ws *websocket.Conn, msg string) {
	payload, _ := json.Marshal(gin.H{"type": "error", "error": msg})
	ws.WriteMessage(websocket.TextMessage, payload)
}

// lokiTailResponse represents a Loki tail WebSocket message.
type lokiTailResponse struct {
	Streams []struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	} `json:"streams"`
	DroppedEntries []struct {
		Labels    map[string]string `json:"labels"`
		Timestamp string            `json:"timestamp"`
	} `json:"dropped_entries"`
}

func parseLokiTailMessage(data []byte) []lokiFindXRecord {
	var tailResp lokiTailResponse
	if err := json.Unmarshal(data, &tailResp); err != nil {
		return nil
	}

	records := make([]lokiFindXRecord, 0, 16)
	for _, stream := range tailResp.Streams {
		streamName := ""
		level := "info"
		if v, ok := stream.Stream["stream"]; ok {
			streamName = v
		}
		if v, ok := stream.Stream["level"]; ok {
			level = v
		} else if v, ok := stream.Stream["severity"]; ok {
			level = v
		}

		for _, entry := range stream.Values {
			if len(entry) < 2 {
				continue
			}
			ts := entry[0]
			// Try to parse nanosecond timestamp to RFC3339
			if nsec, err := parseInt64(ts); err == nil {
				ts = time.Unix(0, nsec).UTC().Format(time.RFC3339Nano)
			}
			records = append(records, lokiFindXRecord{
				Timestamp: ts,
				Message:   sanitizeLogString(entry[1]),
				Level:     level,
				Labels:    stream.Stream,
				Stream:    streamName,
			})
		}
	}
	return records
}

func parseInt64(s string) (int64, error) {
	n := int64(0)
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errNotInt
		}
		n = n*10 + int64(c-'0')
	}
	return n, nil
}

var errNotInt = &parseError{"not an integer"}

type parseError struct{ msg string }

func (e *parseError) Error() string { return e.msg }
