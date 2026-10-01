# Azure DNS Record

Deploys one record set in an Azure public DNS zone — every value the zone answers for one (name, type) pair. The record type is declared by which typed payload the spec carries: exactly one of `a`, `aaaa`, `cname`, `mx`, `srv`, `caa`, `txt`, `ns`, or `ptr`, each shaped the way DNS actually defines that type (MX entries are preference+exchange pairs, SRV entries are priority/weight/port/target, CAA entries are flags/tag/value) — a record can never be declared with a shape its type cannot hold. Foreign key references wire the resource group and zone name to upstream Cloud Resources via ValueFromRef.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DNS Record Set** -- one record set of the payload's type in the target zone, carrying every value listed (multiple A/AAAA addresses round-robin)
- **Alias wiring** (A/AAAA/CNAME only) -- when the payload carries `targetResourceId` instead of literal values, Azure keeps the answer in sync with the referenced resource: a Public IP's address change follows automatically, with no drift window. Alias records also work at the zone apex, where DNS forbids CNAME
- **Azure Tags** -- your governance tags merged over the Planton-derived resource tags (organization, environment, resource id), stored as ARM record-set metadata

**One record set per (name, type)**: Azure stores all values for a (name, type) pair as one record set. A second AzureDnsRecord with the same name and type in the same zone conflicts with this one rather than merging into it — declare all values in one resource.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure DNS Zone** the record joins. Reference an AzureDnsZone Cloud Resource's `zone_name` output via ValueFromRef, or pass the name of a zone managed outside Planton as a literal.
- **The zone's Resource Group** -- Azure addresses record sets by (resource group, zone name, type, record name), so this must be the SAME group the zone lives in.

## Deploy

### Console

Open the deployment store, find **Azure DNS Record**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the record's four steps: the zone context, the nine-way type selector with per-type name guidance, the typed value editor with a live zone-file-style answer preview, and caching/tags. Start from the **Web App A Record** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDnsRecord
metadata:
  name: www-a-record
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-prod-rg"
  zoneName:
    valueFrom:
      kind: AzureDnsZone
      name: example-com-zone
      fieldPath: status.outputs.zone_name
  name: www
  a:
    addresses:
      - value: "203.0.113.10"
```

```shell
planton apply -f dns-record.yaml
```

This creates an A record answering `www.example.com` with one IPv4 address and the platform's default 300-second TTL. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the record joins its zone through the reference graph — the zone deploys first, the record follows:

```yaml
spec:
  zoneName:
    valueFrom:
      kind: AzureDnsZone
      name: example-com-zone
      fieldPath: status.outputs.zone_name
```

Reference-capable values compose further: a TXT payload can reference another resource's deploy-time output (an AzureFrontDoorCustomDomain's `validation_token`), and alias payloads reference ARM-id outputs (an AzurePublicIp's `public_ip_id`).

## Key Configuration

These are the most important decisions when configuring a DNS record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The payload IS the type** -- Set exactly one of the nine payload fields; there is no separate type enum. `a`/`aaaa` carry addresses XOR an alias target, `cname` a hostname XOR an alias target, `mx`/`srv`/`caa` typed entry lists, `txt` reference-capable text values, `ns` child-delegation name servers, `ptr` reverse-DNS hostnames.

**Record name** (`name`) -- Relative to the zone: `@` for the apex, `www` for a host, `*` for wildcards, and underscore-led names for the conventions that live there (`_dmarc`, `_sip._tcp` SRV service names, `_dnsauth.<host>` verification names). Changing the name — or the zone or resource group — replaces the record.

**Alias vs. literal** (A/AAAA/CNAME) -- Literal values are yours to keep current; an alias (`targetResourceId`) delegates the answer to Azure, which tracks the referenced resource with zero drift. The apex escape hatch: DNS forbids CNAME at `@`, but an alias A/AAAA record points the apex at an Azure resource anyway.

**TTL** (`ttlSeconds`) -- How long resolvers cache the answer, which is also how long a change takes to propagate. Azure has no server-side default; Planton applies 300 when unset. 0 is a real value (do not cache). Low (60-300) for records that change during deploys, 3600+ for stable records like MX.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDnsZone** | `zoneName` | `status.outputs.zone_name` |
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzurePublicIp** (alias A records) | `a.targetResourceId` | `status.outputs.public_ip_id` |
| **AzureFrontDoorCustomDomain** (verification TXT) | `txt[]` | `status.outputs.validation_token` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `fqdn` | The fully-qualified name this record answers for (e.g., `www.example.com.`) | Verification tooling, records that chain to this name |
| `record_id` | Azure Resource Manager ID of the record set | Azure Policy assignments, diagnostic settings |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web app A record** -- `www` answering with your app's IPv4 addresses; multiple addresses round-robin. Start from the **Web App A Record** preset.

**Apex alias** -- `@` pointed at an Azure Public IP via an alias A record — the answer follows the IP through every change, where a CNAME would be illegal. Start from the **Apex Alias Record** preset.

**Mail MX records** -- The apex MX set with preference-ordered mail servers, declared as typed preference+exchange pairs in one record set. Start from the **Mail MX Records** preset.

**Domain verification TXT** -- The ownership-proof token custom-domain flows require: Container Apps checks `asuid.<host>`, Front Door checks `_dnsauth.<host>`. Reference the validating resource's token output so the record carries the value minted at deploy time. Start from the **Domain Verification TXT** preset.

## Works With

- [**Azure DNS Zone**](/cloud-catalog/azure-dns-zone) -- the zone this record joins, referenced by `zone_name`
- [**Azure Resource Group**](/cloud-catalog/azure-resource-group) -- the zone's resource group, half of the record's management-plane address
- [**Azure Public IP**](/cloud-catalog/azure-public-ip) -- the classic alias A target: the apex follows the IP automatically
