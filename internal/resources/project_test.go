package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

func TestProjectLifecycle(t *testing.T) {
	for _, minimal := range []bool{true, false} {
		name := "full"
		if minimal {
			name = "minimal"
		}
		t.Run(name, func(t *testing.T) {
			var stored map[string]any
			var methods []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				methods = append(methods, r.Method)
				switch r.Method {
				case "POST":
					if err := json.NewDecoder(r.Body).Decode(&stored); err != nil {
						t.Error(err)
					}
					if stored["name"] != "Application" {
						t.Error("missing name")
					}
					if minimal {
						if _, ok := stored["slug"]; ok {
							t.Error("omitted slug was sent")
						}
						if _, ok := stored["platform"]; ok {
							t.Error("omitted platform was sent")
						}
						stored["slug"] = "generated-slug"
						stored["platform"] = "other"
					} else if stored["slug"] != "application" || stored["platform"] != "go" {
						t.Error("configured fields not sent")
					}
					stored["id"] = 123
					stored["dsn"] = "https://example.com/123"
					w.WriteHeader(http.StatusCreated)
				case "PATCH":
					var patch map[string]any
					if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
						t.Error(err)
					}
					if patch["name"] != "Renamed" || patch["slug"] != "renamed" || patch["platform"] != "python" {
						t.Errorf("unexpected patch: %v", patch)
					}
					if _, ok := patch["dsn"]; ok {
						t.Error("computed DSN sent in update")
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
			c := client.NewRustrakClient(server.URL, "test")
			r := Project()
			raw := map[string]any{"name": "Application"}
			if !minimal {
				raw["slug"] = "application"
				raw["platform"] = "go"
			}
			d := schema.TestResourceDataRaw(t, r.Schema, raw)
			ctx := context.Background()
			if diags := r.CreateContext(ctx, d, c); diags.HasError() {
				t.Fatal(diags)
			}
			if d.Id() != "123" || d.Get("dsn") != "https://example.com/123" {
				t.Fatal("create did not refresh ID and DSN")
			}
			if minimal && (d.Get("slug") != "generated-slug" || d.Get("platform") != "other") {
				t.Fatal("server defaults not stored")
			}
			_ = d.Set("name", "Renamed")
			_ = d.Set("slug", "renamed")
			_ = d.Set("platform", "python")
			if diags := r.UpdateContext(ctx, d, c); diags.HasError() {
				t.Fatal(diags)
			}
			imported := schema.TestResourceDataRaw(t, r.Schema, nil)
			imported.SetId("123")
			states, err := r.Importer.StateContext(ctx, imported, c)
			if err != nil || len(states) != 1 {
				t.Fatalf("import: %v", err)
			}
			if diags := r.ReadContext(ctx, states[0], c); diags.HasError() {
				t.Fatal(diags)
			}
			for _, key := range []string{"name", "slug", "platform", "dsn"} {
				if imported.Get(key) != d.Get(key) {
					t.Errorf("imported %s differs", key)
				}
			}
			stored["name"] = "Remote rename"
			if diags := r.ReadContext(ctx, d, c); diags.HasError() {
				t.Fatal(diags)
			}
			if d.Get("name") != "Remote rename" {
				t.Error("remote drift not refreshed")
			}
			if diags := r.DeleteContext(ctx, d, c); diags.HasError() {
				t.Fatal(diags)
			}
			if d.Id() != "" {
				t.Error("delete retained ID")
			}
			if diags := r.ReadContext(ctx, imported, c); diags.HasError() {
				t.Fatal(diags)
			}
			if imported.Id() != "" {
				t.Error("missing project retained ID")
			}
			want := []string{"POST", "GET", "PATCH", "GET", "GET", "GET", "DELETE", "GET"}
			if !reflect.DeepEqual(methods, want) {
				t.Errorf("requests = %v; want %v", methods, want)
			}
		})
	}
}

func TestProjectResourceErrors(t *testing.T) {
	r := Project()
	for name, fn := range map[string]func(context.Context, *schema.ResourceData, any) diag.Diagnostics{
		"create": r.CreateContext, "read": r.ReadContext, "update": r.UpdateContext, "delete": r.DeleteContext,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
			defer server.Close()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"name": "Application"})
			wantID := "123"
			if name == "create" {
				wantID = ""
			}
			d.SetId(wantID)
			if diags := fn(context.Background(), d, client.NewRustrakClient(server.URL, "test")); !diags.HasError() {
				t.Fatal("expected diagnostic")
			}
			if d.Id() != wantID {
				t.Errorf("failed %s changed ID", name)
			}
		})
	}
}

func TestProjectCreateRefreshFailureRetainsID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":123}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	r := Project()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]any{"name": "Application"})
	diags := r.CreateContext(context.Background(), d, client.NewRustrakClient(server.URL, "test"))
	if !diags.HasError() || d.Id() != "123" {
		t.Fatal("refresh failure must report an error and retain the created ID")
	}
}
