# DigitalOcean Database Firewall

Declares the inbound trusted sources of a DigitalOcean managed database cluster: IPv4 addresses and IPv4 CIDR blocks (DigitalOcean's database firewall accepts no IPv6 shape -- the spec refuses it at validation), Droplets, Kubernetes clusters, App Platform apps, and Droplet tags -- each in its own typed list, with platform resources wired by reference instead of hand-copied ids. The rule set is a property of the cluster, not a standalone object: there is at most one per cluster, every apply replaces the full set, and destroying this resource clears the set -- after which the cluster accepts connections from anywhere again.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Trusted-Sources Rule Set** -- the cluster's complete inbound allowlist, fanned out from the five typed lists to DigitalOcean's `{type, value}` rule rows (`ip_addr`, `droplet`, `k8s`, `app`, `tag`)

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **A DigitalOceanDatabaseCluster** -- the cluster to protect, referenced by name (or an existing cluster's UUID as a literal).
- **Referenced sources** -- any Droplets, DOKS clusters, or Apps you trust by reference must exist (or deploy in the same chart).

### DigitalOcean Account

- Nothing beyond the cluster: trusted-source rules are free.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Database Firewall**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Network + Tagged Fleet** preset in the [Presets](#presets) tab to close the cluster to everything but your VPC range and tagged Droplets.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseFirewall
metadata:
  name: orders-postgres-firewall
  org: acme-corp
  env: prod
spec:
  cluster:
    value: "2f6a8f0e-3b1c-4c8e-9f2d-7a5b4c3d2e1f"
  ipRules:
    - 10.10.0.0/16
  tags:
    - backend
```

```shell
planton apply -f database-firewall.yaml
```

This locks the referenced cluster down to two source classes: the private `10.10.0.0/16` range and every Droplet carrying the `backend` tag -- nothing on the public internet can reach it. A Stack Job tracks the provisioning in real time.

### InfraChart

When the cluster and its trusted platform resources deploy in the same InfraPipeline, wire them all by reference:

```yaml
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  kubernetesClusterIds:
    - valueFrom:
        kind: DigitalOceanKubernetesCluster
        name: workloads-doks
        fieldPath: status.outputs.cluster_id
  appIds:
    - valueFrom:
        kind: DigitalOceanApp
        name: orders-api
        fieldPath: status.outputs.app_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster, DOKS cluster, and app first, then applies the rule set with the resolved ids.

## Key Configuration

These are the most important decisions when configuring a database firewall. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroying the firewall OPENS the database** -- "delete" is not a deletion. Destroy PUTs an empty rule list, after which the cluster accepts connections from anywhere, exactly as it did before the firewall existed. Never remove this resource as cleanup while the cluster lives; remove it only when the cluster goes with it, or replace it with the successor rule set in the same change.

**Exactly one rule set per cluster** -- DigitalOcean holds one trusted-sources list per cluster. Two of these resources pointing at the same cluster do not merge -- each apply overwrites the other's rules, and they will flap forever. Declare every trusted source in one resource per cluster, owned by one manifest.

**A read replica has its own firewall** -- a replica is its own cluster to DigitalOcean, with its own trusted-sources list that starts EMPTY; the primary's rules never reach it. Declare a second resource per replica with `cluster` pointing at the replica (`kind: DigitalOceanDatabaseReplica`, `fieldPath: status.outputs.replica_id`), or the replica stays open to the public endpoint with the primary's credentials.

**Every apply replaces the full set** -- there is no per-rule add or remove; each apply PUTs the complete list. That makes review easy (the manifest IS the allowlist) and partial edits impossible. At least one source across the five lists is required at validation time -- an empty set is rejected before any provisioner runs.

**Choose the list by the source's lifecycle** -- `tags` track every Droplet carrying the tag, so membership updates itself as fleets scale; references (`dropletIds`, `kubernetesClusterIds`, `appIds`) resolve live resource ids at deploy time; `ipRules` literals are for fixed points -- bastions, offices, private CIDR ranges. If you find yourself editing IP lists weekly, the fleet wants a tag.

**Broad CIDRs defeat the purpose** -- `0.0.0.0/0` as an `ipRules` entry is accepted by DigitalOcean and re-admits the whole internet deliberately. Validation requires at least one rule; it cannot require the rule be narrow.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanDatabaseCluster** | `cluster` | `status.outputs.cluster_id` |
| **DigitalOceanDroplet** (optional, per entry) | `dropletIds[]` | `status.outputs.droplet_id` |
| **DigitalOceanKubernetesCluster** (optional, per entry) | `kubernetesClusterIds[]` | `status.outputs.cluster_id` |
| **DigitalOceanApp** (optional, per entry) | `appIds[]` | `status.outputs.app_id` |

Every reference also accepts a literal id in its place -- a cluster UUID, numeric Droplet id, DOKS cluster UUID, or app UUID.

### What This Component Provides

After provisioning, `status.outputs` carries a single value: `cluster_id`, an echo of the resolved cluster reference. DigitalOcean mints no stable standalone id for the rule set -- it is a property of its cluster, and the cluster UUID is the only durable identity -- so there is nothing here for downstream Cloud Resources to consume.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private network plus tagged fleet** -- trust the VPC's private range and every Droplet carrying a fleet tag; membership tracks the tag as the fleet scales, and nothing on the public internet can reach the cluster. Start from the **Private Network + Tagged Fleet** preset.

**Platform workloads only** -- trust a Kubernetes cluster and an App Platform app, both by reference, and nothing else: the database becomes reachable exclusively from platform-managed compute, with zero IP management. Start from the **Platform Workloads Only** preset.

## Works With

- [**DigitalOcean Database Cluster**](/cloud-catalog/digital-ocean-database-cluster) -- the cluster whose inbound sources this rule set defines
- [**DigitalOcean Droplet**](/cloud-catalog/digital-ocean-droplet) -- trusted individually via `dropletIds` or as a tagged fleet via `tags`
- [**DigitalOcean Kubernetes Cluster**](/cloud-catalog/digital-ocean-kubernetes-cluster) -- trusted via `kubernetesClusterIds` so cluster workloads reach the database
- [**DigitalOcean App Platform App**](/cloud-catalog/digital-ocean-app) -- trusted via `appIds` for App Platform services
