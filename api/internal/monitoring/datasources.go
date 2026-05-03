package monitoring

import (
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const DefaultPrometheusDatasourceID = "prometheus-default"

func PrometheusDatasourcesFromConfig() []Datasource {
	var raw []struct {
		ID   string `mapstructure:"id"`
		Type string `mapstructure:"type"`
		URL  string `mapstructure:"url"`
	}
	if err := viper.UnmarshalKey("data_sources", &raw); err != nil {
		log.WithError(err).Warn("monitoring: data_sources config parse failed")
		raw = nil
	}
	promURL := strings.TrimSpace(viper.GetString("prometheus.url"))
	out := make([]Datasource, 0, len(raw)+1)
	hasPrometheus := false
	for _, item := range raw {
		if !strings.EqualFold(item.Type, "prometheus") {
			continue
		}
		hasPrometheus = true
		dsURL := strings.TrimSpace(item.URL)
		if dsURL == "" && promURL != "" {
			dsURL = promURL
		}
		out = append(out, Datasource{ID: strings.TrimSpace(item.ID), Type: "prometheus", URL: dsURL})
	}
	if !hasPrometheus && promURL != "" {
		out = append([]Datasource{{ID: DefaultPrometheusDatasourceID, Type: "prometheus", URL: promURL}}, out...)
	}
	return out
}

func ResolvePrometheusDatasourceFromConfig(datasourceID string) (string, string, error) {
	return ResolvePrometheusDatasource(
		PrometheusDatasourcesFromConfig(),
		strings.TrimSpace(datasourceID),
		DefaultPrometheusDatasourceID,
		viper.GetString("prometheus.url"),
	)
}

func DatasourceURLReady(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}
