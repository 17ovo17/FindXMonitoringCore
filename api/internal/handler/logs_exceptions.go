package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// In-memory exception store for tracking error groups.
var (
	exceptionsMu sync.RWMutex
	exceptions   = map[string]*LogException{}
)

// LogException represents a grouped exception from logs.
type LogException struct {
	ID          string                   `json:"id"`
	Signature   string                   `json:"signature"`
	ErrorType   string                   `json:"error_type"`
	Message     string                   `json:"message"`
	Service     string                   `json:"service"`
	Count       int                      `json:"count"`
	FirstSeen   time.Time                `json:"first_seen"`
	LastSeen    time.Time                `json:"last_seen"`
	Status      string                   `json:"status"` // active, resolved, muted
	Occurrences []LogExceptionOccurrence `json:"occurrences,omitempty"`
}

// LogExceptionOccurrence represents a single occurrence of an exception.
type LogExceptionOccurrence struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Message    string            `json:"message"`
	Stacktrace string            `json:"stacktrace,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// ListLogExceptions returns exception groups.
// GET /api/v1/logs/exceptions
func ListLogExceptions(c *gin.Context) {
	service := strings.TrimSpace(c.Query("service"))
	status := strings.TrimSpace(c.Query("status"))

	exceptionsMu.RLock()
	items := make([]*LogException, 0, len(exceptions))
	for _, ex := range exceptions {
		if service != "" && ex.Service != service {
			continue
		}
		if status != "" && ex.Status != status {
			continue
		}
		// Return without occurrences in list view
		cp := *ex
		cp.Occurrences = nil
		items = append(items, &cp)
	}
	exceptionsMu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].LastSeen.After(items[j].LastSeen)
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"items":  items,
		"total":  len(items),
	})
}

// GetLogException returns exception detail with occurrences.
// GET /api/v1/logs/exceptions/:id
func GetLogException(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exception id is required"})
		return
	}

	exceptionsMu.RLock()
	ex, ok := exceptions[id]
	exceptionsMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "exception not found"})
		return
	}

	c.JSON(http.StatusOK, ex)
}

// ResolveLogException marks an exception as resolved.
// POST /api/v1/logs/exceptions/:id/resolve
func ResolveLogException(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exception id is required"})
		return
	}

	exceptionsMu.Lock()
	ex, ok := exceptions[id]
	if !ok {
		exceptionsMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "exception not found"})
		return
	}
	ex.Status = "resolved"
	exceptionsMu.Unlock()

	c.JSON(http.StatusOK, ex)
}

// ReportLogException ingests a new exception occurrence and groups it.
// POST /api/v1/logs/exceptions
func ReportLogException(c *gin.Context) {
	var input struct {
		ErrorType  string            `json:"error_type"`
		Message    string            `json:"message"`
		Service    string            `json:"service"`
		Stacktrace string            `json:"stacktrace"`
		Labels     map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exception payload"})
		return
	}
	if strings.TrimSpace(input.ErrorType) == "" && strings.TrimSpace(input.Message) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error_type or message is required"})
		return
	}

	sig := computeExceptionSignature(input.ErrorType, input.Message, input.Service)
	now := time.Now()
	occurrence := LogExceptionOccurrence{
		ID:         newExceptionID(),
		Timestamp:  now,
		Message:    sanitizeLogString(input.Message),
		Stacktrace: sanitizeLogString(input.Stacktrace),
		Labels:     input.Labels,
	}

	exceptionsMu.Lock()
	// Find existing exception by signature
	var existing *LogException
	for _, ex := range exceptions {
		if ex.Signature == sig {
			existing = ex
			break
		}
	}

	if existing != nil {
		existing.Count++
		existing.LastSeen = now
		existing.Occurrences = append(existing.Occurrences, occurrence)
		// Keep only last 50 occurrences
		if len(existing.Occurrences) > 50 {
			existing.Occurrences = existing.Occurrences[len(existing.Occurrences)-50:]
		}
		// If it was resolved, reopen it
		if existing.Status == "resolved" {
			existing.Status = "active"
		}
		exceptionsMu.Unlock()
		c.JSON(http.StatusOK, existing)
		return
	}

	// Create new exception group
	ex := &LogException{
		ID:          newExceptionID(),
		Signature:   sig,
		ErrorType:   strings.TrimSpace(input.ErrorType),
		Message:     sanitizeLogString(input.Message),
		Service:     strings.TrimSpace(input.Service),
		Count:       1,
		FirstSeen:   now,
		LastSeen:    now,
		Status:      "active",
		Occurrences: []LogExceptionOccurrence{occurrence},
	}
	exceptions[ex.ID] = ex
	exceptionsMu.Unlock()

	c.JSON(http.StatusOK, ex)
}

func computeExceptionSignature(errorType, message, service string) string {
	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(errorType)))
	h.Write([]byte("|"))
	h.Write([]byte(normalizeExceptionMessage(message)))
	h.Write([]byte("|"))
	h.Write([]byte(strings.TrimSpace(service)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// normalizeExceptionMessage strips variable parts (numbers, UUIDs) to group similar errors.
func normalizeExceptionMessage(msg string) string {
	msg = strings.TrimSpace(msg)
	// Replace UUIDs with placeholder
	parts := strings.Fields(msg)
	normalized := make([]string, 0, len(parts))
	for _, p := range parts {
		if looksLikeUUID(p) {
			normalized = append(normalized, "<id>")
		} else if looksLikeNumber(p) {
			normalized = append(normalized, "<n>")
		} else {
			normalized = append(normalized, p)
		}
	}
	return strings.Join(normalized, " ")
}

func looksLikeUUID(s string) bool {
	if len(s) < 32 {
		return false
	}
	dashes := 0
	hexChars := 0
	for _, c := range s {
		if c == '-' {
			dashes++
		} else if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			hexChars++
		} else {
			return false
		}
	}
	return hexChars >= 32 && dashes <= 4
}

func looksLikeNumber(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func newExceptionID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "ex_" + hex.EncodeToString(b)
}
