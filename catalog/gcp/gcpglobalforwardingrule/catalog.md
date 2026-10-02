# GCP Global Forwarding Rule

Deploys a Compute Engine forwarding rule — the VIP node of a load balancer, global (the default) or regional when `region` is set. It binds an IP address and port to a target proxy (HTTP or HTTPS) — or, for the passthrough Network Load Balancers, straight to a regional backend service — which is where client traffic enters. With the load-balancing scheme set to `NONE`, the same resource becomes a Private Service Connect entry point for Google APIs (`all-apis` / `vpc-sc`) or a producer service attachment. The target is mutable in place — repointing a live VIP at a new proxy is GCP's zero-downtime frontend swap — while the IP, protocol, port range, and scheme are immutable.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine Forwarding Rule** -- global, or regional when `region` is set; bound to the configured target (or backend service), IP address, protocol, ports, and load-balancing scheme
- **Compute Engine API enablement** -- `compute.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the forwarding rule will be created.
- **A target proxy** (GcpTargetHttpsProxy or GcpTargetHttpProxy) for application frontends, or a reserved address for production VIPs.
- For PSC: a VPC network and an internal address reserved for the frontend.

## Deploy

### Console

Open the deployment store, find **GCP Global Forwarding Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **HTTPS Frontend VIP** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGlobalForwardingRule
metadata:
  name: https-frontend
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  target:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/targetHttpsProxies/web-https-proxy"
  ipAddress:
    value: "34.120.1.2"
  portRange: "443"
  loadBalancingScheme: EXTERNAL_MANAGED
```

```shell
planton apply -f global-forwarding-rule.yaml
```

This creates the serving half of a production frontend: a reserved static IP on port 443 pointing at the HTTPS proxy. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the chain:

```yaml
spec:
  target:
    valueFrom:
      kind: GcpTargetHttpsProxy
      name: web-https-proxy
      fieldPath: status.outputs.self_link
  ipAddress:
    valueFrom:
      kind: GcpGlobalAddress
      name: web-vip
      fieldPath: status.outputs.address
```

The InfraPipeline resolves the dependency graph — address and proxy first, then this forwarding rule. Pair with a port-80 rule on the same IP pointing at a GcpTargetHttpProxy for the http→https redirect half.

## Key Configuration

These are the most important decisions when configuring a global forwarding rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- `region` empty builds the global rule (global external ALB, cross-region internal ALB, Traffic Director, PSC to Google APIs); a region name builds the regional rule (regional external and internal ALBs, internal and external passthrough Network Load Balancers, PSC consumer endpoints), whose target, backend service, and address must be regional in the same region. Immutable.

**Load balancer family** -- `EXTERNAL` / `EXTERNAL_MANAGED` (internet edge), `INTERNAL_*` (VPC-facing), or `NONE` (PSC). Immutable and the controlling fork: it decides whether network, PSC, and Traffic Director filter steps apply.

**Target** -- Required; mutable in place (setTarget) — repointing a live VIP at a new proxy causes zero downtime. Defaults to GcpTargetHttpsProxy; use an explicit kind for HTTP proxies, or type `all-apis` / `vpc-sc` for PSC Google APIs.

**IP address** -- Prefer a reserved GcpGlobalAddress so DNS never chases an ephemeral VIP. Required for PSC.

**Port / protocol** -- Port 443 + TCP for HTTPS; port 80 for the redirect half. Blank protocol records no choice (GCP's TCP default).

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpTargetHttpsProxy** | `target` | `status.outputs.self_link` |
| **GcpTargetHttpProxy** | `target` (explicit kind) | `status.outputs.self_link` |
| **GcpGlobalAddress** | `ipAddress` | `status.outputs.address` |
| **GcpVpcNetwork** | `network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** | `subnetwork` | `status.outputs.subnetwork_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ip_address` | The VIP literal DNS points at | DNS A/AAAA records, inventory |
| `self_link` | Self-link URI of the forwarding rule | Audit, reverse references |
| `forwarding_rule_name` | Name as it exists in GCP | gcloud commands, log filters |
| `psc_connection_id` | PSC connection id (scheme NONE only) | PSC diagnostics |
| `psc_connection_status` | PENDING / ACCEPTED / REJECTED / CLOSED | PSC readiness |
| `region` | Region of a regional rule; empty for global | Scope checks on downstream blocks |
| `service_name` | Internal DNS name of an internal passthrough NLB with `serviceLabel` | Client configuration inside the VPC |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTPS frontend** -- Port 443 → HTTPS proxy on EXTERNAL_MANAGED. Start from the **HTTPS Frontend VIP** preset.

**HTTP redirect frontend** -- Port 80 → HTTP proxy (redirect URL map) sharing the same reserved IP. Start from the **HTTP Redirect VIP (Shared IP)** preset.

**PSC Google APIs** -- Scheme NONE with target `all-apis` or `vpc-sc`. Start from the **Private Service Connect to Google APIs** preset.

**Regional external ALB VIP** -- A regional rule on EXTERNAL_MANAGED with STANDARD tier in front of a regional proxy. Start from the **Regional External ALB VIP** preset.

**Internal passthrough NLB** -- A regional rule on INTERNAL naming a regional backend service directly, with a stable internal DNS name and global access. Start from the **Internal Passthrough NLB VIP** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the rule is created
- [**GCP Target HTTPS Proxy**](/cloud-catalog/gcp-target-https-proxy) -- the default target kind for port-443 rules
- [**GCP Target HTTP Proxy**](/cloud-catalog/gcp-target-http-proxy) -- the port-80 redirect half
- [**GCP Global Address**](/cloud-catalog/gcp-global-address) -- the reserved static IP this VIP binds
- [**GCP URL Map**](/cloud-catalog/gcp-url-map) -- sits behind the target proxy in the serving chain
