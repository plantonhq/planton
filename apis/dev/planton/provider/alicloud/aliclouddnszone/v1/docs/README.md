# AliCloudDnsZone -- Research Documentation

## Service Overview

Alibaba Cloud DNS (Alidns) is a globally distributed authoritative DNS hosting service. It provides DNS resolution for domain names with features including domain management, traffic management, DNS security, and monitoring.

### Key Concepts

- **Domain**: A top-level or subdomain registered in the Alidns service. Registering a domain in Alidns does not purchase or transfer it -- it creates a hosted zone so that DNS records can be managed.
- **DNS Servers**: Alibaba Cloud assigns authoritative nameservers when a domain is added. The domain owner must update their registrar's NS records to point to these servers.
- **Domain Group**: An organizational construct for grouping related domains in the Alidns console. Groups do not affect DNS resolution.
- **Punycode**: The ASCII-compatible encoding of internationalized domain names (IDN) that contain non-ASCII characters.

### Alidns vs. PrivateZone

Alidns is for **public DNS resolution** -- queries from the internet resolve against Alidns records. PrivateZone is for **private DNS resolution** within a VPC -- queries from VPC resources resolve against PrivateZone records. These are separate services with separate APIs and components.

## Provider Implementation

### Terraform Resource

- **Resource**: `alicloud_alidns_domain`
- **Legacy alias**: `alicloud_dns_domain` (deprecated)
- **Source**: `alicloud/resource_alicloud_alidns_domain.go`

#### Schema

| Field | Type | Required | Computed | ForceNew | Description |
|-------|------|----------|----------|----------|-------------|
| `domain_name` | string | Yes | No | Yes | Domain name (resource ID) |
| `group_id` | string | No | No | No | Domain group ID |
| `lang` | string | No | No | No | API language (internal) |
| `remark` | string | No | No | No | Domain remark |
| `resource_group_id` | string | No | Yes | Yes | Resource group ID |
| `tags` | map | No | No | No | Tags |
| `dns_servers` | set(string) | No | Yes | No | DNS server addresses |
| `domain_id` | string | No | Yes | No | Domain ID |
| `group_name` | string | No | Yes | No | Domain group name |
| `puny_code` | string | No | Yes | No | Punycode representation |

#### API Operations

- **Create**: `alidns.AddDomain` -- registers the domain; optionally sets group_id and resource_group_id
- **Read**: `alidns.DescribeDomainInfo` -- retrieves domain details and computed fields
- **Update**: `UpdateDomainRemark` (remark), `ChangeDomainGroup` (group_id), `SetResourceTags` (tags)
- **Delete**: `alidns.DeleteDomain` -- removes the domain from Alidns

#### ForceNew Fields

- `domain_name` -- changing the domain name requires destroying and recreating the resource
- `resource_group_id` -- resource group assignment cannot be changed after creation

### Pulumi Resource

- **Type**: `dns.AlidnsDomain`
- **Module**: `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/dns`
- **Deprecated alternatives**: `dns.DnsDomain` (v1.95.0), `dns.Domain` (earlier)

#### Input Args (`AlidnsDomainArgs`)

| Field | Type | Description |
|-------|------|-------------|
| `DomainName` | `pulumi.StringInput` | Domain name (required) |
| `GroupId` | `pulumi.StringPtrInput` | Domain group ID |
| `Lang` | `pulumi.StringPtrInput` | API language |
| `Remark` | `pulumi.StringPtrInput` | Domain remark |
| `ResourceGroupId` | `pulumi.StringPtrInput` | Resource group ID |
| `Tags` | `pulumi.StringMapInput` | Tags |

#### Outputs

| Field | Type | Description |
|-------|------|-------------|
| `DnsServers` | `pulumi.StringArrayOutput` | DNS server names |
| `DomainId` | `pulumi.StringOutput` | Domain ID |
| `DomainName` | `pulumi.StringOutput` | Domain name |
| `GroupName` | `pulumi.StringOutput` | Domain group name |
| `PunyCode` | `pulumi.StringOutput` | Punycode |
| `ResourceGroupId` | `pulumi.StringOutput` | Resource group ID |

## Design Decisions

### Fields Included

- **`region`**: Required for provider initialization even though Alidns is global.
- **`domain_name`**: Core required field. ForceNew -- cannot be changed after creation.
- **`group_id`**: Useful for organizations managing many domains. Input is the group ID (not name); the name is a computed output.
- **`remark`**: A lightweight description field. More appropriate than a full `description` field for domain metadata.
- **`resource_group_id`**: Per DD05, included on resources that support it. ForceNew.
- **`tags`**: Consistent with all other Alibaba Cloud components in the platform.

### Fields Excluded

- **`lang`**: Controls the API response language. Not meaningful for infrastructure-as-code users and would clutter the spec.

### Spec Corrections from T02

The initial T02 resource queue design specified `group_name` as an input field. However, the provider takes `group_id` as input and computes `group_name` as an output. This was corrected during implementation.

## Related Resources

- **`alicloud_alidns_domain_group`**: Manages domain groups. Not included as a separate Planton component -- groups are referenced by ID.
- **`alicloud_alidns_domain_attachment`**: Binds domains to premium DNS instances. Out of scope for the standard component.
- **`alicloud_alidns_record`**: Creates DNS records within a domain. Managed by the AliCloudDnsRecord component.

## Limits and Quotas

- Free Alibaba Cloud DNS supports up to 10 subdomains per domain
- Enterprise DNS plans support unlimited subdomains and advanced features (DNSSEC, traffic management)
- Maximum domain name length: 253 characters (per RFC 1035)
- Maximum tags per resource: 20 (Alibaba Cloud standard limit)
