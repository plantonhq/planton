# DigitalOcean Reserved IP

Reserves a static public IP address (IPv4 or IPv6) in a DigitalOcean region and optionally assigns it to a droplet. The address outlives the droplets behind it: DNS points at the reservation while droplets come and go, and re-pointing between droplets is the classic manual-failover building block. Assignment is also the billing switch -- an assigned reservation is free, while an unassigned reserved IPv4 accrues a monthly charge (prorated hourly) for exactly as long as it holds capacity without a droplet.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Reserved IP** -- the static address in your chosen region (IPv4 by default; IPv6 when `ipVersion: ipv6`)
- **Droplet assignment** -- created only when `droplet` is set. On IPv4 the assignment rides the reservation itself and updates in place; on IPv6 it is a separate assignment resource (the v6 API cannot assign inline), re-pointed by replacing just the assignment -- the address itself never changes either way

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.
- **A droplet (optional)** -- a DigitalOceanDroplet in the SAME region as the reservation, if assigning at create time.

### DigitalOcean Account

- **Billing awareness** -- an unassigned reserved IPv4 bills monthly (prorated hourly) until assigned or destroyed; assigned reservations are free, and IPv6 is free either way. The idle state is the expensive one.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Reserved IP**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Web Front-Door IP** preset in the [Presets](#presets) tab to reserve an address and assign it to your web droplet in one step.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanReservedIp
metadata:
  name: web-frontdoor-ip
  org: acme-corp
  env: prod
spec:
  region: nyc3
  droplet:
    value: "512190123"
```

```shell
planton apply -f do-reserved-ip.yaml
```

This reserves an IPv4 address in `nyc3` and assigns it to the droplet -- the assigned state, which is the free one. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the assignment to a droplet deployed in the same InfraPipeline:

```yaml
spec:
  region: nyc3
  droplet:
    valueFrom:
      kind: DigitalOceanDroplet
      name: web-server
      fieldPath: status.outputs.droplet_id
```

The InfraPipeline resolves the dependency graph, deploys the droplet first, then reserves the address already pointed at it.

## Key Configuration

These are the most important decisions when configuring a reserved IP. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**`ipVersion` is a one-way door** -- Unset means `ipv4`. Switching versions after creation replaces the reservation with a completely different address -- every DNS record naming the old one goes stale. Decide the version before anything points at the address.

**`region` pins where the address can serve** -- A reserved IP assigns only to droplets in ITS region, and changing the region replaces the reservation with a new address. Cross-region failover is not what this kind does (that is DNS or a global load balancer); this is intra-region droplet swapping. Pin the region to where the droplets actually live.

**`droplet` is the failover lever -- and the billing switch** -- Assigning, re-pointing, and unassigning all apply in place; the address never changes. In practice: run a standby droplet, and failover is a one-field manifest change. Leaving the field unset keeps the address reserved but unassigned -- exactly the state that bills for IPv4. Hold capacity deliberately, not by forgetting.

**Destroy releases the address permanently** -- The whole point of reserving is that the address survives droplet churn, so the flip side matters: destroying the reservation returns the address to DigitalOcean's pool, and recreating gets a DIFFERENT one. Destroy is a DNS event -- plan it as one. But do destroy reservations that stop being useful; "keep it around just in case" is the one posture this kind punishes.

**IPv6 deletes deserve a second look** -- The provider's v6 delete swallows every error except 404, so a failed release can look like a success in IaC output. When cleaning up IPv6 reservations by hand, verify with `doctl compute reserved-ipv6 list`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDroplet** (optional) | `droplet` | `status.outputs.droplet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `reserved_ip_address` | The reserved address (IPv4 or IPv6) -- this IS the resource identity; imports, lookups, and assignments all address the reservation by it | DNS A/AAAA record values, application configuration |
| `urn` | The reservation's uniform resource name (e.g. `do:reservedip:203.0.113.10`) | DigitalOcean project membership |

Point DNS at `reserved_ip_address` instead of any droplet's own address -- replacing the droplet later means re-pointing the reservation, not re-publishing DNS.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Front-door address** -- reserve and assign to the serving droplet, point DNS at the reservation, and droplet replacements stop being DNS events. The assigned state is free, so this shape has no carrying cost. Start from the **Web Front-Door IP** preset.

**Standby failover address** -- reserve WITHOUT assigning: capacity held for a failover or migration you have not executed yet. Be deliberate -- this exact shape is the one that bills until assigned or destroyed, so hold it only while the plan that needs it is real. Start from the **Standby Failover IP** preset.

## Works With

- [**DigitalOcean Droplet**](/infra-catalog/digital-ocean-droplet) -- the assignment target; re-pointing between droplets is the failover move
- [**DigitalOcean DNS Zone**](/infra-catalog/digital-ocean-dns-zone) -- zone records pointing at `reserved_ip_address` survive droplet replacements
- [**DigitalOcean DNS Record**](/infra-catalog/digital-ocean-dns-record) -- a standalone A record consuming `reserved_ip_address` via ValueFromRef
- [**DigitalOcean Project**](/infra-catalog/digital-ocean-project) -- project membership lists carry the reservation's `urn`
