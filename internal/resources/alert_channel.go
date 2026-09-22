package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/kitokinha/terraform-provider-rustrak/internal/alertchannel"
	"github.com/kitokinha/terraform-provider-rustrak/internal/client"
)

func AlertChannel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAlertChannelCreate,
		ReadContext:   resourceAlertChannelRead,
		UpdateContext: resourceAlertChannelUpdate,
		DeleteContext: resourceAlertChannelDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Schema: map[string]*schema.Schema{
			"name":          {Type: schema.TypeString, Required: true, Description: "The name of the Rustrak alert channel."},
			"provider_type": {Type: schema.TypeString, Required: true, ForceNew: true, ValidateFunc: validation.StringInSlice([]string{"webhook", "email", "slack", "custom_webhook"}, false), Description: "Notification provider: webhook, email, slack, or custom_webhook. Changes require replacement."},
			"credentials":   {Type: schema.TypeString, Required: true, Sensitive: true, ValidateFunc: validateAlertChannelCredentials, DiffSuppressFunc: structure.SuppressJsonDiff, DiffSuppressOnRefresh: true, Description: "Provider-specific credentials as a JSON object. Use jsonencode()."},
			"is_enabled":    {Type: schema.TypeBool, Optional: true, Default: true, Description: "Whether the alert channel is enabled."},
		},
	}
}

func validateAlertChannelCredentials(v any, key string) ([]string, []error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(v.(string)), &object); err != nil || object == nil {
		return nil, []error{fmt.Errorf("%s must be a JSON object", key)}
	}
	return nil, nil
}

func resourceAlertChannelCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	channel, err := alertchannel.CreateAlertChannel(meta.(*client.RustrakClient), ctx, alertchannel.CreateAlertChannelRequest{
		Name: d.Get("name").(string), ProviderType: d.Get("provider_type").(string),
		Credentials: json.RawMessage(d.Get("credentials").(string)), IsEnabled: d.Get("is_enabled").(bool),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating alert channel: %w", err))
	}
	d.SetId(strconv.Itoa(channel.ID))
	return resourceAlertChannelRead(ctx, d, meta)
}

func resourceAlertChannelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	channel, err := alertchannel.ReadAlertChannel(meta.(*client.RustrakClient), ctx, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading alert channel %q: %w", d.Id(), err))
	}
	if channel == nil {
		d.SetId("")
		return nil
	}
	var credentials map[string]json.RawMessage
	if err := json.Unmarshal(channel.Credentials, &credentials); err != nil || credentials == nil {
		return diag.Errorf("error reading alert channel %q: credentials must be a JSON object", d.Id())
	}
	// Rustrak redacts Slack bot tokens. Preserve a known token in state while
	// still refreshing every other credential field from the server.
	if channel.ProviderType == "slack" && string(credentials["method"]) == `"bot_token"` && string(credentials["token"]) == `"xoxb-****"` {
		var previous map[string]json.RawMessage
		if json.Unmarshal([]byte(d.Get("credentials").(string)), &previous) == nil && string(previous["method"]) == `"bot_token"` {
			if token, ok := previous["token"]; ok {
				credentials["token"] = token
			}
		}
	}
	encoded, err := json.Marshal(credentials)
	if err != nil {
		return diag.FromErr(err)
	}
	for key, value := range map[string]any{"name": channel.Name, "provider_type": channel.ProviderType, "credentials": string(encoded), "is_enabled": channel.IsEnabled} {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceAlertChannelUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	input := alertchannel.UpdateAlertChannelRequest{Name: d.Get("name").(string), IsEnabled: d.Get("is_enabled").(bool)}
	if d.HasChange("credentials") {
		input.Credentials = json.RawMessage(d.Get("credentials").(string))
	}
	_, err := alertchannel.UpdateAlertChannel(meta.(*client.RustrakClient), ctx, d.Id(), input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating alert channel %q: %w", d.Id(), err))
	}
	return resourceAlertChannelRead(ctx, d, meta)
}

func resourceAlertChannelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	if err := alertchannel.DeleteAlertChannel(meta.(*client.RustrakClient), ctx, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error deleting alert channel %q: %w", d.Id(), err))
	}
	d.SetId("")
	return nil
}
