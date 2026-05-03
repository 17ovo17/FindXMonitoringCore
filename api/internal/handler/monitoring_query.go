package handler

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"ai-workbench-api/internal/monitoring"

	"github.com/gin-gonic/gin"
)

const (
	defaultMonitoringDatasourceID = "prometheus-default"
	promInstantTimeout            = 5 * time.Second
	promRangeTimeout              = 10 * time.Second
)

type monitoringPromRequest struct {
	DatasourceID   string  `json:"datasource_id"`
	Query          string  `json:"query"`
	Time           float64 `json:"time"`
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
	Step           float64 `json:"step"`
	TimeoutMS      float64 `json:"timeout_ms"`
	TimeoutSeconds float64 `json:"timeout_seconds"`
}

type promCallResult struct {
	Body       map[string]any
	Data       map[string]any
	Stats      monitoring.ResultStats
	LatencyMS  int64
	StatusCode int
	Warnings   []string
	OK         bool
}

func ListMonitorDatasources(c *gin.Context) {
	items := []gin.H{}
	for _, ds := range loadDataSources() {
		if strings.EqualFold(ds.Type, "prometheus") {
			items = append(items, gin.H{
				"id": ds.ID, "name": ds.Name, "type": ds.Type, "url": sanitizeDatasourceURL(ds.URL),
				"username": ds.Username, "database": ds.Database, "default": ds.ID == defaultMonitoringDatasourceID,
			})
		}
	}
	sort.Slice(items, func(i, j int) bool { return fmt.Sprint(items[i]["id"]) < fmt.Sprint(items[j]["id"]) })
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func MonitorQuery(c *gin.Context) {
	runMonitorPromQuery(c, false)
}

func MonitorQueryRange(c *gin.Context) {
	runMonitorPromQuery(c, true)
}

func runMonitorPromQuery(c *gin.Context, ranged bool) {
	req, ok := bindAndValidateMonitorQuery(c, ranged)
	if !ok {
		return
	}
	dsID, base, ok := resolveMonitoringPrometheus(c, req.DatasourceID)
	if !ok {
		return
	}
	path, params, timeout := monitorPromRequestTarget(req, ranged)
	queryHash := monitorQueryHash(req.Query)
	result := callPrometheus(c.Request.Context(), base, path, params, timeout)
	if !result.OK {
		auditMonitorPromQuery(c, ranged, dsID, queryHash, result.LatencyMS, "failed", nil)
		writeMonitorError(c, result.StatusCode, "prometheus query failed")
		return
	}
	body := result.Body
	if ranged {
		body = sortPrometheusMatrix(body)
	}
	data, _ := body["data"].(map[string]any)
	stats := prometheusResultStats(data)
	auditMonitorPromQuery(c, ranged, dsID, queryHash, result.LatencyMS, "success", stats)
	c.JSON(http.StatusOK, gin.H{
		"datasource_id": dsID, "query_hash": queryHash, "status": "success",
		"latency_ms": result.LatencyMS, "warnings": result.Warnings,
		"data":  gin.H{"resultType": data["resultType"], "result": data["result"]},
		"stats": stats,
	})
}

func monitorQueryHash(query string) string {
	return monitoring.QueryHash(query)
}

func auditMonitorPromQuery(c *gin.Context, ranged bool, datasourceID, queryHash string, latencyMS int64, status string, stats gin.H) {
	action := "monitor.query"
	if ranged {
		action = "monitor.query_range"
	}
	resultType, seriesCount, sampleCount := "", 0, 0
	if stats != nil {
		resultType, _ = stats["result_type"].(string)
		seriesCount, _ = stats["series_count"].(int)
		sampleCount, _ = stats["sample_count"].(int)
	}
	detail := fmt.Sprintf("query_hash=%s status=%s latency_ms=%d result_type=%s series_count=%d sample_count=%d", queryHash, status, latencyMS, resultType, seriesCount, sampleCount)
	auditEvent(c, action, datasourceID, "medium", status, detail, c.GetHeader("X-Test-Batch-Id"))
}

func prometheusResultStats(data map[string]any) gin.H {
	stats := monitoring.ResultStatsForData(data)
	return gin.H{"series_count": stats.SeriesCount, "sample_count": stats.SampleCount, "result_type": stats.ResultType}
}

func bindAndValidateMonitorQuery(c *gin.Context, ranged bool) (monitoringPromRequest, bool) {
	var req monitoringPromRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeMonitorError(c, http.StatusBadRequest, "invalid request payload")
		return req, false
	}
	req.DatasourceID = normalizeMonitoringDatasourceID(req.DatasourceID)
	if err := monitoring.ValidatePromQL(req.Query); err != nil {
		writeMonitorError(c, http.StatusBadRequest, err.Error())
		return req, false
	}
	if ranged && !validatePromRange(c, req.Start, req.End, req.Step) {
		return req, false
	}
	return req, true
}

func validatePromRange(c *gin.Context, start, end, step float64) bool {
	switch {
	case end <= start:
		writeMonitorError(c, http.StatusBadRequest, "end must be greater than start")
	case step <= 0:
		writeMonitorError(c, http.StatusBadRequest, "step must be greater than zero")
	case int(math.Floor((end-start)/step))+1 > 11000:
		writeMonitorError(c, http.StatusBadRequest, "range points exceed limit")
	default:
		return true
	}
	return false
}

func monitorPromRequestTarget(req monitoringPromRequest, ranged bool) (string, url.Values, time.Duration) {
	params := url.Values{"query": []string{strings.TrimSpace(req.Query)}}
	timeout := monitorPromRequestTimeout(req, promInstantTimeout)
	if !ranged {
		if req.Time > 0 {
			params.Set("time", strconv.FormatFloat(req.Time, 'f', -1, 64))
		}
		return "/api/v1/query", params, timeout
	}
	params.Set("start", strconv.FormatFloat(req.Start, 'f', -1, 64))
	params.Set("end", strconv.FormatFloat(req.End, 'f', -1, 64))
	params.Set("step", strconv.FormatFloat(req.Step, 'f', -1, 64))
	return "/api/v1/query_range", params, monitorPromRequestTimeout(req, promRangeTimeout)
}

func monitorPromRequestTimeout(req monitoringPromRequest, fallback time.Duration) time.Duration {
	if req.TimeoutMS > 0 {
		return timeoutOrDefault(req.TimeoutMS/1000, fallback)
	}
	return timeoutOrDefault(req.TimeoutSeconds, fallback)
}

func ListMonitorLabels(c *gin.Context) {
	dsID, base, ok := resolveMonitoringPrometheus(c, c.Query("datasource_id"))
	if ok {
		result := callPrometheus(c.Request.Context(), base, "/api/v1/labels", prometheusMatchParams(c), promInstantTimeout)
		writePrometheusStringList(c, dsID, "", result, boundedPositiveInt(c.DefaultQuery("limit", "200"), 200, 1000))
	}
}

func ListMonitorLabelValues(c *gin.Context) {
	label := strings.TrimSpace(c.Query("label"))
	if ok, _ := regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`, label); !ok {
		writeMonitorError(c, http.StatusBadRequest, "invalid label")
		return
	}
	dsID, base, ok := resolveMonitoringPrometheus(c, c.Query("datasource_id"))
	if ok {
		path := "/api/v1/label/" + url.PathEscape(label) + "/values"
		result := callPrometheus(c.Request.Context(), base, path, prometheusMatchParams(c), promInstantTimeout)
		writePrometheusStringList(c, dsID, label, result, boundedPositiveInt(c.DefaultQuery("limit", "500"), 500, 2000))
	}
}
