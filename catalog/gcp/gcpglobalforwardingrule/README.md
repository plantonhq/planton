# GCP Global Forwarding Rule

Deploys a Compute Engine forwarding rule — the VIP node of a load balancer. The forwarding rule is where traffic enters: it binds an IP address and port to a target proxy, or — for the passthrough Network Load Balancers — straight to a backend service. It is also the entry point for Private Service Connect, forwarding a VPC's traffic privately to Google APIs or a producer's service attachment.

One kind, two scopes. The kind is named for the GLOBAL forwarding rule it began as: leave `region` empty for `google_compute_global_forwarding_rule` (the global external ALB, the cross-region internal ALB, Traffic Director, PSC to Google APIs), or set it for the REGIONAL `google_compute_forwarding_rule` — the front door of the regional external and internal Application Load Balancers (`target` = a regional proxy), of the internal and external passthrough Network Load Balancers (`backendService` instead of `target`), and of a PSC consumer endpoint (`target` = a service attachment). The regional rule adds `ports`, `allPorts`, `allowGlobalAccess`, `allowPscGlobalAccess`, `serviceLabel`, `isMirroringCollector`, `ipCollection`, `recreateClosedPsc`, `sourceIpRanges`, the `L3_DEFAULT` protocol, the `INTERNAL` scheme, and the `STANDARD` tier; the global rule alone carries `metadataFilters`, the `INTERNAL_SELF_MANAGED` scheme, and the backend-bucket migration canary.

## What Gets Created

A single forwarding rule in the chosen project — global, or regional when `region` is set — the resource DNS records point at (via its `ip_address` output).

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **A target proxy** — a `GcpTargetHttpsProxy` or `GcpTargetHttpProxy` (or another global target's URI)
- **Recommended: a reserved static IP** — a `GcpGlobalAddress`, so the VIP survives frontend rebuilds

## Quick Start

Create a file `forwarding-rule.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGlobalForwardingRule
metadata:
  name: web-frontend-443
spec:
  projectId:
    value: my-gcp-project-123
  target:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/targetHttpsProxies/web-https-frontend
  portRange: "443"
```

Deploy:

```shell
planton apply -f forwarding-rule.yaml
```

This creates a port-443 frontend with a Google-assigned ephemeral IP; add `ipAddress` referencing a `GcpGlobalAddress` for a stable production VIP.

## Configuration Reference

### The frontend

| Field | Description |
|-------|-------------|
| `region` | Empty for a global rule; a region name (`us-central1`) for a regional one, whose target, backend service, and address must be regional in the same region. Immutable |
| `target` | The proxy receiving traffic (`GcpTargetHttpsProxy` ref by default; HTTP proxy self-link; PSC bundles `all-apis`/`vpc-sc` on a global rule; service attachment URIs on a regional rule). Exactly one of `target` and `backendService`. **Mutable in place** — the zero-downtime frontend swap |
| `backendService` | The regional `GcpBackendService` a passthrough Network Load Balancer sends traffic to directly (scheme `INTERNAL` or `EXTERNAL`, no proxy). Regional rules only; exactly one of `target` and `backendService`. Immutable |
| `ipAddress` | The VIP: a `GcpGlobalAddress` ref (its reserved IP), a regional `GcpAddress` ref with an explicit `valueFrom.kind` for a regional rule, a literal IP, or empty for an ephemeral IP. Immutable |
| `portRange` | Port or contiguous range (`"443"`, `"8080-8090"`); non-overlapping ranges are how two rules share one IP. At most one of `portRange`, `ports`, `allPorts`. Immutable |
| `ports` | Up to five individual ports or ranges for the passthrough Network Load Balancers (TCP/UDP/SCTP). Regional rules only. Immutable |
| `allPorts` | Forward every port (and port-less packets) — passthrough NLBs and protocol forwarding; required with `L3_DEFAULT`. Regional rules only. Immutable |
| `ipProtocol` | Default `TCP` (what every proxy-based LB and PSC uses); `UDP`, `ESP`, `AH`, `SCTP`, `ICMP` for passthrough; `L3_DEFAULT` (every protocol, regional only, with `allPorts`). Immutable |
| `ipVersion` | `IPV4` (default) or `IPV6` for auto-assigned ephemeral IPs. Immutable |
| `loadBalancingScheme` | `EXTERNAL` (default on both scopes), `EXTERNAL_MANAGED`, `INTERNAL_MANAGED`, `INTERNAL` (internal passthrough NLB, regional only), `INTERNAL_SELF_MANAGED` (Traffic Director, global only), or `NONE` (Private Service Connect). Immutable |

### Network wiring (internal schemes + PSC)

| Field | Description |
|-------|-------------|
| `network` | VPC ref — required for PSC, used by internal schemes and by the regional external ALB (`EXTERNAL_MANAGED` with `region`); rejected on the global external load balancers and the external passthrough NLB (CEL) |
| `subnetwork` | Subnetwork ref for internal load balancing (required on a custom-mode network) and IPv6 external passthrough |
| `networkTier` | `PREMIUM` (default) on both scopes; `STANDARD` on regional rules only (CEL) |
| `allowGlobalAccess` | Let clients in every region reach an internal passthrough NLB. Regional `INTERNAL` rules only. Mutable |
| `serviceLabel` | DNS label prefix giving an internal passthrough NLB a stable internal name (the `service_name` output). Regional `INTERNAL` rules only |
| `isMirroringCollector` | Mark an internal passthrough NLB as a Packet Mirroring collector. Regional `INTERNAL` rules only |
| `sourceIpRanges` | Up to 64 source IPs/CIDRs the rule forwards from — the external passthrough NLB's allowlist. Regional `EXTERNAL` rules only |
| `ipCollection` | BYOIP: the PublicDelegatedPrefix an IPv6 external passthrough NLB draws its address from. Regional rules only |

### Traffic Director / PSC extras

| Field | Description |
|-------|-------------|
| `metadataFilters` | xDS client scoping — `INTERNAL_SELF_MANAGED` (global) only (CEL) |
| `serviceDirectoryRegistration` | Register a PSC endpoint in Service Directory — scheme `NONE` only (CEL); `serviceDirectoryRegion` on a global rule, `service` on a regional one |
| `noAutomateDnsZone` | Skip the auto-created PSC DNS zone — scheme `NONE` only (CEL) |
| `allowPscGlobalAccess` | Let clients in every region reach a PSC consumer endpoint. Regional `NONE` rules only. Mutable |
| `recreateClosedPsc` | Recreate a PSC consumer endpoint whose connection Google reports CLOSED. Regional `NONE` rules only |

### Lifecycle

| Field | Description |
|-------|-------------|
| `labels` | Organize/bill the rule. Mutable |
| `externalManagedBackendBucketMigrationState` / `...TestingPercentage` | The EXTERNAL → EXTERNAL_MANAGED backend-bucket canary migration, without recreating the VIP. Global rules only |
| `deletionPolicy` | What destroy does: `DELETE` (default), `PREVENT` (refuse), or `ABANDON` (keep serving, drop from management) |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `ip_address` | `string` | The VIP — the value DNS records point at (always the literal IP) |
| `self_link` | `string` | Self-link URI of the rule (`regions/{region}` in place of `global` for a regional rule) |
| `forwarding_rule_name` | `string` | Name of the rule in GCP |
| `forwarding_rule_id` | `string` | Server-assigned numeric ID |
| `psc_connection_id` | `string` | PSC connection id (PSC frontends only) |
| `region` | `string` | Region of a regional rule; empty for a global one |
| `service_name` | `string` | Internal DNS name of an internal passthrough NLB that set `serviceLabel`; empty otherwise |
| `psc_connection_status` | `string` | PSC connection status — `ACCEPTED` means the producer admitted the connection |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Only `target` and `labels` mutate in place.** Everything else recreates the rule — and an ephemeral VIP changes on recreate, which is why production frontends reference a reserved `GcpGlobalAddress`.
- **The pair pattern**: a port-80 rule (→ `GcpTargetHttpProxy` serving a redirect map) and a port-443 rule (→ `GcpTargetHttpsProxy`) share one static IP.
- **PSC naming**: Private Service Connect rules for Google APIs are limited to 20-character letter/digit names.
- **Scope is immutable and chain-wide**: a rule cannot move between global and regional; a regional rule's proxy, backend service, and address must be regional in the same region, and a regional external ALB needs a proxy-only subnet (`GcpSubnetwork` with `purpose: REGIONAL_MANAGED_PROXY`) in the region before the rule can be created.
- **An unset scheme is `EXTERNAL` on both scopes** — both engines send it explicitly, so a manifest means the same thing wherever it lives; set `INTERNAL` outright for an internal passthrough NLB.

## Related Components

- [GcpTargetHttpsProxy](/docs/catalog/gcp/gcptargethttpsproxy) — the default target
- [GcpTargetHttpProxy](/docs/catalog/gcp/gcptargethttpproxy) — the port-80 redirect target
- [GcpGlobalAddress](/docs/catalog/gcp/gcpglobaladdress) — the reserved static VIP
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork) — the network for internal/PSC frontends
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the rule

## Additional Resources

- [Forwarding rule concepts](https://cloud.google.com/load-balancing/docs/forwarding-rule-concepts)
- [Private Service Connect for Google APIs](https://cloud.google.com/vpc/docs/configure-private-service-connect-apis)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
