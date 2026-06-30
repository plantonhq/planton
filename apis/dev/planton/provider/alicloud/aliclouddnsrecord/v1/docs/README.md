# AliCloudDnsRecord -- Research Documentation

## Service Overview

Alibaba Cloud DNS (Alidns) manages DNS records within domains hosted in the Alidns service. Each record maps a host record (subdomain) and type to a value. Records are scoped to a parent domain that must already exist in Alidns.

### Key Concepts

- **Resource Record (rr)**: The subdomain part of the DNS name. `@` represents the apex (bare domain), `*` is a wildcard, and any other value is a specific subdomain label.
- **Record Type**: Determines how the value is interpreted -- A for IPv4 addresses, AAAA for IPv6, CNAME for domain aliases, MX for mail routing, TXT for text verification, NS for delegation, SRV for service location, CAA for certificate authority authorization.
- **Resolution Line**: Controls which clients receive this record based on ISP or geography. The `default` line serves all clients. Advanced lines (telecom, unicom, mobile, overseas) enable intelligent DNS routing for China multi-ISP deployments.
- **Record Status**: Records can be set to ENABLE (served in DNS responses) or DISABLE (exists in Alidns but not resolved). This allows staging records or temporarily removing them without deletion.

### Record Type Reference

| Type | Value Format | Notes |
|------|-------------|-------|
| `A` | IPv4 address | e.g., `203.0.113.10` |
| `AAAA` | IPv6 address | e.g., `2001:db8::1` |
| `CNAME` | Domain name | No trailing dot; e.g., `cdn.provider.com` |
| `MX` | Domain name | Requires `priority` (1-10); e.g., `mx1.example.com` |
| `TXT` | Text string | SPF, DKIM, domain verification; e.g., `v=spf1 ...` |
| `NS` | Domain name | Nameserver delegation; no trailing dot |
| `SRV` | Priority weight port target | Service locator |
| `CAA` | Flags tag value | Certificate authority; e.g., `0 issue "letsencrypt.org"` |
| `REDIRECT_URL` | URL | Explicit URL redirect (not available on international sites) |
| `FORWORD_URL` | URL | Implicit URL forward; `line` must be `default`. Note: "FORWORD" is the Alibaba Cloud API's actual spelling. |

## Provider Implementation

### Terraform Resource

- **Resource**: `alicloud_alidns_record`
- **Legacy alias**: `alicloud_dns_record` (deprecated since v1.85.0)
- **Source**: `alicloud/resource_alicloud_alidns_record.go`

#### Schema

| Field | Type | Required | Computed | ForceNew | Default | Description |
|-------|------|----------|----------|----------|---------|-------------|
| `domain_name` | string | Yes | No | Yes | - | Parent domain name |
| `rr` | string | Yes | No | No | - | Host record (subdomain) |
| `type` | string | Yes | No | No | - | DNS record type |
| `value` | string | Yes | No | No | - | Record value |
| `ttl` | int | No | No | No | `600` | Time-to-live in seconds |
| `priority` | int | No | No | No | - | MX priority (1-10), required for MX records |
| `line` | string | No | No | No | `"default"` | DNS resolution line |
| `status` | string | No | No | No | `"ENABLE"` | ENABLE or DISABLE |
| `remark` | string | No | No | No | - | Record description |
| `lang` | string | No | No | No | - | API response language (excluded from Planton spec) |
| `user_client_ip` | string | No | No | No | - | Client IP for API (excluded from Planton spec) |

#### Diff Suppression

- **`value`**: For NS, MX, CNAME, SRV records, trailing dots and whitespace are trimmed before comparison.
- **`priority`**: Diff suppressed for non-MX records (priority is irrelevant for those types).

#### ForceNew Fields

- `domain_name` -- changing the parent domain requires destroying and recreating the record.

### Pulumi Resource

- **Type**: `dns.AlidnsRecord`
- **Module**: `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/dns`
- **Token**: `alicloud:dns/alidnsRecord:AlidnsRecord`

## Design Decisions

### Fields Included

- **`region`**: Required for provider initialization, consistent with all other Alibaba Cloud components.
- **`domain_name`**: Plain string (not StringValueOrRef) because domain names are human-readable values the user always knows. Using StringValueOrRef would add complexity without meaningful benefit.
- **`rr`**: Matches the provider field name. "rr" stands for "Resource Record" in the Alibaba Cloud API. While `host_record` would be more descriptive, matching the provider naming avoids confusion when users cross-reference with Terraform/Pulumi docs.
- **`type`**: String with CEL validation instead of a proto enum. This avoids enum-to-string conversion and keeps values provider-authentic (exact casing the API expects).
- **`line`**: Important for China multi-ISP deployments. Default `"default"` is correct for most users.
- **`status`**: Enables the common pattern of disabling a record without deleting it.
- **`remark`**: Consistent with AliCloudDnsZone which also has a remark field.

### Fields Excluded

- **`lang`**: Controls the API response language. Not a resource configuration attribute.
- **`user_client_ip`**: Client IP for API request tracking. Not a resource configuration attribute.

### No Tags

The `alicloud_alidns_record` resource does not support tags (unlike `alicloud_alidns_domain`). No tag computation is performed in locals.

### Spec Corrections from T02

- **`host_record` -> `rr`**: T02 used `host_record` but the provider field is `rr`. Per coding guidelines, field names match the provider.
- **Added `region`**: Needed for provider initialization; not in T02 spec.
- **Added `line`**: DNS resolution line; not in T02 spec.
- **Added `status`**: Record enable/disable; not in T02 spec.
- **Added `remark`**: Description field; not in T02 spec.
- **Expanded `type` values**: T02 listed 7 types; provider supports 10.

## Related Resources

- **`alicloud_alidns_domain`**: The parent domain. Must exist before records can be created.
- **`alicloud_pvtz_zone_record`**: Private zone records (different service, different API).

## Limits and Quotas

- Free Alidns plan: up to 10 subdomains per domain, TTL range 600-86400
- Basic plan: TTL range 120-86400
- Standard plan: TTL range 60-86400
- Ultimate/Exclusive plans: TTL range down to 1 second
- Maximum host record (rr) length: 253 characters
- Maximum domain name length: 253 characters (RFC 1035)
