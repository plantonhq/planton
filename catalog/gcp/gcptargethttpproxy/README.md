# GCP Target HTTP Proxy

Deploys a Compute Engine target HTTP proxy — the plaintext-HTTP frontend adapter of an Application Load Balancer. It binds a forwarding rule (the VIP) to a URL map (the routing brain); the standard production role is serving the http→https redirect on port 80 while the target HTTPS proxy serves the application on 443.

One kind, two scopes: leave `region` empty for the GLOBAL proxy (`google_compute_target_http_proxy` — the global external ALB, the cross-region internal ALB, Traffic Director), or set it for the REGIONAL proxy (`google_compute_region_target_http_proxy` — the regional external ALB and the regional internal ALB). The same manifest shape serves both; every link in a regional chain must be regional in the same region.

## What Gets Created

A single target HTTP proxy in the chosen project — global, or regional when `region` is set. Forwarding rules reference its `self_link`; the proxy references a URL map in the same scope.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **A URL map** — a `GcpUrlMap` (or its self-link) for the proxy to route through
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `http-proxy.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpProxy
metadata:
  name: web-http-frontend
spec:
  projectId:
    value: my-gcp-project-123
  urlMap:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/urlMaps/http-redirect
```

Deploy:

```shell
planton apply -f http-proxy.yaml
```

This creates a proxy ready for a port-80 global forwarding rule to bind.

## Configuration Reference

| Field | Description |
|-------|-------------|
| `projectId` | Project owning the proxy (literal or `GcpProject` reference); empty uses the provider's default project |
| `proxyName` | Cloud-side name (RFC1035); defaults to `metadata.name`. Immutable |
| `description` | What this proxy fronts. Immutable |
| `region` | Empty for a global proxy; a region name (`us-central1`) for a regional one, whose URL map and forwarding rule must be regional in the same region. Immutable |
| `urlMap` | The URL map to route through (required; `GcpUrlMap` reference or self-link; a regional proxy takes a regional URL map). **Mutable in place** — repointing a live frontend causes no downtime |
| `httpKeepAliveTimeoutSec` | Idle client keep-alive, 5-1200s; only honored by `EXTERNAL_MANAGED` load balancers (GCP default 610). Immutable |
| `proxyBind` | Bind to Traffic Director mesh VIPs instead of Google's edge (`INTERNAL_SELF_MANAGED` only). Global proxies only — rejected when `region` is set. Immutable |
| `deletionPolicy` | What destroy does: `DELETE` (default) removes the proxy, `PREVENT` fails the destroy to protect a production frontend, `ABANDON` leaves it serving unmanaged |

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

- **Only `urlMap` is mutable** — every other field forces destroy-and-recreate, briefly breaking any forwarding rule referencing the old `self_link`.
- **Scope is immutable and chain-wide** — a proxy cannot move between global and regional, and a regional proxy's URL map (and the forwarding rule in front of it) must be regional in the same region. The regional proxy has no `proxyBind`; the spec rejects it when `region` is set.
- **The pair pattern**: point this proxy at a redirect-only URL map (`defaultUrlRedirect` with `httpsRedirect: true`) and let a `GcpTargetHttpsProxy` serve the real application — two forwarding rules share one static IP on ports 80 and 443.
- **TLS never lives here** — for HTTPS termination use `GcpTargetHttpsProxy`.

## Related Components

- [GcpUrlMap](/docs/catalog/gcp/gcpurlmap) — the routing table this proxy consults
- [GcpGlobalForwardingRule](/docs/catalog/gcp/gcpglobalforwardingrule) — the VIP that binds to this proxy
- [GcpTargetHttpsProxy](/docs/catalog/gcp/gcptargethttpsproxy) — the TLS-terminating sibling
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the proxy

## Additional Resources

- [Target proxies overview](https://cloud.google.com/load-balancing/docs/target-proxies)
- [Setting up HTTP-to-HTTPS redirect](https://cloud.google.com/load-balancing/docs/https/setting-up-http-https-redirect)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
