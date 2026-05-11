package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

func resolveIdent(ip string) string {
	target := discoverPromTarget(ip)
	if target.LabelKey == "ident" {
		return target.LabelVal
	}
	return ""
}

func extractIP(text string) string {
	return ipRe.FindString(text)
}

func discoverPromTarget(ip string) promTarget {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" || ip == "" {
		return promTarget{}
	}
	labelPriority := []string{"ident", "instance", "ip", "host", "hostname", "agent_hostname", "target", "node", "nodename", "exported_instance", "address"}
	best := promTarget{}
	bestScore := -1
	seenMatcher := map[string]bool{}
	for _, key := range labelPriority {
		for _, value := range promLabelValues(base, key) {
			if !labelContainsExactIP(value, ip) {
				continue
			}
			matcherKey := key + "=" + value
			if seenMatcher[matcherKey] {
				continue
			}
			seenMatcher[matcherKey] = true
			series := promSeriesByMatcher(base, fmt.Sprintf(`{%s="%s"}`, key, value))
			if len(series) == 0 {
				continue
			}
			metrics := metricNamesFromSeries(series)
			score := scorePromTarget(key, value, ip, metrics, len(series))
			if score > bestScore {
				bestScore = score
				best = promTarget{LabelKey: key, LabelVal: value, Series: series, Metrics: metrics, TargetOnly: isTargetOnlyMetrics(metrics)}
			}
		}
	}
	if best.LabelKey == "" {
		matched := promSeriesContainingIP(base, ip)
		if len(matched) == 0 {
			return promTarget{}
		}
		key, value := mostCommonIPLabel(matched, ip)
		metrics := metricNamesFromSeries(matched)
		best = promTarget{LabelKey: key, LabelVal: value, Series: matched, Metrics: metrics, TargetOnly: isTargetOnlyMetrics(metrics)}
	}
	best.Categories = categorizeMetrics(best.Metrics)
	return best
}

func scorePromTarget(key, value, ip string, metrics []string, seriesCount int) int {
	score := seriesCount
	if labelIsExactIPTarget(value, ip) {
		score += 100000
	}
	if !isTargetOnlyMetrics(metrics) {
		score += 50000
	}
	switch key {
	case "ident":
		score += 20000
	case "instance":
		score += 10000
	}
	return score
}

func labelContainsExactIP(value, ip string) bool {
	if value == "" || ip == "" {
		return false
	}
	for _, match := range ipRe.FindAllString(value, -1) {
		if match == ip {
			return true
		}
	}
	return false
}

func labelIsExactIPTarget(value, ip string) bool {
	return value == ip || strings.HasPrefix(value, ip+":") || strings.HasSuffix(value, "-"+ip) || strings.HasSuffix(value, "_"+ip)
}

func isTargetOnlyMetrics(metrics []string) bool {
	if len(metrics) == 0 {
		return false
	}
	for _, metric := range metrics {
		if targetOnlyMetrics[metric] || strings.HasPrefix(metric, "scrape_") {
			continue
		}
		return false
	}
	return true
}

func promLabelValues(base, label string) []string {
	if base == "" {
		return nil
	}
	resp, err := http.Get(base + "/api/v1/label/" + url.PathEscape(label) + "/values")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if json.Unmarshal(body, &result) != nil || result.Status != "success" {
		return nil
	}
	return result.Data
}

func promSeriesByMatcher(base, matcher string) []map[string]string {
	if base == "" || matcher == "" {
		return nil
	}
	resp, err := http.Get(base + "/api/v1/series?match[]=" + url.QueryEscape(matcher))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string              `json:"status"`
		Data   []map[string]string `json:"data"`
	}
	if json.Unmarshal(body, &result) != nil || result.Status != "success" {
		return nil
	}
	return result.Data
}

func promSeriesContainingIP(base, ip string) []map[string]string {
	all := promSeriesByMatcher(base, `{__name__=~".+"}`)
	matched := []map[string]string{}
	seen := map[string]bool{}
	for _, item := range all {
		for key, value := range item {
			if key == "__name__" || !labelContainsExactIP(value, ip) {
				continue
			}
			sig := seriesSignature(item)
			if !seen[sig] {
				seen[sig] = true
				matched = append(matched, item)
			}
			break
		}
	}
	return matched
}

func seriesSignature(series map[string]string) string {
	keys := make([]string, 0, len(series))
	for key := range series {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+series[key])
	}
	return strings.Join(parts, "|")
}

func mostCommonIPLabel(series []map[string]string, ip string) (string, string) {
	counts := map[string]map[string]int{}
	for _, item := range series {
		for key, value := range item {
			if key == "__name__" || !labelContainsExactIP(value, ip) {
				continue
			}
			if counts[key] == nil {
				counts[key] = map[string]int{}
			}
			counts[key][value]++
		}
	}
	bestKey, bestVal, bestCount := "", "", 0
	for key, values := range counts {
		for value, count := range values {
			if count > bestCount || (count == bestCount && labelIsExactIPTarget(value, ip)) {
				bestKey, bestVal, bestCount = key, value, count
			}
		}
	}
	return bestKey, bestVal
}

func metricNamesFromSeries(series []map[string]string) []string {
	seen := map[string]bool{}
	names := []string{}
	for _, item := range series {
		name := item["__name__"]
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func discoveredPromHosts(limit int) []string {
	base := strings.TrimRight(viper.GetString("prometheus.url"), "/")
	if base == "" {
		return nil
	}
	summary := map[string]*promHostSummary{}
	for _, label := range []string{"ident", "instance", "ip", "host", "hostname", "target", "address"} {
		for _, value := range promLabelValues(base, label) {
			for _, ip := range ipRe.FindAllString(value, -1) {
				if summary[ip] == nil {
					summary[ip] = &promHostSummary{IP: ip}
				}
				summary[ip].Labels = appendUnique(summary[ip].Labels, label+"="+value)
			}
		}
	}
	ips := make([]string, 0, len(summary))
	for ip := range summary {
		ips = append(ips, ip)
	}
	sort.Strings(ips)
	if limit > 0 && len(ips) > limit {
		ips = ips[:limit]
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, ip+" ("+strings.Join(summary[ip].Labels, "; ")+")")
	}
	return out
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}
