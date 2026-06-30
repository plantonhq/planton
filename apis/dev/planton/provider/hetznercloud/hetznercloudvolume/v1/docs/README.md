# HetznerCloud Volume — Research Documentation

## Introduction

A Hetzner Cloud Volume is a network-attached block storage device that can be attached to exactly one server at a time. Volumes persist independently of any server — detaching or deleting the server does not affect the volume's data. This separation of compute and storage lifecycles is the primary reason volumes exist: databases, application state, uploaded files, and any data that must survive server replacement belongs on a volume rather than the server's local disk.

The `HetznerCloudVolume` component provisions a block storage volume and an optional server attachment. Unlike the server's local disk (which is destroyed on server replacement), a volume can be detached from one server and reattached to another in the same location. This makes volumes the standard mechanism for persistent state in Hetzner Cloud architectures — the data layer that decouples "what stores the data" from "what processes the data."

Planton bundles the volume and its optional attachment into a single component. When `serverId` is set, an `hcloud_volume_attachment` resource is created alongside the volume. When omitted, the volume is created unattached and available for later attachment. This two-resource design matches the Hetzner Cloud API's separation between the volume (a storage device) and its attachment (a relationship to a server).

## Historical Context

### Block Storage in Cloud Computing

Network-attached block storage is one of the earliest cloud computing primitives, introduced by Amazon EBS alongside EC2 in 2008. The core abstraction has remained stable across every cloud provider: a virtual disk of a specified size, backed by network-attached storage hardware, presented to a server as a standard block device. The server's OS sees it as a local disk and can partition, format, and mount it like any physical drive.

What makes block storage interesting in cloud environments is the lifecycle separation:

- **Server lifecycle**: Create, resize, replace, delete. Servers are meant to be ephemeral — replaceable units of compute.
- **Volume lifecycle**: Create, attach, detach, reattach, resize, snapshot. Volumes are meant to be durable — persistent units of storage.

This separation enables the "cattle not pets" pattern for servers while keeping state durable. A database server can be replaced (new OS, new server type, new location within the same datacenter) by detaching the volume, destroying the server, creating a new one, and reattaching the volume.

### HetznerCloud Volumes

Hetzner Cloud launched volumes as a block storage product backed by SSD storage in their datacenters. Key characteristics:

- **Size range**: 10 GB to 10,240 GB (10 TB)
- **Performance**: SSD-backed, with throughput scaled to volume size
- **Pricing**: Flat rate per GB/month (currently €0.0440/GB/month), no per-IOPS charges
- **Location-bound**: A volume exists in a specific location (e.g., `fsn1`) and can only be attached to servers in the same location
- **Single attachment**: A volume can be attached to exactly one server at a time
- **Resize**: Size can be increased online (no detach needed), but can never be decreased — the Hetzner Cloud API rejects size reductions
- **Format options**: Volumes can be pre-formatted with ext4 or xfs at creation time, or created raw for manual formatting

The pricing model is simple compared to hyperscalers. There are no IOPS tiers, no throughput classes, no burst credits. You pay for provisioned size, and the performance is what the underlying SSD hardware delivers.

### Volume Attachment as a Separate Concept

The Hetzner Cloud API treats volume creation and volume attachment as separate operations. The Terraform provider mirrors this with two resources:

1. **`hcloud_volume`** — The storage device itself. Has a name, size, location, optional format, labels, and delete protection.
2. **`hcloud_volume_attachment`** — The relationship between a volume and a server. Has a volume ID, server ID, and automount flag.

This separation exists because volumes outlive servers. If the attachment were embedded in the volume resource, deleting a server would either leave the volume in an inconsistent state or force the volume to be recreated. The separate attachment resource means destroying the attachment (detaching) does not affect the volume, and destroying the volume does not require updating the server.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Volumes** in the left sidebar
3. Click **Create Volume**
4. Select a location (must match the server you plan to attach to)
5. Enter a size in GB (10–10,240)
6. Choose a filesystem format:
   - **ext4** — general-purpose Linux filesystem
   - **xfs** — high-performance filesystem for large files
   - **No formatting** — raw block device, format manually
7. Optionally select a server to attach to immediately
8. If attaching, choose whether to automount
9. Enter a name and optional labels
10. Click **Create & Buy now**

After creation, to attach or detach: navigate to the volume's detail page, click "Attach" or "Detach", and select the target server.

**Pros:**
- Immediate visual feedback on volume size, cost, and location
- Attach/detach is a single click
- Format selection is straightforward

**Cons:**
- No version control for volume configurations
- Cannot express the volume-to-server relationship declaratively
- Attach/detach is a separate manual step from creation
- No enforcement of naming or labeling standards
- Cannot reproduce volume configurations across environments

**Verdict:** Suitable for quick experiments and one-off storage. Not viable for environments where volume configurations must be reproducible or where volumes are part of a larger infrastructure deployment.

### Level 1: CLI (`hcloud`)

```bash
# Create an unattached volume with ext4 filesystem
hcloud volume create \
  --name db-data \
  --size 100 \
  --location fsn1 \
  --format ext4 \
  --label env=production \
  --label role=database

# Create and immediately attach to a server with automount
hcloud volume create \
  --name app-storage \
  --size 50 \
  --server my-server \
  --format ext4 \
  --automount

# Attach an existing volume to a server
hcloud volume attach --server my-server db-data

# Detach a volume
hcloud volume detach db-data

# Resize a volume (online, no detach needed)
hcloud volume resize --size 200 db-data

# Enable delete protection
hcloud volume enable-protection db-data delete

# Inspect
hcloud volume describe db-data
hcloud volume list

# Delete (must disable protection first, must detach first)
hcloud volume disable-protection db-data delete
hcloud volume detach db-data
hcloud volume delete db-data
```

**Key CLI behaviors:**
- `--server` and `--location` are mutually exclusive on `create` — providing `--server` infers the location from the server
- `--automount` requires `--server`
- Resize is online: the volume is resized while attached, but the filesystem inside must be resized separately by the server's OS (the API resizes the block device, not the filesystem)
- Delete fails if the volume is attached — you must detach first

**Pros:**
- Scriptable, can be embedded in CI/CD pipelines
- Single command for creation with all options
- Online resize without detaching

**Cons:**
- No state tracking or drift detection
- Multi-step workflow for create + attach + protect
- Filesystem resize after volume resize is a separate manual step on the server
- No declarative relationship between volume and server

**Verdict:** Good for ad-hoc operations and scripted provisioning. The multi-step workflow (create, attach, protect) is error-prone without state management.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides `hcloud_volume` and `hcloud_volume_attachment`:

```hcl
terraform {
  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.60"
    }
  }
}

resource "hcloud_server" "app" {
  name        = "app-01"
  server_type = "cx22"
  image       = "ubuntu-24.04"
  location    = "fsn1"
}

resource "hcloud_volume" "data" {
  name              = "app-data"
  size              = 100
  location          = "fsn1"
  format            = "ext4"
  delete_protection = true
  labels = {
    environment = "production"
    role        = "app-data"
  }
}

resource "hcloud_volume_attachment" "data" {
  volume_id = hcloud_volume.data.id
  server_id = hcloud_server.app.id
  automount = true
}

output "volume_id" {
  value = hcloud_volume.data.id
}

output "linux_device" {
  value = hcloud_volume.data.linux_device
}
```

**Key provider behaviors:**
- `hcloud_volume`: `location` is `ForceNew` — changing the location destroys and recreates the volume (data loss). `size` can be increased in-place but not decreased. `format` is create-time-only — the provider does not read it back after creation. `name` and `labels` can be updated in-place.
- `hcloud_volume_attachment`: Both `volume_id` and `server_id` are `ForceNew` — moving a volume to a different server requires destroying the attachment and creating a new one. `automount` is also `ForceNew`.
- `location` and `server_id` on `hcloud_volume` are mutually exclusive — you can create a volume by specifying a location (unattached) or by specifying a server (attached in the same operation). However, using `server_id` on the volume resource itself conflicts with `hcloud_volume_attachment`. For managed attachments, always use `location` on the volume and a separate attachment resource.
- `linux_device` is a computed output that gives the device path (e.g., `/dev/disk/by-id/scsi-0HC_Volume_12345678`). This path is stable and suitable for `/etc/fstab` entries.

**Pros:**
- Full state tracking and drift detection
- Automatic dependency resolution (attachment depends on volume and server)
- Plan/apply previews changes before execution
- `linux_device` output enables scripted mount configuration

**Cons:**
- Must coordinate `location` and `server_id` mutually exclusive fields
- Two resources for a volume attached to a server (volume + attachment)
- Volume resize triggers a provider-level resize action, but filesystem resize must still be handled outside Terraform (via remote-exec or manual SSH)
- `format` is a "write-only" argument — Terraform does not track it in state after creation, so changing it in the config has no effect

**Verdict:** Production-grade for Terraform teams. The two-resource pattern (volume + attachment) is well-understood. The main pain point is the `format` field's write-only behavior.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `Volume` and `VolumeAttachment`:

```go
package main

import (
    "strconv"

    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        vol, err := hcloud.NewVolume(ctx, "data", &hcloud.VolumeArgs{
            Name:             pulumi.String("app-data"),
            Size:             pulumi.Int(100),
            Location:         pulumi.StringPtr("fsn1"),
            Format:           pulumi.StringPtr("ext4"),
            DeleteProtection: pulumi.Bool(true),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        // Volume ID is IDOutput (string), but attachment expects IntInput
        volIdInt := vol.ID().ApplyT(func(id pulumi.ID) (int, error) {
            return strconv.Atoi(string(id))
        }).(pulumi.IntOutput)

        _, err = hcloud.NewVolumeAttachment(ctx, "data-attachment",
            &hcloud.VolumeAttachmentArgs{
                VolumeId:  volIdInt,
                ServerId:  pulumi.Int(12345678),
                Automount: pulumi.BoolPtr(true),
            },
        )
        if err != nil {
            return err
        }

        ctx.Export("volumeId", vol.ID())
        ctx.Export("linuxDevice", vol.LinuxDevice)
        return nil
    })
}
```

**The ID type mismatch:** Same pattern as every other Hetzner Cloud Pulumi resource. The volume's `ID()` returns a string (`IDOutput`), but `VolumeAttachmentArgs.VolumeId` expects an `IntInput`. The `ApplyT(strconv.Atoi)` conversion is required. The `ServerId` field also expects an integer — if the server ID comes from another resource's output, another conversion is needed.

**The `format` write-only behavior:** Same as Terraform. The `Format` field is passed during creation but is not tracked in state afterward. Changing it in code after the initial `pulumi up` has no effect on the existing volume.

**Pros:**
- Full programming language with type safety
- Automatic dependency tracking via output references
- Conditional logic is native Go (not HCL `count`/`for_each`)

**Cons:**
- ID type conversion boilerplate (`ApplyT(strconv.Atoi)`)
- `Format` write-only behavior inherited from the Terraform bridge
- More verbose than HCL for a simple volume

**Verdict:** Good for teams already using Pulumi. The ID conversion is the main friction point that Planton eliminates.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | Planton |
|--------|---------|-----|-----------|--------|---------|
| State tracking | No | No | Yes | Yes | Yes |
| Drift detection | No | No | Yes | Yes | Yes |
| Server attachment | Separate click | Separate command | Separate resource | Separate resource | Single `serverId` field |
| ID type handling | N/A | By name | Integer attributes | `ApplyT(strconv.Atoi)` | Automatic |
| Format handling | Dropdown | `--format` flag | Write-only arg | Write-only arg | Enum field (create-time) |
| Delete protection | Separate step | Separate command | Boolean field | Boolean field | Boolean field |
| Automount | Checkbox | `--automount` flag | `automount` on attachment | `Automount` on attachment | `automount` field on spec |
| Location validation | Visual (dropdown) | Error on mismatch | Error on mismatch | Error on mismatch | Documented constraint |
| Resize | Click + manual fs resize | CLI + manual fs resize | `size` change + manual fs resize | `Size` change + manual fs resize | `size` change + manual fs resize |

Planton's key differentiators for the volume resource:

1. **Single manifest, optional attachment**: One YAML declares the volume and its server relationship. The IaC module conditionally creates the attachment resource.
2. **StringValueOrRef for server reference**: The `serverId` field accepts a literal ID or a `valueFrom` reference to a `HetznerCloudServer` output. No integer conversion.
3. **Format as a proto enum**: The format field uses a proto enum (`ext4`, `xfs`, `format_unspecified`) rather than an arbitrary string. Invalid values are caught at validation time, not at cloud API call time.
4. **Attachment lifecycle managed automatically**: When `serverId` is added or removed from the spec, the IaC module creates or destroys the attachment resource without user intervention.

## The Planton Approach

### Manifest Format

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
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
  deleteProtection: true
```

### What Planton Automates

1. **Naming**: The volume name in Hetzner Cloud is derived from `metadata.name`.
2. **Labeling**: Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels (standard labels take precedence).
3. **ID type conversion**: The volume's Pulumi ID (string) is converted to integer for the attachment resource's `VolumeId` field. The server ID from `StringValueOrRef` is converted to integer for the attachment's `ServerId` field. Users pass strings; the module handles the conversions.
4. **Conditional attachment**: When `serverId` is set and non-empty, the module creates an `hcloud_volume_attachment`. When omitted, only the volume is created. No separate resource declaration by the user.
5. **Format enum validation**: The proto enum restricts format values to `ext4`, `xfs`, or unspecified (raw). Invalid formats are caught during manifest validation, before any cloud API call.
6. **Provider configuration**: The Hetzner Cloud API token is resolved from provider config or `HCLOUD_TOKEN`, not hardcoded in the manifest.
7. **Dual IaC**: The same manifest drives both Pulumi and Terraform backends.

### The 80/20 Principle

The Hetzner Cloud volume API surface is small — the Planton component exposes nearly all of it. The few exclusions are deliberate.

**Included — Required fields:**

| Field | Rationale |
|-------|-----------|
| `size` | The most fundamental attribute: how much storage. Must be explicit. Validated to the Hetzner Cloud range (10–10,240 GB). |
| `location` | Determines the physical datacenter. Must match any server the volume will be attached to. Changing it forces volume replacement (data loss). |

**Included — Optional fields:**

| Field | Rationale |
|-------|-----------|
| `format` | Pre-formatting saves the manual step of SSH-ing into the server to format the volume. The enum restricts values to the two supported filesystems. |
| `serverId` | Enables declarative attachment. Accepts literal IDs or `valueFrom` references. The conditional attachment resource is a core Planton convenience. |
| `automount` | When attaching at creation, automount avoids a manual mount step inside the server. Only meaningful when `serverId` is set. |
| `deleteProtection` | Prevents accidental deletion of a volume that may contain critical data. Essential for production volumes. |

**Handled by the platform (hardcoded or derived):**
- `name` — Derived from `metadata.name`.
- `labels` — Computed from metadata per CG01 pattern.

**Deliberately excluded:**

| Field | Rationale |
|-------|-----------|
| `server_id` on the volume resource itself | The Terraform `hcloud_volume` resource accepts a `server_id` for create-time attachment. Planton uses a separate `hcloud_volume_attachment` resource instead, because it provides a cleaner lifecycle: the volume's location is always explicit, and the attachment can be created or destroyed independently. The provider's mutually exclusive `location`/`server_id` constraint makes mixing the two approaches error-prone. |

### API Design Decisions

**Format as a proto enum (not a string):**

The Hetzner Cloud API accepts `"ext4"` or `"xfs"` as format strings. A freeform string field would pass any value to the API, producing a cryptic error from the cloud provider. The proto enum restricts the field to known values, with `format_unspecified` (the zero value) meaning "no formatting" — a raw block device. This makes the default behavior (raw) explicit rather than relying on an empty string.

**The format field is create-time-only:** The Hetzner Cloud provider does not read the volume's filesystem format back from the API after creation. Changing the format in the spec after the initial apply has no effect on the existing volume. This is documented in the spec.proto comments and the component README. It is not a bug — the cloud API simply does not expose the volume's current filesystem format as a readable attribute.

**ServerId as StringValueOrRef (not embedded attachment):**

The `serverId` field uses `StringValueOrRef` with `default_kind = HetznerCloudServer` and `default_kind_field_path = "status.outputs.server_id"`. This enables the `valueFrom` shorthand where users reference a server by name and the field path is auto-resolved. When set, the IaC module creates a separate `hcloud_volume_attachment` resource. When unset, the volume is created unattached.

The alternative — using the `hcloud_volume` resource's built-in `server_id` field — was rejected because it makes `location` and `server_id` mutually exclusive. Since Planton always requires `location` (it determines the physical datacenter and is critical for location affinity validation), the separate attachment resource is the only consistent approach.

**Automount's limited scope:**

Automount is a create-time-only setting that tells Hetzner Cloud to mount the volume after the initial attachment. It is not tracked in state after creation. Despite this limitation, it is exposed because it eliminates a common post-deployment manual step. The field is meaningfully guarded: the Pulumi module only passes `automount` to the attachment resource when `serverId` is set, so the field is inert when the volume is unattached.

**Two stack outputs:**

The component exports `volume_id` (for programmatic reference by other systems) and `linux_device` (the stable device path for mounting). The `linux_device` output is particularly useful: it provides the `/dev/disk/by-id/scsi-0HC_Volume_*` path that can be used directly in `/etc/fstab` or mount scripts. This path is stable across server reboots and reattachments.

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Created When | Description |
|------------|----------|-------|--------------|-------------|
| Pulumi | `hcloud.Volume` | 1 | Always | Block storage volume with name, size, location, optional format, labels, delete protection |
| Pulumi | `hcloud.VolumeAttachment` | 0 or 1 | When `serverId` is set | Attaches the volume to the specified server with optional automount |
| Terraform | `hcloud_volume` | 1 | Always | Same as Pulumi |
| Terraform | `hcloud_volume_attachment` | 0 or 1 | When `server_id` is set | Conditional via `count` |

### ID Type Conversions (Pulumi Module)

The Pulumi hcloud SDK requires integer inputs for volume and server IDs. The module performs two conversions:

| Conversion | Input Source | Target | Method |
|------------|-------------|--------|--------|
| Volume ID | `createdVolume.ID()` | `VolumeAttachmentArgs.VolumeId` (IntOutput) | `ApplyT(strconv.Atoi)` at deployment time |
| Server ID | `spec.ServerId.GetValue()` | `VolumeAttachmentArgs.ServerId` (Int) | `strconv.Atoi` at creation time |

The volume ID uses `ApplyT` because it depends on the volume's actual ID, which is only available after the volume is created. The server ID uses plain `strconv.Atoi` because the value is known before resource creation (resolved from `StringValueOrRef` during stack input loading).

### Format Enum Mapping

The proto enum maps to provider values:

| Proto Enum Value | Provider Value | Behavior |
|-----------------|----------------|----------|
| `format_unspecified` (0) | not set (`nil`) | Raw block device, no filesystem |
| `ext4` (1) | `"ext4"` | ext4 filesystem created at volume creation |
| `xfs` (2) | `"xfs"` | XFS filesystem created at volume creation |

In the Pulumi module, the mapping is a simple check: if the format is not `format_unspecified`, the enum's `.String()` method provides the value directly. The Terraform module checks for `null` and `"format_unspecified"` strings and converts to `null` to omit the argument.

### Conditional Attachment Logic

Both IaC modules create the attachment resource conditionally:

**Pulumi** (`volume.go`):
```go
if spec.ServerId != nil && spec.ServerId.GetValue() != "" {
    // Convert volume ID and server ID to integers
    // Create hcloud.NewVolumeAttachment with automount
}
```

**Terraform** (`main.tf`):
```hcl
resource "hcloud_volume_attachment" "this" {
  count = var.spec.server_id != null ? 1 : 0
  ...
}
```

The guard conditions are slightly different due to language idioms, but the behavior is identical: when the server ID is absent or empty, the attachment resource is not created.

### Label Management

Both modules apply standard labels using the CG01 pattern:

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"db-data"` |
| `kind` | Constant | `"HetznerCloudVolume"` |
| `org` | `metadata.org` | `"acme-corp"` |
| `env` | `metadata.env` | `"production"` |
| `id` | `metadata.id` | `"hcvol-abc123"` |

User-specified `metadata.labels` are merged in. Standard labels take precedence on key conflicts. Labels are applied only to the volume resource — the attachment resource does not support labels in the Hetzner Cloud API.

## Production Best Practices

### Sizing Strategy

- **Start with the minimum viable size.** Volume size can be increased online (no downtime, no detach) but can never be decreased. Start conservatively and scale up as data grows.
- **Monitor usage, not just capacity.** Use `df -h` or monitoring agents to track actual utilization. Resize before reaching 85% capacity to avoid emergency expansions.
- **Account for filesystem overhead.** A 100 GB volume formatted with ext4 provides ~93 GB of usable space. XFS overhead is slightly lower.
- **The 10 GB minimum is meaningful.** For small configuration files or application state, 10 GB is sufficient. Don't over-provision — Hetzner charges per GB/month regardless of utilization.

### Filesystem Format Selection

| Filesystem | Best For | Key Characteristics |
|------------|----------|---------------------|
| **ext4** | General-purpose workloads, databases, web applications | Most widely supported, mature, good default for most use cases. Handles small and large files well. |
| **xfs** | Large files, high-throughput sequential I/O, media storage | Optimized for large file operations. Better performance for workloads like video transcoding, log aggregation, and backup storage. |
| **Raw (no format)** | Custom filesystems, LVM, DRBD, ZFS, raw block access | For advanced use cases where the application manages the block device directly. |

**Recommendation:** Use ext4 unless you have a specific reason to choose xfs. If you need ZFS, btrfs, or LVM, create a raw volume and format it yourself on the server.

### Location Affinity

- **Volume and server must be in the same location.** This is enforced by the Hetzner Cloud API. A volume in `fsn1` cannot be attached to a server in `hel1`.
- **Plan location early.** Moving a volume to a different location requires creating a new volume in the target location, copying data (via snapshot, rsync, or application-level export), and deleting the old volume.
- **Use the same location for all related resources.** If your server is in `fsn1`, create volumes, Primary IPs, and Floating IPs in `fsn1`. Location mismatches between resources are the most common configuration error in Hetzner Cloud.

### Attachment and Mounting Patterns

**Automount vs. manual mount:**
- Use `automount: true` for initial deployments where the server and volume are created together. Hetzner Cloud mounts the volume automatically after the first attachment.
- For subsequent attachments (e.g., moving a volume to a new server), automount uses the last known mount point. If the mount point conflicts with an existing path on the new server, the mount may fail. In this case, mount manually.

**Using the `linux_device` output:**
The volume's `linux_device` output provides a stable path like `/dev/disk/by-id/scsi-0HC_Volume_12345678`. This path is:
- Stable across reboots (unlike `/dev/sdb` which can change)
- Suitable for `/etc/fstab` entries
- Available immediately after attachment

Example fstab entry using the linux_device path:
```
/dev/disk/by-id/scsi-0HC_Volume_12345678 /mnt/data ext4 defaults 0 2
```

**Detach before server deletion:** If a volume is attached and the server is deleted, the Hetzner Cloud API automatically detaches the volume. However, for clean lifecycle management, explicitly detach volumes (remove `serverId` from the spec) before deleting the server.

### Backup Strategy

Hetzner Cloud Volumes do not have built-in backup or snapshot capabilities at the volume level. To protect volume data:

- **Application-level backups**: pg_dump, mysqldump, rsync to another volume or object storage. This is the most reliable approach.
- **Server snapshots**: A server snapshot includes all attached volumes. However, snapshots capture the server and volumes at a point in time — for database consistency, stop writes before snapshotting.
- **Cross-volume replication**: For critical data, maintain a replica on a second volume attached to a different server (possibly in a different location for disaster recovery).

### Delete Protection

- **Enable `deleteProtection` for any volume containing production data.** This prevents accidental deletion via the API, CLI, console, or IaC destroy operations.
- **Protection must be disabled before deletion.** This is a two-step process: update the spec to set `deleteProtection: false`, apply, then delete the volume. This deliberate friction prevents hasty deletions.
- **Combine with server `deleteProtection`.** If the server has attached volumes, protect both the server and the volumes to prevent cascading data loss from an accidental server deletion.

### Resize Workflow

1. **Increase the volume size** in the Planton spec and apply. The IaC module calls the Hetzner Cloud resize API, which increases the block device size online.
2. **Resize the filesystem** on the server. The block device is larger, but the filesystem still occupies the original size:
   - For ext4: `resize2fs /dev/disk/by-id/scsi-0HC_Volume_*`
   - For xfs: `xfs_growfs /mnt/data` (where `/mnt/data` is the mount point)
3. **Verify** with `df -h`.

The filesystem resize step cannot be automated by the IaC module — it must be performed inside the running server. Cloud-init, configuration management (Ansible), or a manual SSH session can handle this.

**The size can never be decreased.** Plan sizing carefully. If you over-provision a volume and want to shrink it, the only option is:
1. Create a new smaller volume
2. Copy data from the old volume to the new one
3. Swap the attachment
4. Delete the old volume

## References

- [Hetzner Cloud Volumes Documentation](https://docs.hetzner.cloud/#volumes)
- [Hetzner Cloud API — Volumes](https://docs.hetzner.cloud/#volumes-get-all-volumes)
- [Hetzner Cloud API — Volume Actions](https://docs.hetzner.cloud/#volume-actions)
- [Terraform hcloud_volume Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/volume)
- [Terraform hcloud_volume_attachment Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/volume_attachment)
- [Pulumi hcloud.Volume Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/volume/)
- [Pulumi hcloud.VolumeAttachment Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/volumeattachment/)
