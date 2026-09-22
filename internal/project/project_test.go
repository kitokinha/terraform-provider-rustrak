package project

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

func TestProjectAPI(t *testing.T) {
	want := Project{ID: 123, Name: "Application", Slug: "application", Platform: "go", DSN: "https://example.com/123"}
	tests := []struct {
		name, method, path, body string
		call                     func(*client.RustrakClient) (*Project, error)
	}{
		{"create", "POST", "/api/projects", `{"name":"Application","slug":"application","platform":"go"}`, func(c *client.RustrakClient) (*Project, error) {
			return CreateProject(c, context.Background(), CreateProjectRequest{Name: want.Name, Slug: want.Slug, Platform: want.Platform})
		}},
		{"minimal create", "POST", "/api/projects", `{"name":"Application"}`, func(c *client.RustrakClient) (*Project, error) {
			return CreateProject(c, context.Background(), CreateProjectRequest{Name: want.Name})
		}},
		{"read", "GET", "/api/projects/123", "", func(c *client.RustrakClient) (*Project, error) { return ReadProject(c, context.Background(), "123") }},
		{"update", "PATCH", "/api/projects/123", `{"name":"Application","slug":"application","platform":"go"}`, func(c *client.RustrakClient) (*Project, error) {
			return UpdateProject(c, context.Background(), "123", UpdateProjectRequest{Name: want.Name, Slug: want.Slug, Platform: want.Platform})
		}},
		{"partial update", "PATCH", "/api/projects/123", `{"name":"Application"}`, func(c *client.RustrakClient) (*Project, error) {
			return UpdateProject(c, context.Background(), "123", UpdateProjectRequest{Name: want.Name})
		}},
		{"delete", "DELETE", "/api/projects/123", "", func(c *client.RustrakClient) (*Project, error) {
			return nil, DeleteProject(c, context.Background(), "123")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("Accept") != "application/json" {
					t.Error("missing auth or accept header")
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if tt.body == "" {
					if len(body) != 0 {
						t.Errorf("unexpected body: %s", body)
					}
				} else {
					if r.Header.Get("Content-Type") != "application/json" {
						t.Error("missing JSON content type")
					}
					var gotBody, wantBody any
					if err := json.Unmarshal(body, &gotBody); err != nil {
						t.Error(err)
					}
					_ = json.Unmarshal([]byte(tt.body), &wantBody)
					if !reflect.DeepEqual(gotBody, wantBody) {
						t.Errorf("body = %s; want %s", body, tt.body)
					}
				}
				if tt.method == "DELETE" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				if tt.method == "POST" {
					w.WriteHeader(http.StatusCreated)
				}
				_ = json.NewEncoder(w).Encode(want)
			}))
			defer server.Close()
			got, err := tt.call(client.NewRustrakClient(server.URL, "test-token"))
			if err != nil {
				t.Fatal(err)
			}
			if tt.method != "DELETE" && (got == nil || *got != want) {
				t.Fatalf("project = %#v; want %#v", got, want)
			}
		})
	}
}

func TestProjectAPIErrors(t *testing.T) {
	for _, status := range []int{400, 401, 403, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
			defer server.Close()
			c := client.NewRustrakClient(server.URL, "test")
			ctx := context.Background()
			_, createErr := CreateProject(c, ctx, CreateProjectRequest{Name: "test"})
			_, readErr := ReadProject(c, ctx, "123")
			_, updateErr := UpdateProject(c, ctx, "123", UpdateProjectRequest{Name: "test"})
			deleteErr := DeleteProject(c, ctx, "123")
			for name, err := range map[string]error{"create": createErr, "read": readErr, "update": updateErr, "delete": deleteErr} {
				if err == nil {
					t.Errorf("%s accepted HTTP %d", name, status)
				}
			}
		})
	}
}

func TestProjectMissing(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	c := client.NewRustrakClient(server.URL, "test")
	got, err := ReadProject(c, context.Background(), "123")
	if err != nil || got != nil {
		t.Fatalf("missing project = %#v, %v; want nil, nil", got, err)
	}
	if err := DeleteProject(c, context.Background(), "123"); err != nil {
		t.Fatalf("delete missing project: %v", err)
	}
}

func TestProjectRequestFailures(t *testing.T) {
	t.Run("invalid response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`not JSON`)) }))
		defer server.Close()
		_, err := ReadProject(client.NewRustrakClient(server.URL, "test"), context.Background(), "123")
		if err == nil || !strings.Contains(err.Error(), "decode response") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("cancelled request", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := ReadProject(client.NewRustrakClient("http://example.invalid", "test"), ctx, "123")
		if err == nil || !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("invalid URL", func(t *testing.T) {
		_, err := ReadProject(client.NewRustrakClient(":invalid", "test"), context.Background(), "123")
		if err == nil || !strings.Contains(err.Error(), "create request") {
			t.Fatalf("error = %v", err)
		}
	})
}
