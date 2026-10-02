# GcpCertManagerTrustConfig Terraform Module

This Terraform module creates one Certificate Manager trust config and enables the Certificate Manager API on the target project.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest trust-config.yaml
```

### Standalone Usage

```hcl
module "trust_config" {
  source = "./path/to/module"

  metadata = {
    name = "partner-mtls-trust"
  }

  spec = {
    # StringValueOrRef fields are flattened to plain strings by the tfvars
    # converter before the module sees them.
    project_id = "my-gcp-project"
    trust_stores = [{
      trust_anchors = [file("partner-root-ca.pem")]
    }]
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
| spec.trust_config_name | Trust config name (defaults to metadata.name) | string | no |
| spec.description | Human-readable description | string | no |
| spec.location | Certificate Manager location (empty = global) | string | no |
| spec.trust_stores | At most one store of `trust_anchors` and `intermediate_cas`, each a list of PEM certificates | list(object) | no |
| spec.allowlisted_certificates | PEM certificates accepted regardless of chain | list(string) | no |
| spec.labels | User labels (platform attribution labels win on conflicts) | map(string) | no |
| spec.deletion_policy | DELETE, PREVENT, or ABANDON (empty = provider default) | string | no |

Each PEM renders as its own `trust_anchors`, `intermediate_cas`, or `allowlisted_certificates` block.

## Outputs

| Name | Description |
|------|-------------|
| trust_config_id | Full resource name — what a server TLS policy and a backend authentication config take |
| trust_config_name | Trust config name in GCP |
| location | The Certificate Manager location (`global` unless set) |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.
