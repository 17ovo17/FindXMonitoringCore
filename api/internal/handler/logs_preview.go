package handler

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"ai-workbench-api/internal/model"
)

func previewLogSamples(req model.LogPipelinePreviewRequest) model.LogPipelinePreviewResult {
	samples := append([]string{}, req.Samples...)
	if strings.TrimSpace(req.Sample) != "" {
		samples = append([]string{req.Sample}, samples...)
	}
	parser := previewParser(req)
	fields, warnings := extractPreviewFields(parser, req.Pattern, samples)
	return model.LogPipelinePreviewResult{
		OK: len(warnings) == 0, Status: "preview_only", Parser: parser,
		SampleCount: len(samples), ExtractedFields: fields, Warnings: warnings,
	}
}

func previewParser(req model.LogPipelinePreviewRequest) string {
	if trimmed := strings.ToLower(strings.TrimSpace(req.Parser)); trimmed != "" {
		return trimmed
	}
	var config map[string]any
	if len(req.Pipeline.Config) > 0 && json.Unmarshal(req.Pipeline.Config, &config) == nil {
		if value, ok := config["parser"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.ToLower(strings.TrimSpace(value))
		}
	}
	return "json"
}

func extractPreviewFields(parser, pattern string, samples []string) ([]model.LogPreviewField, []string) {
	switch parser {
	case "json":
		return previewJSONFields(samples)
	case "logfmt":
		return previewLogfmtFields(samples), nil
	case "regex":
		return previewRegexFields(pattern, samples)
	default:
		return nil, []string{"unsupported parser"}
	}
}

func previewJSONFields(samples []string) ([]model.LogPreviewField, []string) {
	fields := map[string]model.LogPreviewField{}
	warnings := []string{}
	for _, sample := range samples {
		var obj map[string]any
		if err := json.Unmarshal([]byte(sample), &obj); err != nil {
			warnings = append(warnings, "sample is not a json object")
			continue
		}
		for key, value := range obj {
			fields[key] = model.LogPreviewField{Key: key, Type: previewValueType(value), Value: previewValue(value)}
		}
	}
	return sortedPreviewFields(fields), warnings
}

func previewLogfmtFields(samples []string) []model.LogPreviewField {
	fields := map[string]model.LogPreviewField{}
	for _, sample := range samples {
		for _, token := range strings.Fields(sample) {
			parts := strings.SplitN(token, "=", 2)
			if len(parts) != 2 || parts[0] == "" {
				continue
			}
			fields[parts[0]] = model.LogPreviewField{Key: parts[0], Type: "string", Value: strings.Trim(parts[1], `"`)}
		}
	}
	return sortedPreviewFields(fields)
}

func previewRegexFields(pattern string, samples []string) ([]model.LogPreviewField, []string) {
	if strings.TrimSpace(pattern) == "" {
		return nil, []string{"regex pattern is required"}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, []string{"regex pattern is invalid"}
	}
	fields := map[string]model.LogPreviewField{}
	for _, sample := range samples {
		match := re.FindStringSubmatch(sample)
		for idx, name := range re.SubexpNames() {
			if idx > 0 && name != "" && idx < len(match) {
				fields[name] = model.LogPreviewField{Key: name, Type: "string", Value: match[idx]}
			}
		}
	}
	return sortedPreviewFields(fields), nil
}

func sortedPreviewFields(fields map[string]model.LogPreviewField) []model.LogPreviewField {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]model.LogPreviewField, 0, len(keys))
	for _, key := range keys {
		out = append(out, fields[key])
	}
	return out
}

func previewValueType(value any) string {
	switch value.(type) {
	case bool:
		return "bool"
	case float64:
		return "number"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	default:
		return "string"
	}
}

func previewValue(value any) string {
	switch typed := value.(type) {
	case string:
		return sanitizeLogString(typed)
	default:
		data, _ := json.Marshal(typed)
		return sanitizeLogString(string(data))
	}
}
