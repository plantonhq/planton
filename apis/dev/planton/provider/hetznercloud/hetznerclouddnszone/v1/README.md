# HetznerCloudDnsZone

The **HetznerCloudDnsZone** resource provisions a DNS zone in Hetzner Cloud with its associated record sets. The component supports two modes: **primary** (Hetzner Cloud is the authoritative nameserver, records managed via the manifest) and **secondary** (Hetzner Cloud synchronizes records from an external primary nameserver via zone transfer). Record values support cross-component references via `StringValueOrRef`, enabling DNS records that automatically resolve to the IP addresses of other Hetzner Cloud resources.

## What It Represents

A [Hetzner Cloud DNS Zone](https://docs.hetzner.cloud/#dns) is a hosted zone on Hetzner Cloud's authoritative nameservers that maps a domain (e.g., `example.com`) to DNS records. After creating the zone and configuring NS delegation at your domain registrar, Hetzner Cloud serves the zone's records to the public internet.

Hetzner Cloud supports two zone modes:

- **Primary**: Hetzner Cloud is the authoritative nameserver. You manage records directly through record sets in the spec. Hetzner Cloud assigns three authoritative nameservers (e.g., `helium.ns.hetzner.de`, `hydrogen.ns.hetzner.com`, `oxygen.ns.hetzner.com`). You must configure these at your registrar.

- **Secondary**: Hetzner Cloud acts as a slave nameserver, pulling records from one or more external primary nameservers via AXFR/IXFR zone transfer. Records cannot be managed directly in the manifest — they are read-only copies from the primary. Optionally authenticated with TSIG.

The mode and domain name are immutable (ForceNew). Changing either forces replacement of the zone.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_zone` | 1 | Always | The DNS zone itself. Establishes the domain on Hetzner Cloud nameservers. |
| `hcloud_zone_rrset` | 0–N | One per entry in `recordSets` (primary mode only) | A record set for each unique (name, type) pair. Groups all records that share the same name and type. |

In primary mode, a zone with 8 record sets (e.g., A, AAAA, CNAME, MX, TXT, CAA, plus two TXT entries for SPF and DMARC) creates 1 zone + 8 rrset resources. In secondary mode, only the zone resource is created — records are pulled from the primary nameserver.

## Key Features

### Primary and Secondary Modes

The `mode` field controls the zone's operating behavior. Proto-level CEL validation enforces mode-dependent constraints:

- Primary mode: `recordSets` is valid; `primaryNameservers` is forbidden.
- Secondary mode: `primaryNameservers` is required (at least one); `recordSets` is forbidden.

Invalid combinations (e.g., a primary zone with `primaryNameservers`, or a secondary zone with `recordSets`) fail validation before any cloud API call.

### Record Sets with StringValueOrRef

Each record set groups all DNS records sharing the same (name, type) pair. Record values use `StringValueOrRef`, which accepts either a literal string or a `valueFrom` reference to another component's output.

Literal value:
```yaml
records:
  - value: "93.184.216.34"
```

Cross-component reference:
```yaml
records:
  - value:
      valueFrom:
        kind: HetznerCloudServer
        name: web-1
        fieldPath: status.outputs.ipv4_address
```

This enables DNS records that automatically track infrastructure changes — when a server's IP changes, the DNS record updates on the next deployment without manual intervention.

Records also support optional comments for documentation:
```yaml
records:
  - value: "93.184.216.34"
    comment: "primary web server"
```

### TSIG Authentication (Secondary Mode)

Secondary zones can authenticate zone transfers using Transaction Signatures (TSIG). Each primary nameserver entry supports `tsigAlgorithm` (e.g., `hmac-sha256`) and `tsigKey` (the shared secret). TSIG prevents unauthorized zone transfers and is recommended when the primary nameserver is accessible over the public internet.

### Per-Record-Set TTL Override

The zone has a default TTL (`ttl` field, defaults to 3600 seconds). Individual record sets can override this with their own `ttl` value. This is useful for records with different caching needs — for example, a low TTL on A records for failover-capable services and a high TTL on MX records for stable mail servers.

### Delete Protection

The `deleteProtection` field prevents accidental deletion of the zone via the Hetzner Cloud API. When enabled, the zone cannot be deleted until protection is explicitly removed. The IaC modules handle this automatically on destroy operations.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the zone from metadata following the CG01 pattern. User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

## Upstream Dependencies (What This Resource Needs)

This component has no upstream dependencies. It does not reference other Planton components as required inputs. However, record values can optionally reference outputs from other components via `StringValueOrRef.valueFrom`.

## Downstream Dependents (What References This Resource)

No components in the current Hetzner Cloud catalog reference `HetznerCloudDnsZone` outputs as required inputs. The `zone_id` and `nameservers` outputs are informational — the zone ID is used to configure NS delegation at the domain registrar, and the nameservers list tells you which NS records to set.

In infra-chart compositions, DNS zones typically consume outputs from other components (server IPs, floating IPs, load balancer IPs) rather than the other way around.

## Stack Outputs

| Output | Description |
|---|---|
| `zone_id` | The Hetzner Cloud numeric ID of the created zone (as a string). Can be referenced by other components via `StringValueOrRef`. |
| `nameservers` | The authoritative Hetzner nameservers assigned to the zone (e.g., `["helium.ns.hetzner.de", "hydrogen.ns.hetzner.com", "oxygen.ns.hetzner.com"]`). Configure these at your domain registrar to activate the zone. |

## References

- [Hetzner Cloud DNS Documentation](https://docs.hetzner.cloud/#dns)
- [Terraform hcloud_zone Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/zone)
- [Terraform hcloud_zone_rrset Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/zone_rrset)
- [Pulumi hcloud.Zone Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/zone/)
- [Pulumi hcloud.ZoneRrset Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/zonerrset/)
