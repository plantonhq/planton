# DigitalOceanDatabaseFirewall

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDatabaseFirewallSpec models the full
digitalocean_database_firewall resource surface: the inbound trusted
sources of a DigitalOcean managed database cluster.

DigitalOcean's API takes one polymorphic rule list of {type, value}
rows; this spec replaces it with one TYPED list per source kind, so a
value can never be paired with the wrong type and resources are wired by
reference instead of hand-copied ids. The modules fan the lists back out
to the provider's rows.

The rule set is a PROPERTY of the cluster, not a standalone object:
there is at most one per cluster (declare all trusted sources in one
resource -- two resources on one cluster overwrite each other), every
update replaces the full set, and "destroy" clears the set (an empty
rule list), after which the cluster accepts connections from anywhere
again -- destroying the firewall OPENS the database, a fact the
verification tooling asserts as an empty rule list rather than a
deleted object.

## Example

```yaml
# Reference manifest for DigitalOceanDatabaseFirewall -- protovalidate-
# valid, embedded as the reference page's Example block, and the document
# the offline tofu plan renders. Exercises all five typed rule lists (the
# offline plan proves the full fan-out; live lanes prove the ip and droplet
# arms).
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseFirewall
metadata:
  name: orders-postgres-firewall
spec:
  # Literal cluster UUID; use valueFrom to reference a
  # DigitalOceanDatabaseCluster resource instead.
  cluster:
    value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
  ipRules:
    - 203.0.113.10
    - 10.10.0.0/16
  dropletIds:
    - value: "123456789"
  kubernetesClusterIds:
    - value: bbbbbbbb-cccc-dddd-eeee-ffffffffffff
  appIds:
    - value: cccccccc-dddd-eeee-ffff-000000000000
  tags:
    - backend
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanDatabaseCluster (`status.outputs.cluster_id`) |
| `spec.ipRules` | `[]string` |  |  |  |
| `spec.dropletIds` | `[]string \| valueFrom` |  |  | DigitalOceanDroplet (`status.outputs.droplet_id`) |
| `spec.kubernetesClusterIds` | `[]string \| valueFrom` |  |  | DigitalOceanKubernetesCluster (`status.outputs.cluster_id`) |
| `spec.appIds` | `[]string \| valueFrom` |  |  | DigitalOceanApp (`status.outputs.app_id`) |
| `spec.tags` | `[]string` |  |  |  |

## Field Details

### spec.cluster

`string | valueFrom` · required

The database cluster whose inbound sources these rules define. Use a
literal cluster UUID or a reference to a DigitalOceanDatabaseCluster
resource. Changing it moves the rule set to another cluster (replace).

- references: DigitalOceanDatabaseCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDatabaseCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.ipRules

`[]string`

(Optional) IP addresses or CIDR blocks trusted to reach the cluster.

- rule: {"repeated":{"items":{"cel":[{"id":"ip_or_cidr","message":"must be an IP address or CIDR block","expression":"this.isIp() || this.isIpPrefix()"}]}}}

### spec.dropletIds

`[]string | valueFrom`

(Optional) Droplets trusted to reach the cluster, as literal numeric
Droplet IDs or references to DigitalOceanDroplet resources.

- references: DigitalOceanDroplet (`status.outputs.droplet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanDroplet, name: <that resource's name>, fieldPath: status.outputs.droplet_id}} -- a bare string does not parse

### spec.kubernetesClusterIds

`[]string | valueFrom`

(Optional) Kubernetes clusters trusted to reach the cluster, as
literal cluster UUIDs or references to DigitalOceanKubernetesCluster
resources.

Containment-exempt: this is the database cluster's inbound trusted-
sources list; a Kubernetes cluster named here is a client the database
ADMITS, never the place the rule set lives. On a diagram the rule set
stays with its database cluster, with a line to each trusted cluster.

- references: DigitalOceanKubernetesCluster (`status.outputs.cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanKubernetesCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.appIds

`[]string | valueFrom`

(Optional) App Platform apps trusted to reach the cluster, as literal
app UUIDs or references to DigitalOceanApp resources.

- references: DigitalOceanApp (`status.outputs.app_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanApp, name: <that resource's name>, fieldPath: status.outputs.app_id}} -- a bare string does not parse

### spec.tags

`[]string`

(Optional) Droplet tags trusted to reach the cluster: every Droplet
carrying a listed tag is trusted, and membership tracks the tag
automatically -- prefer tags over droplet_ids for dynamic fleets.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

## Validation Rules

- `spec.at_least_one_rule`: at least one trusted source (ip_rules, droplet_ids, kubernetes_cluster_ids, app_ids, or tags) must be specified

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDatabaseFirewall, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | UUID of the database cluster whose inbound sources this rule set defines. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanDatabaseCluster | `status.outputs.cluster_id` |
| `spec.dropletIds` | DigitalOceanDroplet | `status.outputs.droplet_id` |
| `spec.kubernetesClusterIds` | DigitalOceanKubernetesCluster | `status.outputs.cluster_id` |
| `spec.appIds` | DigitalOceanApp | `status.outputs.app_id` |

## See Also

- [Overview](../README.md)
