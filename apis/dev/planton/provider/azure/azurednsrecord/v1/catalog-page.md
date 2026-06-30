# Azure DNS Record

Deploys an individual DNS record (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, or CAA) within an existing Azure DNS Zone. The component supports all standard record types with configurable TTL, multiple record values for round-robin behavior, and MX priority for mail exchange records.

## What Gets Created

When you deploy an AzureDnsRecord resource, Planton provisions:

- **DNS Record** -- one of the following Pulumi Azure DNS resources based on the specified `type`: `dns.ARecord`, `dns.AaaaRecord`, `dns.CNameRecord`, `dns.MxRecord`, `dns.TxtRecord`, `dns.SrvRecord`, `dns.NsRecord`, `dns.PtrRecord`, or `dns.CaaRecord`
- **Azure Tags** -- resource metadata tags applied to the record for tracking and governance, including resource name, kind, organization, and environment

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** containing the DNS Zone (can reference an AzureResourceGroup resource)
- **An Azure DNS Zone** where the record will be created (can reference an AzureDnsZone resource)

## Quick Start

Create a file `dns-record.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: my-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureDnsRecord.my-a-record
spec:
  resourceGroup: my-rg
  zoneName: example.com
  type: A
  name: www
  values:
    - "192.0.2.1"
```

Deploy:

```shell
planton apply -f dns-record.yaml
```

This creates an A record for `www.example.com` pointing to `192.0.2.1` with a default TTL of 300 seconds (5 minutes).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group containing the DNS Zone. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `zoneName` | `StringValueOrRef` | Name of the DNS Zone where the record will be created (e.g., `example.com`). Can reference an AzureDnsZone resource via `valueFrom`. | Required |
| `type` | `enum` | DNS record type. Values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, `NS`, `PTR`, `CAA`. | Required, must be a defined enum value |
| `name` | `string` | Record name relative to the zone. Use `@` for zone apex, `*` for wildcard, or a valid DNS label (e.g., `www`, `api.v1`). | Required, must match `@`, `*`, or lowercase alphanumeric with hyphens and dots |
| `values` | `string[]` | Record values. Format depends on type: IPv4 for A, IPv6 for AAAA, hostname for CNAME, mail server for MX, text for TXT, `priority weight port target` for SRV, `flags tag value` for CAA. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ttlSeconds` | `int32` | `300` | Time to live in seconds. Determines how long resolvers cache this record. Range: 1--2147483647. Common values: 60 (1 min), 300 (5 min), 3600 (1 hour), 86400 (1 day). |
| `mxPriority` | `int32` | `10` | Priority value for MX records. Lower values indicate higher priority. Only applicable when `type` is `MX`. Range: 0--65535. |

## Examples

### A Record for a Subdomain

Point a subdomain to one or more IPv4 addresses with round-robin behavior:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: web-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsRecord.web-a-record
spec:
  resourceGroup: prod-rg
  zoneName: example.com
  type: A
  name: www
  values:
    - "192.0.2.1"
    - "192.0.2.2"
  ttlSeconds: 3600
```

### CNAME Record for an Alias

Create an alias from one hostname to another:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: app-cname
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsRecord.app-cname
spec:
  resourceGroup: prod-rg
  zoneName: example.com
  type: CNAME
  name: app
  values:
    - "myapp.azurewebsites.net"
  ttlSeconds: 300
```

### MX Records for Email Routing

Configure mail exchange records with priority for primary and secondary mail servers:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: mail-mx-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsRecord.mail-mx-record
spec:
  resourceGroup: prod-rg
  zoneName: example.com
  type: MX
  name: "@"
  values:
    - "mail1.example.com"
    - "mail2.example.com"
  ttlSeconds: 3600
  mxPriority: 10
```

### TXT Record for Domain Verification

Add SPF or domain-verification TXT records at the zone apex:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: spf-txt-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsRecord.spf-txt-record
spec:
  resourceGroup: prod-rg
  zoneName: example.com
  type: TXT
  name: "@"
  values:
    - "v=spf1 include:_spf.google.com ~all"
  ttlSeconds: 3600
```

### Using Foreign Key References

Reference Planton-managed resources instead of hardcoding the resource group and zone name. The `resourceGroup` field defaults to kind `AzureResourceGroup` with field path `status.outputs.resource_group_name`. The `zoneName` field defaults to kind `AzureDnsZone` with field path `status.outputs.zone_name`.

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsRecord
metadata:
  name: ref-a-record
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsRecord.ref-a-record
spec:
  resourceGroup:
    valueFrom:
      name: my-rg
  zoneName:
    valueFrom:
      name: my-azure-zone
  type: A
  name: api
  values:
    - "10.0.1.50"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `record_id` | `string` | Azure Resource Manager ID of the DNS record (format: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Network/dnsZones/{zone}/{type}/{name}`) |
| `fqdn` | `string` | Fully qualified domain name for this record (e.g., `www.example.com`) |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group containing the DNS Zone
- [AzureDnsZone](/docs/catalog/azure/azurednszone) -- provides the DNS Zone where records are created
- [AzurePublicIp](/docs/catalog/azure/azurepublicip) -- public IP addresses that A records can point to
- [AzureLoadBalancer](/docs/catalog/azure/azureloadbalancer) -- load balancer frontend IPs that DNS records can target
