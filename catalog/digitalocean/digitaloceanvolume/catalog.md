# DigitalOcean Volume

Deploys a DigitalOcean block storage volume with configurable size, region, and optional filesystem pre-formatting. Volumes provide persistent, network-attached storage that can be attached to Droplets independently of Droplet lifecycle. Size is a one-way ratchet -- volumes only expand -- and everything else is create-only: changing the filesystem, label, description, snapshot source, or region replaces the volume and the data on it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Block Storage Volume** -- a `digitalocean_volume` resource in the specified region with the given size, optional filesystem formatting (with an optional filesystem label), optional description, and tags
- **DigitalOcean Tags** -- tags from the spec, merged with the standard Planton labels, applied directly to the volume resource for organizational tracking and cost allocation

The volume is created in a detached state. Attach it from the Droplet side: the DigitalOcean Droplet kind's `volumeIds` list consumes this volume's `volume_id` output.

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A target region** -- the volume must be created in the same region as any Droplet that will attach to it. Volumes cannot cross regions. Check available regions via the DigitalOcean API or control panel.
- **Sufficient block storage quota** in the target region for the requested volume size.
- **A volume snapshot ID** if creating the volume from an existing snapshot (optional). The new volume inherits the snapshot's region and minimum size -- `sizeGib` must be at least the snapshot's size.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Volume**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **General-Purpose ext4 Volume** preset in the [Presets](#presets) tab for a ready-to-use storage volume.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVolume
metadata:
  name: app-data
  org: acme-corp
  env: prod
spec:
  volumeName: app-data
  region: nyc1
  sizeGib: 50
```

```shell
planton apply -f do-volume.yaml
```

This creates a 50 GiB unformatted block storage volume in DigitalOcean's NYC1 region, detached and ready for a Droplet's `volumeIds` list to claim it. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a DigitalOcean storage volume. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Volume size** -- The `sizeGib` field sets the volume capacity in GiB (DigitalOcean caps volumes at 16 TiB). Volumes can only be EXPANDED after creation -- a shrink fails at plan time -- and expansion is online, but the filesystem inside does not grow itself: after expanding, resize it from the Droplet (`resize2fs` for ext4, `xfs_growfs` for xfs). Budget sizes with headroom; "grow later" is easy, "grow now under pressure" is an incident.

**Region** -- The `region` field accepts a DigitalOcean region slug (e.g., `nyc1`, `sfo3`, `fra1`, `lon1`). The volume must be in the same region as any Droplet you plan to attach it to, and changing the region later replaces the volume.

**Filesystem type** -- The `filesystemType` field pre-formats the volume with `ext4` or `xfs` exactly once, at creation; leave it unset and the volume arrives raw, with `mkfs` from the Droplet on you. Pick one path per volume and stay on it: declaring `ext4` later on a hand-formatted volume plans a replacement that destroys the data. Use `ext4` for general-purpose storage and `xfs` for database workloads with heavy sequential writes.

**Filesystem label** -- The `initialFilesystemLabel` field labels the filesystem when formatting (e.g. `pgdata`), so the Droplet can mount by label (`LABEL=pgdata`) instead of by device path, which shifts across reboots and attach order. If you format at creation, set the label too -- retrofitting one later means re-formatting or hand-running `e2label`/`xfs_admin` on the Droplet.

**Description** -- The `description` field adds a human-readable description stored on the DigitalOcean resource. It is create-only at the current provider pin: editing it later REPLACES the volume, so write it right the first time -- and read plans on volume changes carefully, because a replaced data volume is data loss unless you snapshot first.

**Snapshot source** -- The `snapshotId` field seeds a NEW volume from a point-in-time capture -- there is no in-place restore. The volume inherits the snapshot's data, and `sizeGib` must be at least the snapshot's size. Useful for disaster recovery or cloning environments from production data.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `volume_id` | UUID of the created DigitalOcean volume | The Droplet kind's `volumeIds` attachment list |
| `urn` | The volume's uniform resource name (`do:volume:<uuid>`) | DigitalOcean project membership |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General-purpose ext4 volume** -- 50 GiB volume pre-formatted with ext4 for application data, logs, and file storage. ext4 is the most widely compatible Linux filesystem and a safe default for most workloads. The volume is immediately mountable after attaching to a Droplet. Start from the **General-Purpose ext4 Volume** preset.

**Database XFS volume** -- 100 GiB volume pre-formatted with XFS and mounted by label, optimized for database workloads such as PostgreSQL and MySQL data directories. XFS provides superior write performance for sequential and random I/O patterns. The larger starting size accommodates typical database growth. Start from the **Database XFS Volume** preset.

## Works With

- [**DigitalOcean Droplet**](/infra-catalog/digital-ocean-droplet) -- attaches this volume via its `volumeIds` list consuming the `volume_id` output