package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
	"github.com/kitokinha/terraform-provider-rustrak/internal/config"
)

func TestProvider(t *testing.T) {
	p := Provider()
	if err := p.InternalValidate(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"rustrak_project", "rustrak_alert_channel"} {
		if p.ResourcesMap[name] == nil {
			t.Errorf("missing resource %s", name)
		}
	}
	if !p.Schema["rustrak_token"].Sensitive {
		t.Error("authentication token must be sensitive")
	}
}

func TestProviderConfigure(t *testing.T) {
	for _, tt := range []struct {
		name, host, token, service, pat string
		raw                             map[string]any
		wantHost, wantToken             string
	}{
		{name: "defaults", pat: "pat", wantHost: config.DefaultHostURL, wantToken: "pat"},
		{name: "service before PAT", host: "https://env.example.com", service: "service", pat: "pat", wantHost: "https://env.example.com", wantToken: "service"},
		{name: "primary before service", token: "primary", service: "service", pat: "pat", wantHost: config.DefaultHostURL, wantToken: "primary"},
		{name: "explicit overrides environment", host: "https://env.example.com", token: "environment", raw: map[string]any{"host": "https://explicit.example.com", "rustrak_token": "explicit"}, wantHost: "https://explicit.example.com", wantToken: "explicit"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RUSTRAK_HOST", tt.host)
			t.Setenv("RUSTRAK_TOKEN", tt.token)
			t.Setenv("RUSTRAK_SERVICE_TOKEN", tt.service)
			t.Setenv("RUSTRAK_PAT_TOKEN", tt.pat)
			p := Provider()
			d := schema.TestResourceDataRaw(t, p.Schema, tt.raw)
			meta, diags := p.ConfigureContextFunc(context.Background(), d)
			if diags.HasError() {
				t.Fatal(diags)
			}
			c, ok := meta.(*client.RustrakClient)
			if !ok {
				t.Fatalf("client type = %T", meta)
			}
			if c.HostURL != tt.wantHost || c.Token != tt.wantToken || c.HTTPClient == nil {
				t.Error("provider did not configure the expected host, token, and HTTP client")
			}
		})
	}
}
