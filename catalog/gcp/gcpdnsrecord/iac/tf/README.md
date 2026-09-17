# GcpDnsRecord Terraform Module

This Terraform module creates one DNS record set in a Google Cloud DNS
managed zone — static values (round-robin) or exactly one routing policy
(weighted round robin, geolocation, or primary/backup failover). It also
enables the Cloud DNS API on the target project.

## Usage

### With Planton CLI

```bash
planton tofu apply --manifest dns-record.yaml
```

### Standalone Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-example-com"
  }

  spec = {
    # StringValueOrRef fields are flattened to plain strings by the tfvars
    # converter before the module sees them.
    project_id   = "my-gcp-project"
    managed_zone = "example-zone"
    type         = "A"
    name         = "www.example.com."
    values       = ["192.0.2.1"]
    ttl_seconds  = 300
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
| spec.managed_zone | Name of the Cloud DNS Managed Zone (zone resource name) | string | yes |
| spec.type | DNS record type, uppercase (any type the API supports) | string | yes |
| spec.name | FQDN for the record (must end with dot) | string | yes |
| spec.values | Static record values (round-robin); mutually exclusive with routing_policy | list(string) | one of |
| spec.routing_policy | Query steering: exactly one of wrr / geo / primary_backup | object | one of |
| spec.ttl_seconds | TTL in seconds (default: 300) | number | no |

## Outputs

| Name | Description |
|------|-------------|
| fqdn | The fully qualified domain name of the record |
| record_type | The DNS record type |
| managed_zone | The managed zone containing the record |
| project_id | The GCP project ID the record was created in |
| ttl_seconds | The TTL in seconds |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege
permission set the deploying principal needs.

## Examples

### A Record

```hcl
spec = {
  managed_zone = "example-zone"
  type         = "A"
  name         = "www.example.com."
  values       = ["192.0.2.1"]
}
```

### Weighted Canary

```hcl
spec = {
  managed_zone = "example-zone"
  type         = "A"
  name         = "api.example.com."
  routing_policy = {
    wrr = [
      { weight = 95, values = ["192.0.2.1"] },
      { weight = 5, values = ["192.0.2.2"] },
    ]
  }
  ttl_seconds = 60
}
```

### Geolocation Routing

```hcl
spec = {
  managed_zone = "example-zone"
  type         = "A"
  name         = "app.example.com."
  routing_policy = {
    geo = [
      { location = "us-east1", values = ["192.0.2.1"] },
      { location = "europe-west3", values = ["192.0.2.2"] },
    ]
  }
}
```
