# Hetzner Cloud Snapshot — Research Documentation

## Introduction

A Hetzner Cloud snapshot is a point-in-time disk image captured from a server. The snapshot records the complete state of the server's local disk — the operating system, installed packages, application code, configuration files, and data. Once created, the snapshot is stored as a Hetzner Cloud Image (type `snapshot`) and can be used to create new servers that boot into the exact state captured at snapshot time.

Snapshots solve three problems that every infrastructure team encounters:

1. **Rollback** — Before a risky upgrade or configuration change, snapshot the server. If the change fails, create a new server from the snapshot to restore the previous state.
2. **Golden images** — Configure a server once (install packages, harden the OS, deploy the application), snapshot it, and stamp out identical servers from that snapshot. This is faster and more reliable than running configuration management on every new server.
3. **Server cloning** — Need a second server with the same setup? Snapshot the first and create a new server from it. Cheaper and faster than rebuilding from scratch.

The `HetznerCloudSnapshot` component wraps the `hcloud_snapshot` Terraform resource (and the equivalent `hcloud.Snapshot` Pulumi resource) into a single OpenMCF manifest. The manifest declares which server to snapshot and an optional description. The IaC modules handle labeling, ID type conversions, and provider configuration.

## Snapshots in the Hetzner Cloud Ecosystem

### Images, Snapshots, and Backups

Hetzner Cloud uses the **Image** API as the storage layer for three distinct concepts:

| Type | Created By | Lifecycle | Use Case |
|------|-----------|-----------|----------|
| **System image** | Hetzner (pre-built) | Permanent, managed by Hetzner | OS installation (`ubuntu-24.04`, `debian-12`, `rocky-9`) |
| **Snapshot** | User (on-demand) | Persists until explicitly deleted | Golden images, rollback points, server cloning |
| **Backup** | Hetzner (automated, if enabled) | Rotated automatically (7 backups max) | Disaster recovery for servers with backup pricing enabled |

All three are stored as Images in the API. The `type` field distinguishes them: `system`, `snapshot`, or `backup`. When the Terraform provider creates an `hcloud_snapshot`, it calls the Server API's "Create Image" action with `type=snapshot`. The resulting image appears in `hcloud image list` alongside system images and backups.

This distinction matters because:

- **Snapshots are user-managed.** You create them, you delete them, you pay for them until deleted. There is no automatic rotation.
- **Backups are provider-managed.** Hetzner creates them on a schedule, rotates old ones automatically, and charges a percentage of the server price. Backups are tied to the server — deleting the server deletes its backups.
- **Snapshots persist independently.** Deleting the source server does not delete snapshots taken from it. The snapshot exists as a standalone image with its own lifecycle.

### Pricing

Snapshots are billed based on the disk size of the source server, not the amount of data actually written to disk. The pricing (as of early 2026) is approximately EUR 0.0119/GB/month. A snapshot of a CX22 server (40 GB local disk) costs ~EUR 0.48/month regardless of disk utilization.

This pricing model has implications for snapshot management:

- Snapshots of large servers (e.g., CCX63 with 360 GB disk) cost ~EUR 4.28/month each. Accumulating snapshots without a retention strategy adds up.
- There is no incremental or differential snapshot pricing. Every snapshot stores the full disk image.
- Deleting a snapshot stops billing immediately. There is no minimum retention period.

### Creation Time

Snapshot creation is an asynchronous server action. The Hetzner Cloud API accepts the request, returns an action ID, and the provider polls until the action completes. The time depends on the server's disk size and current load on the snapshot infrastructure:

- Small servers (20-40 GB disk): typically 1-5 minutes
- Medium servers (80-160 GB disk): typically 5-15 minutes
- Large servers (240-360 GB disk): can take 30-60+ minutes

The Terraform provider sets a 90-minute timeout for snapshot creation. This generous timeout accounts for worst-case scenarios (large disk, busy snapshot infrastructure). In practice, most snapshots complete well within this window.

The server remains running during snapshot creation. The snapshot captures a crash-consistent disk state — equivalent to pulling the power cord and reading the disk. For most workloads (stateless web servers, application servers), this is sufficient. For databases and other stateful workloads, see the consistency section in Production Best Practices.

## Deployment Methods Landscape

### Level 0: Manual (Hetzner Cloud Console)

1. Log in to [console.hetzner.cloud](https://console.hetzner.cloud)
2. Navigate to **Servers** in the left sidebar
3. Click on the target server
4. Switch to the **Snapshots** tab
5. Click **Create Snapshot**
6. Enter a description (optional)
7. Click **Create Snapshot**
8. Wait for the snapshot to appear in the list (status changes from "creating" to the image size)

To use the snapshot: when creating a new server, select "Snapshots" in the image picker and choose the snapshot.

**Pros:**
- Visual confirmation of snapshot progress and size
- Easy to browse existing snapshots per server
- Image picker shows snapshots directly when creating servers

**Cons:**
- No version control or audit trail for when/why snapshots were taken
- No way to enforce naming or labeling standards
- Manual process for multi-server environments
- No programmatic relationship between snapshot and new servers created from it
- Cannot reproduce the snapshot workflow across environments

**Verdict:** Suitable for one-off snapshots during manual maintenance. Not viable when snapshots are part of a repeatable deployment workflow.

### Level 1: CLI (`hcloud`)

```bash
# Create a snapshot from a server
hcloud server create-image \
  --type snapshot \
  --description "pre-upgrade baseline" \
  12345678

# The command returns the image ID immediately, but the snapshot
# creation continues asynchronously. Check status:
hcloud image describe 98765432

# List all snapshots
hcloud image list --type snapshot

# Create a new server from the snapshot
hcloud server create \
  --name worker-01 \
  --type cx22 \
  --image 98765432 \
  --location fsn1

# Update snapshot description
hcloud image update --description "golden image v2.1" 98765432

# Delete a snapshot
hcloud image delete 98765432
```

**Key CLI behaviors:**
- `hcloud server create-image` takes a server ID (positional argument), not a server name. You must look up the server ID first or use `$(hcloud server describe my-server -o format='{{.ID}}')`.
- The `--type` flag defaults to `snapshot`. Backups cannot be created via CLI — they are managed by Hetzner's automated backup system.
- The command returns immediately with the image ID, but the snapshot creation runs asynchronously. Use `hcloud image describe` to check completion.
- Labels are passed as `--label key=value` flags (repeatable).

**Pros:**
- Scriptable, can be embedded in CI/CD pipelines or cron jobs
- Single command to create a snapshot
- Immediate return of image ID for downstream use

**Cons:**
- No state tracking or drift detection
- Multi-step workflow to create snapshot + use it for new servers
- Must manually track which snapshots exist and which are still needed
- No declarative relationship between server and its snapshots

**Verdict:** Good for scripted snapshot workflows (e.g., nightly golden image builds). The lack of state management means cleanup and retention must be handled separately.

### Level 2: IaC — Terraform

The `hcloud` Terraform provider (`hetznercloud/hcloud ~> 1.60`) provides `hcloud_snapshot`:

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

resource "hcloud_snapshot" "app_baseline" {
  server_id   = hcloud_server.app.id
  description = "app server baseline"
  labels = {
    environment = "production"
    purpose     = "golden-image"
  }
}

# Use the snapshot to create a new server
resource "hcloud_server" "worker" {
  name        = "worker-01"
  server_type = "cx22"
  image       = hcloud_snapshot.app_baseline.id
  location    = "fsn1"
}

output "snapshot_id" {
  value = hcloud_snapshot.app_baseline.id
}
```

**Key provider behaviors:**
- `server_id` is `ForceNew` — changing the source server destroys the existing snapshot and creates a new one from the new server. This is the most important behavior to understand: the snapshot is tied to the server that created it.
- `description` and `labels` can be updated in-place without replacing the snapshot.
- The provider creates the snapshot by calling the Server API's "Create Image" action and waits for the action to complete (up to 90 minutes).
- The snapshot's `id` is the Hetzner Cloud image ID. It can be used directly in `hcloud_server.image`.
- Snapshot creation requires the server to exist. Terraform's dependency graph handles this automatically when `server_id = hcloud_server.app.id`.

**Pros:**
- Full state tracking — Terraform knows which snapshots exist and manages their lifecycle
- Automatic dependency resolution between server and snapshot
- The snapshot ID can be referenced by other resources (servers, modules)
- Plan/apply previews changes before execution

**Cons:**
- `ForceNew` on `server_id` means any change to the server reference destroys the snapshot. If the server is replaced (e.g., server type change), the snapshot is also destroyed and recreated.
- No built-in retention mechanism — old snapshots must be managed via `terraform state rm` or by removing the resource from config.
- Snapshot creation blocks the apply until complete. For large servers, this can add 30+ minutes to the apply cycle.

**Verdict:** Production-grade for teams managing snapshots as part of their Terraform state. The `ForceNew` behavior on `server_id` is the main surprise — it's correct but catches users who expect snapshots to be immutable once created.

### Level 3: IaC — Pulumi

The `pulumi-hcloud` SDK (bridged from the Terraform provider) exposes `Snapshot`:

```go
package main

import (
    "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        server, err := hcloud.NewServer(ctx, "app", &hcloud.ServerArgs{
            Name:       pulumi.String("app-01"),
            ServerType: pulumi.String("cx22"),
            Image:      pulumi.String("ubuntu-24.04"),
            Location:   pulumi.StringPtr("fsn1"),
        })
        if err != nil {
            return err
        }

        snap, err := hcloud.NewSnapshot(ctx, "baseline", &hcloud.SnapshotArgs{
            ServerId:    pulumi.Int(server.ID().ApplyT(strconv.Atoi).(pulumi.IntOutput)),
            Description: pulumi.StringPtr("app server baseline"),
            Labels: pulumi.StringMap{
                "environment": pulumi.String("production"),
            },
        })
        if err != nil {
            return err
        }

        ctx.Export("snapshotId", snap.ID())
        return nil
    })
}
```

**The ID type mismatch:** The `SnapshotArgs.ServerId` field expects an `IntInput`, but server IDs are typically available as string outputs (`IDOutput`). When referencing another Pulumi resource's ID, the `ApplyT(strconv.Atoi)` conversion is required. When using a literal server ID, `pulumi.Int(12345678)` works directly.

**Bridged behaviors from Terraform:**
- `ServerId` is `ForceNew` — same replacement behavior as Terraform
- `Description` and `Labels` are updatable in-place
- The 90-minute creation timeout is inherited

**Pros:**
- Full programming language with type safety and conditionals
- Automatic dependency tracking via output references
- Better error messages when type conversions fail (compile-time vs. runtime)

**Cons:**
- ID type conversion boilerplate (`ApplyT(strconv.Atoi)`) for server references
- More verbose than HCL for a simple snapshot
- The bridged provider inherits all Terraform behaviors, including the ForceNew semantics

**Verdict:** Good for teams already using Pulumi. The ID conversion is the main friction point that OpenMCF eliminates.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | OpenMCF |
|--------|---------|-----|-----------|--------|---------|
| State tracking | No | No | Yes | Yes | Yes |
| Drift detection | No | No | Yes | Yes | Yes |
| Server reference | Click on server | Positional server ID | `server_id` (int) | `ServerId` (IntInput) | `serverId` (StringValueOrRef) |
| ID type handling | N/A | N/A | Native int | `ApplyT(strconv.Atoi)` | Automatic |
| Description update | Console edit | `hcloud image update` | In-place update | In-place update | In-place update |
| Label management | Manual | `--label` flags | Map attribute | `StringMap` | Derived from metadata |
| Snapshot retention | Manual delete | Manual/scripted | State management | State management | State management |
| Golden image workflow | Multi-step manual | Multi-step scripted | Resource references | Output references | Manifest references |

OpenMCF's key differentiators for the snapshot resource:

1. **StringValueOrRef for server reference**: The `serverId` field accepts a literal string ID or a `valueFrom` reference. No integer conversion needed. The `default_kind` annotation means the user only needs to specify `name` and the field path is auto-resolved.
2. **Automatic labeling**: Standard labels are derived from metadata. No manual label maps in the manifest.
3. **Dual IaC backends**: The same manifest drives both Pulumi and Terraform. The IaC modules handle provider-specific details (integer conversion for Pulumi, `tonumber()` for Terraform).

## The OpenMCF Approach

### Manifest Format

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudSnapshot
metadata:
  name: app-baseline
  org: acme-corp
  env: production
spec:
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: app-server
      fieldPath: status.outputs.server_id
  description: "app server baseline before v3.2 upgrade"
```

### What OpenMCF Automates

1. **Labeling**: Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are computed from metadata and merged with user-specified labels. Standard labels take precedence on key conflicts.
2. **ID type conversion**: The `serverId` string from `StringValueOrRef` is converted to the integer type required by both the Pulumi SDK (`strconv.Atoi`) and the Terraform provider (`tonumber()`). Users pass strings; the modules handle the conversion.
3. **Provider configuration**: The Hetzner Cloud API token is resolved from provider config or `HCLOUD_TOKEN`, not hardcoded in the manifest.
4. **Dual IaC**: The same manifest drives both Pulumi and Terraform backends. The user chooses the backend via provisioner labels; the manifest itself is backend-agnostic.

### The 80/20 Principle

The Hetzner Cloud snapshot resource surface is small. The OpenMCF component exposes the essential fields and handles the rest automatically.

**Included — Required fields:**

| Field | Rationale |
|-------|-----------|
| `serverId` | The most fundamental attribute: which server to snapshot. Must be explicit. Uses `StringValueOrRef` to support both literal IDs and component references. |

**Included — Optional fields:**

| Field | Rationale |
|-------|-----------|
| `description` | Human-readable identification. Snapshots accumulate over time, and descriptions are the primary way to distinguish them in listings. Updatable post-creation. |

**Handled by the platform (hardcoded or derived):**

| Field | How It's Handled |
|-------|-----------------|
| `labels` | Derived from `metadata` per the CG01 label pattern. Standard labels take precedence over user-specified `metadata.labels`. |
| `type` | Always `snapshot`. The Hetzner Cloud API also supports `backup` as an image type, but backups are managed by Hetzner's automated backup system, not user-created. |

**Deliberately excluded:**

There are no fields in the provider's `hcloud_snapshot` resource that OpenMCF excludes. The provider exposes `server_id`, `description`, and `labels` — all three are covered. The `type` field is implicit (always `snapshot` for this resource type).

### API Design Decisions

**ServerId as StringValueOrRef (not an integer):**

The Hetzner Cloud API and both IaC providers use integer server IDs. OpenMCF uses `StringValueOrRef` for consistency with all other components that reference foreign resources. The string-to-integer conversion is handled in the IaC modules. This means the same `valueFrom` pattern used for volume-to-server references, floating-IP-to-server references, and other cross-component relationships works identically for snapshot-to-server references.

**Single output (snapshot_id only):**

The `hcloud_snapshot` resource has computed attributes beyond the ID (e.g., `created`, `os_flavor`, `os_version`, `disk_size`). These are not exposed as stack outputs because they are informational attributes available via `hcloud image describe` and not typically consumed by other IaC resources. The `snapshot_id` is the only output that other components need — it is the image ID used to boot new servers from the snapshot.

**No retention or rotation fields:**

The component creates a single snapshot. It does not manage retention policies, rotation schedules, or multi-snapshot workflows. This is deliberate: retention is an operational concern that varies by team and environment, not a property of the snapshot itself. Teams that need automated retention should manage it at the orchestration layer (CI/CD pipelines, cron jobs, or a future infra chart).

## Implementation Landscape

### Resources Created

| IaC Engine | Resource | Count | Description |
|------------|----------|-------|-------------|
| Pulumi | `hcloud.Snapshot` | 1 | Creates a server snapshot stored as a Hetzner Cloud Image |
| Terraform | `hcloud_snapshot` | 1 | Same as Pulumi |

No conditional resources. This is the simplest IaC module in the Hetzner Cloud component catalog.

### ID Type Conversion (Pulumi Module)

The Pulumi `hcloud.SnapshotArgs.ServerId` field requires an `IntInput`. The module performs a single conversion:

| Conversion | Input Source | Target | Method |
|------------|-------------|--------|--------|
| Server ID | `spec.ServerId.GetValue()` | `SnapshotArgs.ServerId` (Int) | `strconv.Atoi` at creation time |

This is a creation-time conversion (plain `strconv.Atoi`), not a deployment-time conversion (`ApplyT`). The server ID is resolved from `StringValueOrRef` during stack input loading, so the value is known before resource creation begins. This is the same pattern used for other components that reference foreign IDs via literal values.

### Label Management

Both modules apply standard labels using the CG01 pattern:

| Label Key | Source | Example |
|-----------|--------|---------|
| `resource` | Constant | `"true"` |
| `name` | `metadata.name` | `"app-baseline"` |
| `kind` | Constant | `"HetznerCloudSnapshot"` |
| `org` | `metadata.org` (if set) | `"acme-corp"` |
| `env` | `metadata.env` (if set) | `"production"` |
| `id` | `metadata.id` (if set) | `"hcsnp-abc123"` |

User-specified `metadata.labels` are merged in. Standard labels take precedence on key conflicts.

## Production Best Practices

### Snapshot Before Upgrades

The most common snapshot use case: capture a known-good state before a risky change.

**Workflow:**
1. Deploy the snapshot manifest (referencing the server to upgrade)
2. Apply — the snapshot captures the current server state
3. Perform the upgrade on the server
4. If the upgrade fails: create a new server from the snapshot to restore the previous state
5. If the upgrade succeeds: keep the snapshot as a rollback point or delete it to stop billing

**Timing:** Snapshot the server immediately before the upgrade, not hours or days in advance. The snapshot captures the disk state at the moment of creation — any changes between snapshot and upgrade are not captured.

### Golden Image Pipeline

For teams managing server fleets, the golden image pattern eliminates configuration drift:

1. **Build** — Start with a base OS image, install packages, harden the OS, deploy application code. This can be done manually, with cloud-init, or with configuration management (Ansible, Chef, etc.).
2. **Snapshot** — Capture the configured server as a snapshot.
3. **Deploy** — Create new servers from the snapshot. Each server boots into the exact state captured in step 2.
4. **Rotate** — When a new version is needed, repeat steps 1-3. Delete old snapshots when no servers are using them.

This is faster than running configuration management on every new server (boot from snapshot takes seconds; running Ansible takes minutes). It also eliminates the "works on my machine" problem — every server is byte-for-byte identical.

### Consistency Considerations

Snapshots capture a **crash-consistent** disk state. The server remains running during snapshot creation, and the snapshot captures whatever state the disk is in at that moment. This is equivalent to pulling the power cord and reading the disk.

**Stateless workloads** (web servers, application servers, workers): Crash consistency is sufficient. The application can recover from an unclean shutdown, and the snapshot captures a usable state.

**Databases** (PostgreSQL, MySQL, MongoDB): Crash consistency means the database's write-ahead log (WAL) or journal will replay uncommitted transactions on recovery. This works for most databases but may result in:
- Loss of transactions that were in-flight at snapshot time
- A recovery period on first boot from the snapshot (WAL replay)

For databases, the safest approach is:
1. Stop the database service (or put it in backup mode)
2. Create the snapshot
3. Restart the database service

If stopping the database is not acceptable, rely on the database's crash recovery mechanism. PostgreSQL, MySQL/InnoDB, and MongoDB all handle crash-consistent snapshots correctly — they recover to the last committed transaction.

**Filesystems with volatile caches** (XFS, ext4 with writeback mode): The disk state may not reflect the latest writes that are still in the filesystem cache. For critical data, run `sync` before snapshotting or mount filesystems with `barrier=1` (the default for most modern filesystems).

### The ForceNew Trap

Changing the `serverId` field forces replacement of the snapshot. This is because the Hetzner Cloud provider marks `server_id` as `ForceNew` — there is no API to "re-point" a snapshot to a different server.

**What this means in practice:**
- If the source server is replaced (e.g., server type change that triggers ForceNew), the snapshot is also destroyed and recreated from the new server.
- The old snapshot's image ID becomes invalid. Any servers that were created from that snapshot continue to run, but the snapshot itself no longer exists for creating new servers.
- If you need to preserve a snapshot permanently (regardless of server lifecycle changes), consider managing it outside the IaC state or using a separate manifest that references the server by literal ID rather than `valueFrom`.

### Snapshot Cost Management

Snapshots are billed per GB/month based on the source server's disk size. Without active management, costs accumulate:

- **Tag snapshots with purpose and date** — Use the `description` field to record why the snapshot was taken and when.
- **Delete snapshots after use** — Pre-upgrade snapshots should be deleted once the upgrade is confirmed successful.
- **Avoid snapshotting large servers unnecessarily** — A snapshot of a 360 GB disk server costs ~EUR 4.28/month. Consider whether the data on the server warrants a full-disk snapshot or whether application-level backups are more appropriate.
- **Monitor snapshot count** — Use `hcloud image list --type snapshot` to audit existing snapshots periodically.

### When Not to Use Snapshots

Snapshots are not the right tool for every backup scenario:

| Scenario | Better Alternative |
|----------|-------------------|
| Continuous database backup | `pg_dump`/`mysqldump` to object storage |
| File-level backup | `rsync` or `borgbackup` to a backup volume or external storage |
| Volume data protection | Hetzner Cloud does not support volume-level snapshots — use application-level backups |
| Disaster recovery across locations | Snapshots are location-bound; use application-level replication for cross-location DR |
| Frequent incremental backups | Snapshots are full-disk and non-incremental; use block-level backup tools |

Snapshots are best suited for point-in-time captures of full server state — the "before and after" of a change, or the "golden image" for fleet deployment.

## References

- [Hetzner Cloud Server Actions — Create Image](https://docs.hetzner.cloud/#server-actions-create-image)
- [Hetzner Cloud Images Documentation](https://docs.hetzner.cloud/#images)
- [Hetzner Cloud Images API — Get All Images](https://docs.hetzner.cloud/#images-get-all-images)
- [Hetzner Cloud Pricing](https://docs.hetzner.cloud/#pricing)
- [Terraform hcloud_snapshot Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/snapshot)
- [Pulumi hcloud.Snapshot Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/snapshot/)
