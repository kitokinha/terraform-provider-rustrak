package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
	"github.com/kitokinha/terraform-provider-rustrak/internal/config"
	"github.com/kitokinha/terraform-provider-rustrak/internal/resources"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("RUSTRAK_HOST", config.DefaultHostURL),
				Description: "The host URL for the rustrak API. Can be set with RUSTRAK_HOST environment variable.",
			},
			"rustrak_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"RUSTRAK_TOKEN", "RUSTRAK_SERVICE_TOKEN", "RUSTRAK_PAT_TOKEN"}, nil),
				Description: "The token for authenticating with rustrak. Can be a service token or a personal access token (PAT). Can be set with RUSTRAK_TOKEN, RUSTRAK_SERVICE_TOKEN, or RUSTRAK_PAT_TOKEN environment variables.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"rustrak_project": resources.Project(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
	host := d.Get("host").(string)
	rustrakToken := d.Get("rustrak_token").(string)
	client := client.NewRustrakClient(host, rustrakToken)
	return client, nil
}
