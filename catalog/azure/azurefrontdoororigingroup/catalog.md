# Azure Front Door Origin Group

Deploys a Front Door origin group -- the load-balanced pool of backends a route sends traffic to, carrying the pool-level behavior: health probing, latency-based origin selection, session affinity, and the traffic-restore ramp for recovered origins. The backends themselves are first-class Azure Front Door Origin resources referencing this group, so a regional stamp can add its backend to a shared group without touching the group or any other region's origins. Renaming the group replaces it AND every origin nested under it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Origin Group** -- a named child of the profile that owns the pool's load-balancing and health-probe policy
- **Load-balancing settings** -- always sent (Azure requires them on every group): sample size, successful samples required, and the latency window, with Azure's defaults (4 / 3 / 50 ms) when the spec is silent
- **Health probe** (created only when `healthProbe` is set) -- protocol, interval, request type, and path; absent probe settings mean probing disabled and every origin assumed healthy
- **Session affinity and restore timing** -- cookie-based stickiness (Azure default on) and the 0-50 minute ramp before a healed or new origin takes its full traffic share

ARM does not support tags on origin groups, so no Azure tags are applied here.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Profile** the group nests under (`profileId`). Reference an AzureFrontDoorProfile via ValueFromRef, or provide the profile ARM ID directly.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Origin Group**, and click **Deploy**. The wizard walks you through the parent profile and name, health probing, then load balancing and session affinity. Start from the **Single-Origin Group** preset in the [Presets](#presets) tab when probes would only add load, or **Multi-Region Probed Backends** when failover matters.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorOriginGroup
metadata:
  name: my-app-backends
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  originGroupName: app-backends
```

```shell
planton apply -f front-door-origin-group.yaml
```

This creates a group with Azure's load-balancing defaults and no health probe -- the right shape for a single origin; add origins next, then attach a route. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the full Front Door composition in one InfraPipeline, add the probe the moment the group will hold more than one origin:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  originGroupName: app-backends
  healthProbe:
    protocol: HTTPS
    intervalInSeconds: 30
    requestType: HEAD
    path: /healthz
```

The InfraPipeline resolves the dependency graph, deploys the profile first, then provisions the group with the resolved ARM ID.

## Key Configuration

These are the most important decisions when configuring an origin group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Health probe** -- omit it for a single-origin group: probes only add origin load when traffic has nowhere else to go. Configure it whenever the group has two or more origins -- probes are what take an unhealthy origin out of rotation. Probe over HTTPS when origins serve TLS (the probe then exercises certificate validity as part of health), and point `path` at a dedicated endpoint like `/healthz` so probes prove the application, not just the web server.

**Probe interval** -- `intervalInSeconds` (1-255) sets failure-detection speed, but every Front Door edge location probes every origin, so an aggressive interval can be a meaningful fraction of a small origin's traffic. 30-120 seconds suits most production workloads; with the default sampling (4 samples, 3 required), 30-second probes eject a failed origin in roughly two minutes.

**Load balancing** -- Front Door keeps origins whose recent probe samples pass, then routes among healthy origins whose latency is within `additionalLatencyInMilliseconds` of the fastest one. Small windows pin traffic to the closest origin; wider windows (e.g. 100 ms) spread it across geographically dispersed backends. Lower `successfulSamplesRequired` relative to `sampleSize` to tolerate flaky probes; raise it to eject faster.

**Session affinity** -- `sessionAffinityEnabled` defaults to true (Azure's default): cookie-based stickiness, correct for backends holding sessions in origin memory. Disable it for stateless APIs so traffic spreads evenly.

**Restore ramp** -- `restoreTrafficTimeToHealedOrNewEndpointInMinutes` (0-50, default 10) gradually shifts traffic to an origin that just recovered or joined, avoiding cold-starting a backend with its full share at once. Lengthen it for backends with heavy warmup (in-process caches, JIT, connection pools); 0 shifts immediately.

**ForceNew fields** -- `profileId` and `originGroupName` fix the group's ARM identity at creation; renaming replaces the group and every origin nested under it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `origin_group_id` | ARM resource ID of the origin group | AzureFrontDoorOrigin's `originGroupId` (parent) and AzureFrontDoorRoute's `originGroupId` (destination) |
| `origin_group_name` | The group's name within its profile | Operator tooling, portal cross-reference |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single backend, no probe** -- the honest shape for one origin: probing is deliberately absent because a single origin is served regardless of probe results. The group grows into a probed multi-origin pool when a second region arrives, without replacement. Start from the **Single-Origin Group** preset.

**Multi-region failover pool** -- HTTPS probes against a dedicated health endpoint, a latency window wide enough to spread traffic across regions, and affinity off for stateless APIs. The trade on probe aggressiveness: faster ejection costs more probe traffic against every origin. Start from the **Multi-Region Probed Backends** preset.

**Stateful web app pool** -- sticky sessions on (the load-bearing choice, spelled explicitly), gentle HEAD probes, and a 20-minute restore ramp -- twice Azure's default -- so a recovered origin receives a growing slice instead of an avalanche. Start from the **Stateful Web App Backends** preset.

## Works With

- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container the group nests under
- [**Azure Front Door Origin**](/infra-catalog/azure-front-door-origin) -- each backend inside this group, referencing `origin_group_id`
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- forwards matched requests to this group by `origin_group_id`
