# GcpCertManagerCert Terraform Module

This Terraform module creates one Certificate Manager certificate —
Google-managed (auto-renewed) or self-managed (uploaded PEM) — and enables
the Certificate Manager API on the target project.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest certificate.yaml
```

### Standalone Usage

```hcl
module "certificate" {
  source = "./path/to/module"

  metadata = {
    name = "web-cert"
  }

  spec = {
    # StringValueOrRef fields are flattened to plain strings by the tfvars
    # converter before the module sees them.
    project_id = "my-gcp-project"
    managed = {
      domains            = ["app.example.com"]
      dns_authorizations = ["projects/my-gcp-project/locations/global/dnsAuthorizations/app-auth"]
    }
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
| spec.cert_name | Certificate name (defaults to metadata.name) | string | no |
| spec.location | Certificate Manager location (empty = global) | string | no |
| spec.scope | DEFAULT / EDGE_CACHE / ALL_REGIONS / CLIENT_AUTH | string | no |
| spec.managed | Managed arm: domains, dns_authorizations (auth IDs), issuance_config | object | one of |
| spec.self_managed | Uploaded arm: pem_certificate + pem_private_key (sensitive) | object | one of |
| spec.labels | User labels (platform attribution labels win on conflicts) | map(string) | no |

## Outputs

| Name | Description |
|------|-------------|
| certificate_id | Fully-qualified resource ID |
| certificate_name | Certificate name — consumed by target HTTPS proxies |
| san_dnsnames | SANs in the issued certificate |
| location | The Certificate Manager location |
| managed_state | PROVISIONING/FAILED/ACTIVE for managed; empty for self-managed |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.

## Notes

- DNS authorizations are separate `GcpCertManagerDnsAuthorization`
  resources; this module never creates them or their DNS records.
- A managed certificate stays PROVISIONING until its validation records
  resolve publicly — creation succeeds regardless.
