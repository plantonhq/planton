# DigitalOcean Volume

A DigitalOcean block storage volume described once in a Planton manifest: expandable network-attached storage for Droplets, optionally pre-formatted (ext4/xfs) with a filesystem label, created empty or from a snapshot, and tagged. Attachment is a property of the Droplet — its `volumeIds` list consumes this kind's `volume_id` output — never of the volume itself.

## What this component models

The spec maps onto DigitalOcean's `digitalocean_volume` in full:

| Spec field | What it controls |
|---|---|
| `volumeName` | The volume's name (lowercase, hyphens; create-only) |
| `region` | Where the volume lives — must match the Droplet's region to attach |
| `sizeGib` | Size in GiB; can only be EXPANDED after creation (a shrink fails at plan time) |
| `description` | Free-form description; create-only at the current provider pin (a change REPLACES the volume) |
| `filesystemType` | Optional one-time formatting at creation (`ext4`/`xfs`); leave unset to format yourself |
| `initialFilesystemLabel` | The label applied when formatting (e.g. `pgdata`); only meaningful with `filesystemType` |
| `snapshotId` | Create the volume from a volume snapshot (inherits its region and minimum size) |
| `tags` | User tags; both provisioners merge them with the standard Planton labels |

## Quick start

The smallest real volume:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanVolume
metadata:
  name: my-volume
spec:
  volumeName: scratch-data
  region: nyc3
  sizeGib: 50
```

A production database volume, formatted and labeled:

```yaml
spec:
  volumeName: postgres-data
  description: PostgreSQL data volume for production
  region: nyc3
  sizeGib: 500
  filesystemType: xfs
  initialFilesystemLabel: pgdata
  tags:
    - env:prod
    - service:postgres
```

Attach it from the Droplet:

```yaml
# on the DigitalOceanDroplet spec
volumeIds:
  - valueFrom:
      kind: DigitalOceanVolume
      name: my-volume
      fieldPath: status.outputs.volume_id
```

## Behavior worth knowing

- **Size only grows** — the provider rejects a shrink at plan time; DigitalOcean caps volumes at 16 TiB.
- **Formatting happens exactly once** — `filesystemType`, `initialFilesystemLabel`, and `snapshotId` act at creation and are never reported back by the API; after import they stay empty in state, and both modules ignore later edits to them, so an adopted volume's first apply is a no-op instead of a data-destroying replacement.
- **Description is create-only** — at the current provider pin, editing it replaces the volume. Write it right the first time.
- **One Droplet at a time** — DigitalOcean volumes attach to a single Droplet; regional, not zonal.
- **Volumes import by UUID** — the `volume_id` output is the resource identity.

## Outputs

| Output | Meaning |
|---|---|
| `volume_id` | The volume's UUID — what the Droplet kind's `volumeIds` consumes (`status.outputs.volume_id`) |
| `urn` | The uniform resource name (`do:volume:<uuid>`) |

## See also

- `GUIDE.md` — operational judgment calls (expand-only sizing, formatting, snapshots)
- `presets/` — general-purpose ext4 and database xfs starting points
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
