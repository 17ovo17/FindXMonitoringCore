package tracing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSWClientMissingEndpointReturnsTypedError(t *testing.T) {
	resetSkyWalkingClientConfig(t)
	client := NewSWClient()
	if err := client.Query(context.Background(), "query { ping }", nil, nil); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("missing endpoint should return ErrNotConfigured, got %v", err)
	}
}

func TestSWClientSanitizesHTTPAndGraphQLErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "http body", body: `token=secret-cookie dsn=mysql://root:pass@db/db`},
		{name: "graphql errors", body: `{"errors":[{"message":"token=secret-cookie dsn=mysql://root:pass@db/db"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := http.StatusOK
			if tt.name == "http body" {
				status = http.StatusBadGateway
			}
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer upstream.Close()

			client := &SWClient{baseURL: upstream.URL, http: upstream.Client()}
			err := client.Query(context.Background(), "query { ping }", nil, nil)
			if err == nil {
				t.Fatal("expected sanitized upstream error")
			}
			errText := err.Error()
			for _, forbidden := range []string{"secret-cookie", "root:pass", "mysql://", "dsn="} {
				if strings.Contains(errText, forbidden) {
					t.Fatalf("error leaked sensitive marker %q: %s", forbidden, errText)
				}
			}
		})
	}
}

func resetSkyWalkingClientConfig(t *testing.T) {
	t.Helper()
	oldGraphQL := viper.GetString("skywalking.graphql_url")
	t.Setenv("SKYWALKING_GRAPHQL_URL", "")
	t.Setenv("SKYWALKING_OAP_GRAPHQL_URL", "")
	t.Setenv("SKYWALKING_OAP_URL", "")
	viper.Set("skywalking.graphql_url", "")
	t.Cleanup(func() {
		viper.Set("skywalking.graphql_url", oldGraphQL)
	})
}
