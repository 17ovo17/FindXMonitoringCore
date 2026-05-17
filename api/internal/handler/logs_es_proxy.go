package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const esProxyTimeout = 30 * time.Second

var esHTTPClient = &http.Client{Timeout: esProxyTimeout}

func getESURL() string {
	if u := strings.TrimRight(strings.TrimSpace(viper.GetString("elasticsearch.url")), "/"); u != "" {
		return u
	}
	return strings.TrimRight(strings.TrimSpace(os.Getenv("ES_URL")), "/")
}

func getESIndex() string {
	if idx := strings.TrimSpace(viper.GetString("elasticsearch.index")); idx != "" {
		return idx
	}
	if idx := strings.TrimSpace(os.Getenv("ES_INDEX")); idx != "" {
		return idx
	}
	return "logs-*"
}

func LogsESQuery(c *gin.Context) {
	esURL := getESURL()
	if esURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "elasticsearch datasource is not configured", "code": "es_unavailable"})
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	size := 100
	if s := strings.TrimSpace(c.Query("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 1000 {
			size = n
		}
	}
	from := 0
	if p := strings.TrimSpace(c.Query("page")); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 1 {
			from = (n - 1) * size
		}
	}

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now
	if s := strings.TrimSpace(c.Query("start")); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			startTime = time.Unix(ts, 0)
		}
	}
	if s := strings.TrimSpace(c.Query("end")); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		}
	}

	esQuery := buildESSearchQuery(query, startTime, endTime, size, from, c.Query("severity"))
	index := getESIndex()
	target := esURL + "/" + index + "/_search"

	body, err := json.Marshal(esQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "build es query failed"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), esProxyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create es request failed"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := esHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("es: query proxy failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "elasticsearch upstream request failed", "code": "es_upstream_error"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "read es response failed", "code": "es_upstream_error"})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logrus.WithField("status", resp.StatusCode).Warn("es: upstream returned error")
		c.JSON(http.StatusBadGateway, gin.H{"error": "elasticsearch returned error", "code": "es_upstream_error"})
		return
	}

	records, total := parseESSearchResponse(respBody)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"items":  records,
		"total":  total,
	})
}

func LogsESAggregate(c *gin.Context) {
	esURL := getESURL()
	if esURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "elasticsearch datasource is not configured", "code": "es_unavailable"})
		return
	}

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now
	if s := strings.TrimSpace(c.Query("start")); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			startTime = time.Unix(ts, 0)
		}
	}
	if s := strings.TrimSpace(c.Query("end")); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			endTime = time.Unix(ts, 0)
		}
	}

	interval := esAutoInterval(startTime, endTime)
	query := strings.TrimSpace(c.Query("query"))

	esQuery := buildESAggregateQuery(query, startTime, endTime, c.Query("severity"), interval)
	index := getESIndex()
	target := esURL + "/" + index + "/_search"

	body, err := json.Marshal(esQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "build es aggregate query failed"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), esProxyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create es request failed"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := esHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("es: aggregate proxy failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "elasticsearch upstream request failed", "code": "es_upstream_error"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "read es response failed", "code": "es_upstream_error"})
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "elasticsearch returned error", "code": "es_upstream_error"})
		return
	}

	buckets := parseESAggregateResponse(respBody)
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"buckets": buckets,
	})
}

func LogsESFields(c *gin.Context) {
	esURL := getESURL()
	if esURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "elasticsearch datasource is not configured", "code": "es_unavailable"})
		return
	}

	index := getESIndex()
	target := esURL + "/" + index + "/_mapping"

	ctx, cancel := context.WithTimeout(c.Request.Context(), esProxyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create es request failed"})
		return
	}

	resp, err := esHTTPClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "elasticsearch upstream request failed", "code": "es_upstream_error"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "read es response failed"})
		return
	}

	fields := parseESMappingFields(respBody)
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"fields": fields,
	})
}

// --- ES query builders ---

func buildESSearchQuery(queryStr string, start, end time.Time, size, from int, severity string) map[string]interface{} {
	must := []map[string]interface{}{
		{"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte":    start.Format(time.RFC3339),
				"lte":    end.Format(time.RFC3339),
				"format": "strict_date_optional_time",
			},
		}},
	}

	if queryStr != "" {
		must = append(must, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":            queryStr,
				"default_operator": "AND",
			},
		})
	}

	if severity != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"level": strings.ToLower(severity),
			},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{"must": must},
		},
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]interface{}{"order": "desc"}},
		},
		"size": size,
		"from": from,
	}
}

func buildESAggregateQuery(queryStr string, start, end time.Time, severity, interval string) map[string]interface{} {
	must := []map[string]interface{}{
		{"range": map[string]interface{}{
			"@timestamp": map[string]interface{}{
				"gte":    start.Format(time.RFC3339),
				"lte":    end.Format(time.RFC3339),
				"format": "strict_date_optional_time",
			},
		}},
	}

	if queryStr != "" {
		must = append(must, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":            queryStr,
				"default_operator": "AND",
			},
		})
	}

	if severity != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{"level": strings.ToLower(severity)},
		})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{"must": must},
		},
		"size": 0,
		"aggs": map[string]interface{}{
			"timeline": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":          "@timestamp",
					"fixed_interval": interval,
				},
			},
		},
	}
}

func esAutoInterval(start, end time.Time) string {
	duration := end.Sub(start)
	switch {
	case duration <= 30*time.Minute:
		return "1m"
	case duration <= 3*time.Hour:
		return "5m"
	case duration <= 24*time.Hour:
		return "30m"
	default:
		return "1h"
	}
}

// --- ES response parsers ---

type esFindXRecord struct {
	Timestamp string            `json:"timestamp"`
	Message   string            `json:"message"`
	Level     string            `json:"level"`
	Labels    map[string]string `json:"labels"`
	Stream    string            `json:"stream"`
	ID        string            `json:"id"`
}

func parseESSearchResponse(data []byte) ([]esFindXRecord, int) {
	var resp struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				ID     string                 `json:"_id"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		logrus.WithError(err).Warn("es: failed to parse search response")
		return []esFindXRecord{}, 0
	}

	records := make([]esFindXRecord, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		record := esFindXRecord{ID: hit.ID}
		record.Timestamp = esExtractString(hit.Source, "@timestamp")
		record.Message = esExtractMessage(hit.Source)
		record.Level = esExtractLevel(hit.Source)
		record.Stream = esExtractString(hit.Source, "service.name")
		if record.Stream == "" {
			record.Stream = esExtractString(hit.Source, "host.name")
		}
		record.Labels = esExtractLabels(hit.Source)
		records = append(records, record)
	}
	return records, resp.Hits.Total.Value
}

func parseESAggregateResponse(data []byte) []map[string]interface{} {
	var resp struct {
		Aggregations struct {
			Timeline struct {
				Buckets []struct {
					KeyAsString string `json:"key_as_string"`
					Key         int64  `json:"key"`
					DocCount    int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"timeline"`
		} `json:"aggregations"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		logrus.WithError(err).Warn("es: failed to parse aggregate response")
		return []map[string]interface{}{}
	}

	buckets := make([]map[string]interface{}, 0, len(resp.Aggregations.Timeline.Buckets))
	for _, b := range resp.Aggregations.Timeline.Buckets {
		buckets = append(buckets, map[string]interface{}{
			"timestamp": b.KeyAsString,
			"count":     b.DocCount,
		})
	}
	return buckets
}

func parseESMappingFields(data []byte) []map[string]string {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return []map[string]string{}
	}

	fields := make([]map[string]string, 0, 32)
	for _, indexData := range raw {
		indexMap, ok := indexData.(map[string]interface{})
		if !ok {
			continue
		}
		mappings, ok := indexMap["mappings"].(map[string]interface{})
		if !ok {
			continue
		}
		properties, ok := mappings["properties"].(map[string]interface{})
		if !ok {
			continue
		}
		for fieldName, fieldData := range properties {
			fieldMap, ok := fieldData.(map[string]interface{})
			if !ok {
				continue
			}
			fieldType, _ := fieldMap["type"].(string)
			if fieldType == "" {
				fieldType = "object"
			}
			fields = append(fields, map[string]string{
				"name": fieldName,
				"type": fieldType,
			})
		}
		break
	}
	return fields
}

func esExtractString(source map[string]interface{}, key string) string {
	parts := strings.Split(key, ".")
	current := source
	for i, part := range parts {
		if i == len(parts)-1 {
			if v, ok := current[part]; ok {
				switch val := v.(type) {
				case string:
					return val
				case float64:
					return strconv.FormatFloat(val, 'f', -1, 64)
				}
			}
			return ""
		}
		nested, ok := current[part].(map[string]interface{})
		if !ok {
			return ""
		}
		current = nested
	}
	return ""
}

func esExtractMessage(source map[string]interface{}) string {
	for _, key := range []string{"message", "msg", "log", "body", "content"} {
		if v, ok := source[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	b, _ := json.Marshal(source)
	if len(b) > 500 {
		return string(b[:500])
	}
	return string(b)
}

func esExtractLevel(source map[string]interface{}) string {
	for _, key := range []string{"level", "severity", "log.level", "loglevel"} {
		if v := esExtractString(source, key); v != "" {
			return strings.ToLower(v)
		}
	}
	return "info"
}

func esExtractLabels(source map[string]interface{}) map[string]string {
	labels := make(map[string]string, 8)
	for _, key := range []string{"host.name", "service.name", "container.name", "agent.name", "source"} {
		if v := esExtractString(source, key); v != "" {
			labels[key] = v
		}
	}
	if v, ok := source["tags"]; ok {
		if s, ok := v.(string); ok {
			labels["tags"] = s
		}
	}
	return labels
}
