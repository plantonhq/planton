# Hetzner Cloud Snapshot

Creates a point-in-time disk image from a Hetzner Cloud server, stored as a Hetzner Cloud Image. The snapshot captures the full disk state of the source server and can be used as a boot source when creating new servers — enabling golden image workflows, pre-upgrade rollback points, and server cloning.

## What Gets Created

- **Server Snapshot** — an `hcloud_snapshot` resource that captures the source server's disk as a Hetzner Cloud Image (type `snapshot`). The snapshot persists independently of the source server, with labels computed from resource metadata.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **An existing server** to snapshot — either a pre-existing server referenced by its numeric ID, or a `HetznerCloudServer` component referenced via `valueFrom`

## Quick Start

Create a file `snapshot.yaml`:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: my-snapshot
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudSnapshot.my-snapshot
spec:
  serverId:
    value: "12345678"
```

Deploy:

```shell
openmcf apply -f snapshot.yaml
```

This creates a snapshot of server `12345678`. The snapshot's image ID is available in the stack outputs as `snapshot_id`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `serverId` | `StringValueOrRef` | Server to create the snapshot from. Accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource via `valueFrom`. Changing this value forces replacement of the snapshot (the existing snapshot is destroyed and a new one is created from the new server). | required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | empty | Human-readable description of the snapshot. Useful for identifying the purpose (e.g., "pre-upgrade baseline", "golden image v2.1"). Can be updated after creation without replacing the snapshot. |

## Examples

### Snapshot with Description

A snapshot with a description for identification. The description appears in the Hetzner Cloud console and API listings.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: pre-upgrade-baseline
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: maintenance
    pulumi.openmcf.org/stack.name: production.HetznerCloudSnapshot.pre-upgrade-baseline
spec:
  serverId:
    value: "12345678"
  description: "pre-upgrade baseline before v3.2 rollout"
```

### Snapshot Referencing a Server Component

Using `valueFrom` to reference a `HetznerCloudServer` component's output. The snapshot waits for the server to be created before attempting to capture its disk.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: app-server-snapshot
  org: acme-corp
  env: staging
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: app-platform
    pulumi.openmcf.org/stack.name: staging.HetznerCloudSnapshot.app-server-snapshot
spec:
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: app-server
      fieldPath: status.outputs.server_id
  description: "staging app server baseline"
```

### Golden Image for Fleet Deployment

Snapshot a configured template server, then use the snapshot's image ID to create identical worker servers. This is faster and more consistent than running configuration management on every new server.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: golden-image-v1
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: fleet-management
    pulumi.openmcf.org/stack.name: production.HetznerCloudSnapshot.golden-image-v1
spec:
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: golden-template
      fieldPath: status.outputs.server_id
  description: "golden image v1 - ubuntu 24.04 with nginx and app v3.2"
```

After deployment, the `snapshot_id` output provides the image ID to pass as the `image` field when creating new `HetznerCloudServer` instances.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `snapshot_id` | `string` | The Hetzner Cloud image ID of the created snapshot. Usable as the `image` parameter when creating new servers from this snapshot. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — The source server to snapshot (referenced via `serverId`) and the consumer of snapshot images (via the `image` field)
