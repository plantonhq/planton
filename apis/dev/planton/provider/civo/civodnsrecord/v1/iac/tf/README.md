# CivoDnsRecord Terraform Module

This Terraform module provisions a Civo DNS record.

## Prerequisites

- Terraform 1.0+
- Civo API key with DNS management permissions

## Usage

### As Part of Planton

This module is typically invoked through the Planton CLI:

```bash
planton apply -f manifest.yaml --engine terraform
```

### Standalone Usage

```hcl
module "dns_record" {
  source = "./iac/tf"

  metadata = {
    name = "www-record"
    org  = "my-org"
    env  = "production"
  }

  spec = {
    zone_id = "your-zone-id"
    name    = "www"
    type    = "A"
    value   = "192.0.2.1"
    ttl     = 3600
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `metadata` | Resource metadata | `object` | Yes |
| `spec` | CivoDnsRecordSpec | `object` | Yes |

### spec Object

| Field | Description | Type | Required | Default |
|-------|-------------|------|----------|---------|
| `zone_id` | Civo Zone ID | `string` | Yes | - |
| `name` | Record name | `string` | Yes | - |
| `type` | Record type (A, AAAA, CNAME, MX, TXT, SRV, NS) | `string` | Yes | - |
| `value` | Record value | `string` | Yes | - |
| `ttl` | Time to live (60-86400 seconds) | `number` | No | 3600 |
| `priority` | Priority for MX/SRV records | `number` | No | 0 |

## Outputs

| Name | Description |
|------|-------------|
| `record_id` | Civo DNS record ID |
| `hostname` | Record hostname |
| `record_type` | DNS record type |
| `account_id` | Civo account ID |

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `CIVO_TOKEN` | Civo API key | Yes |

## Troubleshooting

### "authentication failed"

Verify your Civo API key is valid and has DNS permissions:

```bash
export CIVO_TOKEN="your-api-key"
```

### "zone not found"

Verify the `zone_id` matches an existing Civo DNS zone.

### "invalid record type"

Ensure `type` is one of: A, AAAA, CNAME, MX, TXT, SRV, NS.
