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

If no host is specified, the provider uses the default Rustrak API host.

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

## Import

Existing Rustrak projects can be imported using their project ID:

```bash
terraform import rustrak_project.example 123
```

After importing, refresh the Terraform state:

```bash
terraform plan
```

## Supported Resources

| Resource          | Description               |
| ----------------- | ------------------------- |
| `rustrak_project` | Manages a Rustrak project |

More detailed resource documentation is available in [`docs/resources/project.md`](docs/resources/project.md).

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

Run the tests:

```bash
go test ./...
```

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
│       └── project.md
├── internal/
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
