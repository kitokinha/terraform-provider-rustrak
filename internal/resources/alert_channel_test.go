package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

func TestAlertChannelLifecycle(t *testing.T) {
	var stored map[string]any
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing authorization")
		}
		if r.Method == "POST" {
			if r.URL.Path != "/api/integrations" {
				t.Errorf("wrong create path: %s", r.URL.Path)
			}
		} else if r.URL.Path != "/api/integrations/42" {
			t.Errorf("wrong resource path: %s", r.URL.Path)
		}
		switch r.Method {
		case "POST":
			if err := json.NewDecoder(r.Body).Decode(&stored); err != nil {
				t.Error(err)
			}
			if stored["is_enabled"] != false {
				t.Error("create must send explicit false")
			}
			if _, ok := stored["credentials"].(map[string]any); !ok {
				t.Error("credentials must be an object")
			}
			stored["id"] = 42
			w.WriteHeader(http.StatusCreated)
		case "PATCH":
			var patch map[string]any
			if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
				t.Error(err)
			}
			if _, ok := patch["provider_type"]; ok {
				t.Error("provider_type cannot be updated")
			}
			for k, v := range patch {
				stored[k] = v
			}
		case "DELETE":
			stored = nil
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if stored == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(stored)
	}))
	defer server.Close()
	c := client.NewRustrakClient(server.URL, "test-token")
	resource := AlertChannel()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{"name": "Ops", "provider_type": "webhook", "credentials": `{"url":"https://example.com","headers":{"X-Test":"one"}}`, "is_enabled": false})
	ctx := context.Background()
	if diags := resource.CreateContext(ctx, data, c); diags.HasError() {
		t.Fatal(diags)
	}
	if data.Id() != "42" {
		t.Fatalf("unexpected ID %s", data.Id())
	}
	_ = data.Set("name", "New name")
	_ = data.Set("credentials", `{"url":"https://example.com/new"}`)
	if diags := resource.UpdateContext(ctx, data, c); diags.HasError() {
		t.Fatal(diags)
	}
	if stored["name"] != "New name" || stored["is_enabled"] != false {
		t.Fatalf("bad update: %v", stored)
	}
	imported := schema.TestResourceDataRaw(t, resource.Schema, nil)
	imported.SetId("42")
	states, err := resource.Importer.StateContext(ctx, imported, c)
	if err != nil || len(states) != 1 {
		t.Fatalf("import: %v", err)
	}
	if diags := resource.ReadContext(ctx, states[0], c); diags.HasError() {
		t.Fatal(diags)
	}
	if imported.Get("name") != "New name" || imported.Get("provider_type") != "webhook" {
		t.Fatal("import failed to refresh fields")
	}
	if diags := resource.DeleteContext(ctx, data, c); diags.HasError() {
		t.Fatal(diags)
	}
	if data.Id() != "" {
		t.Fatal("delete did not clear ID")
	}
	if diags := resource.ReadContext(ctx, imported, c); diags.HasError() {
		t.Fatal(diags)
	}
	if imported.Id() != "" {
		t.Fatal("404 did not clear ID")
	}
	expected := []string{"POST /api/integrations", "GET /api/integrations/42", "PATCH /api/integrations/42", "GET /api/integrations/42", "GET /api/integrations/42", "DELETE /api/integrations/42", "GET /api/integrations/42"}
	if !reflect.DeepEqual(requests, expected) {
		t.Fatalf("requests: %v", requests)
	}
}

func TestAlertChannelRedactedToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":42,"name":"Slack","provider_type":"slack","is_enabled":true,"credentials":{"method":"bot_token","token":"xoxb-****","channel":"new-channel"}}`))
	}))
	defer server.Close()
	resource := AlertChannel()
	data := schema.TestResourceDataRaw(t, resource.Schema, map[string]any{"credentials": `{"method":"bot_token","token":"xoxb-secret","channel":"old-channel"}`})
	data.SetId("42")
	if diags := resource.ReadContext(context.Background(), data, client.NewRustrakClient(server.URL, "test")); diags.HasError() {
		t.Fatal(diags)
	}
	var credentials map[string]any
	_ = json.Unmarshal([]byte(data.Get("credentials").(string)), &credentials)
	if credentials["token"] != "xoxb-secret" || credentials["channel"] != "new-channel" {
		t.Fatal("refresh must preserve secret but detect other credential drift")
	}
}

func TestAlertChannelErrors(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusUnauthorized, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"error":"secret-credential"}`))
		}))
		resource := AlertChannel()
		for _, action := range []string{"read", "delete"} {
			data := schema.TestResourceDataRaw(t, resource.Schema, nil)
			data.SetId("42")
			fn := resource.ReadContext
			if action == "delete" {
				fn = schema.ReadContextFunc(resource.DeleteContext)
			}
			diags := fn(context.Background(), data, client.NewRustrakClient(server.URL, "test"))
			if status == http.StatusNotFound {
				if diags.HasError() || data.Id() != "" {
					t.Errorf("404 %s: %v", action, diags)
				}
			} else {
				if !diags.HasError() || data.Id() != "42" {
					t.Errorf("%d %s must retain state and report error", status, action)
				}
				for _, d := range diags {
					if strings.Contains(d.Summary+d.Detail, "secret-credential") {
						t.Error("leaked credentials")
					}
				}
			}
		}
		server.Close()
	}
}

func TestAlertChannelSchema(t *testing.T) {
	r := AlertChannel()
	for _, value := range []string{`null`, `[]`, `"text"`, `invalid`} {
		if _, errs := r.Schema["credentials"].ValidateFunc(value, "credentials"); len(errs) == 0 {
			t.Errorf("accepted %s", value)
		}
	}
	if _, errs := r.Schema["credentials"].ValidateFunc(`{"nested":[true,1]}`, "credentials"); len(errs) > 0 {
		t.Fatal(errs)
	}
	if !r.Schema["credentials"].DiffSuppressFunc("credentials", `{"a":1,"b":2}`, "{\n\"b\":2,\"a\":1}", nil) {
		t.Error("equivalent JSON must not cause a diff")
	}
	if !r.Schema["provider_type"].ForceNew || !r.Schema["credentials"].Sensitive {
		t.Fatal("invalid schema flags")
	}
	if _, errs := r.Schema["provider_type"].ValidateFunc("invalid", "provider_type"); len(errs) == 0 {
		t.Error("accepted invalid provider")
	}
	if data := schema.TestResourceDataRaw(t, r.Schema, nil); data.Get("is_enabled") != true {
		t.Error("must default to enabled")
	}
}
