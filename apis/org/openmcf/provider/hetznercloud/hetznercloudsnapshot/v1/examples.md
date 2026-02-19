# HetznerCloudSnapshot Examples

## Minimal Snapshot

The simplest configuration: snapshot a server by its literal ID. No description, no org/env metadata. The snapshot captures the full disk of server `12345678` at the moment of creation.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: quick-backup
spec:
  serverId:
    value: "12345678"
```

---

## Snapshot with Description

Adding a description makes the snapshot identifiable in the Hetzner Cloud console and API listings. Descriptions can be updated after creation without replacing the snapshot.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: pre-upgrade-baseline
  org: acme-corp
  env: production
spec:
  serverId:
    value: "12345678"
  description: "pre-upgrade baseline before v3.2 rollout"
```

---

## Snapshot Referencing a Server Component

Using `valueFrom` to reference a `HetznerCloudServer` component's output establishes a dependency edge — the snapshot waits for the server to be created before attempting to capture its disk.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: app-server-snapshot
  org: acme-corp
  env: staging
spec:
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: app-server
      fieldPath: status.outputs.server_id
  description: "staging app server baseline"
```

The companion server manifest (deployed separately or in the same infra chart):

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-server
  org: acme-corp
  env: staging
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

---

## Golden Image Workflow

The golden image pattern: configure a server once, snapshot it, then stamp out identical servers from the snapshot. This is the core use case for snapshots in fleet management.

**Step 1** — Create and configure the template server (install packages, harden the OS, deploy application code):

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: golden-template
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

**Step 2** — Snapshot the configured server:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: golden-image-v1
  org: acme-corp
  env: production
spec:
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: golden-template
      fieldPath: status.outputs.server_id
  description: "golden image v1 - ubuntu 24.04 with nginx and app v3.2"
```

**Step 3** — Create worker servers from the snapshot. Each server boots with the exact disk state captured in the snapshot:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: worker-01
  org: acme-corp
  env: production
spec:
  serverType: cx22
  image: "98765432"  # snapshot_id from golden-image-v1 outputs
  location: fsn1
```

The `image` field accepts any Hetzner Cloud image ID, including snapshot IDs. After the snapshot is created, its `snapshot_id` output (available at `status.outputs.snapshot_id`) provides the image ID to pass to new servers.
