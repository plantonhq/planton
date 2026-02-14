---
title: "Server Group"
description: "Server Group deployment documentation"
icon: "package"
order: 100
componentName: "openstackservergroup"
---

# OpenStack Server Group

Deploys an OpenStack Compute server group, which defines a placement policy that controls whether compute instances are co-located on the same hypervisor or spread across different hypervisors. Server groups are referenced by instances via scheduler hints to enforce affinity or anti-affinity constraints.

## What Gets Created

When you deploy an OpenStackServerGroup resource, OpenMCF provisions:

- **Compute Server Group** — an `openstack_compute_servergroup_v2` resource with the specified placement policy (affinity, anti-affinity, soft-affinity, or soft-anti-affinity)

All fields on a server group are immutable. Changing any field recreates the server group, which orphans existing member instances from the old group.

## Prerequisites

- **OpenStack credentials** configured via environment variables or OpenMCF provider config
- **Nova Compute API 2.15+** if using `soft-affinity` or `soft-anti-affinity` policies
- **Sufficient hypervisor hosts** when using hard `anti-affinity` (one host required per instance in the group)

## Quick Start

Create a file `server-group.yaml`:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: my-server-group
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OpenstackServerGroup.my-server-group
spec:
  policy: anti-affinity
```

Deploy:

```shell
openmcf apply -f server-group.yaml
```

This creates a server group named `my-server-group` with an anti-affinity policy, ensuring that instances referencing this group are placed on different hypervisors.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `policy` | `string` | Placement policy for instances in this server group. Valid values: `affinity`, `anti-affinity`, `soft-affinity`, `soft-anti-affinity`. Changing this recreates the server group. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `region` | `string` | provider default | Overrides the region from the provider config for this server group. |

### Policy Reference

| Policy | Behavior | API Requirement |
|--------|----------|-----------------|
| `affinity` | All instances on the same hypervisor (hard constraint) | Any Nova version |
| `anti-affinity` | All instances on different hypervisors (hard constraint) | Any Nova version |
| `soft-affinity` | Prefer same hypervisor, no failure if unavailable | Nova API 2.15+ |
| `soft-anti-affinity` | Prefer different hypervisors, no failure if unavailable | Nova API 2.15+ |

## Examples

### Anti-Affinity for Database Replicas

Spread database replicas across different hypervisors for high availability. If a hypervisor fails, at most one replica is affected:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: db-anti-affinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenstackServerGroup.db-anti-affinity
    openmcf.org/stack.jobId: prod.OpenstackServerGroup.db-anti-affinity
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackservergroup/v1/iac/pulumi/module
spec:
  policy: anti-affinity
```

### Affinity for Tightly Coupled Services

Co-locate application and cache instances on the same hypervisor to minimize network latency between them:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: app-cache-affinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenstackServerGroup.app-cache-affinity
    openmcf.org/stack.jobId: prod.OpenstackServerGroup.app-cache-affinity
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackservergroup/v1/iac/pulumi/module
spec:
  policy: affinity
```

### Soft Anti-Affinity for Large Clusters

Use a soft policy when the cluster may not have enough distinct hypervisors to satisfy a hard anti-affinity constraint. Instances are spread on a best-effort basis without scheduling failures:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: worker-soft-spread
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OpenstackServerGroup.worker-soft-spread
    openmcf.org/stack.jobId: staging.OpenstackServerGroup.worker-soft-spread
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackservergroup/v1/iac/pulumi/module
spec:
  policy: soft-anti-affinity
```

### Server Group in a Specific Region

Override the provider region for a server group that must be created in a particular OpenStack region:

```yaml
apiVersion: openstack.openmcf.org/v1
kind: OpenStackServerGroup
metadata:
  name: regional-anti-affinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OpenstackServerGroup.regional-anti-affinity
    openmcf.org/stack.jobId: prod.OpenstackServerGroup.regional-anti-affinity
    openmcf.org/stack.module.source: github.com/plantonhq/openmcf//apis/org/openmcf/provider/openstack/openstackservergroup/v1/iac/pulumi/module
spec:
  policy: anti-affinity
  region: RegionTwo
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `server_group_id` | `string` | UUID of the created server group. Used as a foreign key by OpenStackInstance via scheduler hints. |
| `name` | `string` | Name of the server group, derived from `metadata.name`. |
| `members` | `string[]` | List of instance UUIDs that belong to this server group. Empty at creation; populated as instances are launched with scheduler hints referencing this group. |
| `region` | `string` | OpenStack region where the server group was created. |

## Related Components

- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — references the server group via scheduler hints to enforce the placement policy
- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — provides networking for instances placed by the server group
- [OpenStackSecurityGroup](/docs/catalog/openstack/openstacksecuritygroup) — controls traffic to instances within the server group
