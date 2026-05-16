package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LogsFieldsDiscovery returns available log fields with value counts by querying Loki labels.
// GET /api/v1/logs/fields/discover
func LogsFieldsDiscovery(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"fields": builtinLogFields().Fields,
			"source": "builtin",
		})
		return
	}

	start := strings.TrimSpace(c.Query("start"))
	end := strings.TrimSpace(c.Query("end"))

	params := url.Values{}
	if start != "" {
		params.Set("start", start)
	}
	if end != "" {
		params.Set("end", end)
	}

	target := lokiURL + "/loki/api/v1/labels"
	if qs := params.Encode(); qs != "" {
		target += "?" + qs
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), lokiProxyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki request failed"})
		return
	}
	copyLokiAuthHeaders(c, req)

	resp, err := lokiHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("loki: fields discovery failed")
		c.JSON(http.StatusOK, gin.H{
			"status": "fallback",
			"fields": builtinLogFields().Fields,
			"source": "builtin",
			"error":  "loki unreachable, showing builtin fields",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		c.JSON(http.StatusOK, gin.H{
			"status": "fallback",
			"fields": builtinLogFields().Fields,
			"source": "builtin",
			"error":  "loki returned non-success status",
		})
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read loki response failed"})
		return
	}

	var lokiLabels struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal(data, &lokiLabels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "parse loki labels failed"})
		return
	}

	fields := make([]gin.H, 0, len(lokiLabels.Data))
	for _, label := range lokiLabels.Data {
		fields = append(fields, gin.H{
			"key":      label,
			"type":     "string",
			"category": "loki",
			"source":   "loki",
			"indexed":  true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"fields": fields,
		"source": "loki",
		"total":  len(fields),
	})
}

// LogsAggregateLoki aggregates logs by a field using Loki metric queries.
// GET /api/v1/logs/aggregate
// Params: query (LogQL), group_by, start, end
func LogsAggregateLoki(c *gin.Context) {
	lokiURL := getLokiURL()
	if lokiURL == "" {
		blockLokiContract(c, http.StatusServiceUnavailable, "loki datasource is not configured")
		return
	}

	query := strings.TrimSpace(c.Query("query"))
	groupBy := strings.TrimSpace(c.Query("group_by"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}
	if groupBy == "" {
		groupBy = "level"
	}

	// Build a count_over_time metric query grouped by the field
	metricQuery := `sum by (` + groupBy + `) (count_over_time(` + query + ` [5m]))`

	params := url.Values{}
	params.Set("query", metricQuery)
	if start := strings.TrimSpace(c.Query("start")); start != "" {
		params.Set("start", start)
	} else {
		params.Set("start", strconv.FormatInt(time.Now().Add(-1*time.Hour).UnixNano(), 10))
	}
	if end := strings.TrimSpace(c.Query("end")); end != "" {
		params.Set("end", end)
	} else {
		params.Set("end", strconv.FormatInt(time.Now().UnixNano(), 10))
	}
	params.Set("step", "300")

	target := lokiURL + "/loki/api/v1/query_range?" + params.Encode()
	ctx, cancel := context.WithTimeout(c.Request.Context(), lokiProxyTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create loki request failed"})
		return
	}
	copyLokiAuthHeaders(c, req)

	resp, err := lokiHTTPClient.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("loki: aggregate query failed")
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		blockLokiUpstream(c, http.StatusBadGateway)
		return
	}

	buckets := parseLokiAggregateResponse(data, groupBy)
	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Count > buckets[j].Count
	})

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"group_by": groupBy,
		"buckets":  buckets,
		"total":    len(buckets),
	})
}

type lokiAggBucket struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

func parseLokiAggregateResponse(data []byte, groupBy string) []lokiAggBucket {
	var resp struct {
		Data struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		logrus.WithError(err).Warn("loki: failed to parse aggregate response")
		return []lokiAggBucket{}
	}

	buckets := make([]lokiAggBucket, 0, len(resp.Data.Result))
	for _, series := range resp.Data.Result {
		key := series.Metric[groupBy]
		if key == "" {
			key = "unknown"
		}
		total := 0
		for _, val := range series.Values {
			if len(val) >= 2 {
				if s, ok := val[1].(string); ok {
					n, _ := strconv.Atoi(s)
					total += n
				}
			}
		}
		buckets = append(buckets, lokiAggBucket{
			Key:   key,
			Label: key,
			Count: total,
		})
	}
	return buckets
}
