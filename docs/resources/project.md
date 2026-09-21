# rustrak_project Resource

The `rustrak_project` resource allows you to create and manage projects in Rustrak.

## Example Usage

```hcl
resource "rustrak_project" "example" {
  name     = "My Project"
  slug     = "my-project"
  platform = "go"
}
```

## Minimal Configuration

Only `name` is required.

```hcl
resource "rustrak_project" "example" {
  name = "My Project"
}
```

## Arguments

### `name`

The name of the Rustrak project.

**Type:** `string`

**Required:** Yes

```hcl
resource "rustrak_project" "example" {
  name = "My Project"
}
```

Changing the name updates the existing Rustrak project.

---

### `slug`

The slug of the Rustrak project.

**Type:** `string`

**Required:** No

**Computed:** Yes

The slug can be explicitly configured:

```hcl
resource "rustrak_project" "example" {
  name = "My Project"
  slug = "my-project"
}
```

If omitted, Rustrak can generate or determine the project's slug.

Because this attribute is both `Optional` and `Computed`, Terraform can use the value returned by the Rustrak API when it is not explicitly configured.

---

### `platform`

The platform associated with the Rustrak project.

**Type:** `string`

**Required:** No

**Computed:** Yes

Example:

```hcl
resource "rustrak_project" "example" {
  name     = "My Go Application"
  platform = "go"
}
```

Like `slug`, the platform can be returned by the Rustrak API when it is not explicitly configured.

---

## Attributes

### `id`

The unique identifier of the Rustrak project.

The resource ID is assigned by Rustrak when the project is created.

Terraform stores the project ID as a string.

Example:

```hcl
output "project_id" {
  value = rustrak_project.example.id
}
```

### `dsn`

The DSN used to send events to the Rustrak project.

**Type:** `string`

**Computed:** Yes

The DSN is returned by the Rustrak API after the project is created or read.

Example:

```hcl
output "project_dsn" {
  value     = rustrak_project.example.dsn
  sensitive = true
}
```

## Import

Existing Rustrak projects can be imported into Terraform using their project ID.

```bash
terraform import rustrak_project.example 123
```

Where `123` is the Rustrak project ID.

After importing the resource, run:

```bash
terraform plan
```

Terraform will read the project from Rustrak and populate its attributes.

## Lifecycle

The resource supports the complete Terraform resource lifecycle:

* Create
* Read
* Update
* Delete
* Import

### Create

Terraform creates a new Rustrak project using the configured `name`, `slug`, and `platform`.

```hcl
resource "rustrak_project" "example" {
  name     = "My Project"
  slug     = "my-project"
  platform = "go"
}
```

After creation, Terraform stores the project ID and reads the complete project state from Rustrak.

### Read

Terraform retrieves the project from Rustrak using its ID.

The following attributes are populated from the API response:

* `name`
* `slug`
* `platform`
* `dsn`

If the project no longer exists, Terraform removes it from the Terraform state.

### Update

Changes to the following attributes are sent to Rustrak:

* `name`
* `slug`
* `platform`

After the update, Terraform refreshes the resource state from Rustrak.

### Delete

Destroying the Terraform resource deletes the corresponding project from Rustrak.

```bash
terraform destroy
```

After successful deletion, the project is removed from Terraform state.

## Complete Example

```hcl
terraform {
  required_providers {
    rustrak = {
      source = "kitokinha/rustrak"
    }
  }
}

provider "rustrak" {}

resource "rustrak_project" "application" {
  name     = "My Application"
  slug     = "my-application"
  platform = "go"
}

output "project_id" {
  value = rustrak_project.application.id
}

output "project_dsn" {
  value     = rustrak_project.application.dsn
  sensitive = true
}
```
