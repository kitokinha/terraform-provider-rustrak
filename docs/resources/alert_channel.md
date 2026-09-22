# rustrak_alert_channel Resource

Manages an alert channel through the current Rustrak `/api/integrations` API.
Channels are global integrations; managing them requires an admin token.
This resource follows the [Alert Channels API reference](https://rustrak.github.io/rustrak/api-reference#tag/alert-channels).

## Example Usage

```hcl
resource "rustrak_alert_channel" "webhook" {
  name          = "Operations"
  provider_type = "webhook"
  is_enabled    = true
  credentials = jsonencode({
    url = "https://example.com/alerts"
  })
}
```

## Arguments

* `name` (required, string): Name of the channel.
* `provider_type` (required, string): `webhook`, `email`, `slack`, or `custom_webhook`. Changing this replaces the channel because the API does not support updating its provider type.
* `credentials` (required, sensitive string): Provider-specific JSON object, normally built with `jsonencode()`. Supports nested objects and arrays. JSON formatting and key order do not cause differences.
* `is_enabled` (optional, boolean): Whether the channel is enabled. Defaults to `true`.

Credentials depend on the selected provider. Routing overrides and alert rules are configured separately from integrations.
Sensitive values are hidden in Terraform output but remain in Terraform state.
Slack bot tokens masked by the API are preserved from the configured state; remote token changes cannot be detected.

## Attributes

* `id`: Numeric integration ID, stored as a string.

## Import

```bash
terraform import rustrak_alert_channel.webhook 123
```

After importing, supply matching configuration and run `terraform plan`.
For Slack bot-token integrations, supply the original token in `credentials`: the API returns only a masked token, so import cannot recover it.

## Lifecycle

Supports create, read, update, delete, and import. Creation and updates refresh state from the API.
Updates send credentials only when they change. A missing channel is removed from state;
deleting an already missing channel succeeds.
