---
title: "Server Snapshot"
description: "This preset captures a point-in-time disk image of a Hetzner Cloud server. The resulting snapshot is stored as a Hetzner Cloud Image (type \"snapshot\") and can be used to create new servers from the..."
type: "preset"
rank: "01"
presetSlug: "01-server-snapshot"
componentSlug: "hetzner-cloud-snapshot"
componentTitle: "Hetzner Cloud Snapshot"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Server Snapshot

This preset captures a point-in-time disk image of a Hetzner Cloud server. The resulting snapshot is stored as a Hetzner Cloud Image (type "snapshot") and can be used to create new servers from the captured state. Snapshots persist independently of the source server -- deleting the server does not remove its snapshots.

This is the only preset for HetznerCloudSnapshot because the component has no configuration variance. Every snapshot manifest points at a server and optionally includes a description.

## When to Use

- Creating a backup before a risky operation (OS upgrade, major application update, configuration change)
- Capturing a fully configured server as a reusable golden image for launching identical servers
- Preserving server state before decommissioning for audit or compliance purposes
- Taking a known-good baseline that can be restored if a deployment goes wrong

## Key Configuration Choices

- **Server reference** (`serverId`) -- the numeric Hetzner Cloud server ID to snapshot; changing this value forces replacement of the snapshot resource (the old snapshot is destroyed and a new one is created from the new server)
- **Description included** (`description`) -- a human-readable label that makes snapshots identifiable in the Console and API listings; especially important when managing multiple snapshots across servers

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<server-id>` | Numeric ID of the Hetzner Cloud server to snapshot | The `status.outputs.server_id` of your HetznerCloudServer resource, or the Server details page in the Hetzner Cloud Console |
| `<snapshot-description>` | Human-readable purpose of the snapshot (e.g., "pre-upgrade baseline 2026-02-19", "golden image v2.1") | Your own naming convention |
