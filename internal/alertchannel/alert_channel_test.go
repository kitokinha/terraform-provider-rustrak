package alertchannel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

func TestUpdateOmitsUnchangedCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if _, ok := body["credentials"]; ok {
			t.Error("unchanged credentials must be omitted")
		}
		if body["is_enabled"] != false {
			t.Error("false must not be omitted")
		}
		_, _ = w.Write([]byte(`{"id":42,"name":"renamed","provider_type":"slack","credentials":{},"is_enabled":false}`))
	}))
	defer server.Close()
	_, err := UpdateAlertChannel(client.NewRustrakClient(server.URL, "test"), context.Background(), "42", UpdateAlertChannelRequest{Name: "renamed", IsEnabled: false})
	if err != nil {
		t.Fatal(err)
	}
}
