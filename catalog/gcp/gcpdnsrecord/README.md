# GcpDnsRecord

## Overview

`GcpDnsRecord` manages one DNS record set inside a Google Cloud DNS managed
zone. A record set is the unit Cloud DNS manages: one (name, type) pair that
answers queries either with a static list of values (round-robin) or with
exactly one routing policy — weighted round robin, geolocation, or
primary/backup failover with health-checked targets.

## Purpose

- **Declarative records**: define DNS records as code, with the zone owned
  separately by `GcpDnsZone` — records and zones have independent lifecycles
  and owners.
- **All record types**: `type` accepts any record type the Cloud DNS API
  supports (A, AAAA, CNAME, MX, TXT, SRV, NS, PTR, CAA, SOA, HTTPS, SVCB,
  DS, DNSKEY, TLSA, SSHFP, NAPTR, ...), so new types need no kind
  change.
- **Traffic steering**: weighted round robin for canary rollouts, geo
  routing for latency-sensitive multi-region serving, and primary/backup
  failover with health-checked internal load balancers or external
  endpoints.
- **Composable references**: the zone, project, health check, internal load
  balancer VIPs (GcpAddress), and networks (GcpVpcNetwork) are all
  referenceable resources, not hardcoded strings.

## Key Features

- Static values (round-robin across multiple values) OR one routing policy
  per record — the same exclusive contract the Cloud DNS API enforces,
  validated before deploy.
- Weighted round robin with zero-weight staging entries.
- Geolocation routing with optional fencing (unhealthy locations keep
  answering instead of spilling to the next-closest location).
- Primary/backup failover with a trickle ratio to keep backup paths warm.
- Health-checked targets: internal load balancer frontends (implicit health
  signal) or public endpoints (via a referenced GcpHealthCheck).
- Wildcard (`*.example.com.`) and underscore service labels
  (`_dmarc.example.com.`) supported.
- Deletion policy: `DELETE` (default), `PREVENT` (destroy refuses), or
  `ABANDON` (the record keeps answering but leaves management).

## Example Usage

### Basic A Record

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: www-example-com
spec:
  managedZone:
    value: example-zone
  type: A
  name: www.example.com.
  values:
    - value: 192.0.2.1
```

`projectId` may be omitted — the record is then created in the provider's
default project (ambient credentials decide).

### Weighted Canary Rollout

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDnsRecord
metadata:
  name: api-canary
spec:
  managedZone:
    valueFrom:
      kind: GcpDnsZone
      name: example-com
      fieldPath: status.outputs.zone_name
  type: A
  name: api.example.com.
  routingPolicy:
    wrr:
      - weight: 95
        values: ["203.0.113.10"]
      - weight: 5
        values: ["203.0.113.20"]
```

### Deploy with CLI

```bash
planton pulumi up --manifest dns-record.yaml
# or
planton tofu apply --manifest dns-record.yaml
```

## Best Practices

1. **Always use FQDNs**: record names end with a trailing dot
   (`www.example.com.`).
2. **TTL matches change cadence**: 60–300s for records that move during
   failover or canaries; 3600+ for stable infrastructure records.
3. **Keep one owner per (name, type) pair**: Cloud DNS stores one record
   set per pair — two resources managing the same pair will fight.
4. **Prefer references over literals**: point `managedZone` at the
   GcpDnsZone resource and internal load balancer targets at GcpAddress /
   GcpVpcNetwork resources so renames and rebuilds propagate.

## Related Kinds

- **GcpDnsZone**: the managed zone this record lives in (zone shell only —
  records belong here).
- **GcpHealthCheck**: health checks for routing policies with external
  endpoints.
- **GcpAddress** / **GcpVpcNetwork**: the VIPs and networks referenced by
  health-checked internal load balancer targets.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
