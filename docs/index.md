# Rustrak Provider

The Rustrak Terraform Provider allows you to manage Rustrak resources using Terraform.

## Example Usage

```hcl
terraform {
  required_providers {
    rustrak = {
      source  = "kitokinha/rustrak"
      version = "~> 0.1"
    }
  }
}

provider "rustrak" {
  rustrak_token = var.rustrak_token
}
```

The authentication token can also be provided through an environment variable:

```bash
export RUSTRAK_TOKEN="your-token"
```

Then the provider can be configured without specifying the token directly:

```hcl
provider "rustrak" {}
```

## Authentication

The provider supports authentication using either a Rustrak service token or a personal access token (PAT).

The following environment variables are supported:

* `RUSTRAK_TOKEN`
* `RUSTRAK_SERVICE_TOKEN`
* `RUSTRAK_PAT_TOKEN`

The provider checks these environment variables in the order listed above.

### Example

```bash
export RUSTRAK_TOKEN="your-token"
```

Or:

```bash
export RUSTRAK_SERVICE_TOKEN="your-service-token"
```

Or:

```bash
export RUSTRAK_PAT_TOKEN="your-personal-access-token"
```

The token is marked as sensitive by the provider and should not be committed to source control.

## Provider Configuration

The provider supports the following arguments.

### `host`

The URL of the Rustrak API.

**Type:** `string`

**Required:** No

**Environment variable:** `RUSTRAK_HOST`

If `host` is not explicitly configured, the provider uses its default API host.

```hcl
provider "rustrak" {
  host = "https://api.rustrak.com"
}
```

Alternatively:

```bash
export RUSTRAK_HOST="https://api.rustrak.com"
```

### `rustrak_token`

The token used to authenticate with the Rustrak API.

**Type:** `string`

**Required:** Yes

**Sensitive:** Yes

**Environment variables:**

* `RUSTRAK_TOKEN`
* `RUSTRAK_SERVICE_TOKEN`
* `RUSTRAK_PAT_TOKEN`

Example:

```hcl
provider "rustrak" {
  rustrak_token = var.rustrak_token
}
```

## Resources

The provider currently supports the following resources:

* [`rustrak_project`](resources/project.md)

## Complete Example

```hcl
terraform {
  required_providers {
    rustrak = {
      source = "kitokinha/rustrak"
    }
  }
}

variable "rustrak_token" {
  type      = string
  sensitive = true
}

provider "rustrak" {
  rustrak_token = var.rustrak_token
}

resource "rustrak_project" "example" {
  name     = "My Project"
  slug     = "my-project"
  platform = "go"
}
```

The `dsn` returned by the project can then be referenced by other Terraform resources:

```hcl
output "project_dsn" {
  value     = rustrak_project.example.dsn
  sensitive = true
}
```
