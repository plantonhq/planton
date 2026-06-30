# AliCloud DNS Domain

Registers and manages an Alibaba Cloud DNS domain in the Alidns service with optional group assignment, resource group placement, domain remarks, and automatic tag management. The domain is the prerequisite for creating DNS records (A, AAAA, CNAME, MX, TXT, etc.) via the AliCloudDnsRecord component.

## What Gets Created

When you deploy an AliCloudDnsZone resource, Planton provisions:

- **Alidns Domain** -- an `alicloud_alidns_domain` resource (Pulumi: `dns.AlidnsDomain`) that registers the domain in the Alidns hosted zone
- **DNS Servers** -- Alibaba Cloud assigns a set of authoritative nameservers; point your domain registrar's NS records to these servers for Alidns to serve queries
- **Tags** -- system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or Planton provider config
- **Domain ownership** -- you must own or control the domain at your registrar to point NS records to the Alibaba Cloud DNS servers
- **Planton CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `dns-zone.yaml`:

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudDnsZone
metadata:
  name: my-domain
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudDnsZone.my-domain
spec:
  region: cn-hangzhou
  domainName: example.com
```

Deploy:

```shell
planton apply -f dns-zone.yaml
```

This registers the domain in Alidns. After deployment, retrieve the `dns_servers` output and update your domain registrar's NS records.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for provider initialization (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`). Alidns is a global service, but the provider requires a region. | Required; non-empty |
| `domainName` | `string` | The domain name to register in Alidns (e.g., `example.com`). Cannot be changed after creation. | Required; 1-253 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `groupId` | `string` | `""` | Alidns domain group ID. Groups organize domains in the console. If omitted, the domain is placed in the default group. |
| `remark` | `string` | `""` | Description or notes for the domain. Visible in the Alidns console. |
| `resourceGroupId` | `string` | `""` | Alibaba Cloud resource group ID for access control and cost attribution. Cannot be changed after creation. |
| `tags` | `map<string, string>` | `{}` | User-defined key-value tags. Merged with system tags; user values take precedence on key conflict. |

## Examples

### Basic Domain Registration

Register a domain with only the required fields. Suitable for development or simple DNS hosting.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudDnsZone
metadata:
  name: dev-domain
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AliCloudDnsZone.dev-domain
spec:
  region: cn-hangzhou
  domainName: dev.example.com
```

### Production Domain with Tags

A production domain with resource group placement and organizational tags for governance.

```yaml
apiVersion: alicloud.planton.dev/v1
kind: AliCloudDnsZone
metadata:
  name: prod-domain
  org: my-org
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AliCloudDnsZone.prod-domain
spec:
  region: cn-shanghai
  domainName: platform.example.com
  remark: Primary platform domain for production services
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | The domain ID assigned by Alibaba Cloud. |
| `domain_name` | `string` | The domain name as registered in Alidns. |
| `dns_servers` | `repeated string` | DNS server names assigned by Alibaba Cloud. Point your registrar's NS records to these servers. |
| `group_name` | `string` | The domain group name (computed from the `groupId` input). Empty when in the default group. |
| `puny_code` | `string` | Punycode representation for internationalized domain names containing non-ASCII characters. |

## Related Components

- [AliCloudDnsRecord](/docs/catalog/alicloud/aliclouddnsrecord) -- creates DNS records (A, AAAA, CNAME, MX, TXT, NS, SRV) within this domain
- [AliCloudPrivateDnsZone](/docs/catalog/alicloud/alicloudprivatednszone) -- manages private DNS zones for VPC-internal resolution (separate from public Alidns)
