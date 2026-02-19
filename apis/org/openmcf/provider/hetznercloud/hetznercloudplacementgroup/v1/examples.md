# HetznerCloudPlacementGroup Examples

## Minimal (Default Spread)

The simplest configuration: an empty spec. The `type` field defaults to `spread`, which is the only strategy Hetzner Cloud supports.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-group
spec: {}
```

---

## Explicit Spread Type

Identical behavior to the minimal example, but with the type explicitly set. Useful for documentation clarity in manifests that are reviewed by teams unfamiliar with the defaults.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
spec:
  type: spread
```

---

## Production HA with Organizational Metadata

A placement group for database replicas in a production environment. The org and env metadata drive label generation for resource tracking and cost allocation.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
  org: acme-corp
  env: production
  labels:
    role: database
    tier: critical
spec:
  type: spread
```

---

## InfraChart Composition with valueFrom

In an infra chart, a `HetznerCloudServer` references the placement group via `valueFrom` so the placement group ID is resolved from stack outputs. This eliminates hardcoded IDs and establishes a dependency edge in the DAG.

Placement group manifest:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
  org: acme-corp
  env: production
spec:
  type: spread
```

Server manifest referencing the placement group output:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: db-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: ha-db-group
      fieldPath: status.outputs.placement_group_id
```

The `valueFrom` reference ensures that:
1. The placement group is created before the server
2. The correct numeric ID is passed without manual lookup
3. If the placement group is replaced (e.g., type change), dependent servers are aware on the next apply
