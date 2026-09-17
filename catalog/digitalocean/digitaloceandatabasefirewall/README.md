# DigitalOcean Database Firewall

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_firewall` resource at the pinned provider version.

## What this component models

The inbound trusted-sources rule set of a DigitalOcean managed database cluster. DigitalOcean's API takes one polymorphic rule list of `{type, value}` rows; this component replaces it with one TYPED list per source kind, so a value can never be paired with the wrong type and platform resources are wired by reference:

- `cluster` -- the cluster whose inbound sources these rules define (by UUID or reference)
- `ip_rules` -- IPv4 addresses or IPv4 CIDR blocks (IPv6 is refused at validation because DigitalOcean's database firewall rejects it at apply)
- `droplet_ids` -- Droplets, by numeric id or `DigitalOceanDroplet` reference
- `kubernetes_cluster_ids` -- DOKS clusters, by UUID or `DigitalOceanKubernetesCluster` reference
- `app_ids` -- App Platform apps, by UUID or `DigitalOceanApp` reference
- `tags` -- Droplet tags (membership tracks the tag automatically)

At least one rule across the five lists is required -- the provider rejects an empty set, and validation here catches it before any deploy.

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseFirewall
metadata:
  name: orders-postgres-firewall
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  ipRules:
    - 10.10.0.0/16
  tags:
    - backend
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | UUID of the cluster this rule set protects (the firewall's only durable identity) |

## Behavior worth knowing

- **One rule set per cluster.** The rule set is a property of the cluster. Declare ALL trusted sources in one resource -- two resources targeting one cluster overwrite each other.
- **A read replica has its own, initially EMPTY rule set.** The primary's firewall does not reach its replicas; declare a second resource per replica with `cluster` pointing at the replica's `replica_id`.
- **Updates replace the full set.** Every apply PUTs the complete list; there is no per-rule lifecycle.
- **Destroy OPENS the database.** Deleting this resource clears the rule set, after which the cluster accepts connections from anywhere again. Treat this resource's lifecycle as part of the cluster's security posture.
- **No stable state id.** The provider mints a random state identifier at create; the cluster UUID is the real identity (imports take the bare cluster UUID).

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines fan the five typed lists out to the provider's rule rows identically; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
