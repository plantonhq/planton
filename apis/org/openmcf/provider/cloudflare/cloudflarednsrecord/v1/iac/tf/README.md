# CloudflareDnsRecord Terraform Module

This Terraform module provisions a Cloudflare DNS record.

## Prerequisites

- Terraform 1.0+
- Cloudflare API token with DNS:Edit permissions

## Usage

### As Part of OpenMCF

This module is typically invoked through the OpenMCF CLI:

```bash
planton apply -f manifest.yaml --iac terraform
```

### Standalone Usage

```hcl
module "dns_record" {
  source = "./path/to/module"

  metadata = {
    name = "www-a-record"
  }

  spec = {
    zone_id = { value = "your-zone-id" }
    name    = "www"
    type    = "A"
    content = "192.0.2.1"
    proxied = true
    ttl     = 1
    comment = "Primary web server"
  }
}
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token | Yes |

## Inputs

### metadata

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Resource name |
| `id` | string | No | Optional resource ID |
| `org` | string | No | Organization |
| `env` | string | No | Environment |
| `labels` | map(string) | No | Labels |
| `tags` | list(string) | No | Tags |

### spec

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `zone_id` | string | Yes | - | Cloudflare Zone ID (StringValueOrRef, flattened to a string) |
| `name` | string | Yes | - | DNS record name |
| `type` | string | Yes | - | Record type (any of the 21 Cloudflare types) |
| `content` | string | Cond. | "" | Value for simple types (A/AAAA/CNAME/MX/NS/PTR/TXT/OPENPGPKEY) |
| `data` | object | Cond. | - | Typed block for structured types (one of caa/cert/dnskey/ds/https/loc/naptr/smimea/srv/sshfp/svcb/tlsa/uri) |
| `proxied` | bool | No | false | Proxy through Cloudflare (A/AAAA/CNAME only) |
| `ttl` | number | No | 1 | TTL in seconds (0/1 = auto, or 30-86400) |
| `priority` | number | No | 0 | Priority for MX records |
| `comment` | string | No | "" | Comment |
| `tags` | list(string) | No | [] | Custom record tags |
| `settings` | object | No | - | `ipv4_only`, `ipv6_only`, `flatten_cname` (proxied records) |

Exactly one of `content` or a `data` block must be set, matching the record `type`.

## Outputs

| Name | Description |
|------|-------------|
| `record_id` | Cloudflare DNS record ID |
| `record_name` | The record name as stored by Cloudflare |
| `record_type` | DNS record type |
| `proxied` | Whether record is proxied |

## Examples

### A Record

```hcl
spec = {
  zone_id = { value = "abc123" }
  name    = "www"
  type    = "A"
  content = "192.0.2.1"
  proxied = true
}
```

### SRV Record (structured data)

```hcl
spec = {
  zone_id = { value = "abc123" }
  name    = "_sip._tcp"
  type    = "SRV"
  data = {
    srv = {
      priority = 10
      weight   = 5
      port     = 5060
      target   = "sip.example.com"
    }
  }
}
```

## Troubleshooting

### "authentication failed"

Ensure `CLOUDFLARE_API_TOKEN` environment variable is set with a valid token.

### "zone not found"

Verify the `zone_id` matches an existing Cloudflare zone.

### "invalid record type"

Ensure `type` is one of the 21 supported Cloudflare record types, and that you supply `content` for simple types or the matching `data` block for structured types.

## Validation

```bash
terraform init
terraform validate
```
