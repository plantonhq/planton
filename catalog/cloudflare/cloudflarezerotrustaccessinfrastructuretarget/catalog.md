# Cloudflare Zero Trust Access Infrastructure Target

Registers an infrastructure target: a server, identified by hostname and private IP, that Access infrastructure applications broker short-lived SSH access to through the account's tunnels. Targets are the inventory layer — applications select them by hostname or IP, which makes the hostname itself the access-control surface. A target is a plain CRUD object (real create, update, delete), unlike the Zero Trust settings singletons.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Infrastructure Target** — one target on the account binding `hostname` to its IPv4 and/or IPv6 address, each inside a virtual network

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **Zero Trust enabled on the account** — the organization (team name) onboarding step must be done (a CloudflareZeroTrustOrganization Infra Component).
- **A tunnel path to the server** (for live SSH brokering) — a CloudflareZeroTrustTunnel whose CloudflareZeroTrustTunnelRoute covers the target's address, in the same virtual network. Registration alone does not make the target reachable.
- **A virtual network** (only for overlapping CIDRs) — a CloudflareZeroTrustTunnelVirtualNetwork per network segment when two sites reuse the same private ranges.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Access Infrastructure Target**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the account, the hostname, and the per-family addressing. Start from the **IPv4 target in the default virtual network** preset in the [Presets](#presets) tab to pre-populate the everyday shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustAccessInfrastructureTarget
metadata:
  name: prod-db-1
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  hostname: prod-db-1
  ip:
    ipv4:
      ipAddr: 10.0.10.5
```

```shell
planton apply -f target.yaml
```

This registers `prod-db-1` at `10.0.10.5` in the account's default virtual network — inventory only; SSH reachability still needs a tunnel route covering the address. An Infra Job tracks the provisioning in real time.

### InfraChart

When the virtual network is deployed in the same InfraPipeline, wire `virtualNetworkId` with ValueFromRef:

```yaml
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  hostname: prod-db-1
  ip:
    ipv4:
      ipAddr: 10.0.10.5
      virtualNetworkId:
        valueFrom:
          kind: CloudflareZeroTrustTunnelVirtualNetwork
          name: dc-east
          fieldPath: status.outputs.virtual_network_id
```

The InfraPipeline resolves the dependency graph, deploys the virtual network first, then registers the target inside it.

## Key Configuration

These are the most important decisions when configuring an infrastructure target. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The hostname is the permission scheme** — infrastructure applications grant access by hostname pattern (`prod-db-*`), so a target named casually (`box7`) either falls outside every pattern and is unreachable, or inside the wrong one and is over-granted. Decide the naming scheme first — environment-role-index (`prod-db-1`) works — and register every target inside it. Hostnames allow letters, digits, dashes, and periods, up to 255 characters, starting and ending alphanumeric.

**The default virtual network is a real choice** — omitting `virtualNetworkId` places the address in the account's default virtual network: fine for a flat estate, wrong for overlapping CIDRs. If two datacenters both use 10.0.10.0/24, each needs its own CloudflareZeroTrustTunnelVirtualNetwork, and every target must say which one it means. An omitted virtual network never drifts the plan — silence is stable, but it is still a choice.

**Inventory, not connectivity** — registering a target does not make it reachable. SSH sessions ride the account's tunnels, and a tunnel must route the target's network in the same virtual network. Register the target, route its network, then point an infrastructure application at it — three resources, one path.

**Address families** — declare the IPv4 arm, the IPv6 arm, or both; each binds one address inside one virtual network, and a declared family must carry its address. Destroy is a real delete: the target leaves the inventory immediately.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareZeroTrustTunnelVirtualNetwork** (optional, per family) | `ip.ipv4.virtualNetworkId` / `ip.ipv6.virtualNetworkId` | `status.outputs.virtual_network_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `target_id` | The Cloudflare-assigned UUID of the target | Target lookups in the Access API and import recipes |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single server, default network** — one hostname, one private IPv4 address, default virtual network by omission; the everyday shape for a flat estate. Start from the **IPv4 target in the default virtual network** preset.

**Overlapping CIDRs, isolated networks** — both address families pinned into a specific virtual network so two sites reusing 10.0.10.0/24 stay unambiguous. Start from the **Dual-stack target in an isolated virtual network** preset.

**Fleet registration by scheme** — one manifest per server, all following the same hostname scheme, so a single infrastructure application pattern (`prod-db-*`) grants exactly the intended set.

## Works With

- [**Cloudflare Zero Trust Tunnel Virtual Network**](/infra-catalog/cloudflare-zero-trust-tunnel-virtual-network) — the segment the address lives in; disambiguates overlapping CIDRs
- [**Cloudflare Zero Trust Tunnel Route**](/infra-catalog/cloudflare-zero-trust-tunnel-route) — the route that makes the target reachable
- [**Cloudflare Zero Trust Tunnel**](/infra-catalog/cloudflare-zero-trust-tunnel) — the data path SSH sessions ride
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) — the infrastructure application that selects targets and grants SSH
