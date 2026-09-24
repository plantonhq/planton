# DigitalOcean VPC Peering

Deploys a VPC peering connection between two DigitalOcean VPCs, so resources in both networks reach each other over private addresses without touching the public internet. The two members are modeled as named references (`vpc_1`, `vpc_2`) because the API requires exactly two -- any other cardinality is unrepresentable -- and the pair is symmetric: which VPC is which does not matter. Only the peering's name updates in place; changing either VPC replaces the connection.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **VPC peering connection** -- one `digitalocean_vpc_peering` resource linking the two referenced VPCs. The module waits for the peering to reach ACTIVE before exporting outputs, and deletes retry through DigitalOcean's transient 403 responses while the peering settles.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Two VPCs** -- DigitalOceanVpc resources (or literal VPC UUIDs) with non-overlapping IP ranges; DigitalOcean rejects overlapping peers.

### DigitalOcean Account

- **Disjoint CIDR plans** -- a VPC's IP range is create-only, so an overlap discovered at peering time means rebuilding one of the networks. If two environments might ever need to talk, give their VPCs disjoint ranges on day one.

## Deploy

### Console

Open the deployment store, find **DigitalOcean VPC Peering**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **App-to-Data Peering** preset in the [Presets](#presets) tab to link an application VPC with a data VPC.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVpcPeering
metadata:
  name: app-to-data-peering
  org: acme-corp
  env: prod
spec:
  peeringName: app-to-data
  vpc_1:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  vpc_2:
    value: ffffffff-1111-2222-3333-444444444444
```

```shell
planton apply -f do-vpc-peering.yaml
```

This peers the two VPCs by literal UUID; the link is active within minutes and needs no route-table work on either side. A Stack Job tracks the provisioning in real time.

### InfraChart

When the two VPCs deploy in the same InfraPipeline, wire both members by reference:

```yaml
spec:
  peeringName: app-to-data
  vpc_1:
    valueFrom:
      kind: DigitalOceanVpc
      name: vpc-app
      fieldPath: status.outputs.vpc_id
  vpc_2:
    valueFrom:
      kind: DigitalOceanVpc
      name: vpc-data
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys both VPCs first, then provisions the peering with the resolved UUIDs -- recreating a VPC re-wires the peering in the same apply.

## Key Configuration

These are the most important decisions when configuring a VPC peering. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**CIDR planning happens before the first VPC, not before the peering** -- DigitalOcean rejects peerings between VPCs with overlapping IP ranges, and a VPC's range is create-only. Discovering an overlap at peering time means rebuilding one of the networks, so plan disjoint ranges when the VPCs are born.

**Only the name is mutable** -- `peeringName` updates in place; changing `vpc_1` or `vpc_2` REPLACES the peering, and the replacement is a traffic event on the old pair. Treat the two members as fixed once workloads depend on the link.

**All-or-nothing reachability** -- a peering exposes every private address in each VPC to the other, with no per-CIDR filtering on the link itself; narrowing access is droplet-firewall work on each host. Peer networks that genuinely trust each other; for one-service access, prefer exposing that service properly instead of peering whole networks.

**No transit -- peer pairwise** -- DigitalOcean peering is non-transitive: A-B and B-C do not give A-C. A hub-VPC topology needs an explicit peering per spoke pair that must communicate, and each pair is its own instance of this kind -- charts compose them cleanly.

**Deleting a peering is a traffic event** -- the moment the peering goes, cross-VPC routes vanish and everything that talked across it starts timing out; there is no drain or grace period. Trace what depends on the link before destroying it. Settling also takes minutes: creates wait for ACTIVE and deletes ride out transient 403s, so back-to-back create/destroy cycles on the same VPC pair can briefly collide with a peering still DELETING -- give teardowns a moment before re-peering.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanVpc** | `vpc_1` | `status.outputs.vpc_id` |
| **DigitalOceanVpc** | `vpc_2` | `status.outputs.vpc_id` |

### What This Component Provides

`status.outputs` carries one value: `peering_id`, the connection's UUID (its API identity and import id). The lifecycle status is deliberately not an output -- both provisioners wait for ACTIVE before the apply succeeds, so a stored status could only ever read ACTIVE and would go stale the moment DigitalOcean moved the peering; anyone who needs it reads it live (`GET /v2/vpcs/peerings/{id}`), which is also how the E2E verifier asserts the peering. No downstream Cloud Resource consumes a peering by reference, so there is no ValueFromRef story to teach.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**App-to-data tier split** -- a stateless application VPC peered to a locked-down data VPC: app droplets reach databases over private addresses while the two networks keep separate lifecycles and firewall postures. Start from the **App-to-Data Peering** preset.

**Cross-region private link** -- peering is region-agnostic, so multi-region deployments replicate and coordinate privately without a VPN; only the CIDRs must be disjoint. Private does not mean local -- design replication and timeouts for the inter-region round trip. Start from the **Cross-Region Link** preset.

## Works With

- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- the two networks this connection links, wired by their `vpc_id` outputs
- [**DigitalOcean Cloud Firewall**](/cloud-catalog/digital-ocean-firewall) -- per-host access control across the link; the peering itself filters nothing
- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- the workloads reaching across the peering by private address
- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the classic far side: databases in a data VPC reached privately from the app VPC
