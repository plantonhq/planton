# GcpDnsRecord Terraform Module

This Terraform module creates and manages DNS records in Google Cloud DNS Managed Zones.

## Usage

### With Planton CLI

```bash
planton terraform apply --manifest dns-record.yaml
```

### Standalone Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-example-com"
  }

  spec = {
    project_id = {
      value = "my-gcp-project"
    }
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
| google | 6.19.0 |

## Providers

| Name | Version |
|------|---------|
| google | 6.19.0 |

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| metadata | Resource metadata including name | object | yes |
| spec.project_id | GCP project ID (StringValueOrRef) | object | yes |
| spec.managed_zone | Name of the Cloud DNS Managed Zone | string | yes |
| spec.type | DNS record type (A, AAAA, CNAME, etc.) | string | yes |
| spec.name | FQDN for the record (must end with dot) | string | yes |
| spec.values | Record values (supports multiple for round-robin) | list(string) | yes |
| spec.ttl_seconds | TTL in seconds (default: 300) | number | no |

## Outputs

| Name | Description |
|------|-------------|
| fqdn | The fully qualified domain name of the record |
| record_type | The DNS record type |
| managed_zone | The managed zone containing the record |
| project_id | The GCP project ID |
| ttl_seconds | The TTL in seconds |

## Required Permissions

The service account running Terraform needs:
- `roles/dns.admin` - Full DNS management

Or more restrictive:
- `roles/dns.recordset.editor` - Record set operations only

## Examples

### A Record

```hcl
spec = {
  project_id   = { value = "my-project" }
  managed_zone = "example-zone"
  type         = "A"
  name         = "www.example.com."
  values       = ["192.0.2.1"]
}
```

### CNAME Record

```hcl
spec = {
  project_id   = { value = "my-project" }
  managed_zone = "example-zone"
  type         = "CNAME"
  name         = "blog.example.com."
  values       = ["example.github.io."]
}
```

### Round-Robin A Record

```hcl
spec = {
  project_id   = { value = "my-project" }
  managed_zone = "example-zone"
  type         = "A"
  name         = "api.example.com."
  values       = ["192.0.2.1", "192.0.2.2", "192.0.2.3"]
  ttl_seconds  = 60
}
```
