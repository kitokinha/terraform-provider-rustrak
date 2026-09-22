# Terraform Provider Rustrak

[![Go Reference](https://pkg.go.dev/badge/github.com/kitokinha/terraform-provider-rustrak.svg)](https://pkg.go.dev/github.com/kitokinha/terraform-provider-rustrak)
[![Go Report Card](https://goreportcard.com/badge/github.com/kitokinha/terraform-provider-rustrak)](https://goreportcard.com/report/github.com/kitokinha/terraform-provider-rustrak)

Terraform provider for managing [Rustrak](https://rustrak.com) resources.

## Requirements

* [Terraform](https://developer.hashicorp.com/terraform) >= 1.0
* A Rustrak account
* A Rustrak authentication token

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    rustrak = {
      source  = "kitokinha/rustrak"
      version = "~> 0.1"
    }
  }
}
```

Then initialize Terraform:

```bash
terraform init
```

## Authentication

The provider supports Rustrak service tokens and personal access tokens (PATs).

The recommended way to authenticate is through an environment variable:

```bash
export RUSTRAK_TOKEN="your-token"
```

The following environment variables are supported:

* `RUSTRAK_TOKEN`
* `RUSTRAK_SERVICE_TOKEN`
* `RUSTRAK_PAT_TOKEN`

The provider checks them in that order.

You can then configure the provider without putting the token in your Terraform configuration:

```hcl
provider "rustrak" {}
```

## Provider Configuration

The provider supports the following configuration options:

| Argument        | Type   | Required | Sensitive | Environment Variable                                          |
| --------------- | ------ | -------- | --------- | ------------------------------------------------------------- |
| `host`          | string | No       | No        | `RUSTRAK_HOST`                                                |
| `rustrak_token` | string | Yes      | Yes       | `RUSTRAK_TOKEN`, `RUSTRAK_SERVICE_TOKEN`, `RUSTRAK_PAT_TOKEN` |

### Custom API Host

The API host can be configured directly:

```hcl
provider "rustrak" {
  host = "https://api.rustrak.com"
}
```

Or through the `RUSTRAK_HOST` environment variable:

```bash
export RUSTRAK_HOST="https://api.rustrak.com"
```

If neither `host` nor `RUSTRAK_HOST` is set, the provider uses `https://api.rustrak.dev`. Explicit provider arguments take precedence over environment variables.

## Usage

### Create a Project

```hcl
resource "rustrak_project" "example" {
  name     = "My Application"
  slug     = "my-application"
  platform = "go"
}
```

Only `name` is required:

```hcl
resource "rustrak_project" "example" {
  name = "My Application"
}
```

### Access Project Attributes

The project ID and DSN are available as resource attributes:

```hcl
output "project_id" {
  value = rustrak_project.example.id
}

output "project_dsn" {
  value     = rustrak_project.example.dsn
  sensitive = true
}
```

### Create an Alert Channel

```hcl
resource "rustrak_alert_channel" "operations" {
  name          = "Operations"
  provider_type = "webhook"
  is_enabled    = true
  credentials = jsonencode({
    url = "https://example.com/alerts"
  })
}
```

Alert channels use the current `/api/integrations` API and require an admin token.
See the [alert channel reference](docs/resources/alert_channel.md) for provider types, credentials, and import behavior.

## Import

Existing Rustrak projects can be imported using their project ID:

```bash
terraform import rustrak_project.example 123
```

Existing alert integrations can be imported using their integration ID:

```bash
terraform import rustrak_alert_channel.operations 456
```

After importing, provide matching resource configuration and review the plan:

```bash
terraform plan
```

## Supported Resources

| Resource          | Description               |
| ----------------- | ------------------------- |
| `rustrak_project` | Manages a Rustrak project |
| `rustrak_alert_channel` | Manages an alert integration |

More detailed resource documentation is available in [`docs/resources/project.md`](docs/resources/project.md).

Alert channel usage and import examples are available in [`docs/resources/alert_channel.md`](docs/resources/alert_channel.md).

## Development

Clone the repository:

```bash
git clone https://github.com/kitokinha/terraform-provider-rustrak.git
cd terraform-provider-rustrak
```

Install dependencies:

```bash
go mod tidy
```

### Tests

Run the automated suite:

```bash
go test ./...
```

Tests use local mock HTTP servers and do not require a Rustrak instance, API token, or Terraform installation. They cover:

* Project API methods, paths, authentication headers, request bodies, response decoding, and error handling.
* Project create, refresh, update, delete, import, server defaults, remote changes, and missing-project behavior.
* Alert channel lifecycle, credentials, Slack token masking, schema validation, and API failures.
* Provider configuration, environment-variable precedence, resource registration, and client error classification.

Run race detection and coverage:

```bash
go test -race -cover ./...
```

Run static checks:

```bash
go vet ./...
```

The suite does not exercise a live Rustrak server or Terraform CLI acceptance workflow.

Build the provider:

```bash
go build ./...
```

## Project Structure

```text
.
├── docs/
│   ├── index.md
│   └── resources/
│       ├── alert_channel.md
│       └── project.md
├── internal/
│   ├── alertchannel/
│   ├── client/
│   ├── config/
│   ├── project/
│   ├── provider/
│   └── resources/
├── main.go
├── go.mod
└── README.md
```

## License

This project is licensed under the terms of the repository's license.
