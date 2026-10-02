# GcpCertManagerIssuanceConfig Terraform Module

This Terraform module creates one Certificate Manager certificate issuance config and enables the Certificate Manager API on the target project.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest issuance-config.yaml
```

### Standalone Usage

```hcl
module "issuance_config" {
  source = "./path/to/module"

  metadata = {
    name = "internal-tls-issuance"
  }

  spec = {
    # StringValueOrRef fields are flattened to plain strings by the tfvars
    # converter before the module sees them.
    project_id                 = "my-gcp-project"
    ca_pool                    = "projects/my-gcp-project/locations/us-central1/caPools/internal-pool"
    key_algorithm              = "ECDSA_P256"
    lifetime                   = "2592000s"
    rotation_window_percentage = 66
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
| spec.issuance_config_name | Issuance config name (defaults to metadata.name) | string | no |
| spec.description | Human-readable description | string | no |
| spec.location | Certificate Manager location (empty = global) | string | no |
| spec.ca_pool | Full CA pool name, `projects/*/locations/*/caPools/*` | string | yes |
| spec.key_algorithm | RSA_2048 or ECDSA_P256 | string | yes |
| spec.lifetime | Certificate lifetime, 1814400s–2592000s | string | yes |
| spec.rotation_window_percentage | Renewal point as a percentage of lifetime | number | yes |
| spec.labels | User labels (platform attribution labels win on conflicts) | map(string) | no |
| spec.deletion_policy | DELETE, PREVENT, or ABANDON (empty = provider default) | string | no |

## Outputs

| Name | Description |
|------|-------------|
| issuance_config_id | Full resource name — what a certificate's `managed.issuance_config` takes |
| issuance_config_name | Issuance config name in GCP |
| location | The Certificate Manager location (`global` unless set) |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.
