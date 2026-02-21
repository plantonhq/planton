# HetznerCloudVolume Examples

## Minimal Unattached Volume

The simplest configuration: a 10 GB raw volume in Falkenstein with no filesystem and no server attachment. The volume is created unattached and available for later use.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: scratch-data
spec:
  size: 10
  location: fsn1
```

---

## Formatted Volume

A 50 GB volume pre-formatted with ext4. The filesystem is created at volume creation time — no need to SSH into the server to format the block device. The volume is still unattached; it can be attached to a server later.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: app-data
  org: acme-corp
  env: staging
spec:
  size: 50
  location: fsn1
  format: ext4
```

---

## Volume Attached to a Server

A 100 GB ext4 volume attached to a server via a `valueFrom` reference. The `serverId` references a `HetznerCloudServer` component's output, establishing a dependency edge in the deployment DAG — the volume attachment waits for the server to be created.

Automount is enabled, so Hetzner Cloud mounts the volume automatically after the initial attachment.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: db-data
  org: acme-corp
  env: production
spec:
  size: 100
  location: fsn1
  format: ext4
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: db-primary
      fieldPath: status.outputs.server_id
  automount: true
```

The companion server manifest (deployed in the same infra chart):

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: db-primary
  org: acme-corp
  env: production
spec:
  serverType: ccx13
  image: ubuntu-24.04
  location: fsn1
```

---

## Production Volume with Protection

A large XFS volume for high-throughput workloads, attached to a server with delete protection enabled. XFS is preferred here for its performance characteristics with large sequential writes (log aggregation, media storage, backups).

Delete protection prevents accidental deletion — the protection must be explicitly removed before the volume can be destroyed.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudVolume
metadata:
  name: media-storage
  org: acme-corp
  env: production
  labels:
    role: media
    tier: hot
spec:
  size: 500
  location: fsn1
  format: xfs
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: media-processor
      fieldPath: status.outputs.server_id
  automount: true
  deleteProtection: true
```

The `valueFrom` reference ensures:
1. The volume attachment waits for the server to exist
2. The correct numeric server ID is passed without manual lookup
3. Replacing the server (e.g., upgrading the server type) propagates the new ID to the volume attachment on the next apply
