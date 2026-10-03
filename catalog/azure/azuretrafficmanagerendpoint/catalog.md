# Azure Traffic Manager Endpoint

Deploys one destination of a Traffic Manager profile: a public Azure resource by ARM ID (`azure`), a DNS name or IP address (`external`), or another profile composing routing trees (`nested`) -- exactly one variant per endpoint. Shared fields (weight, priority, enabled, geo and subnet claims, probe headers) live at the spec root, and which of them matter depends on the referenced profile's routing method. Endpoints are free at rest; probes and queries bill on the profile.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **One Traffic Manager endpoint** of the type your spec's variant declares, inside the referenced profile -- exactly one of the three typed endpoint resources materializes, addressed as `{profile_id}/{TYPE}/{name}`

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An Azure Traffic Manager Profile** -- the endpoint is created inside a referenced profile; `profileId` defaults to referencing its `traffic_manager_profile_id` output, which also orders the deploy.

### Azure Subscription

- **Which fields matter depends on the PROFILE's routing method** -- weight (Weighted), priority (Priority), geo claims (Geographic: every code claimed by exactly one endpoint), subnet claims (Subnet: no overlaps) -- Azure evaluates them at apply time.
- **Azure endpoints need a PUBLIC address on the target** -- the referenced resource must hold a public IP (Standard tier for Public IPs; Basic is not steerable); external targets under Performance routing need an explicit `endpointLocation`.
- **Endpoints are free at rest** -- probes and queries bill on the profile.

## Deploy

### Console

Open the deployment store, find **Azure Traffic Manager Endpoint**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the profile reference, and the endpoint variant with its routing fields. Start from the **Azure Endpoint** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureTrafficManagerEndpoint
metadata:
  name: eastus-web
  org: acme-corp
  env: prod
spec:
  profileId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-global/providers/Microsoft.Network/trafficManagerProfiles/web-traffic-manager
  name: eastus-web
  priority: 10
  azure:
    targetResourceId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-eastus/providers/Microsoft.Network/publicIPAddresses/web-eastus-pip
```

```shell
planton apply -f endpoint.yaml
```

This adds the `web-eastus-pip` Public IP as an azure-type endpoint of the profile at failover priority 10; the profile starts probing it immediately, and it enters DNS answers once probes pass. An Infra Job tracks the provisioning in real time.

### InfraChart

When the profile and the target are Infra Components in the same chart, wire both by reference:

```yaml
spec:
  profileId:
    valueFrom:
      name: web-traffic-manager
  name: eastus-web
  priority: 10
  azure:
    targetResourceId:
      valueFrom:
        kind: AzurePublicIp
        name: web-eastus-pip
        fieldPath: status.outputs.public_ip_id
```

The InfraPipeline resolves the dependency graph, provisioning the profile and the target before the endpoint that joins them.

## Key Configuration

These are the most important decisions when configuring an Azure Traffic Manager Endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Azure endpoints follow the resource; external endpoints follow the string.** An `azure` endpoint tracks its target resource -- if the Public IP's address changes, Traffic Manager follows automatically, and the resource's region feeds Performance routing with no location field at all. An `external` endpoint is a frozen string: retargeting is an edit, and Performance routing needs the explicit `endpointLocation` you gave it. Prefer azure endpoints for Azure resources; use external only for what genuinely lives elsewhere.

**Set priorities explicitly, or creation order owns your failover plan.** Unset priority lets Azure assign the next free value in creation order -- fine until someone recreates an endpoint and it silently moves to the back of the failover line. On Priority-routed profiles, give every endpoint an explicit value with gaps (10, 20, 30) so inserting a tier later never renumbers the plan.

**Drain with `enabled: false`; reserve always-serve for broken probes.** Disabling is the maintenance drain: the endpoint leaves DNS answers while its configuration stays. `alwaysServeEnabled` is the opposite tool -- it disables health checking and keeps the endpoint in answers even when probes fail, for targets probes cannot reach. Never leave always-serve on a target you expect health-based failover to protect; it opts that endpoint out of failover entirely.

**Geographic claims are exclusive and validated live.** Every geographic code (`WORLD`, `GEO-EU`, country codes) must be claimed by exactly ONE endpoint in the profile -- Azure rejects overlaps at apply time against its live hierarchy. Claim `WORLD` on a catch-all endpoint; a Geographic profile with unclaimed regions returns no answer for those callers.

**Nested trees: the child floor is your blast-radius dial.** `minimumChildEndpoints` decides when a whole child profile counts as down: with a floor of 1, the parent keeps sending traffic to a region running on its last healthy instance; with a floor near the child's endpoint count, one instance failure fails the region over entirely. Set it to the child's genuine minimum serving capacity. The child profile must not use MultiValue routing.

**Subnet claims are fixed at creation.** `subnets` (source-IP claims for Subnet-routed profiles) is the one shared field whose change replaces the endpoint -- and ranges must not overlap across the profile's endpoints. Everything else on the endpoint updates in place except its name, profile, and type.

**Name for the monitor view.** Uniqueness is per (profile, endpoint type) -- an azure and an external endpoint may legally share a name. Do not lean on that: names that say what they front (`eastus-pip`, `onprem-dc2`) make the monitor view and the ARM IDs read at a glance.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Traffic Manager Profile | `profileId` | `status.outputs.traffic_manager_profile_id` |
| Azure Traffic Manager Profile (the child, for nested endpoints) | `nested.targetProfileId` | `status.outputs.traffic_manager_profile_id` |
| Azure Public IP (or any public Azure resource) | `azure.targetResourceId` | `status.outputs.public_ip_id` (kind declared explicitly) |
| Any component with a hostname output (external targets) | `external.target` | declared explicitly per kind |

### What This Kind Provides

`status.outputs` carries the endpoint's ARM ID (`endpoint_id`) and its name within the profile (`endpoint_name`). Nothing downstream consumes an endpoint by reference -- it is a leaf destination inside its profile -- so these outputs exist for identification and import rather than composition.

## Common Patterns

**Regional Azure fleet** -- One azure-type endpoint per region's Public IP behind a Performance or Weighted profile; the address and region stay tracked by Azure. Start from the **Azure Endpoint** preset.

**Hybrid active/passive** -- An azure endpoint as the priority-10 primary and an external endpoint (on-premises or another cloud) at priority 20; together they form a failover pair on a Priority profile. Start from the **External Endpoint** preset.

**Routing trees** -- A nested endpoint per region pointing at regional Weighted child profiles under a Performance parent: latency picks the region, weights split within it, and the child floor decides when a whole region fails over.

## Works With

- [**Azure Traffic Manager Profile**](/infra-catalog/azure-traffic-manager-profile) -- the profile this endpoint belongs to, and (for nested endpoints) the child it targets.
- [**Azure Public IP**](/infra-catalog/azure-public-ip) -- the most common azure-variant target; reference its `public_ip_id` output (Standard tier).
