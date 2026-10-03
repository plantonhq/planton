# Azure Traffic Manager Profile

Deploys a Traffic Manager profile -- Azure's DNS-based traffic director, which answers lookups on its `{relative-name}.trafficmanager.net` name with the address of one of its endpoints, chosen by routing method and endpoint health. Because the steering happens in DNS, Traffic Manager fronts anything with a resolvable address -- across regions, clouds, and on-premises -- and is never in the data path. The profile is a global object; endpoints are separate Azure Traffic Manager Endpoint resources referencing it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Traffic Manager profile** -- the global routing object with its DNS identity and health-probe configuration; the provider pins the ARM location to `global`, which is why the spec carries no region

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure Resource Group** -- the profile's metadata record lives in a referenced resource group (the profile itself is global).

### Azure Subscription

- **The DNS relative name is globally unique across ALL of Azure** -- the trafficmanager.net namespace is shared; Azure rejects a taken name at apply time, so prefix with your organization.
- **Billing is per million DNS queries plus per-endpoint health probes** -- fast-interval probes and Traffic View bill extra; the profile object itself is cheap at rest.

## Deploy

### Console

Open the deployment store, find **Azure Traffic Manager Profile**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the routing method, the DNS identity, and the health-probe settings. Start from the **Performance Routing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureTrafficManagerProfile
metadata:
  name: web-traffic-manager
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: prod-global
  name: web-traffic-manager
  routingMethod: Performance
  dnsConfig:
    relativeName: acme-corp-web-prod
    ttlSeconds: 60
  monitorConfig:
    protocol: HTTPS
    port: 443
    path: /healthz
```

```shell
planton apply -f profile.yaml
```

This creates a Performance-routed profile answering on `acme-corp-web-prod.trafficmanager.net`, probing each endpoint over HTTPS on `/healthz` every 30 seconds and expecting a 200. An Infra Job tracks the provisioning in real time.

### InfraChart

When the resource group is an Infra Component in the same chart, wire it by reference:

```yaml
spec:
  resourceGroup:
    valueFrom:
      name: prod-global
  name: web-traffic-manager
  routingMethod: Performance
  dnsConfig:
    relativeName: acme-corp-web-prod
  monitorConfig:
    protocol: HTTPS
    port: 443
    path: /healthz
```

The InfraPipeline resolves the dependency graph, provisioning the resource group before the profile -- and the profile before the endpoint resources that reference its `traffic_manager_profile_id` output.

## Key Configuration

These are the most important decisions when configuring an Azure Traffic Manager Profile. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Match the routing method to the topology.** Performance sends each caller to the lowest-latency endpoint (the everyday multi-region front); Priority is active/passive failover; Weighted spreads proportionally (canaries, migrations); Geographic pins callers' regions to specific endpoints; Subnet steers by source IP; MultiValue returns all healthy addresses at once. The method is updatable in place -- switching re-steers traffic without touching endpoints.

**The relative name is a global, permanent choice.** `{relativeName}.trafficmanager.net` is shared across every Azure customer and fixed at creation -- renaming replaces the profile and breaks every CNAME pointing at it. Prefix with your organization, treat the generated name as infrastructure, and give users only your own domain, CNAMEd to it.

**The TTL is your failover clock.** Traffic Manager stops handing out an unhealthy endpoint immediately, but clients keep using cached answers until the TTL drains. Detection time is `interval x tolerated failures + timeout`; add the TTL on top for the real user-visible failover window -- a 60-second TTL with default probing means roughly two minutes worst-case.

**Pick the probe that proves the SERVICE, not the socket.** TCP probes pass when the port accepts a handshake -- a hung application behind a live listener stays "healthy" forever. Probe HTTP/HTTPS with a real health path, scope `expectedStatusCodeRanges` deliberately (expecting only 200 while the app answers 301 to probes is the classic all-endpoints-degraded false alarm), and send a probe `Host` header when the target serves name-based virtual hosts.

**The fast interval is a paid trio, not one field.** `intervalInSeconds: 10` (billed extra per endpoint) requires an explicit `timeoutInSeconds` of 5-9 -- the default 10 no longer fits inside the probe window, and validation rejects the pairing. Combine it with a low tolerated-failure count and a short TTL to buy sub-minute failover.

**Know the all-degraded fallback.** When EVERY endpoint probes unhealthy, Traffic Manager answers as if all were healthy -- a monitoring misconfiguration degrades to "no steering" rather than "no answers". If your endpoints work while the portal shows everything degraded, fix the probe: real failover is silently broken.

**MultiValue is for resolver-side logic, not load balancing.** It answers with up to `maxReturn` healthy addresses at once (required, 1-8) and its endpoints must be literal external IP addresses. Use it when the client retries across addresses; for actual traffic distribution, Weighted is the tool.

**Disabling beats deleting.** `enabled: false` parks the profile -- its name stops steering -- without losing the globally unique relative name to a window where anyone can claim it. Maintenance windows and kill switches are one-field flips.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Resource Group | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `traffic_manager_profile_id` | The profile's ARM resource ID | What each Azure Traffic Manager Endpoint references; what alias DNS records target |
| `fqdn` | The profile's public DNS name (`{relative_name}.trafficmanager.net`) | The CNAME target for your own domain |

`traffic_manager_profile_name` is also exported for identification.

## Common Patterns

**Nearest-region front** -- A Performance profile over one endpoint per region: each user's lookup answers with the lowest-latency healthy endpoint, probed over HTTPS on a real health path. Start from the **Performance Routing** preset.

**Active/passive disaster recovery** -- A Priority profile where traffic holds on the primary until health fails it over, tuned with the fast-interval trio for a worst-case failover near 50 seconds. Start from the **Priority Failover** preset.

**Weighted canary** -- A Weighted profile with a heavy weight on the incumbent and a light one on the new deployment; shifting the split is an endpoint-weight edit, and pulling the canary is one endpoint disable.

## Works With

- [**Azure Traffic Manager Endpoint**](/infra-catalog/azure-traffic-manager-endpoint) -- the destinations this profile steers to; each references the profile's `traffic_manager_profile_id` output.
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- holds the profile's metadata record; reference its `resource_group_name` output.
- [**Azure DNS Record**](/infra-catalog/azure-dns-record) -- points your own domain at the profile: a CNAME to the `fqdn` output, or an alias record targeting the profile's ARM ID.
