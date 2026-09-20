# GCP Target HTTP Proxy

Deploys a Compute Engine target HTTP proxy — the plaintext-HTTP frontend adapter of an Application Load Balancer, global (the default) or regional when `region` is set. The proxy binds a forwarding rule (the VIP) to a URL map (the routing brain): the rule delivers client connections, the proxy consults the map for every request. It is deliberately thin — TLS lives on the target HTTPS proxy, routing on the URL map, traffic policy on the backend service. The standard production pattern is a PAIR sharing one static IP: this proxy serves a redirect-only URL map (http→https 301) while the HTTPS proxy serves the application.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine Target HTTP Proxy** -- global, or regional when `region` is set; bound to the configured URL map, with optional keep-alive tuning and (global only) Traffic Director bind
- **Compute Engine API enablement** -- `compute.googleapis.com` is enabled in the target project; tearing down the proxy never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the proxy will be created. The module enables the Compute Engine API itself, so the connection's principal needs permission to enable services on a fresh project.
- **A URL map** (GcpUrlMap) — the proxy's one required dependency; for the redirect pattern, a redirect-only map.

## Deploy

### Console

Open the deployment store, find **GCP Target HTTP Proxy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **HTTPS Redirect Frontend** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTargetHttpProxy
metadata:
  name: web-http-proxy
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  urlMap:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/urlMaps/http-redirect-map"
```

```shell
planton apply -f target-http-proxy.yaml
```

This creates the redirect half: a port-80 forwarding rule pointing here upgrades every request to HTTPS. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the URL map:

```yaml
spec:
  urlMap:
    valueFrom:
      kind: GcpUrlMap
      name: http-redirect-map
      fieldPath: status.outputs.self_link
```

The InfraPipeline resolves the dependency graph — the URL map first, then this proxy — and a downstream GcpGlobalForwardingRule references this proxy's `self_link` as its target.

## Key Configuration

These are the most important decisions when configuring a target HTTP proxy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**URL map** -- The one REQUIRED field, and the ONLY mutable one: GCP swaps it in place (a dedicated setUrlMap call), so repointing a live frontend at a new routing table causes zero downtime — the blue/green lever for whole routing schemes.

**Scope** -- `region` empty builds the global proxy (global external ALB, cross-region internal ALB, Traffic Director); a region name builds the regional proxy (regional external and internal ALBs), whose URL map and forwarding rule must be regional in the same region. Immutable.

**Keep-alive timeout** -- 5-1200 seconds; only honored by the envoy-based EXTERNAL_MANAGED scheme (GCP default 610s). Raise it above your clients' own keep-alive so the LB never closes first. Immutable.

**Traffic Director bind** -- `proxyBind` attaches the proxy to the mesh's private IPs instead of Google's edge; only meaningful behind an INTERNAL_SELF_MANAGED forwarding rule. Immutable.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpUrlMap** | `urlMap` | `status.outputs.self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the proxy | GcpGlobalForwardingRule `target` |
| `proxy_name` | Name as it exists in GCP | Audit, fleet inventory |
| `proxy_id` | Server-assigned numeric ID | Diagnostics |
| `fingerprint` | Optimistic-concurrency token (empty for a regional proxy) | Out-of-band gcloud updates |
| `region` | Region of a regional proxy; empty for global | Scope checks on downstream blocks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTPS redirect frontend** -- This proxy + a redirect-only URL map: the port-80 half of every production frontend. Start from the **HTTPS Redirect Frontend** preset.

**Plain HTTP frontend** -- Serving an application over plain HTTP (internal tools, pre-TLS testing). Start from the **Plain HTTP Frontend** preset.

**Traffic Director mesh** -- The proxy bound to mesh-private IPs for INTERNAL_SELF_MANAGED frontends. Start from the **Traffic Director Mesh Frontend** preset.

**Regional HTTP frontend** -- The regional proxy of a regional external Application Load Balancer, pointing at a regional URL map. Start from the **Regional HTTP Frontend** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the proxy is created
- [**GCP URL Map**](/cloud-catalog/gcp-url-map) -- the routing table this proxy consults
- [**GCP Global Forwarding Rule**](/cloud-catalog/gcp-global-forwarding-rule) -- consumes this proxy's `self_link` as its target
- [**GCP Target HTTPS Proxy**](/cloud-catalog/gcp-target-https-proxy) -- the TLS sibling serving the application half
