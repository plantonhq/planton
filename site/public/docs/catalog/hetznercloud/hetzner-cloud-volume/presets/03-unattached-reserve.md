---
title: "Unattached Reserve Volume"
description: "This preset creates a Hetzner Cloud block storage volume that is formatted and ready to use but not attached to any server. It provisions a single `hcloud_volume` resource with delete protection..."
type: "preset"
rank: "03"
presetSlug: "03-unattached-reserve"
componentSlug: "hetzner-cloud-volume"
componentTitle: "Hetzner Cloud Volume"
provider: "hetznercloud"
icon: "package"
order: 3
---

# Unattached Reserve Volume

This preset creates a Hetzner Cloud block storage volume that is formatted and ready to use but not attached to any server. It provisions a single `hcloud_volume` resource with delete protection enabled and no `hcloud_volume_attachment`. The volume sits in a specific location waiting to be claimed by a server deployed later.

This pattern is common in infrastructure-as-code workflows where storage is provisioned ahead of the compute that will consume it -- for example, when a database volume must exist before the database server is deployed, or when preparing disaster recovery capacity in a target location.

## When to Use

- Pre-provisioning storage before the target server is deployed (the server's manifest can reference this volume later)
- Disaster recovery preparation where volumes are staged in a recovery location ahead of a failover event
- Migration workflows where data is restored to a volume before the replacement server is created
- Capacity reservation in a specific Hetzner Cloud location

## Key Configuration Choices

- **No server attachment** -- `serverId` is omitted, so no `hcloud_volume_attachment` is created; add `serverId` to the spec later to trigger attachment
- **ext4 pre-formatted** (`format: ext4`) -- the volume is formatted at creation time so it can be mounted immediately when eventually attached, with no manual formatting needed
- **Delete protection enabled** (`deleteProtection: true`) -- a reserved volume exists for a reason; protection prevents accidental cleanup by automation or human error before the volume is claimed
- **50 GB** (`size: 50`) -- a starting size; adjust based on the workload that will eventually consume the volume
- **Falkenstein location** (`location: fsn1`) -- must match the location where the target server will be deployed; volumes can only attach to servers in the same location

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired resource name.

## Related Presets

- **01-attached-ext4** -- general-purpose ext4 volume immediately attached to a server
- **02-database-storage** -- XFS-formatted volume with delete protection for database workloads
