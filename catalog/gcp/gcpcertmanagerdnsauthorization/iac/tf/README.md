# GcpCertManagerDnsAuthorization Terraform Module

This Terraform module creates one Certificate Manager DNS authorization
and enables the Certificate Manager API on the target project.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest dns-authorization.yaml
```

### Standalone Usage

```hcl
module "dns_authorization" {
  source = "./path/to/module"

  metadata = {
    name = "example-com-auth"
  }

  spec = {
    # StringValueOrRef fields are flattened to plain strings by the tfvars
    # converter before the module sees them.
    project_id = "my-gcp-project"
    domain     = "example.com"
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
| spec.project_id | GCP project ID; empty falls back to the provider's default project | string | no |
| spec.authorization_name | Authorization name (defaults to metadata.name) | string | no |
| spec.domain | The domain being authorized (covers its wildcard too) | string | yes |
| spec.location | Certificate Manager location (empty = global) | string | no |
| spec.type | FIXED_RECORD or PER_PROJECT_RECORD (empty = GCP default) | string | no |
| spec.labels | User labels (platform attribution labels win on conflicts) | map(string) | no |

## Outputs

| Name | Description |
|------|-------------|
| authorization_id | Fully-qualified resource ID — consumed by certificates |
| authorization_name | Authorization name in GCP |
| domain | The authorized domain |
| dns_record_name | Fully-qualified name of the validation record |
| dns_record_type | Validation record type (CNAME) |
| dns_record_data | Validation record data — the CNAME target |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.
