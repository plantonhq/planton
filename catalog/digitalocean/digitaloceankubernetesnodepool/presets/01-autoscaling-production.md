# Autoscaling Production Pool

This preset adds an autoscaling application pool to an existing DOKS cluster: 3 nodes initially, scaling between 2 and 6 as demand moves, with a `workload: app` node label for scheduling and a production tag for DigitalOcean-side grouping.

## When to Use

- Application workloads whose demand varies (web tiers, API backends, queue consumers)
- Growing a cluster beyond its default pool without resizing it (default-pool size changes replace the whole cluster)
- Separating application workloads from system workloads with labels

## Key Configuration Choices

- **Autoscaling with coherent bounds** (`autoScale` + `minNodes`/`maxNodes`) -- DigitalOcean's cluster-autoscaler manages the node count between 2 and 6; the pool starts at `minNodes` and no `nodeCount` is stated, because a stated count would be re-applied against the autoscaler on every update.
- **General-purpose sizing** (`size: s-4vcpu-8gb`) -- balanced CPU/RAM for typical application pods. Changing the size later replaces the pool (the nodes are recreated), so schedule size changes deliberately.
- **Node label** (`workload: app`) -- target this pool from Kubernetes with a `nodeSelector` or node affinity. The standard Planton identity labels are always applied alongside.
- **Cluster reference** (`cluster.valueFrom`) -- resolves the owning cluster's UUID from a `DigitalOceanKubernetesCluster` resource's outputs at deploy time.

## Placeholders to Replace

- `metadata.name` / `nodePoolName` -- your pool's name (unique within the cluster).
- `cluster.valueFrom.name` -- the name of your `DigitalOceanKubernetesCluster` resource (or replace the block with `value: <the cluster UUID>` for a cluster created outside Planton).
