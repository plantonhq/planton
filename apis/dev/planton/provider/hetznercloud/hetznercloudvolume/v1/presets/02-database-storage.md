# Database Storage Volume

This preset creates a production-grade Hetzner Cloud block storage volume optimized for database workloads. It uses the XFS filesystem for high-throughput sequential writes, enables delete protection to guard against accidental data loss, and attaches the volume to a server with automount. The IaC module provisions an `hcloud_volume` and an `hcloud_volume_attachment` resource.

XFS is the standard filesystem choice for PostgreSQL (WAL segments), MySQL/MariaDB (InnoDB redo logs), MongoDB (WiredTiger journal), and other database engines that perform large sequential writes. It delivers higher sustained throughput than ext4 under these access patterns and handles large files more efficiently.

## When to Use

- Database servers (PostgreSQL, MySQL, MariaDB, MongoDB, ClickHouse) that store data on a dedicated volume separate from the OS disk
- Any stateful workload where accidental volume deletion would cause unrecoverable data loss
- Production environments where the volume holds business-critical data and must be protected from both accidental deletion and undersizing

## Key Configuration Choices

- **XFS filesystem** (`format: xfs`) -- optimized for large sequential writes and sustained throughput; the standard choice for database engines that write WAL/redo logs and perform sequential I/O
- **100 GB** (`size: 100`) -- a more realistic starting size for database storage; adjust based on your data growth projections (size can be increased online but never decreased)
- **Delete protection enabled** (`deleteProtection: true`) -- prevents accidental volume destruction via the API or Console; must be explicitly disabled before the volume can be removed
- **Automount enabled** (`automount: true`) -- Hetzner Cloud mounts the volume automatically after attachment; the database engine's data directory should point to this mount path
- **Falkenstein location** (`location: fsn1`) -- change to match the database server's location; volume and server must be co-located

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<server-id>` | Numeric ID of the Hetzner Cloud server running the database engine | The `status.outputs.server_id` of your HetznerCloudServer resource, or the Servers page in the Hetzner Cloud Console |

## Related Presets

- **01-attached-ext4** -- general-purpose ext4 volume without delete protection, for non-critical workloads
- **03-unattached-reserve** -- pre-provisioned volume not attached to any server, for infrastructure preparation or migration staging
