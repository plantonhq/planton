# Azure DNS Zone

Deploys an Azure DNS Zone with an optional set of pre-populated DNS records. The component creates the zone in a specified resource group and supports A, AAAA, CNAME, MX, TXT, NS, CAA, SRV, and PTR record types, each with configurable TTL.

## What Gets Created

When you deploy an AzureDnsZone resource, Planton provisions:

- **DNS Zone** -- a `dns.Zone` resource in the specified resource group, representing the authoritative zone for the given domain name
- **DNS Records** -- one Azure DNS record resource per entry in `records`, created as the appropriate type (`dns.ARecord`, `dns.AaaaRecord`, `dns.CNameRecord`, `dns.MxRecord`, `dns.TxtRecord`, `dns.NsRecord`, `dns.CaaRecord`, `dns.SrvRecord`, `dns.PtrRecord`)
- **Azure Tags** -- resource metadata tags applied to the zone and all records for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the DNS zone will be created (can reference an AzureResourceGroup resource)
- **Domain ownership** -- you must own or control the domain to point its NS records at the Azure-assigned name servers after deployment

## Quick Start

Create a file `dnszone.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: my-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureDnsZone.my-zone
spec:
  zoneName: example.com
  resourceGroup: my-rg
```

Deploy:

```shell
planton apply -f dnszone.yaml
```

This creates an empty DNS zone for `example.com`. After deployment, update your domain registrar to use the name servers returned in `status.outputs.nameservers`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `zoneName` | `string` | DNS zone name (e.g., `example.com`). Do not include a trailing dot. | Required. Must match a valid DNS domain pattern. |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name where the zone will be created. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `records` | `AzureDnsRecord[]` | `[]` | DNS records to pre-populate in the zone. If omitted, the zone is created empty. |
| `records[].recordType` | `enum` | -- | DNS record type. Values: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `CAA`, `SRV`, `PTR`. |
| `records[].name` | `string` | -- | Record name. Use a fully qualified domain name ending with a dot (e.g., `www.example.com.`) or a relative name within the zone (e.g., `www`). Use `@` for the zone root. |
| `records[].values` | `string[]` | -- | Record values. IP addresses for A/AAAA, hostnames for CNAME (trailing dot), mail exchangers for MX, etc. Minimum 1 entry. |
| `records[].ttlSeconds` | `int` | `60` | Time To Live for the record, in seconds. |

## Examples

### Empty Zone

A minimal zone with no records, useful when DNS records are managed by an external system:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: empty-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureDnsZone.empty-zone
spec:
  zoneName: example.com
  resourceGroup: dev-rg
```

### Zone with A and CNAME Records

A zone that maps the apex domain and `www` subdomain to IP addresses and an alias:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: web-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsZone.web-zone
spec:
  zoneName: example.com
  resourceGroup: prod-rg
  records:
    - recordType: A
      name: example.com.
      values:
        - "203.0.113.10"
        - "203.0.113.11"
      ttlSeconds: 300
    - recordType: CNAME
      name: www.example.com.
      values:
        - "example.com."
      ttlSeconds: 3600
```

### Zone with MX and TXT Records for Email

A zone configured for email delivery with MX records and an SPF TXT record:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: mail-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsZone.mail-zone
spec:
  zoneName: example.com
  resourceGroup: prod-rg
  records:
    - recordType: MX
      name: example.com.
      values:
        - "mail1.example.com."
        - "mail2.example.com."
      ttlSeconds: 3600
    - recordType: TXT
      name: example.com.
      values:
        - "v=spf1 include:_spf.example.com ~all"
      ttlSeconds: 3600
```

### Zone with Mixed Record Types

A comprehensive zone with A, AAAA, CNAME, CAA, and NS records:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: full-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsZone.full-zone
spec:
  zoneName: example.com
  resourceGroup: prod-rg
  records:
    - recordType: A
      name: example.com.
      values:
        - "203.0.113.10"
      ttlSeconds: 300
    - recordType: AAAA
      name: example.com.
      values:
        - "2001:db8::1"
      ttlSeconds: 300
    - recordType: CNAME
      name: cdn.example.com.
      values:
        - "example.azureedge.net."
      ttlSeconds: 3600
    - recordType: CAA
      name: example.com.
      values:
        - "letsencrypt.org"
      ttlSeconds: 3600
    - recordType: NS
      name: subdomain.example.com.
      values:
        - "ns1.delegated.example.com."
        - "ns2.delegated.example.com."
      ttlSeconds: 86400
```

### Using Foreign Key References

Reference an Planton-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureDnsZone
metadata:
  name: ref-zone
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureDnsZone.ref-zone
spec:
  zoneName: example.com
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  records:
    - recordType: A
      name: example.com.
      values:
        - "203.0.113.10"
      ttlSeconds: 300
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `zoneId` | `string` | Azure Resource Manager ID of the DNS zone |
| `zoneName` | `string` | DNS zone name |
| `nameservers` | `string[]` | Name server addresses assigned to the DNS zone by Azure. Update your domain registrar to delegate to these name servers. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group where the DNS zone is created
- [AzureDnsRecord](/docs/catalog/azure/azurednsrecord) -- manages individual DNS records as standalone resources outside of the zone spec
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the virtual network that services resolved by this zone may run in
