# Hetzner Cloud Placement Group

Creates a placement group in Hetzner Cloud that controls the physical distribution of servers across infrastructure. Servers assigned to a spread placement group are guaranteed to run on different physical hosts, providing fault tolerance for high-availability workloads.

## What Gets Created

- **Placement Group** — an `hcloud_placement_group` resource with the specified strategy (defaults to `spread`), a name derived from `metadata.name`, and standard labels computed from resource metadata. The group is referenced by servers via its numeric ID at creation time.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config

## Quick Start

Create a file `placement-group.yaml`:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-group
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudPlacementGroup.ha-group
spec: {}
```

Deploy:

```shell
openmcf apply -f placement-group.yaml
```

This creates a spread placement group named `ha-group`. Servers assigned to this group are guaranteed to run on separate physical hosts.

## Configuration Reference

### Required Fields

This component has no required spec fields. An empty `spec` block (or `spec: {}`) creates a valid spread placement group.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `enum` (`spread`) | `spread` | Placement group strategy. Hetzner Cloud currently supports only `spread`, which distributes servers across different physical hosts. |

## Examples

### Minimal Spread Group

The simplest deployment: an empty spec defaults to the `spread` strategy.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-group
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudPlacementGroup.ha-group
spec: {}
```

### Production HA Group with Org and Environment

A placement group for database replicas scoped to a specific organization and environment. The metadata drives label generation for resource tracking.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudPlacementGroup.ha-db-group
    role: database
spec:
  type: spread
```

### Server Composition via valueFrom

A placement group referenced by a HetznerCloudServer using `valueFrom`. The server receives the group's numeric ID from the placement group's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudPlacementGroup
metadata:
  name: ha-db-group
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudPlacementGroup.ha-db-group
spec:
  type: spread
```

The server references this placement group:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: db-01
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudServer.db-01
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

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `placement_group_id` | `string` | Hetzner Cloud numeric ID of the created placement group. Referenced by HetznerCloudServer via `placementGroupId`. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — References placement group IDs for anti-affinity server placement
- [HetznerCloudSshKey](/docs/catalog/hetznercloud/hetznercloudsshkey) — Commonly deployed alongside placement groups as a foundation for server provisioning
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetznercloudfirewall) — Security boundaries applied to servers in placement groups
