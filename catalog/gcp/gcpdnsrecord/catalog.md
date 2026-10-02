# GCP DNS Record

Deploys a DNS record set in an existing Google Cloud DNS Managed Zone — one (name, type) pair answering either with static values (round-robin) or with exactly one routing policy that steers each query by weight, caller geography, or target health. Any record type the Cloud DNS API supports passes through (A, AAAA, CNAME, MX, TXT, SRV, NS, CAA, and newer types like HTTPS and SVCB). The name and values are reference-capable, so certificate-validation records compose directly from a GcpCertManagerDnsAuthorization's outputs.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Cloud DNS API enablement** -- `dns.googleapis.com` is enabled in the target project (never disabled on destroy, so tearing down one record cannot break the rest of the project)
- **Cloud DNS Record Set** -- a record set in the specified Managed Zone, configured with the chosen record type, fully qualified domain name, and TTL
- **Static values (round-robin)** -- when `values` are provided, a multi-value record set answering every resolver identically
- **Routing policy** -- when `routingPolicy` is set instead, weighted round-robin, geolocation, or primary/backup failover steering, optionally health-checking internal load balancer frontends and external endpoints

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** containing the Cloud DNS Managed Zone where the record will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef; the module enables the Cloud DNS API itself.
- **An existing Cloud DNS Managed Zone** for the target domain. Provide the zone's RESOURCE name (e.g. `prod-example-zone`, not the DNS name) directly or reference a GcpDnsZone Cloud Resource via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP DNS Record**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **A Record** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: api-a-record
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  managedZone:
    value: "example-zone"
  type: A
  name:
    value: "api.example.com."
  values:
    - value: "34.120.0.1"
```

```shell
planton apply -f dns-record.yaml
```

This creates an A record pointing `api.example.com.` to `34.120.0.1` with the default 300-second TTL. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the DNS record to a GCP project and DNS zone deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  managedZone:
    valueFrom:
      kind: GcpDnsZone
      name: example-zone
      fieldPath: status.outputs.zone_name
```

The InfraPipeline resolves the dependency graph, deploys the project and DNS zone first, then provisions the DNS record with the resolved values.

## Key Configuration

These are the most important decisions when configuring a DNS record. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Record type** -- Uppercase, e.g. `A`, `CNAME`, `TXT` via the `type` field. Cloud DNS accepts any type its API supports — the string passes through as-is, so newer types (HTTPS, SVCB, TLSA) need no platform change. The type determines the format of values: IPv4 addresses for A records, hostnames with trailing dots for CNAME records, priority-prefixed servers for MX records.

**Fully qualified domain name** -- The `name` field is a `StringValueOrRef`: write a literal as `name: {value: "www.example.com."}` — it must end with a trailing dot — or wire `valueFrom` to any kind's output (the field names no default kind). Wildcard records use a `*.` prefix; leading underscores support service labels like `_dmarc` and `_acme-challenge`. Point it at a GcpCertManagerDnsAuthorization's `dns_record_name` output to compose certificate-validation records with zero hand-copying.

**Static values XOR routing policy** -- Exactly one. Each `values` entry is also a `StringValueOrRef` (`- value: "34.120.0.1"` or a `valueFrom` block to any kind's output, e.g. a GcpGlobalAddress's `address`, a GcpDnsZone's `nameservers.0` for a delegation, or an authorization's `dns_record_data`); multiple entries answer as a round-robin set. `routingPolicy` instead steers each query: `wrr` splits by weight (a weight of 0 stages an entry — the DNS canary lever), `geo` answers with the entry nearest the caller, `primaryBackup` serves primaries while healthy and falls back to regional backups (with an optional `trickleRatio` keeping the backup path warm). Health-checked internal load balancer targets carry their own health signal; external endpoints need the policy-level `healthCheck` reference.

**TTL** -- `ttlSeconds` defaults to 300 seconds when omitted; an explicit 0 disables caching. Use lower values (60s) during migrations and for failover policies, and higher values (3600s) for stable production records to reduce DNS query load.

**Deletion policy** -- destroying a record stops answers for its name as resolver caches expire — often long after the apply, and often for systems nobody remembers (mail routing, domain verification, service discovery). Set `deletionPolicy: PREVENT` on names other systems resolve; `ABANDON` removes the record from management while it keeps answering in GCP.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpDnsZone** | `managedZone` | `status.outputs.zone_name` |
| **GcpCertManagerDnsAuthorization** (optional) | `name` / `values[]` | `status.outputs.dns_record_name` / `status.outputs.dns_record_data` |
| **GcpHealthCheck** (external endpoints only) | `routingPolicy.healthCheck` | `status.outputs.self_link` |
| **GcpAddress** (per ILB target) | `routingPolicy.*.healthCheckedTargets.internalLoadBalancers[].ipAddress` | `status.outputs.address` |
| **GcpVpcNetwork** (per ILB target) | `routingPolicy.*.healthCheckedTargets.internalLoadBalancers[].networkUrl` | `status.outputs.network_self_link` |
| **GcpProject** (per ILB target) | `routingPolicy.*.healthCheckedTargets.internalLoadBalancers[].project` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `fqdn` | Fully qualified domain name of the created record | Application configuration, monitoring targets, another record's CNAME value |

The remaining outputs — `record_type`, `managed_zone`, `project_id`, `ttl_seconds` — echo the record's resolved inputs back for verification and carry no downstream composition story.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**A record** -- Maps a hostname to an IPv4 address with a 5-minute TTL. The most common DNS record type for pointing domains to load balancers, VMs, or external services. Start from the **A Record** preset.

**CNAME record** -- Aliases one hostname to another, allowing the target IP to change without updating the record. Use for `www` subdomains, CDN frontends, or SaaS provider endpoints. Start from the **CNAME Record** preset.

**Weighted canary** -- A `wrr` routing policy splitting answers 95/5 between stable and canary targets; shift by editing weights, stage new entries at weight 0. Start from the **Weighted Canary** preset.

**Certificate validation** -- A CNAME whose `name` and `values` reference a GcpCertManagerDnsAuthorization's outputs — the three-node composition (authorization → this record → GcpCertManagerCert) validates certificates before any load balancer serves the domain.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project containing the DNS Managed Zone
- [**GCP DNS Zone**](/cloud-catalog/gcp-dns-zone) -- provides the Managed Zone where the record is created
- [**GCP Cert Manager DNS Authorization**](/cloud-catalog/gcp-cert-manager-dns-authorization) -- exports the validation CNAME this record serves
- [**GCP Health Check**](/cloud-catalog/gcp-health-check) -- probes external endpoints for routing-policy health withdrawal
- [**GCP Address**](/cloud-catalog/gcp-address) -- the reserved internal VIP a health-checked load-balancer target answers on
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the network an internal load-balancer target belongs to
