# DigitalOcean Kubernetes Node Pool -- Operational Guide

Live-earned judgment for operating additional DOKS worker pools. The [README](README.md) covers what the component models; this covers how it behaves.

## The default pool is not this kind

Every DOKS cluster carries an inline default pool that belongs to the `DigitalOceanKubernetesCluster` resource. DigitalOcean marks it with the `terraform:default-node-pool` tag and refuses to import it as a standalone pool. Grow a cluster with this kind; never try to adopt the default pool here.

## Size changes replace the pool

`size` (and `gpuPartitionMode`) are ForceNew: changing them tears the pool's nodes down and creates new ones. Workloads on the pool reschedule during the replacement. The safe pattern for a capacity-class change is blue/green at the pool level: create the new pool, drain workloads onto it (labels + selectors), then delete the old one.

## Autoscaling drift is by design

With `autoScale: true`, the live node count moves between `minNodes` and `maxNodes` without touching your manifest -- the provider suppresses the diff by comparing against the pool's actual count. `nodeCount` only seeds the pool. The spec enforces the bounds' coherence early (`minNodes >= 1`, `maxNodes >= minNodes`); DigitalOcean would reject incoherent bounds only at apply time.

## Labels and taints travel with the pool, not the nodes

DOKS reapplies the pool's labels and taints to every node it creates -- including replacements after autoscaler scale-ups and node recycles. Imperative `kubectl label node`/`kubectl taint node` edits on individual nodes are lost on the next node rotation; put them here instead.

## Tags versus labels

`tags` are DigitalOcean-side: they group the pool's Droplets for billing attribution and can be targeted by DigitalOcean Cloud Firewalls. `labels` are Kubernetes-side: they drive scheduling. The provider silently filters DOKS's own machinery tags (`k8s:*`, `terraform:*`) out of state -- never author tags with those prefixes.

## Wire Droplet-scoped resources to the pool through its tags, never its node ids

The pool's node and Droplet ids are deliberately not outputs. DOKS replaces nodes by design: the autoscaler adds and removes them, a cluster upgrade recycles every one, and auto-repair swaps a failed node for a fresh Droplet with a new id. A firewall built from a list of Droplet ids captured at apply time would protect the nodes that existed then and silently miss every node created since. Put a tag on the pool (`tags: [web-workers]`) and target that tag from the `DigitalOceanFirewall` or load balancer: DigitalOcean applies the pool's tags to each node it creates, so the tag follows membership on its own. When you need the live node set for automation, read it from the API (`GET /v2/kubernetes/clusters/{cluster_id}/node_pools/{node_pool_id}`) at the moment you need it; `cluster_id` and `node_pool_id` are the outputs for exactly that.

## Creates and destroys take minutes, not seconds

A pool is "created" only when every node reports `running`: measured 1.5–2 minutes for a one-node `s-1vcpu-2gb` pool on a fresh cluster, whichever engine applies it. A destroy takes 70–80 seconds because DigitalOcean drains and terminates the nodes before the API reports the pool gone; the nodes' Droplets then linger as members of the cluster's VPC for a further minute or two. Budget both when a pool sits inside a larger pipeline, and never treat a pool that is still `provisioning` a minute in as stuck.

## Importing an existing pool

The import id is the plain pool UUID (`doctl kubernetes cluster node-pool list <cluster-id>`); the provider recovers the owning cluster by scanning the account. Two expected non-defects on a blind round-trip: the module's Planton identity labels/tags appear as additions on the first plan (they were not on the manually created pool), and a default pool import is refused outright.

## GPU pools

`gpuPartitionMode` accepts the two AMD partition tokens (`AMD_PARTITION_MODE_SPX_NPS1`, `AMD_PARTITION_MODE_DPX_NPS2`) and only makes sense on AMD GPU size slugs. Both provisioners deploy it, and both replace the pool when it changes — decide the partition layout before the pool carries workloads.

## What is deliberately NOT here

Per-node customization (DOKS nodes are cattle -- the pool is the unit), Kubernetes-side objects (deployments, tolerations -- author them in your workload manifests), and the cluster's own settings (version, addons, maintenance -- the `DigitalOceanKubernetesCluster` kind owns those).
