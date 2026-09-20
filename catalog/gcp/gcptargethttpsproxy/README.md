# GCP Target HTTPS Proxy

Deploys a Compute Engine target HTTPS proxy — the TLS-termination node of an Application Load Balancer. It binds a forwarding rule (the VIP) to a URL map (the routing brain) and owns the client-facing handshake: certificates, SSL policy, QUIC (HTTP/3), and TLS 1.3 early data.

One kind, two scopes: leave `region` empty for the GLOBAL proxy (`google_compute_target_https_proxy` — the global external ALB, the cross-region internal ALB, Traffic Director), or set it for the REGIONAL proxy (`google_compute_region_target_https_proxy` — the regional external ALB and the regional internal ALB). A regional proxy's URL map, certificates, and SSL policy must all be regional in the same region; the certificate map, QUIC, early-data, and mesh-bind levers exist only on the global proxy.

## What Gets Created

A single target HTTPS proxy in the chosen project — global, or regional when `region` is set. Forwarding rules reference its `self_link`; the proxy references a URL map and its certificate source in the same scope.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **A URL map** — a `GcpUrlMap` (or its self-link) for the proxy to route through
- **A certificate source** — typically one or more `GcpManagedSslCertificate` resources
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `https-proxy.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpsProxy
metadata:
  name: web-https-frontend
spec:
  projectId:
    value: my-gcp-project-123
  urlMap:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/web-routing
  sslCertificates:
    - value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/sslCertificates/web-cert
```

Deploy:

```shell
planton apply -f https-proxy.yaml
```

This creates a proxy ready for a port-443 global forwarding rule to bind.

## Configuration Reference

### Certificate source (choose exactly one)

| Field | Description |
|-------|-------------|
| `sslCertificates` | 1-15 classic compute SSL certificates (`GcpManagedSslCertificate` refs or self-links; on a regional proxy, regional `GcpSslCertificate` refs with an explicit `valueFrom.kind`). Rotation swaps the list in place |
| `certificateManagerCertificates` | Certificate Manager certificates — cross-region internal ALB (`INTERNAL_MANAGED`) on a global proxy, or regional certificates on a regional proxy |
| `certificateMap` | Certificate Manager map selecting the cert by SNI hostname — global external ALBs only (rejected when `region` is set); the beyond-15-domains mechanism |

Traffic Director proxies skip certificates and use `serverTlsPolicy` instead.

### TLS behavior

| Field | Description |
|-------|-------------|
| `sslPolicy` | SSL policy self-link constraining TLS versions/ciphers; empty keeps GCP's permissive default (min TLS 1.0). Mutable |
| `serverTlsPolicy` | Network security ServerTlsPolicy — the mTLS lever (demand + validate client certificates). Mutable and clearable |
| `quicOverride` | `NONE` (GCP decides, the default), `ENABLE`, or `DISABLE`. Global proxies only. Mutable |
| `tlsEarlyData` | TLS 1.3 0-RTT: `STRICT` / `PERMISSIVE` / `UNRESTRICTED` / `DISABLED` (GCP default). Global proxies only. Immutable |

### Wiring

| Field | Description |
|-------|-------------|
| `urlMap` | Required. The URL map to route decrypted requests through. Mutable in place |
| `projectId` | Project owning the proxy; empty uses the provider's default project. Immutable |
| `proxyName` | Cloud-side name (RFC1035); defaults to `metadata.name`. Immutable |
| `region` | Empty for a global proxy; a region name (`us-central1`) for a regional one, whose URL map, certificates, and SSL policy must be regional in the same region. Immutable |
| `httpKeepAliveTimeoutSec` | Idle client keep-alive, 5-1200s; `EXTERNAL_MANAGED` only. Immutable |
| `proxyBind` | Bind to Traffic Director mesh VIPs (`INTERNAL_SELF_MANAGED` only). Global proxies only. Immutable |
| `deletionPolicy` | What destroy does: `DELETE` (default) removes the proxy, `PREVENT` fails the destroy to protect a production TLS frontend, `ABANDON` leaves it serving unmanaged |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value a forwarding rule references as its target (`regions/{region}` in place of `global` for a regional proxy) |
| `proxy_name` | `string` | Name of the proxy in GCP |
| `proxy_id` | `string` | Server-assigned numeric ID |
| `fingerprint` | `string` | Fingerprint for optimistic concurrency control (empty for a regional proxy — the regional API has none) |
| `region` | `string` | Region of a regional proxy; empty for a global one |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Zero-downtime certificate rotation**: attach the replacement certificate to `sslCertificates` first, wait for it to go ACTIVE, then remove the old one — the proxy swaps the list in place (`setSslCertificates`).
- **A still-PROVISIONING managed certificate attaches fine** — attachment is in fact required before the certificate can activate (DNS must point at the load balancer's IP).
- **Certificate mechanisms are mutually exclusive** — the spec rejects combinations pre-deploy; GCP would reject them at the API.
- **Scope is immutable and chain-wide** — a proxy cannot move between global and regional; a regional proxy's URL map, certificates, SSL policy, and forwarding rule must be regional in the same region, and `certificateMap`, `quicOverride`, `tlsEarlyData`, and `proxyBind` are rejected when `region` is set (the regional resource has none of them).
- **`urlMap`, certificates, `sslPolicy`, `serverTlsPolicy`, and `quicOverride` update in place**; name, description, keep-alive, `tlsEarlyData`, and `proxyBind` are ForceNew.

## Related Components

- [GcpManagedSslCertificate](/docs/catalog/gcp/gcpmanagedsslcertificate) — Google-managed certificates for `sslCertificates`
- [GcpCertManagerCert](/docs/catalog/gcp/gcpcertmanagercert) — Certificate Manager certificates for the internal-ALB list
- [GcpUrlMap](/docs/catalog/gcp/gcpurlmap) — the routing table this proxy consults
- [GcpGlobalForwardingRule](/docs/catalog/gcp/gcpglobalforwardingrule) — the VIP that binds to this proxy
- [GcpTargetHttpProxy](/docs/catalog/gcp/gcptargethttpproxy) — the port-80 redirect sibling
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the proxy

## Additional Resources

- [Target proxies overview](https://cloud.google.com/load-balancing/docs/target-proxies)
- [SSL certificates overview](https://cloud.google.com/load-balancing/docs/ssl-certificates)
- [Certificate Manager certificate maps](https://cloud.google.com/certificate-manager/docs/certificate-maps)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
