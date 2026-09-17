# GcpProject Terraform Module

This Terraform module creates one Google Cloud project under an
organization or folder — billing linked, labels/tags applied, the default
network suppressed by default, requested APIs pre-enabled, and the
configured deletion policy applied.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest project.yaml
```

### Standalone Usage

```hcl
module "project" {
  source = "./path/to/module"

  metadata = {
    name = "prod-workloads"
  }

  spec = {
    project_id         = "acme-prod-workloads"
    parent_type        = "folder"
    parent_id          = "123456789012"
    billing_account_id = "0123AB-4567CD-89EFGH"
    deletion_policy    = "PREVENT"
    enabled_apis       = ["compute.googleapis.com"]
  }
}
```

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.0 |
| google | ~> 8.3 |

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name | object | yes |
| spec.project_id | Globally-unique project ID (immutable) | string | yes |
| spec.display_name | Console display name (defaults to metadata.name) | string | no |
| spec.parent_type / spec.parent_id | "organization" or "folder" + numeric ID | string | no |
| spec.billing_account_id | Billing account to link | string | no |
| spec.labels | User labels (platform attribution labels win on conflicts) | map(string) | no |
| spec.tags | Resource-manager tags, create-time only | map(string) | no |
| spec.auto_create_network | Whether GCP auto-creates the default VPC (default false) | bool | no |
| spec.enabled_apis | Cloud APIs to pre-enable | list(string) | no |
| spec.deletion_policy | DELETE (default) / PREVENT / ABANDON | string | no |

## Outputs

| Name | Description |
|------|-------------|
| project_id | The unique project ID — resolved by every other kind's project reference |
| project_number | The numeric identifier assigned by Google |
| name | The display name |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs. The scopes are split:
project creation authorizes on the parent organization or folder, and the
billing link authorizes on the billing account itself.

## Notes

- IAM grants are separate `GcpProjectIamMember` resources; this module
  never touches the project's IAM policy.
- API enablement uses `disable_on_destroy = false`: removing an entry (or
  the project resource) never disables a service other tooling depends on.
